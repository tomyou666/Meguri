//go:build !windows

package chromiumfetch

import (
	"os/exec"

	"github.com/chromedp/chromedp"
)

// browserJob は非 Windows 向けのブラウザプロセス管理スタブ。
type browserJob struct{}

// newBrowserJob は no-op の browserJob と ModifyCmdFunc を返す。
func newBrowserJob() (*browserJob, chromedp.ExecAllocatorOption) {
	return &browserJob{}, chromedp.ModifyCmdFunc(func(*exec.Cmd) {})
}

// Close は非 Windows では何もしない。
func (j *browserJob) Close() {}

// killProcessTree は非 Windows では何もしない。
func killProcessTree(_ int) {}
