//go:build dev && !windows

package wails_service

import (
	"os/exec"
	"syscall"
)

func applyDetachedHelperAttrs(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
