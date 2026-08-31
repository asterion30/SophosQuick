package ui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"sophosquick/internal/config"
	"sophosquick/internal/crypto"
	"sophosquick/internal/sophos"
)

// UI represents the main SophosQuick graphical application state.
type UI struct {
	App           fyne.App
	Window        fyne.Window
	Config        *config.Config
	Client        *sophos.Client

	// UI Elements
	statusDot     *canvas.Circle
	statusLabel   *widget.Label
	profileSelect *widget.Select
	totpEntry     *widget.Entry
	connectBtn    *widget.Button
	disconnectBtn *widget.Button
	configPassBtn *widget.Button
	feedbackLabel *widget.Label
	isBusy        bool
}

// New creates and initializes the SophosQuick UI.
func New(cfg *config.Config, client *sophos.Client) *UI {
	a := app.NewWithID("io.github.sophosquick")
	a.Settings().SetTheme(&DarkSlateTheme{})

	w := a.NewWindow("SophosQuick")
	w.Resize(fyne.NewSize(320, 360))
	w.SetFixedSize(true)
	w.CenterOnScreen()

	ui := &UI{
		App:    a,
		Window: w,
		Config: cfg,
		Client: client,
	}

	ui.buildView()
	return ui
}

// ShowAndRun renders the window and starts the event loop.
func (u *UI) ShowAndRun() {
	// Auto-focus the TOTP input upon window show
	u.Window.Canvas().Focus(u.totpEntry)
	u.Window.ShowAndRun()
}

// buildView constructs the compact card layout.
func (u *UI) buildView() {
	// Header Section
	titleLabel := widget.NewLabelWithStyle("SophosQuick", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	subTitleLabel := widget.NewLabelWithStyle("Secure VPN Launcher", fyne.TextAlignLeading, fyne.TextStyle{Italic: true})

	u.statusDot = canvas.NewCircle(ColorDarkSlate600)
	u.statusDot.Resize(fyne.NewSize(8, 8))

	u.statusLabel = widget.NewLabelWithStyle("Desconectado", fyne.TextAlignTrailing, fyne.TextStyle{Monospace: false})

	statusContainer := container.NewHBox(
		layout.NewSpacer(),
		container.NewCenter(u.statusDot),
		u.statusLabel,
	)

	headerLeft := container.NewVBox(titleLabel, subTitleLabel)
	header := container.NewBorder(nil, nil, headerLeft, statusContainer)

	// Connection Selector Section
	profileLabel := widget.NewLabelWithStyle("Perfil de Conexión", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	profiles, _ := sophos.DiscoverConnections(u.Client, u.Config.FallbackConnections, u.Config.ConnectionsDir)
	u.profileSelect = widget.NewSelect(profiles, func(selected string) {
		u.Config.DefaultConnection = selected
		_ = config.SaveConfig(u.Config)
	})

	if len(profiles) > 0 {
		selected := u.Config.DefaultConnection
		found := false
		for _, p := range profiles {
			if p == selected {
				found = true
				break
			}
		}
		if !found {
			selected = profiles[0]
		}
		u.profileSelect.SetSelected(selected)
	}

	refreshBtn := widget.NewButton("🔄", func() {
		u.refreshProfiles()
	})

	profileRow := container.NewBorder(nil, nil, nil, refreshBtn, u.profileSelect)

	// TOTP Input Section
	totpTitle := widget.NewLabelWithStyle("Código MFA / TOTP", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	u.totpEntry = widget.NewEntry()
	u.totpEntry.SetPlaceHolder("Ingresa código de 6 dígitos")
	u.totpEntry.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}

	// Submit on Enter keypress
	u.totpEntry.OnSubmitted = func(text string) {
		if !u.isBusy {
			u.onConnect()
		}
	}

	// Action Buttons
	u.connectBtn = widget.NewButton("🚀 Conectar VPN", func() {
		u.onConnect()
	})
	u.connectBtn.Importance = widget.HighImportance

	u.disconnectBtn = widget.NewButton("🛑 Desconectar", func() {
		u.onDisconnect()
	})

	// Secondary Settings Button
	u.configPassBtn = widget.NewButton("🔑 Configurar Contraseña Base", func() {
		u.showPasswordModal()
	})
	u.configPassBtn.Importance = widget.LowImportance

	// Dynamic Feedback Message Banner
	u.feedbackLabel = widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
	u.feedbackLabel.Wrapping = fyne.TextTruncate

	// Check if password exists; if not, prompt helper feedback
	if !crypto.HasSavedCredential() {
		u.setFeedback("⚠️ Requiere configurar contraseña base por 1ra vez")
	} else {
		u.setFeedback("Listo para conectar")
	}

	// Assemble card body
	formContainer := container.NewVBox(
		profileLabel,
		profileRow,
		widget.NewSeparator(),
		totpTitle,
		u.totpEntry,
		layout.NewSpacer(),
		u.connectBtn,
		u.disconnectBtn,
		widget.NewSeparator(),
		u.configPassBtn,
		u.feedbackLabel,
	)

	// Background card padding
	mainCard := container.NewPadded(
		container.NewBorder(
			container.NewVBox(header, widget.NewSeparator()),
			nil, nil, nil,
			formContainer,
		),
	)

	u.Window.SetContent(mainCard)
}

// refreshProfiles re-scans the environment for VPN profiles.
func (u *UI) refreshProfiles() {
	profiles, _ := sophos.DiscoverConnections(u.Client, u.Config.FallbackConnections, u.Config.ConnectionsDir)
	u.profileSelect.Options = profiles
	if len(profiles) > 0 {
		u.profileSelect.SetSelected(profiles[0])
	}
	u.profileSelect.Refresh()
	u.setFeedback("Perfiles actualizados 🔄")
}

// setStatus updates the top status indicator badge and label.
func (u *UI) setStatus(text string) {
	u.statusLabel.SetText(text)
	switch text {
	case "Conectado":
		u.statusDot.FillColor = ColorEmerald500
	case "Conectando...":
		u.statusDot.FillColor = ColorAmber500
	default:
		u.statusDot.FillColor = ColorDarkSlate600
	}
	u.statusDot.Refresh()
}

// setFeedback updates the bottom notification banner.
func (u *UI) setFeedback(msg string) {
	u.feedbackLabel.SetText(msg)
}

// setBusy toggles the interaction state of buttons during execution.
func (u *UI) setBusy(busy bool) {
	u.isBusy = busy
	if busy {
		u.connectBtn.Disable()
		u.disconnectBtn.Disable()
		u.connectBtn.SetText("⌛ Conectando...")
	} else {
		u.connectBtn.Enable()
		u.disconnectBtn.Enable()
		u.connectBtn.SetText("🚀 Conectar VPN")
	}
}

// onConnect handles the VPN connection workflow.
func (u *UI) onConnect() {
	totp := strings.TrimSpace(u.totpEntry.Text)
	if totp == "" {
		u.setFeedback("⚠️ Ingresa el código TOTP de 6 dígitos")
		dialog.ShowInformation("Código MFA requerido", "Por favor ingresa el código generado por tu app de autenticación (ej. Google Authenticator / Microsoft Authenticator).", u.Window)
		u.Window.Canvas().Focus(u.totpEntry)
		return
	}

	if !crypto.HasSavedCredential() {
		u.setFeedback("⚠️ Configura primero tu contraseña corporativa")
		u.showPasswordModal()
		return
	}

	basePass, err := crypto.LoadPassword()
	if err != nil {
		u.setFeedback("❌ Error al cargar credenciales")
		dialog.ShowError(fmt.Errorf("Error al leer la contraseña base cifrada:\n%v", err), u.Window)
		return
	}

	selectedProfile := u.profileSelect.Selected
	if selectedProfile == "" {
		if len(u.Config.FallbackConnections) > 0 {
			selectedProfile = u.Config.FallbackConnections[0]
		} else {
			selectedProfile = u.Config.DefaultConnection
		}
	}

	username := u.Config.Username
	if username == "" {
		username = config.GetCurrentUsername()
	}

	fullPassword := basePass + totp

	u.setBusy(true)
	u.setStatus("Conectando...")
	u.setFeedback(fmt.Sprintf("Conectando a %s...", selectedProfile))

	go func() {
		output, err := u.Client.Connect(selectedProfile, username, fullPassword)
		time.Sleep(500 * time.Millisecond) // Smooth UX transition

		fyne.Do(func() {
			u.setBusy(false)
			if err != nil {
				u.setStatus("Desconectado")
				u.setFeedback("❌ Error en conexión")
				dialog.ShowError(fmt.Errorf("Fallo al conectar:\n%v", err), u.Window)
				u.Window.Canvas().Focus(u.totpEntry)
			} else {
				u.setStatus("Conectado")
				u.setFeedback("✅ Conexión iniciada con éxito")
				// Clear TOTP entry for security after successful attempt
				u.totpEntry.SetText("")
				dialog.ShowInformation("VPN en Proceso", fmt.Sprintf("Se ha enviado la orden de conexión a Sophos Connect:\n%s\n\nVerifica el agente de Sophos en la bandeja del sistema.", output), u.Window)
			}
		})
	}()
}

// onDisconnect handles the VPN disconnection workflow.
func (u *UI) onDisconnect() {
	selectedProfile := u.profileSelect.Selected
	if selectedProfile == "" {
		if len(u.Config.FallbackConnections) > 0 {
			selectedProfile = u.Config.FallbackConnections[0]
		} else {
			selectedProfile = u.Config.DefaultConnection
		}
	}

	u.setBusy(true)
	u.setFeedback("Desconectando...")

	go func() {
		output, err := u.Client.Disconnect(selectedProfile)
		time.Sleep(300 * time.Millisecond)

		fyne.Do(func() {
			u.setBusy(false)
			u.setStatus("Desconectado")
			if err != nil {
				u.setFeedback("❌ Error al desconectar")
				dialog.ShowError(fmt.Errorf("Error al desconectar:\n%v", err), u.Window)
			} else {
				u.setFeedback("Desconectado")
				dialog.ShowInformation("VPN Desconectada", fmt.Sprintf("Se ha enviado la señal de desconexión:\n%s", output), u.Window)
			}
			u.Window.Canvas().Focus(u.totpEntry)
		})
	}()
}

// showPasswordModal displays a secure modal dialog to store/update the base corporate password.
func (u *UI) showPasswordModal() {
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Contraseña de Active Directory / Windows")

	infoText := widget.NewLabel("Tu contraseña se guardará encriptada con DPAPI de Windows ligada a tu cuenta de usuario.\n\n⚠️ Ingresa únicamente tu contraseña base corporativa (SIN el código TOTP/MFA).")
	infoText.Wrapping = fyne.TextWrapWord

	formContent := container.NewVBox(
		infoText,
		widget.NewLabelWithStyle("Nueva Contraseña Base:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		passwordEntry,
	)

	d := dialog.NewCustomConfirm(
		"Configurar Contraseña Base",
		"Guardar",
		"Cancelar",
		formContent,
		func(confirmed bool) {
			if !confirmed {
				return
			}
			pass := strings.TrimSpace(passwordEntry.Text)
			if pass == "" {
				dialog.ShowError(fmt.Errorf("La contraseña no puede estar vacía."), u.Window)
				return
			}

			err := crypto.SavePassword(pass)
			if err != nil {
				dialog.ShowError(fmt.Errorf("No se pudo guardar la contraseña:\n%v", err), u.Window)
				u.setFeedback("❌ Error al guardar contraseña")
			} else {
				u.setFeedback("🔑 Contraseña guardada correctamente")
				dialog.ShowInformation("Éxito", "¡Contraseña encriptada y guardada con éxito en la bóveda segura del sistema!", u.Window)
				u.Window.Canvas().Focus(u.totpEntry)
			}
		},
		u.Window,
	)

	d.Resize(fyne.NewSize(300, 240))
	d.Show()
	u.Window.Canvas().Focus(passwordEntry)
}
