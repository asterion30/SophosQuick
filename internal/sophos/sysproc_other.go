//go:build !windows

package sophos

import "syscall"

// hideWindowSysProcAttr returns default SysProcAttr for non-Windows platforms.
func hideWindowSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
