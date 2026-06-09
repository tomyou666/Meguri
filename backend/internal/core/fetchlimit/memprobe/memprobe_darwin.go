//go:build darwin

package memprobe

import "golang.org/x/sys/unix"

// darwinVMStatsSyscall は macOS のメモリ統計を sysctl で近似取得する。
func darwinVMStatsSyscall() (total, used uint64, err error) {
	total, err = unix.SysctlUint64("hw.memsize")
	if err != nil {
		return 0, 0, err
	}
	freePages, err := unix.SysctlUint64("vm.page_free_count")
	if err != nil {
		return total, total / 2, nil
	}
	pageSize, err := unix.SysctlUint64("hw.pagesize")
	if err != nil {
		pageSize = 4096
	}
	free := freePages * pageSize
	if free > total {
		free = total
	}
	used = total - free
	return total, used, nil
}
