package sophos

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Default paths for sccli.exe on Windows.
var defaultSccliPaths = []string{
	`C:\Program Files (x86)\Sophos\Connect\sccli.exe`,
	`C:\Program Files\Sophos\Connect\sccli.exe`,
}

// Client manages CLI interaction with Sophos Connect (sccli.exe).
type Client struct {
	sccliPath string
}

// NewClient initializes a new Sophos Client. If customPath is provided and non-empty,
// it is used; otherwise default paths or PATH search is attempted.
func NewClient(customPath ...string) *Client {
	var path string
	if len(customPath) > 0 && customPath[0] != "" {
		path = customPath[0]
	} else {
		path = findDefaultSccli()
	}

	return &Client{
		sccliPath: path,
	}
}

// findDefaultSccli searches for sccli.exe in standard locations or PATH.
func findDefaultSccli() string {
	for _, p := range defaultSccliPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// Try checking environment variables on Windows
	if progFilesX86 := os.Getenv("ProgramFiles(x86)"); progFilesX86 != "" {
		p := filepath.Join(progFilesX86, "Sophos", "Connect", "sccli.exe")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if progFiles := os.Getenv("ProgramFiles"); progFiles != "" {
		p := filepath.Join(progFiles, "Sophos", "Connect", "sccli.exe")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// Try PATH lookup
	if lp, err := exec.LookPath("sccli.exe"); err == nil {
		return lp
	}

	return defaultSccliPaths[0]
}

// GetSccliPath returns the current path configured for sccli.exe.
func (c *Client) GetSccliPath() string {
	return c.sccliPath
}

// SetSccliPath updates the path for sccli.exe.
func (c *Client) SetSccliPath(path string) {
	c.sccliPath = path
}

// IsInstalled checks if the sccli.exe binary exists on the system.
func (c *Client) IsInstalled() bool {
	if runtime.GOOS != "windows" {
		return true // Allow mock/dev mode on non-windows
	}
	_, err := os.Stat(c.sccliPath)
	return err == nil
}

// ListConnections executes `sccli.exe list` and returns parsed connection names.
func (c *Client) ListConnections() ([]string, error) {
	if runtime.GOOS != "windows" {
		return nil, nil // No live sccli on non-windows
	}

	if !c.IsInstalled() {
		return nil, fmt.Errorf("sccli.exe no encontrado en %s", c.sccliPath)
	}

	cmd := exec.Command(c.sccliPath, "list")
	cmd.SysProcAttr = hideWindowSysProcAttr()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar 'sccli list': %w (%s)", err, stderr.String())
	}

	return ParseConnectionList(stdout.String()), nil
}

// ParseConnectionList parses the output of `sccli.exe list` to extract connection names.
func ParseConnectionList(output string) []string {
	var connections []string
	seen := make(map[string]bool)

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Skip decorative header lines
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "===") {
			continue
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "available connections") ||
			strings.HasPrefix(lower, "connection name") ||
			strings.HasPrefix(lower, "nombre de conexión") {
			continue
		}

		// Handle tabular output: "ConnectionName      Status"
		parts := strings.Fields(line)
		if len(parts) > 0 {
			connName := parts[0]
			connLower := strings.ToLower(connName)
			if !seen[connLower] {
				seen[connLower] = true
				connections = append(connections, connName)
			}
		}
	}

	return connections
}

// Connect executes `sccli.exe enable -n <connName> -u <username> -p <fullPassword>`.
func (c *Client) Connect(connName, username, fullPassword string) (string, error) {
	connName = strings.TrimSpace(connName)
	username = strings.TrimSpace(username)

	if connName == "" {
		return "", fmt.Errorf("el nombre de la conexión no puede estar vacío")
	}
	if username == "" {
		return "", fmt.Errorf("el nombre de usuario no puede estar vacío")
	}
	if fullPassword == "" {
		return "", fmt.Errorf("la contraseña no puede estar vacía")
	}

	if runtime.GOOS != "windows" {
		return fmt.Sprintf("[SIMULADO] Conectado exitosamente a '%s' con usuario '%s'", connName, username), nil
	}

	if !c.IsInstalled() {
		return "", fmt.Errorf("sccli.exe no encontrado en %s", c.sccliPath)
	}

	cmd := exec.Command(c.sccliPath, "enable", "-n", connName, "-u", username, "-p", fullPassword)
	cmd.SysProcAttr = hideWindowSysProcAttr()

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
		if out != "" {
			return "", fmt.Errorf("%s", out)
		}
		return "", fmt.Errorf("error al ejecutar conexión: %v", err)
	}

	if out != "" {
		// Sanitize output to prevent leaking plain text password lines
		var sanitizedLines []string
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(strings.ToLower(line), "password") {
				sanitizedLines = append(sanitizedLines, "Password '********' will be used")
			} else {
				sanitizedLines = append(sanitizedLines, line)
			}
		}
		out = strings.Join(sanitizedLines, "\n")
	} else {
		out = fmt.Sprintf("Comando de conexión enviado para '%s'.", connName)
	}
	return out, nil
}

// Disconnect executes `sccli.exe disable -n <connName>`.
func (c *Client) Disconnect(connName string) (string, error) {
	connName = strings.TrimSpace(connName)
	if connName == "" {
		return "", fmt.Errorf("el nombre de la conexión no puede estar vacío")
	}

	if runtime.GOOS != "windows" {
		return fmt.Sprintf("[SIMULADO] Desconectado exitosamente de '%s'", connName), nil
	}

	if !c.IsInstalled() {
		return "", fmt.Errorf("sccli.exe no encontrado en %s", c.sccliPath)
	}

	cmd := exec.Command(c.sccliPath, "disable", "-n", connName)
	cmd.SysProcAttr = hideWindowSysProcAttr()

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
		out = fmt.Sprintf("Desconexión solicitada para '%s'.", connName)
	}
	return out, nil
}

// CheckStatus verifies if the given connection is active.
func (c *Client) CheckStatus(connName string) (bool, string, error) {
	connName = strings.TrimSpace(connName)
	if connName == "" {
		return false, "Desconectado", nil
	}

	if runtime.GOOS != "windows" {
		return false, "Desconectado", nil
	}

	if !c.IsInstalled() {
		return false, "Sophos no detectado", nil
	}

	cmd := exec.Command(c.sccliPath, "status", "-n", connName)
	cmd.SysProcAttr = hideWindowSysProcAttr()

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	_ = cmd.Run()

	out := strings.ToLower(stdout.String())
	if strings.Contains(out, "connected") || strings.Contains(out, "conectado") || strings.Contains(out, "established") {
		return true, "Conectado", nil
	}

	return false, "Desconectado", nil
}
