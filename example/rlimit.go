// +build !windows

package main

import (
	"log"
	"syscall"
)

func init() {
	// Increase resources limitations
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err == nil {
		rLimit.Cur = rLimit.Max
		err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	}
	if err == nil {
		log.Println("syscallSetRlimit: OK")
	} else {
		log.Panicln("syscallSetRlimit:", err)
	}
}
