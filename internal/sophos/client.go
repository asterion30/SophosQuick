package sophos

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"sophosquick/internal/crypto"
)

// Client handles interaction with Sophos Connect CLI (sccli.exe).
type Client struct {
	SccliPath string
	Username  string
}

// NewClient detects the environment and initializes a Sophos client.
func NewClient() *Client {
	username := os.Getenv("USERNAME")
	if username == "" {
		username = os.Getenv("USER")
	}
	if username == "" {
		username = "user"
	}

	sccli := findSccli()
	return &Client{
		SccliPath: sccli,
		Username:  username,
	}
}

// findSccli searches common installation paths for sccli.exe.
func findSccli() string {
	candidates := []string{
		`C:\Program Files (x86)\Sophos\Connect\sccli.exe`,
		`C:\Program Files\Sophos\Connect\sccli.exe`,
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Sophos", "Connect", "sccli.exe"),
		filepath.Join(os.Getenv("ProgramFiles"), "Sophos", "Connect", "sccli.exe"),
	}

	for _, p := range candidates {
		if p != "" {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}

	if lp, err := exec.LookPath("sccli.exe"); err == nil {
		return lp
	}

	return `C:\Program Files (x86)\Sophos\Connect\sccli.exe`
}

// IsInstalled returns true if sccli executable exists on the system.
func (c *Client) IsInstalled() bool {
	if runtime.GOOS != "windows" {
		return true // Simulated mode in development
	}
	_, err := os.Stat(c.SccliPath)
	return err == nil
}

// DiscoverConnections scans known Sophos profile directories and returns connection names.
func (c *Client) DiscoverConnections(defaults []string) []string {
	found := make(map[string]bool)
	var list []string

	// Check Sophos profiles directories
	dirs := []string{
		`C:\Program Files (x86)\Sophos\Connect\connections`,
		`C:\Program Files\Sophos\Connect\connections`,
		`C:\ProgramData\Sophos\Connect\connections`,
		filepath.Join(os.Getenv("ProgramData"), "Sophos", "Connect", "connections"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Sophos", "Connect", "connections"),
	}

	for _, d := range dirs {
		if d == "" {
			continue
		}
		entries, err := os.ReadDir(d)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				name := entry.Name()
				if !found[name] {
					found[name] = true
					list = append(list, name)
				}
			} else {
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if ext == ".apx" || ext == ".tgb" || ext == ".ovpn" || ext == ".pro" {
					name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
					if !found[name] {
						found[name] = true
						list = append(list, name)
					}
				}
			}
		}
	}

	// Always ensure default corporate profiles exist
	for _, d := range defaults {
		if !found[d] {
			found[d] = true
			list = append(list, d)
		}
	}

	return list
}

// Connect sends the credentials and TOTP to Sophos Connect.
func (c *Client) Connect(connectionName string, totpCode string) (string, error) {
	totpCode = strings.TrimSpace(totpCode)
	if totpCode == "" {
		return "", fmt.Errorf("debes ingresar el código TOTP / MFA")
	}

	basePass, err := crypto.LoadPassword()
	if err != nil {
		return "", fmt.Errorf("no se encontró la contraseña base: %w\nPor favor configúrala primero con '🔑 Configurar Contraseña Base'", err)
	}

	combinedPassword := basePass + totpCode

	if runtime.GOOS != "windows" {
		// Mock execution on non-Windows
		return fmt.Sprintf("[SIMULADO] Conectando a '%s' con usuario '%s' (TOTP: %s)", connectionName, c.Username, totpCode), nil
	}

	if !c.IsInstalled() {
		return "", fmt.Errorf("no se encontró sccli.exe en '%s'. Verifica que Sophos Connect esté instalado", c.SccliPath)
	}

	cmd := exec.Command(c.SccliPath, "enable", "-n", connectionName, "-u", c.Username, "-p", combinedPassword)
	configureHiddenWindow(cmd)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	out := strings.TrimSpace(stdout.String())
	errOut := strings.TrimSpace(stderr.String())

	if err != nil {
		if errOut != "" {
			return "", fmt.Errorf("%s (error: %v)", errOut, err)
		}
		if out != "" {
			return "", fmt.Errorf("%s", out)
		}
		return "", fmt.Errorf("error al ejecutar conexión: %v", err)
	}

	if out == "" {
		out = fmt.Sprintf("Comando de conexión enviado para '%s'.", connectionName)
	}
	return out, nil
}

// Disconnect sends the disable signal for the specified connection.
func (c *Client) Disconnect(connectionName string) (string, error) {
	if runtime.GOOS != "windows" {
		return fmt.Sprintf("[SIMULADO] Desconectado de '%s'", connectionName), nil
	}

	if !c.IsInstalled() {
		return "", fmt.Errorf("no se encontró sccli.exe en '%s'", c.SccliPath)
	}

	cmd := exec.Command(c.SccliPath, "disable", "-n", connectionName)
	configureHiddenWindow(cmd)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	out := strings.TrimSpace(stdout.String())
	errOut := strings.TrimSpace(stderr.String())

	if err != nil {
		if errOut != "" {
			return "", fmt.Errorf("%s (error: %v)", errOut, err)
		}
		return "", fmt.Errorf("error al desconectar: %v", err)
	}

	if out == "" {
		out = fmt.Sprintf("Desconexión solicitada para '%s'.", connectionName)
	}
	return out, nil
}

// CheckStatus verifies if the given connection is active.
func (c *Client) CheckStatus(connectionName string) (bool, string, error) {
	if runtime.GOOS != "windows" {
		return false, "Desconectado", nil
	}

	if !c.IsInstalled() {
		return false, "Sophos no detectado", nil
	}

	cmd := exec.Command(c.SccliPath, "status", "-n", connectionName)
	configureHiddenWindow(cmd)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	_ = cmd.Run()

	out := strings.ToLower(stdout.String())
	if strings.Contains(out, "connected") || strings.Contains(out, "conectado") || strings.Contains(out, "established") {
		return true, "Conectado", nil
	}

	return false, "Desconectado", nil
}
