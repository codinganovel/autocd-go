package main

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
)

func main() {
	fmt.Printf("Running on: %s\n", runtime.GOOS)

	// Test if syscall.Exec is available
	executable := "/bin/echo"
	args := []string{"echo", "Hello from exec"}

	if runtime.GOOS == "windows" {
		executable = "cmd.exe"
		args = []string{"cmd.exe", "/c", "echo", "Hello from exec"}
	} else if runtime.GOOS == "darwin" {
		executable = "/bin/echo"
	}

	fmt.Println("Attempting syscall.Exec...")
	err := syscall.Exec(executable, args, os.Environ())

	// If we reach here, exec failed (it should replace the process)
	fmt.Printf("syscall.Exec returned error: %v\n", err)
}
