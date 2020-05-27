package utils

import (
	"syscall"
)

/*
 * @Author: ZhenpengDeng(monitor1379)
 * @Date: 2020-05-27 17:18:57
 * @Last Modified by: ZhenpengDeng(monitor1379)
 * @Last Modified time: 2020-05-27 17:25:04
 */

func GetSystemTotalMemory() (int64, error) {
	return getSystemTotalMemory()
}

// refer: https://github.com/pbnjay/memory/blob/master/memory_linux.go
func getSystemTotalMemory() (int64, error) {
	in := &syscall.Sysinfo_t{}
	err := syscall.Sysinfo(in)
	if err != nil {
		return 0, err
	}
	// If this is a 32-bit system, then these fields are
	// uint32 instead of uint64.
	// So we always convert to uint64 to match signature.
	return int64(in.Totalram) * int64(in.Unit), nil
}
