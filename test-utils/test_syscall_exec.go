//go:build windows
// +build windows

package main

import (
	"fmt"
	"syscall"
)

func main() {
	err := syscall.Exec("cmd.exe", []string{"cmd.exe"}, []string{})
	if err != nil {
		fmt.Printf("syscall.Exec error: %v\n", err)
	}
}
