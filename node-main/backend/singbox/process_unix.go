//go:build !windows

package singbox

import (
	"os"
	"os/exec"
	"syscall"
	"time"
)

// setProcAttributes configures the sing-box subprocess for clean group kills.
func setProcAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}
}

// killProcessTree sends SIGTERM/SIGKILL to the whole process group.
func killProcessTree(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	_ = proc.Signal(syscall.SIGTERM)
	if pgid, err := syscall.Getpgid(pid); err == nil && pgid != 0 {
		_ = syscall.Kill(-pgid, syscall.SIGTERM)
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
	}
	_ = proc.Signal(syscall.SIGKILL)
	for i := 0; i < 10; i++ {
		if !isProcessRunning(pid) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func isProcessRunning(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}
