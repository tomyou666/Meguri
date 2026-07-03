//go:build windows

package chromiumfetch

import (
	"os/exec"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"github.com/chromedp/chromedp"
	"golang.org/x/sys/windows"
)

// browserJob は Windows Job Object でブラウザプロセスツリーを管理する。
type browserJob struct {
	// handle は Job Object のハンドル。作成失敗時は 0。
	handle windows.Handle
	// rootPID はブラウザルートプロセスの PID。
	rootPID atomic.Int32
}

// newBrowserJob は Job Object 付きの browserJob と起動時フックを返す。
// Job Object 作成に失敗した場合は PID 追跡のみ行う。
func newBrowserJob() (*browserJob, chromedp.ExecAllocatorOption) {
	job := &browserJob{}
	handle, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return job, chromedp.ModifyCmdFunc(func(cmd *exec.Cmd) {
			trackBrowserPID(&job.rootPID, cmd)
		})
	}
	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
		BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
			LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
		},
	}
	_, _ = windows.SetInformationJobObject(
		handle,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
	)
	job.handle = handle
	return job, chromedp.ModifyCmdFunc(func(cmd *exec.Cmd) {
		trackBrowserPID(&job.rootPID, cmd)
		assignProcessToJobWhenReady(handle, cmd)
	})
}

// Close はプロセスツリーを終了し、Job Object ハンドルを閉じる。
func (j *browserJob) Close() {
	pid := int(j.rootPID.Load())
	if pid > 0 {
		killProcessTree(pid)
	}
	if j.handle != 0 {
		_ = windows.CloseHandle(j.handle)
		j.handle = 0
	}
}

// trackBrowserPID は起動後のプロセス PID を store に記録する。
func trackBrowserPID(store *atomic.Int32, cmd *exec.Cmd) {
	go func() {
		for i := 0; i < 200; i++ {
			if cmd.Process != nil {
				store.Store(int32(cmd.Process.Pid))
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
}

// assignProcessToJobWhenReady はプロセス起動後に Job Object へ割り当てる。
func assignProcessToJobWhenReady(job windows.Handle, cmd *exec.Cmd) {
	go func() {
		for i := 0; i < 200; i++ {
			if cmd.Process != nil {
				pid := uint32(cmd.Process.Pid)
				handle, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, pid)
				if err == nil {
					_ = windows.AssignProcessToJobObject(job, handle)
					_ = windows.CloseHandle(handle)
				}
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
}

// killProcessTree は taskkill で指定 PID のプロセスツリーを強制終了する。
func killProcessTree(pid int) {
	if pid <= 0 {
		return
	}
	taskkill := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/T", "/F")
	taskkill.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = taskkill.Run()
}
