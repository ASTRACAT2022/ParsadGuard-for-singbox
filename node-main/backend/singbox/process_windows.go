//go:build windows

package singbox

import (
	"os"
	"os/exec"
	"time"
)

func setProcAttributes(cmd *exec.Cmd) {}

func killProcessTree(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	_ = proc.Kill()
	for i := 0; i < 10; i++ {
		if proc.Signal(os.Signal(nil)) != nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}
