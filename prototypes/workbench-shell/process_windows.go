package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// 实验手段：只接受当前 WebView2 浏览器进程的唯一 renderer 子进程。
// 多 renderer 时拒绝猜测；生产恢复机制须另行依据实验结果决定。
func panelKillPID(browser int) (int, error) {
	if *killTarget == "browser" {
		return browser, nil
	}
	if *killTarget != "renderer" {
		return 0, fmt.Errorf("unknown kill target %q", *killTarget)
	}
	script := fmt.Sprintf("Get-CimInstance Win32_Process -Filter 'ParentProcessId=%d' | Where-Object { $_.Name -eq 'msedgewebview2.exe' -and $_.CommandLine -match '--type=renderer' } | ForEach-Object { $_.ProcessId }", browser)
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("enumerate renderer: %w", err)
	}
	ids := strings.Fields(string(out))
	if len(ids) != 1 {
		return 0, fmt.Errorf("browser %d has %d renderers; cannot select one", browser, len(ids))
	}
	return strconv.Atoi(ids[0])
}
