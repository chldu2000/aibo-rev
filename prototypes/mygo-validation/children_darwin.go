package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func spawnChildren(out string) {
	type child struct {
		Kind string `json:"kind"`
		PID  int    `json:"pid"`
	}
	var kids []child
	// 1. 读 stdin 的子进程：父进程退出后 stdin 收到 EOF 即退出（ACP Agent 的典型行为）。
	cat := exec.Command("/bin/cat")
	cat.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if _, err := cat.StdinPipe(); err == nil && cat.Start() == nil {
		kids = append(kids, child{"stdin-eof", cat.Process.Pid})
	}
	// 2. 不读 stdin 的子进程。
	sl := exec.Command("/bin/sleep", "600")
	sl.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if sl.Start() == nil {
		kids = append(kids, child{"ignores-stdin", sl.Process.Pid})
	}
	// 3. 读 stdin 的子进程，再派生一个不读 stdin 的孙进程。
	sh := exec.Command("/bin/sh", "-c", "sleep 600 & echo $! ; cat")
	sh.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdout, _ := sh.StdoutPipe()
	if _, err := sh.StdinPipe(); err == nil && sh.Start() == nil {
		kids = append(kids, child{"stdin-eof-parent", sh.Process.Pid})
		buf := make([]byte, 32)
		n, _ := stdout.Read(buf)
		var gpid int
		fmt.Sscan(strings.TrimSpace(string(buf[:n])), &gpid)
		kids = append(kids, child{"grandchild-ignores-stdin", gpid})
	}
	b, _ := json.MarshalIndent(map[string]any{"app": os.Getpid(), "children": kids}, "", "  ")
	_ = os.WriteFile(out, b, 0o644)
	record("children", kids)
}

func runChildrenProbe(out string) bool { return false }
