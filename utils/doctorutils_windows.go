//go:build windows

package utils

import (
	"syscall"
	"unsafe"
)

func getDiskSpace(path string) (uint64, error) {
	var free, total, avail uint64
	h := syscall.MustLoadDLL("kernel32.dll").MustFindProc("GetDiskFreeSpaceExW")
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	_, _, err = h.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&free)),
		uintptr(unsafe.Pointer(&total)),
		uintptr(unsafe.Pointer(&avail)),
	)
	if err != nil && err != syscall.Errno(0) {
		return 0, err
	}
	return free, nil
}
