//go:build linux || bsd || darwin
// +build linux bsd darwin

package graceful

import (
	"os"
	"syscall"
)

var signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP}
