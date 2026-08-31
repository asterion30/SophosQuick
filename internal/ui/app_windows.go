//go:build windows

package ui

import (
	"fmt"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"sophosquick/internal/config"
	"sophosquick/internal/crypto"
	"sophosquick/internal/sophos"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procRegisterClassExW     = user32.NewProc("RegisterClassExW")
	procCreateWindowExW      = user32.NewProc("CreateWindowExW")
	procDefWindowProcW       = user32.NewProc("DefWindowProcW")
	procShowWindow           = user32.NewProc("ShowWindow")
	procUpdateWindow         = user32.NewProc("UpdateWindow")
	procGetMessageW          = user32.NewProc("GetMessageW")
	procTranslateMessage     = user32.NewProc("TranslateMessage")
	procDispatchMessageW     = user32.NewProc("DispatchMessageW")
	procSendMessageW         = user32.NewProc("SendMessageW")
	procSetWindowTextW       = user32.NewProc("SetWindowTextW")
	procGetWindowTextW       = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	procSetFocus             = user32.NewProc("SetFocus")
	procEnableWindow         = user32.NewProc("EnableWindow")
	procPostQuitMessage      = user32.NewProc("PostQuitMessage")
	procMessageBoxW          = user32.NewProc("MessageBoxW")
	procGetSystemMetrics     = user32.NewProc("GetSystemMetrics")
	procSetWindowPos         = user32.NewProc("SetWindowPos")
	procDestroyWindow        = user32.NewProc("DestroyWindow")

	procCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
	procSetTextColor     = gdi32.NewProc("SetTextColor")
	procSetBkColor       = gdi32.NewProc("SetBkColor")
	procSetBkMode        = gdi32.NewProc("SetBkMode")
	procCreateFontW      = gdi32.NewProc("CreateFontW")

	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

const (
	WS_OVERLAPPED   = 0x00000000
	WS_CAPTION      = 0x00C00000
	WS_SYSMENU      = 0x00080000
	WS_MINIMIZEBOX  = 0x00020000
	WS_VISIBLE      = 0x10000000
	WS_CHILD        = 0x40000000
	WS_TABSTOP      = 0x00010000
	WS_BORDER       = 0x00800000

	WS_EX_CLIENTEDGE  = 0x00000200
	WS_EX_TOPMOST     = 0x00000008
	WS_EX_DLGMODALFRAME = 0x00000001

	CBS_DROPDOWNLIST = 0x0003
	CBS_HASSTRINGS   = 0x0200

	ES_LEFT      = 0x0000
	ES_CENTER    = 0x0001
	ES_PASSWORD  = 0x0020
	ES_AUTOHSCROLL = 0x0080
	ES_NUMBER    = 0x2000

	BS_DEFPUSHBUTTON = 0x0001
	BS_PUSHBUTTON    = 0x0000

	SS_LEFT   = 0x0000
	SS_CENTER = 0x0001
	SS_RIGHT  = 0x0002

	WM_CREATE         = 0x0001
	WM_DESTROY        = 0x0002
	WM_CLOSE          = 0x0010
	WM_COMMAND        = 0x0111
	WM_SETFONT        = 0x0030
	WM_CTLCOLORSTATIC = 0x0138
	WM_CTLCOLOREDIT   = 0x0133
	WM_CTLCOLORBTN    = 0x0135
	WM_CTLCOLORDLG    = 0x0136

	CB_ADDSTRING = 0x0143
	CB_SETCURSEL = 0x014E
	CB_GETCURSEL = 0x0147
	CB_GETLBTEXT = 0x0148
	CB_RESETCONTENT = 0x014B

	MB_OK          = 0x00000000
	MB_ICONINFO    = 0x00000040
	MB_ICONERROR   = 0x00000010
	MB_ICONWARNING = 0x00000030

	SM_CXSCREEN = 0
	SM_CYSCREEN = 1
	TRANSPARENT = 1

	// Control IDs
	ID_COMBO_PROFILE   = 1001
	ID_BTN_REFRESH     = 1002
	ID_EDIT_TOTP       = 1003
	ID_BTN_CONNECT     = 1004
	ID_BTN_DISCONNECT  = 1005
	ID_BTN_CONFIG_PASS = 1006
	ID_LBL_STATUS      = 1007
	ID_LBL_FEEDBACK    = 1008

	ID_PASS_EDIT   = 2001
	ID_PASS_SAVE   = 2002
	ID_PASS_CANCEL = 2003
)

type WNDCLASSEXW struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type MSG struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      struct{ x, y int32 }
}

type WindowsUI struct {
	cfg       *config.Config
	client    *sophos.Client
	hwnd      uintptr
	hCombo    uintptr
	hTotp     uintptr
	hBtnConn  uintptr
	hBtnDisc  uintptr
	hBtnPass  uintptr
	hBtnRef   uintptr
	hStatus   uintptr
	hFeedback uintptr

	bgBrush    uintptr
	cardBrush  uintptr
	fontTitle  uintptr
	fontBody   uintptr
	fontBold   uintptr
	fontTotp   uintptr

	isBusy bool
}

var globalUI *WindowsUI

func newPlatformUI(cfg *config.Config, client *sophos.Client) UI {
	ui := &WindowsUI{
		cfg:    cfg,
		client: client,
	}
	globalUI = ui
	return ui
}

func utf16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func (ui *WindowsUI) ShowAndRun() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hInstance, _, _ := procGetModuleHandleW.Call(0)
	className := utf16Ptr("SophosQuickWindowClass")

	// Colors: Dark Slate theme (#0F172A = RGB(15, 23, 42))
	ui.bgBrush, _, _ = procCreateSolidBrush.Call(uintptr(0x002A170F))   // 0x00BBGGRR
	ui.cardBrush, _, _ = procCreateSolidBrush.Call(uintptr(0x003B291E)) // #1E293B

	// Fonts (Segoe UI)
	ui.fontTitle = createFont("Segoe UI", 20, 700)
	ui.fontBody = createFont("Segoe UI", 14, 400)
	ui.fontBold = createFont("Segoe UI", 14, 600)
	ui.fontTotp = createFont("Consolas", 22, 700)

	var wc WNDCLASSEXW
	wc.cbSize = uint32(unsafe.Sizeof(wc))
	wc.lpfnWndProc = syscall.NewCallback(wndProc)
	wc.hInstance = hInstance
	wc.hbrBackground = ui.bgBrush
	wc.lpszClassName = className

	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	width := int32(350)
	height := int32(410)

	screenWidth, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
	screenHeight, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)
	posX := (int32(screenWidth) - width) / 2
	posY := (int32(screenHeight) - height) / 2

	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16Ptr("SophosQuick - VPN Launcher"))),
		WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU|WS_MINIMIZEBOX|WS_VISIBLE,
		uintptr(posX), uintptr(posY), uintptr(width), uintptr(height),
		0, 0, hInstance, 0,
	)

	ui.hwnd = hwnd

	procShowWindow.Call(hwnd, 1)
	procUpdateWindow.Call(hwnd)
	procSetFocus.Call(ui.hTotp)

	var msg MSG
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func createFont(name string, size int32, weight int32) uintptr {
	hFont, _, _ := procCreateFontW.Call(
		uintptr(size), 0, 0, 0, uintptr(weight),
		0, 0, 0, 0, 0, 0, 0, 0,
		uintptr(unsafe.Pointer(utf16Ptr(name))),
	)
	return hFont
}

func wndProc(hwnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	ui := globalUI
	if ui == nil {
		r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
		return r
	}

	switch msg {
	case WM_CREATE:
		ui.buildControls(hwnd)
		return 0

	case WM_COMMAND:
		id := uint16(wParam & 0xFFFF)
		code := uint16((wParam >> 16) & 0xFFFF)

		if code == 0 || code == 1 { // Button click / Menu command
			switch id {
			case ID_BTN_CONNECT:
				if !ui.isBusy {
					ui.onConnect()
				}
			case ID_BTN_DISCONNECT:
				if !ui.isBusy {
					ui.onDisconnect()
				}
			case ID_BTN_REFRESH:
				ui.refreshProfiles()
			case ID_BTN_CONFIG_PASS:
				ui.showPasswordDialog()
			}
		}
		return 0

	case WM_CTLCOLORSTATIC, WM_CTLCOLORDLG:
		hdc := wParam
		// Text Color: White (#F8FAFC)
		procSetTextColor.Call(hdc, uintptr(0x00FCFAF8))
		// Background Color: Dark Slate (#0F172A)
		procSetBkColor.Call(hdc, uintptr(0x002A170F))
		procSetBkMode.Call(hdc, TRANSPARENT)
		return ui.bgBrush

	case WM_CTLCOLOREDIT:
		hdc := wParam
		procSetTextColor.Call(hdc, uintptr(0x00000000))
		procSetBkColor.Call(hdc, uintptr(0x00FFFFFF))
		return ui.cardBrush

	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	}

	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

func (ui *WindowsUI) buildControls(parent uintptr) {
	hInstance, _, _ := procGetModuleHandleW.Call(0)

	// 1. Header Title
	hTitle, _, _ := procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(utf16Ptr("STATIC"))),
		uintptr(unsafe.Pointer(utf16Ptr("SophosQuick"))),
		WS_CHILD|WS_VISIBLE|SS_LEFT,
		20, 15, 180, 26, parent, 0, hInstance, 0,
	)
	procSendMessageW.Call(hTitle, WM_SETFONT, ui.fontTitle, 1)

	// 2. Subtitle
	hSub, _, _ := procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(utf16Ptr("STATIC"))),
		uintptr(unsafe.Pointer(utf16Ptr("Secure VPN Launcher"))),
		WS_CHILD|WS_VISIBLE|SS_LEFT,
		20, 42, 180, 18, parent, 0, hInstance, 0,
	)
	procSendMessageW.Call(hSub, WM_SETFONT, ui.fontBody, 1)

	// 3. Status Label (Top Right)
	ui.hStatus, _, _ = procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(utf16Ptr("STATIC"))),
		uintptr(unsafe.Pointer(utf16Ptr("⚪ Desconectado"))),
		WS_CHILD|WS_VISIBLE|SS_RIGHT,
		200, 18, 120, 20, parent, uintptr(ID_LBL_STATUS), hInstance, 0,
	)
	procSendMessageW.Call(ui.hStatus, WM_SETFONT, ui.fontBold, 1)

	// 4. Label: Perfil de Conexión
	hLblProfile, _, _ := procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(utf16Ptr("STATIC"))),
		uintptr(unsafe.Pointer(utf16Ptr("Perfil de Conexión:"))),
		WS_CHILD|WS_VISIBLE|SS_LEFT,
		20, 72, 290, 18, parent, 0, hInstance, 0,
	)
	procSendMessageW.Call(hLblProfile, WM_SETFONT, ui.fontBold, 1)

	// 5. ComboBox Profile
	ui.hCombo, _, _ = procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(utf16Ptr("COMBOBOX"))),
		0,
		WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_BORDER|CBS_DROPDOWNLIST|CBS_HASSTRINGS,
		20, 93, 255, 200, parent, uintptr(ID_COMBO_PROFILE), hInstance, 0,
	)
	procSendMessageW.Call(ui.hCombo, WM_SETFONT, ui.fontBody, 1)

	// 6. Refresh Button
	ui.hBtnRef, _, _ = procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(utf16Ptr("BUTTON"))),
		uintptr(unsafe.Pointer(utf16Ptr("🔄"))),
		WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON,
		280, 92, 35, 25, parent, uintptr(ID_BTN_REFRESH), hInstance, 0,
	)
	procSendMessageW.Call(ui.hBtnRef, WM_SETFONT, ui.fontBody, 1)

	// Populate profiles
	ui.refreshProfiles()

	// 7. Label: Código MFA / TOTP
	hLblTotp, _, _ := procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(utf16Ptr("STATIC"))),
		uintptr(unsafe.Pointer(utf16Ptr("Código MFA / TOTP (6 dígitos):"))),
		WS_CHILD|WS_VISIBLE|SS_LEFT,
		20, 130, 290, 18, parent, 0, hInstance, 0,
	)
	procSendMessageW.Call(hLblTotp, WM_SETFONT, ui.fontBold, 1)

	// 8. Edit Box TOTP
	ui.hTotp, _, _ = procCreateWindowExW.Call(
		WS_EX_CLIENTEDGE, uintptr(unsafe.Pointer(utf16Ptr("EDIT"))),
		0,
		WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_BORDER|ES_CENTER|ES_NUMBER|ES_AUTOHSCROLL,
		20, 150, 295, 34, parent, uintptr(ID_EDIT_TOTP), hInstance, 0,
	)
	procSendMessageW.Call(ui.hTotp, WM_SETFONT, ui.fontTotp, 1)

	// 9. Primary Connect Button
	ui.hBtnConn, _, _ = procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(utf16Ptr("BUTTON"))),
		uintptr(unsafe.Pointer(utf16Ptr("🚀 Conectar VPN"))),
		WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_DEFPUSHBUTTON,
		20, 198, 295, 38, parent, uintptr(ID_BTN_CONNECT), hInstance, 0,
	)
	procSendMessageW.Call(ui.hBtnConn, WM_SETFONT, ui.fontBold, 1)

	// 10. Disconnect Button
	ui.hBtnDisc, _, _ = procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(utf16Ptr("BUTTON"))),
		uintptr(unsafe.Pointer(utf16Ptr("🛑 Desconectar"))),
		WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON,
		20, 244, 295, 30, parent, uintptr(ID_BTN_DISCONNECT), hInstance, 0,
	)
	procSendMessageW.Call(ui.hBtnDisc, WM_SETFONT, ui.fontBody, 1)

	// 11. Config Password Button
	ui.hBtnPass, _, _ = procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(utf16Ptr("BUTTON"))),
		uintptr(unsafe.Pointer(utf16Ptr("🔑 Configurar Contraseña Base"))),
		WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON,
		20, 282, 295, 28, parent, uintptr(ID_BTN_CONFIG_PASS), hInstance, 0,
	)
	procSendMessageW.Call(ui.hBtnPass, WM_SETFONT, ui.fontBody, 1)

	// 12. Feedback Message Label
	ui.hFeedback, _, _ = procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(utf16Ptr("STATIC"))),
		uintptr(unsafe.Pointer(utf16Ptr("Listo para conectar"))),
		WS_CHILD|WS_VISIBLE|SS_CENTER,
		20, 322, 295, 20, parent, uintptr(ID_LBL_FEEDBACK), hInstance, 0,
	)
	procSendMessageW.Call(ui.hFeedback, WM_SETFONT, ui.fontBody, 1)

	if !crypto.HasSavedCredential() {
		ui.setFeedback("⚠️ Requiere configurar contraseña base por 1ra vez")
	}
}

func (ui *WindowsUI) refreshProfiles() {
	procSendMessageW.Call(ui.hCombo, CB_RESETCONTENT, 0, 0)

	profiles, _ := sophos.DiscoverConnections(ui.client, ui.cfg.FallbackConnections, ui.cfg.ConnectionsDir)
	for _, p := range profiles {
		procSendMessageW.Call(ui.hCombo, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(utf16Ptr(p))))
	}

	if len(profiles) > 0 {
		procSendMessageW.Call(ui.hCombo, CB_SETCURSEL, 0, 0)
	}
}

func (ui *WindowsUI) getSelectedProfile() string {
	idx, _, _ := procSendMessageW.Call(ui.hCombo, CB_GETCURSEL, 0, 0)
	if int32(idx) < 0 {
		return ui.cfg.DefaultConnection
	}

	buf := make([]uint16, 256)
	procSendMessageW.Call(ui.hCombo, CB_GETLBTEXT, idx, uintptr(unsafe.Pointer(&buf[0])))
	return syscall.UTF16ToString(buf)
}

func (ui *WindowsUI) getTotpText() string {
	length, _, _ := procGetWindowTextLengthW.Call(ui.hTotp)
	if length == 0 {
		return ""
	}
	buf := make([]uint16, length+1)
	procGetWindowTextW.Call(ui.hTotp, uintptr(unsafe.Pointer(&buf[0])), length+1)
	return strings.TrimSpace(syscall.UTF16ToString(buf))
}

func (ui *WindowsUI) setStatus(text string) {
	procSetWindowTextW.Call(ui.hStatus, uintptr(unsafe.Pointer(utf16Ptr(text))))
}

func (ui *WindowsUI) setFeedback(text string) {
	procSetWindowTextW.Call(ui.hFeedback, uintptr(unsafe.Pointer(utf16Ptr(text))))
}

func (ui *WindowsUI) setBusy(busy bool) {
	ui.isBusy = busy
	val := uintptr(1)
	if busy {
		val = 0
		procSetWindowTextW.Call(ui.hBtnConn, uintptr(unsafe.Pointer(utf16Ptr("⌛ Conectando..."))))
	} else {
		procSetWindowTextW.Call(ui.hBtnConn, uintptr(unsafe.Pointer(utf16Ptr("🚀 Conectar VPN"))))
	}
	procEnableWindow.Call(ui.hBtnConn, val)
	procEnableWindow.Call(ui.hBtnDisc, val)
	procEnableWindow.Call(ui.hBtnPass, val)
	procEnableWindow.Call(ui.hBtnRef, val)
}

func (ui *WindowsUI) onConnect() {
	totp := ui.getTotpText()
	if totp == "" {
		procMessageBoxW.Call(ui.hwnd,
			uintptr(unsafe.Pointer(utf16Ptr("Por favor ingresa tu código TOTP / MFA de 6 dígitos generado por tu app de autenticación."))),
			uintptr(unsafe.Pointer(utf16Ptr("Código MFA Requerido"))),
			MB_OK|MB_ICONWARNING,
		)
		procSetFocus.Call(ui.hTotp)
		return
	}

	if !crypto.HasSavedCredential() {
		procMessageBoxW.Call(ui.hwnd,
			uintptr(unsafe.Pointer(utf16Ptr("Debes configurar primero tu contraseña corporativa antes de conectar."))),
			uintptr(unsafe.Pointer(utf16Ptr("Configuración Requerida"))),
			MB_OK|MB_ICONWARNING,
		)
		ui.showPasswordDialog()
		return
	}

	basePass, err := crypto.LoadPassword()
	if err != nil {
		procMessageBoxW.Call(ui.hwnd,
			uintptr(unsafe.Pointer(utf16Ptr(fmt.Sprintf("Error al leer la contraseña base cifrada:\n%v", err)))),
			uintptr(unsafe.Pointer(utf16Ptr("Error de Credenciales"))),
			MB_OK|MB_ICONERROR,
		)
		return
	}

	selectedProfile := ui.getSelectedProfile()
	username := ui.cfg.Username
	if username == "" {
		username = config.GetCurrentUsername()
	}

	fullPassword := basePass + totp

	ui.setBusy(true)
	ui.setStatus("🟡 Conectando...")
	ui.setFeedback(fmt.Sprintf("Conectando a %s...", selectedProfile))

	go func() {
		output, err := ui.client.Connect(selectedProfile, username, fullPassword)
		time.Sleep(400 * time.Millisecond)

		ui.setBusy(false)
		if err != nil {
			ui.setStatus("⚪ Desconectado")
			ui.setFeedback("❌ Error en conexión")
			procMessageBoxW.Call(ui.hwnd,
				uintptr(unsafe.Pointer(utf16Ptr(fmt.Sprintf("Fallo al conectar:\n%v", err)))),
				uintptr(unsafe.Pointer(utf16Ptr("Error al Conectar"))),
				MB_OK|MB_ICONERROR,
			)
			procSetFocus.Call(ui.hTotp)
		} else {
			ui.setStatus("🟢 Conectado")
			ui.setFeedback("✅ Conexión iniciada con éxito")
			procSetWindowTextW.Call(ui.hTotp, uintptr(unsafe.Pointer(utf16Ptr(""))))
			procMessageBoxW.Call(ui.hwnd,
				uintptr(unsafe.Pointer(utf16Ptr(fmt.Sprintf("Se ha enviado la orden de conexión a Sophos Connect:\n%s\n\nVerifica el agente de Sophos en la barra de tareas.", output)))),
				uintptr(unsafe.Pointer(utf16Ptr("VPN en Proceso"))),
				MB_OK|MB_ICONINFO,
			)
		}
	}()
}

func (ui *WindowsUI) onDisconnect() {
	selectedProfile := ui.getSelectedProfile()
	ui.setBusy(true)
	ui.setFeedback("Desconectando...")

	go func() {
		output, err := ui.client.Disconnect(selectedProfile)
		time.Sleep(300 * time.Millisecond)

		ui.setBusy(false)
		ui.setStatus("⚪ Desconectado")
		if err != nil {
			ui.setFeedback("❌ Error al desconectar")
			procMessageBoxW.Call(ui.hwnd,
				uintptr(unsafe.Pointer(utf16Ptr(fmt.Sprintf("Error al desconectar:\n%v", err)))),
				uintptr(unsafe.Pointer(utf16Ptr("Error de Desconexión"))),
				MB_OK|MB_ICONERROR,
			)
		} else {
			ui.setFeedback("Desconectado")
			procMessageBoxW.Call(ui.hwnd,
				uintptr(unsafe.Pointer(utf16Ptr(fmt.Sprintf("Se ha enviado la señal de desconexión:\n%s", output)))),
				uintptr(unsafe.Pointer(utf16Ptr("VPN Desconectada"))),
				MB_OK|MB_ICONINFO,
			)
		}
		procSetFocus.Call(ui.hTotp)
	}()
}

func (ui *WindowsUI) showPasswordDialog() {
	hInstance, _, _ := procGetModuleHandleW.Call(0)
	dlgClass := utf16Ptr("SophosQuickPasswordDialogClass")

	var wc WNDCLASSEXW
	wc.cbSize = uint32(unsafe.Sizeof(wc))
	wc.lpfnWndProc = syscall.NewCallback(passDlgWndProc)
	wc.hInstance = hInstance
	wc.hbrBackground = ui.bgBrush
	wc.lpszClassName = dlgClass

	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	width := int32(320)
	height := int32(180)

	screenWidth, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
	screenHeight, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)
	posX := (int32(screenWidth) - width) / 2
	posY := (int32(screenHeight) - height) / 2

	hDlg, _, _ := procCreateWindowExW.Call(
		WS_EX_DLGMODALFRAME|WS_EX_TOPMOST,
		uintptr(unsafe.Pointer(dlgClass)),
		uintptr(unsafe.Pointer(utf16Ptr("Configurar Contraseña Base"))),
		WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU|WS_VISIBLE,
		uintptr(posX), uintptr(posY), uintptr(width), uintptr(height),
		ui.hwnd, 0, hInstance, 0,
	)

	// Label
	hLbl, _, _ := procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(utf16Ptr("STATIC"))),
		uintptr(unsafe.Pointer(utf16Ptr("Ingresa tu contraseña corporativa (SIN TOTP):"))),
		WS_CHILD|WS_VISIBLE|SS_LEFT,
		15, 15, 275, 18, hDlg, 0, hInstance, 0,
	)
	procSendMessageW.Call(hLbl, WM_SETFONT, ui.fontBody, 1)

	// Password edit
	hEdit, _, _ := procCreateWindowExW.Call(
		WS_EX_CLIENTEDGE, uintptr(unsafe.Pointer(utf16Ptr("EDIT"))),
		0,
		WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_BORDER|ES_PASSWORD|ES_AUTOHSCROLL,
		15, 38, 275, 26, hDlg, uintptr(ID_PASS_EDIT), hInstance, 0,
	)
	procSendMessageW.Call(hEdit, WM_SETFONT, ui.fontBody, 1)

	// Save Button
	hSave, _, _ := procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(utf16Ptr("BUTTON"))),
		uintptr(unsafe.Pointer(utf16Ptr("Guardar"))),
		WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_DEFPUSHBUTTON,
		15, 80, 130, 32, hDlg, uintptr(ID_PASS_SAVE), hInstance, 0,
	)
	procSendMessageW.Call(hSave, WM_SETFONT, ui.fontBold, 1)

	// Cancel Button
	hCancel, _, _ := procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(utf16Ptr("BUTTON"))),
		uintptr(unsafe.Pointer(utf16Ptr("Cancelar"))),
		WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON,
		160, 80, 130, 32, hDlg, uintptr(ID_PASS_CANCEL), hInstance, 0,
	)
	procSendMessageW.Call(hCancel, WM_SETFONT, ui.fontBody, 1)

	procSetFocus.Call(hEdit)
}

func passDlgWndProc(hwnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	ui := globalUI

	switch msg {
	case WM_COMMAND:
		id := uint16(wParam & 0xFFFF)
		switch id {
		case ID_PASS_SAVE:
			hEdit, _, _ := user32.NewProc("GetDlgItem").Call(hwnd, uintptr(ID_PASS_EDIT))
			len, _, _ := procGetWindowTextLengthW.Call(hEdit)
			if len == 0 {
				procMessageBoxW.Call(hwnd,
					uintptr(unsafe.Pointer(utf16Ptr("La contraseña no puede estar vacía."))),
					uintptr(unsafe.Pointer(utf16Ptr("Validación"))),
					MB_OK|MB_ICONWARNING,
				)
				return 0
			}
			buf := make([]uint16, len+1)
			procGetWindowTextW.Call(hEdit, uintptr(unsafe.Pointer(&buf[0])), len+1)
			pass := strings.TrimSpace(syscall.UTF16ToString(buf))

			err := crypto.SavePassword(pass)
			if err != nil {
				procMessageBoxW.Call(hwnd,
					uintptr(unsafe.Pointer(utf16Ptr(fmt.Sprintf("No se pudo guardar la contraseña:\n%v", err)))),
					uintptr(unsafe.Pointer(utf16Ptr("Error"))),
					MB_OK|MB_ICONERROR,
				)
			} else {
				ui.setFeedback("🔑 Contraseña guardada correctamente")
				procMessageBoxW.Call(hwnd,
					uintptr(unsafe.Pointer(utf16Ptr("¡Contraseña encriptada y guardada con éxito en la bóveda DPAPI del sistema!"))),
					uintptr(unsafe.Pointer(utf16Ptr("Éxito"))),
					MB_OK|MB_ICONINFO,
				)
				procDestroyWindow.Call(hwnd)
			}
			return 0

		case ID_PASS_CANCEL:
			procDestroyWindow.Call(hwnd)
			return 0
		}

	case WM_CTLCOLORSTATIC, WM_CTLCOLORDLG:
		hdc := wParam
		procSetTextColor.Call(hdc, uintptr(0x00FCFAF8))
		procSetBkColor.Call(hdc, uintptr(0x002A170F))
		procSetBkMode.Call(hdc, TRANSPARENT)
		return ui.bgBrush

	case WM_CLOSE:
		procDestroyWindow.Call(hwnd)
		return 0
	}

	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}
