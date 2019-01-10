// +build windows

package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func isTemporaryX(err error) bool {
	if errno, ok := err.(syscall.Errno); ok {
		return errno == syscall.Errno(10061)
	}
	return false
}

func createSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

func terminateRecursively(pid int) error {
	pe := syscall.ProcessEntry32{}
	pe.Size = uint32(unsafe.Sizeof(pe))

	hSnap, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return fmt.Errorf("could iterate over children of process #%d: %v", pid, err)
	}

	err = syscall.Process32First(hSnap, &pe)
	if err != nil {
		return err
	}

	tryNext := true
	for tryNext {
		if pe.ParentProcessID == uint32(pid) {
			if err := terminate(int(pe.ProcessID)); err != nil {
				return err
			}
		}
		tryNext = syscall.Process32Next(hSnap, &pe) == nil
	}

	return terminate(pid)
}

func terminate(pid int) error {
	h, e := syscall.OpenProcess(syscall.PROCESS_TERMINATE, false, uint32(pid))
	if e != nil {
		return os.NewSyscallError("OpenProcess", e)
	}
	//noinspection GoUnhandledErrorResult
	defer syscall.CloseHandle(h)
	e = syscall.TerminateProcess(h, uint32(1))
	return os.NewSyscallError("TerminateProcess", e)
}
