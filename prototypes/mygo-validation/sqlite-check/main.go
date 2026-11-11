// 纯 Go SQLite（modernc.org/sqlite）在 WAL、批量提交与 FTS5 trigram 下的表现。一次性实验。
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var words = []string{"会话", "队列", "暂停", "恢复", "审批", "引用", "子任务", "工作区", "附件", "取消", "目标", "上下文", "错误", "重试", "模型", "工具", "终端", "补读", "索引", "迁移",
	"session", "queue", "cancel", "approval", "reference", "worker", "timeout", "retry", "stream", "commit"}

func text(r *rand.Rand, n int) string {
	var b strings.Builder
	for b.Len() < n {
		b.WriteString(words[r.Intn(len(words))])
		if r.Intn(4) == 0 {
			b.WriteString("，")
		}
	}
	return b.String()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	dir, _ := os.MkdirTemp("", "aibo-sqlite-")
	defer os.RemoveAll(dir)
	out := map[string]any{}
	db, err := sql.Open("sqlite", "file:"+filepath.Join(dir, "aibo.db")+"?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)")
	must(err)
	db.SetMaxOpenConns(4)
	var ver string
	must(db.QueryRow("select sqlite_version()").Scan(&ver))
	out["sqliteVersion"] = ver
	must2(db.Exec(`create table messages(id integer primary key, session int, seq int, body text)`))
	must2(db.Exec(`create index messages_session on messages(session, seq)`))

	// 1. 模拟 8 个会话并发流式输出：单写入 goroutine，每 100ms 或 64 KiB 合并提交
	r := rand.New(rand.NewSource(1))
	type chunk struct{ session, seq int; body string }
	ch := make(chan chunk, 1024)
	var commits, rows int
	var maxCommit time.Duration
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		batch := []chunk{}
		size := 0
		flush := func() {
			if len(batch) == 0 {
				return
			}
			s := time.Now()
			tx, _ := db.Begin()
			st, _ := tx.Prepare("insert into messages(session, seq, body) values(?,?,?)")
			for _, c := range batch {
				_, _ = st.Exec(c.session, c.seq, c.body)
			}
			st.Close()
			must(tx.Commit())
			if d := time.Since(s); d > maxCommit {
				maxCommit = d
			}
			commits++
			rows += len(batch)
			batch, size = batch[:0], 0
		}
		tick := time.NewTicker(100 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case c, ok := <-ch:
				if !ok {
					flush()
					return
				}
				batch = append(batch, c)
				size += len(c.body)
				if size >= 64<<10 {
					flush()
				}
			case <-tick.C:
				flush()
			}
		}
	}()
	t0 := time.Now()
	for i := 0; i < 200000; i++ {
		ch <- chunk{i % 8, i, text(r, 400)}
	}
	close(ch)
	wg.Wait()
	el := time.Since(t0)
	out["write"] = map[string]any{"rows": rows, "commits": commits, "seconds": el.Seconds(), "rowsPerSec": int(float64(rows) / el.Seconds()), "maxCommitMs": maxCommit.Milliseconds(), "bodyChars": 400}

	// 2. 读：长历史分段读取（某会话最后 50 条、向前翻页）
	s := time.Now()
	for p := 0; p < 100; p++ {
		rs, err := db.Query("select id, body from messages where session=3 and seq < ? order by seq desc limit 50", 200000-p*400)
		must(err)
		for rs.Next() {
		}
		rs.Close()
	}
	out["pageReadAvgMs"] = float64(time.Since(s).Microseconds()) / 100 / 1000

	// 3. 独立索引库：FTS5 trigram
	idx, err := sql.Open("sqlite", "file:"+filepath.Join(dir, "index.db")+"?_pragma=journal_mode(WAL)")
	must(err)
	_, err = idx.Exec(`create virtual table fts using fts5(body, content='', tokenize='trigram')`)
	out["fts5Trigram"] = errStr(err)
	if err == nil {
		s = time.Now()
		tx, _ := idx.Begin()
		st, _ := tx.Prepare("insert into fts(rowid, body) values(?,?)")
		rs, _ := db.Query("select id, body from messages")
		n := 0
		for rs.Next() {
			var id int
			var b string
			_ = rs.Scan(&id, &b)
			_, _ = st.Exec(id, b)
			n++
		}
		rs.Close()
		st.Close()
		must(tx.Commit())
		out["indexBuild"] = map[string]any{"rows": n, "seconds": time.Since(s).Seconds()}
		q := func(term string) map[string]any {
			s := time.Now()
			var c int
			err := idx.QueryRow("select count(*) from fts where fts match ?", `"`+term+`"`).Scan(&c)
			return map[string]any{"hits": c, "ms": float64(time.Since(s).Microseconds()) / 1000, "err": errStr(err)}
		}
		out["queries"] = map[string]any{"子任务补读": q("子任务补读"), "暂停恢复": q("暂停，恢复"), "approval": q("approval"), "两字·会话": q("会话")}
		var fi os.FileInfo
		fi, _ = os.Stat(filepath.Join(dir, "index.db"))
		mi, _ := os.Stat(filepath.Join(dir, "aibo.db"))
		out["sizesMB"] = map[string]any{"aibo.db": mi.Size() >> 20, "index.db": fi.Size() >> 20}
		// 少于 3 字的查询：回退为有上限的扫描
		s = time.Now()
		var c int
		must(db.QueryRow("select count(*) from (select 1 from messages where session=3 and body like '%会话%' limit 200)").Scan(&c))
		out["shortQueryScanMs"] = float64(time.Since(s).Microseconds()) / 1000
	}
	b, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(b))
}

func must2(_ sql.Result, err error) { must(err) }

func errStr(err error) string {
	if err == nil {
		return "ok"
	}
	return err.Error()
}
