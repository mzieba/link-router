//go:build !linux && !windows

package register

import (
	"fmt"
	"runtime"
)

func Register(string) error { return fmt.Errorf("registration not supported on %s", runtime.GOOS) }
func Unregister() error     { return fmt.Errorf("unregistration not supported on %s", runtime.GOOS) }
