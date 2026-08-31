//go:build windows

package sophos

import "syscall"

// CREATE_NO_WINDOW is the Windows process creation flag that prevents console window creation.
const CREATE_NO_WINDOW = 0x08000000

// hideWindowSysProcAttr returns SysProcAttr configured to hide console windows on Windows.
func hideWindowSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: CREATE_NO_WINDOW,
		HideWindow:    true,
	}
}
