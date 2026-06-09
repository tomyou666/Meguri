//go:build windows

package memprobe

import (
	"fmt"
	"syscall"
	"unsafe"
)

type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

// platformAvailableBytes は Windows の利用可能物理メモリを返す。
func platformAvailableBytes() (uint64, error) {
	var st memoryStatusEx
	st.Length = uint32(unsafe.Sizeof(st))
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	globalMemoryStatusEx := kernel32.NewProc("GlobalMemoryStatusEx")
	r, _, err := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&st)))
	if r == 0 {
		if err != syscall.Errno(0) {
			return 0, err
		}
		return 0, fmt.Errorf("GlobalMemoryStatusEx failed")
	}
	return st.AvailPhys, nil
}

// platformUsedRatio は Windows のメモリ使用率を返す。
func platformUsedRatio() (float64, error) {
	var st memoryStatusEx
	st.Length = uint32(unsafe.Sizeof(st))
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	globalMemoryStatusEx := kernel32.NewProc("GlobalMemoryStatusEx")
	r, _, err := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&st)))
	if r == 0 {
		if err != syscall.Errno(0) {
			return 0, err
		}
		return 0, fmt.Errorf("GlobalMemoryStatusEx failed")
	}
	if st.TotalPhys == 0 {
		return 0, nil
	}
	used := st.TotalPhys - st.AvailPhys
	return float64(used) / float64(st.TotalPhys), nil
}
