//go:build !windows

package memprobe

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// platformAvailableBytes は Unix 系 OS の利用可能メモリを返す。
func platformAvailableBytes() (uint64, error) {
	switch runtime.GOOS {
	case "linux":
		return linuxMemAvailable()
	case "darwin":
		return darwinMemAvailable()
	default:
		return 0, fmt.Errorf("memprobe: unsupported GOOS %s", runtime.GOOS)
	}
}

// platformUsedRatio は Unix 系 OS のメモリ使用率を返す。
func platformUsedRatio() (float64, error) {
	switch runtime.GOOS {
	case "linux":
		return linuxUsedRatio()
	case "darwin":
		return darwinUsedRatio()
	default:
		return 0, fmt.Errorf("memprobe: unsupported GOOS %s", runtime.GOOS)
	}
}

func linuxMemAvailable() (uint64, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				break
			}
			kb, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return 0, err
			}
			return kb * 1024, nil
		}
	}
	return 0, fmt.Errorf("MemAvailable not found in /proc/meminfo")
}

func linuxUsedRatio() (float64, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer f.Close()

	var total, available uint64
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		val, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		switch fields[0] {
		case "MemTotal:":
			total = val * 1024
		case "MemAvailable:":
			available = val * 1024
		}
	}
	if total == 0 {
		return 0, fmt.Errorf("MemTotal not found in /proc/meminfo")
	}
	used := total - available
	return float64(used) / float64(total), nil
}

func darwinMemAvailable() (uint64, error) {
	total, used, err := darwinVMStats()
	if err != nil {
		return 0, err
	}
	if used > total {
		return 0, nil
	}
	return total - used, nil
}

func darwinUsedRatio() (float64, error) {
	total, used, err := darwinVMStats()
	if err != nil {
		return 0, err
	}
	if total == 0 {
		return 0, nil
	}
	return float64(used) / float64(total), nil
}

func darwinVMStats() (total, used uint64, err error) {
	// sysctl は cgo なしでは煩雑なため runtime.MemStats とページサイズで近似する。
	// darwin では hw.memsize と host_statistics 相当が理想だが、簡易実装として
	// /proc が無い環境では syscall ベースのフォールバックを使う。
	return darwinVMStatsSyscall()
}
