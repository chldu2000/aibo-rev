package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var probeInputs []io.WriteCloser

// 子进程实验使用自身作为辅助程序，不依赖 shell 或用户安装的工具。
func init() {
	if len(os.Args) < 3 || os.Args[1] != "--probe-child" {
		return
	}
	switch os.Args[2] {
	case "stdin-eof":
		_, _ = io.Copy(io.Discard, os.Stdin)
	case "parent":
		cmd := exec.Command(os.Args[0], "--probe-child", "ignores-stdin")
		if err := cmd.Start(); err != nil {
			panic(err)
		}
		fmt.Println(cmd.Process.Pid)
		_, _ = io.Copy(io.Discard, os.Stdin)
	default:
		for {
			time.Sleep(time.Hour)
		}
	}
	os.Exit(0)
}

func runChildrenProbe(out string) bool {
	spawnChildren(out)
	// 正常退出与强制结束使用同一个实验；崩溃脚本在报告生成后立即结束宿主。
	time.Sleep(8 * time.Second)
	return true
}

func spawnChildren(out string) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		panic(err)
	}
	// 句柄不继承；宿主退出时系统关闭最后一个句柄并结束所有后代。
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err := windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits))); err != nil {
		panic(err)
	}
	if err := windows.AssignProcessToJobObject(job, windows.CurrentProcess()); err != nil {
		panic(err)
	}
	type child struct {
		Kind string `json:"kind"`
		PID  int    `json:"pid"`
	}
	var kids []child
	for _, kind := range []string{"stdin-eof", "ignores-stdin", "parent"} {
		cmd := exec.Command(os.Args[0], "--probe-child", kind)
		input, err := cmd.StdinPipe()
		if err != nil {
			panic(err)
		}
		probeInputs = append(probeInputs, input)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			panic(err)
		}
		if err := cmd.Start(); err != nil {
			panic(err)
		}
		kids = append(kids, child{kind, cmd.Process.Pid})
		if kind == "parent" {
			line, err := bufio.NewReader(stdout).ReadString('\n')
			if err != nil {
				panic(err)
			}
			pid, err := strconv.Atoi(line[:len(line)-1])
			if err != nil {
				panic(err)
			}
			kids = append(kids, child{"grandchild-ignores-stdin", pid})
		}
	}
	b, err := json.MarshalIndent(map[string]any{"app": os.Getpid(), "jobKillOnClose": true, "children": kids}, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(out, b, 0o644); err != nil {
		panic(err)
	}
}
