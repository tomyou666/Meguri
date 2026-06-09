//go:build !windows && !darwin

package memprobe

import "fmt"

// darwinVMStatsSyscall は darwin 以外では未使用のスタブ。
func darwinVMStatsSyscall() (total, used uint64, err error) {
	return 0, 0, fmt.Errorf("darwin only")
}
