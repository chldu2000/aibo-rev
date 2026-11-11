// 一次性实验（不是产品实现）：同一 ACP 连接中并发多个会话的可靠性。
// 用本机 codex 与 @agentclientprotocol/codex-acp。
//
//	(cd adapter && npm install --ignore-scripts)
//	go run . -out reports/run.json
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// ---- 最小 ACP（JSON-RPC 2.0，按行分隔）客户端 ----

type msg struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcErr         `json:"error,omitempty"`
}

type rpcErr struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *rpcErr) Error() string { return fmt.Sprintf("rpc %d: %s %s", e.Code, e.Message, string(e.Data)) }

type update struct {
	At   time.Time
	Kind string
	Text string
	Raw  json.RawMessage
}

type conn struct {
	cmd     *exec.Cmd
	in      io.WriteCloser
	mu      sync.Mutex
	nextID  atomic.Int64
	pending map[int64]chan msg
	wire    *os.File

	umu       sync.Mutex
	updates   map[string][]update // sessionId → 更新
	unknown   []string            // 没有 sessionId 或未知会话的通知
	perms     []string            // 收到的权限请求
	closed    chan struct{}
	closeErr  error
	appServer int
}

func start(dir string, logDir string) (*conn, error) {
	cmd := exec.Command(filepath.Join(dir, "adapter/node_modules/.bin/codex-acp"))
	home, _ := os.UserHomeDir()
	cmd.Env = append(os.Environ(),
		"CODEX_PATH="+filepath.Join(home, ".local/bin/codex"),
		"INITIAL_AGENT_MODE=read-only",
		"NO_BROWSER=1",
		`CODEX_CONFIG={"model_reasoning_effort":"low"}`,
		"APP_SERVER_LOGS="+logDir,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, _ := os.Create(filepath.Join(logDir, "adapter.stderr"))
	cmd.Stderr = stderr
	wire, _ := os.Create(filepath.Join(logDir, "wire.jsonl"))
	c := &conn{cmd: cmd, in: in, pending: map[int64]chan msg{}, wire: wire, updates: map[string][]update{}, closed: make(chan struct{})}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	go c.read(out)
	return c, nil
}

func (c *conn) logWire(dir string, b []byte) {
	c.mu.Lock()
	fmt.Fprintf(c.wire, "{\"t\":%q,\"dir\":%q,\"m\":%s}\n", time.Now().Format("15:04:05.000"), dir, b)
	c.mu.Unlock()
}

func (c *conn) send(m any) error {
	b, _ := json.Marshal(m)
	c.logWire("out", b)
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.in.Write(append(b, '\n'))
	return err
}

func (c *conn) read(r io.Reader) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 64<<20)
	for sc.Scan() {
		line := append([]byte(nil), sc.Bytes()...)
		c.logWire("in", line)
		var m msg
		if json.Unmarshal(line, &m) != nil {
			continue
		}
		switch {
		case m.Method != "" && m.ID != nil:
			go c.handleRequest(m)
		case m.Method != "":
			c.handleNotification(m)
		case m.ID != nil:
			var id int64
			_ = json.Unmarshal(*m.ID, &id)
			c.mu.Lock()
			ch := c.pending[id]
			delete(c.pending, id)
			c.mu.Unlock()
			if ch != nil {
				ch <- m
			}
		}
	}
	c.closeErr = sc.Err()
	if c.closeErr == nil {
		c.closeErr = io.EOF
	}
	close(c.closed)
}

func (c *conn) call(method string, params any, timeout time.Duration) (json.RawMessage, error) {
	id := c.nextID.Add(1)
	ch := make(chan msg, 1)
	c.mu.Lock()
	c.pending[id] = ch
	c.mu.Unlock()
	if err := c.send(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params}); err != nil {
		return nil, err
	}
	select {
	case m := <-ch:
		if m.Error != nil {
			return nil, m.Error
		}
		return m.Result, nil
	case <-c.closed:
		return nil, fmt.Errorf("连接已关闭：%w", c.closeErr)
	case <-time.After(timeout):
		return nil, fmt.Errorf("超时 %s", timeout)
	}
}

func (c *conn) notify(method string, params any) error {
	return c.send(map[string]any{"jsonrpc": "2.0", "method": method, "params": params})
}

func (c *conn) handleNotification(m msg) {
	if m.Method != "session/update" {
		return
	}
	var p struct {
		SessionID string `json:"sessionId"`
		Update    struct {
			SessionUpdate string `json:"sessionUpdate"`
			Content       struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"update"`
	}
	_ = json.Unmarshal(m.Params, &p)
	c.umu.Lock()
	defer c.umu.Unlock()
	if p.SessionID == "" {
		c.unknown = append(c.unknown, string(m.Params))
		return
	}
	c.updates[p.SessionID] = append(c.updates[p.SessionID], update{At: time.Now(), Kind: p.Update.SessionUpdate, Text: p.Update.Content.Text, Raw: m.Params})
}

// handleRequest：只读模式下拒绝所有权限请求；未声明的客户端能力返回 method not found。
func (c *conn) handleRequest(m msg) {
	reply := map[string]any{"jsonrpc": "2.0", "id": m.ID}
	switch m.Method {
	case "session/request_permission":
		var p struct {
			SessionID string `json:"sessionId"`
			Options   []struct {
				OptionID string `json:"optionId"`
				Kind     string `json:"kind"`
			} `json:"options"`
		}
		_ = json.Unmarshal(m.Params, &p)
		choice := ""
		for _, o := range p.Options {
			if strings.HasPrefix(o.Kind, "reject") {
				choice = o.OptionID
				break
			}
		}
		c.umu.Lock()
		c.perms = append(c.perms, p.SessionID)
		c.umu.Unlock()
		if choice == "" {
			reply["result"] = map[string]any{"outcome": map[string]any{"outcome": "cancelled"}}
		} else {
			reply["result"] = map[string]any{"outcome": map[string]any{"outcome": "selected", "optionId": choice}}
		}
	default:
		reply["error"] = map[string]any{"code": -32601, "message": "method not found: " + m.Method}
	}
	_ = c.send(reply)
}

func (c *conn) text(sid string, since int) (string, int) {
	c.umu.Lock()
	defer c.umu.Unlock()
	var b strings.Builder
	us := c.updates[sid]
	for _, u := range us[min(since, len(us)):] {
		if u.Kind == "agent_message_chunk" {
			b.WriteString(u.Text)
		}
	}
	return b.String(), len(us)
}

func (c *conn) kinds(sid string, since int) map[string]int {
	c.umu.Lock()
	defer c.umu.Unlock()
	k := map[string]int{}
	us := c.updates[sid]
	for _, u := range us[min(since, len(us)):] {
		k[u.Kind]++
	}
	return k
}

func (c *conn) mark(sid string) int {
	c.umu.Lock()
	defer c.umu.Unlock()
	return len(c.updates[sid])
}

// ---- 场景 ----

func (c *conn) newSession(cwd string) (string, error) {
	r, err := c.call("session/new", map[string]any{"cwd": cwd, "mcpServers": []any{}}, 60*time.Second)
	if err != nil {
		return "", err
	}
	var v struct {
		SessionID string `json:"sessionId"`
	}
	_ = json.Unmarshal(r, &v)
	return v.SessionID, nil
}

type turn struct {
	Session    string         `json:"session"`
	Expect     string         `json:"expect,omitempty"`
	StopReason string         `json:"stopReason"`
	Err        string         `json:"error,omitempty"`
	Ms         int64          `json:"ms"`
	Reply      string         `json:"reply"`
	OK         bool           `json:"ok"`
	Kinds      map[string]int `json:"updateKinds"`
}

func (c *conn) prompt(sid, text string, timeout time.Duration) turn {
	t0 := time.Now()
	since := c.mark(sid)
	r, err := c.call("session/prompt", map[string]any{"sessionId": sid, "prompt": []any{map[string]any{"type": "text", "text": text}}}, timeout)
	tr := turn{Session: sid, Ms: time.Since(t0).Milliseconds()}
	time.Sleep(200 * time.Millisecond) // 晚到的更新
	tr.Reply, _ = c.text(sid, since)
	tr.Kinds = c.kinds(sid, since)
	if err != nil {
		tr.Err = err.Error()
		return tr
	}
	var v struct {
		StopReason string `json:"stopReason"`
	}
	_ = json.Unmarshal(r, &v)
	tr.StopReason = v.StopReason
	return tr
}

func parallel(n int, fn func(i int) turn) []turn {
	out := make([]turn, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); out[i] = fn(i) }()
	}
	wg.Wait()
	return out
}

func descendants(pgid int) []string {
	out, _ := exec.Command("ps", "-axo", "pid=,pgid=,ppid=,command=").Output()
	var r []string
	for _, l := range strings.Split(string(out), "\n") {
		f := strings.Fields(l)
		if len(f) > 3 && f[1] == fmt.Sprint(pgid) {
			r = append(r, strings.Join(f[:4], " ")+" "+strings.Join(f[3:min(len(f), 6)], " "))
		}
	}
	return r
}

// appServers 列出本实验启动的 codex app-server（命令行含 CODEX_PATH 指向的本机 codex）。
func appServers() []string {
	out, _ := exec.Command("ps", "-axo", "pid=,pgid=,ppid=,command=").Output()
	var r []string
	for _, l := range strings.Split(string(out), "\n") {
		if strings.Contains(l, ".local/bin/codex app-server") {
			r = append(r, strings.TrimSpace(l))
		}
	}
	return r
}

func main() {
	out := flag.String("out", "reports/run.json", "报告")
	rounds := flag.Int("rounds", 2, "并发纯文本轮数")
	nSess := flag.Int("sessions", 4, "并发会话数")
	flag.Parse()
	wd, _ := os.Getwd()
	logDir, _ := filepath.Abs(filepath.Join("reports", time.Now().Format("20060102-150405")))
	_ = os.MkdirAll(logDir, 0o755)
	rep := map[string]any{"logDir": logDir}
	save := func() {
		b, _ := json.MarshalIndent(rep, "", "  ")
		_ = os.WriteFile(*out, b, 0o644)
	}
	defer save()
	fail := func(stage string, err error) {
		rep["fatal"] = stage + ": " + err.Error()
		save()
		fmt.Fprintln(os.Stderr, stage, err)
		os.Exit(1)
	}

	c, err := start(wd, logDir)
	if err != nil {
		fail("start", err)
	}
	rep["adapterPid"] = c.cmd.Process.Pid
	init, err := c.call("initialize", map[string]any{
		"protocolVersion":    1,
		"clientCapabilities": map[string]any{"fs": map[string]any{"readTextFile": false, "writeTextFile": false}, "terminal": false},
		"clientInfo":         map[string]any{"name": "aibo-rev-acp-concurrency-probe", "version": "0.0.1"},
	}, 60*time.Second)
	if err != nil {
		fail("initialize", err)
	}
	var iv map[string]any
	_ = json.Unmarshal(init, &iv)
	rep["initialize"] = map[string]any{"agentInfo": iv["agentInfo"], "protocolVersion": iv["protocolVersion"], "authMethods": iv["authMethods"]}

	// 工作区：每个会话一个目录，各放一个不同内容的 note.txt
	base, _ := os.MkdirTemp("", "aibo-acp-concurrency-")
	rep["workspace"] = base
	dirs := make([]string, *nSess)
	codes := make([]string, *nSess)
	for i := range dirs {
		dirs[i] = filepath.Join(base, fmt.Sprintf("ws%d", i))
		_ = os.MkdirAll(dirs[i], 0o755)
		codes[i] = fmt.Sprintf("NOTE-%d-%04d", i, time.Now().UnixNano()%10000)
		_ = os.WriteFile(filepath.Join(dirs[i], "note.txt"), []byte(codes[i]+"\n"), 0o644)
	}
	sids := make([]string, *nSess)
	t0 := time.Now()
	for i := range sids {
		if sids[i], err = c.newSession(dirs[i]); err != nil {
			fail("session/new", err)
		}
	}
	rep["sessionsCreatedMs"] = time.Since(t0).Milliseconds()
	rep["sessions"] = sids
	rep["descendantsAfterSessions"] = descendants(c.cmd.Process.Pid)
	save()

	// S1 并发纯文本回合：回复只含本会话的标记
	var s1 [][]turn
	for r := 0; r < *rounds; r++ {
		res := parallel(*nSess, func(i int) turn {
			tok := fmt.Sprintf("TOKEN-R%d-S%d", r, i)
			tr := c.prompt(sids[i], "不要使用任何工具。只回复这一行，原样输出：" + tok, 3*time.Minute)
			tr.Expect = tok
			tr.OK = tr.StopReason == "end_turn" && strings.Contains(tr.Reply, tok)
			for j := 0; j < *nSess; j++ {
				if j != i && strings.Contains(tr.Reply, fmt.Sprintf("TOKEN-R%d-S%d", r, j)) {
					tr.OK = false
					tr.Err += " 串入其他会话内容"
				}
			}
			return tr
		})
		s1 = append(s1, res)
		rep["S1_parallelText"] = s1
		save()
	}

	// S2 并发读取各自工作目录下的文件
	rep["S2_cwdIsolation"] = parallel(*nSess, func(i int) turn {
		tr := c.prompt(sids[i], "读取当前工作目录下 note.txt 的内容。只回复文件里的那一行，不要其他文字。", 4*time.Minute)
		tr.Expect = codes[i]
		tr.OK = tr.StopReason == "end_turn" && strings.Contains(tr.Reply, codes[i])
		for j := range codes {
			if j != i && strings.Contains(tr.Reply, codes[j]) {
				tr.OK = false
				tr.Err += " 读到其他会话目录"
			}
		}
		return tr
	})
	save()

	// S3 取消一个会话，其他会话继续
	const count = "不要使用任何工具。从 1 数到 120，每行一个数字，最后一行写 DONE。"
	var s3 []turn
	var cancelAt int64
	res := parallel(2, func(i int) turn {
		if i == 0 {
			go func() {
				t := time.Now()
				for time.Since(t) < 90*time.Second {
					if s, _ := c.text(sids[0], 0); strings.Count(s, "\n") >= 5 {
						break
					}
					time.Sleep(50 * time.Millisecond)
				}
				cancelAt = time.Since(t).Milliseconds()
				_ = c.notify("session/cancel", map[string]any{"sessionId": sids[0]})
			}()
		}
		return c.prompt(sids[i], count, 4*time.Minute)
	})
	res[0].Expect, res[0].OK = "cancelled", res[0].StopReason == "cancelled"
	res[1].Expect, res[1].OK = "DONE", res[1].StopReason == "end_turn" && strings.Contains(res[1].Reply, "DONE")
	s3 = append(s3, res...)
	again := c.prompt(sids[0], "不要使用任何工具。只回复：AGAIN-OK", 3*time.Minute)
	again.Expect, again.OK = "AGAIN-OK", again.StopReason == "end_turn" && strings.Contains(again.Reply, "AGAIN-OK")
	s3 = append(s3, again)
	rep["S3_cancelOne"] = map[string]any{"turns": s3, "cancelSentAfterMs": cancelAt}
	save()

	// S4 适配器进程退出：未完成的回合如何结束、子进程是否残留
	before := descendants(c.cmd.Process.Pid)
	serversBefore := appServers()
	done := make(chan []turn)
	go func() {
		done <- parallel(2, func(i int) turn { return c.prompt(sids[i], count, 2*time.Minute) })
	}()
	time.Sleep(4 * time.Second)
	_ = c.cmd.Process.Signal(syscall.SIGKILL)
	t1 := time.Now()
	s4 := <-done
	time.Sleep(2 * time.Second)
	after := descendants(c.cmd.Process.Pid)
	rep["S4_adapterKilled"] = map[string]any{"turns": s4, "settledAfterKillMs": time.Since(t1).Milliseconds() - 2000,
		"processGroupBefore": before, "processGroupAfter2s": after, "appServersBefore": serversBefore, "appServersAfter2s": appServers()}
	if len(after) > 0 {
		_ = syscall.Kill(-c.cmd.Process.Pid, syscall.SIGKILL)
	}
	_ = c.cmd.Wait()

	c.umu.Lock()
	rep["unroutedNotifications"] = len(c.unknown)
	rep["permissionRequests"] = c.perms
	c.umu.Unlock()
	_ = os.RemoveAll(base)
}
