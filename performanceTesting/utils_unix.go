// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package main

import (
	"syscall"
)

func isTemporaryX(err error) bool {
	return false
}

func createSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}
}

func terminateRecursively(pid int) error {
	if pgid, err := syscall.Getpgid(pid); err == nil && syscall.Kill(-pgid, syscall.SIGKILL) == nil {
		return nil
	}
	return syscall.Kill(-pid, syscall.SIGKILL)
}
