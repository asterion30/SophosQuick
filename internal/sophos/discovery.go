package sophos

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Default paths where Sophos Connect configuration and connection profiles reside.
const (
	DefaultWindowsConnectionsDir = `C:\ProgramData\Sophos\Connect\connections`
	DefaultUnixConnectionsDir    = `/etc/sophos/connect/connections`
)

// Supported connection file extensions used by Sophos Connect.
var supportedExtensions = map[string]bool{
	".tgb":  true,
	".ovpn": true,
	".apx":  true,
	".conf": true,
	".json": true,
}

// GetDefaultConnectionsDir returns the default system directory where connection profiles are stored.
func GetDefaultConnectionsDir() string {
	if runtime.GOOS == "windows" {
		programData := os.Getenv("ProgramData")
		if programData != "" {
			return filepath.Join(programData, "Sophos", "Connect", "connections")
		}
		return DefaultWindowsConnectionsDir
	}
	return DefaultUnixConnectionsDir
}

// ScanConnectionsDir inspects the specified directory and extracts connection names
// from configuration files (.tgb, .ovpn, .apx, .conf, .json).
func ScanConnectionsDir(dirPath string) ([]string, error) {
	if dirPath == "" {
		dirPath = GetDefaultConnectionsDir()
	}

	info, err := os.Stat(dirPath)
	if err != nil || !info.IsDir() {
		return nil, err
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var found []string
	seen := make(map[string]bool)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))

		if supportedExtensions[ext] {
			baseName := strings.TrimSuffix(name, filepath.Ext(name))
			baseName = strings.TrimSpace(baseName)
			if baseName != "" && !seen[strings.ToLower(baseName)] {
				seen[strings.ToLower(baseName)] = true
				found = append(found, baseName)
			}
		}
	}

	return found, nil
}

// DiscoverConnections aggregates connection names from:
// 1. Sophos CLI (`sccli.exe list`) if client is available.
// 2. Directory scan of `%ProgramData%\Sophos\Connect\connections\`.
// 3. Fallback / configured list from application configuration.
//
// Results are deduplicated while preserving order of discovery.
func DiscoverConnections(client *Client, fallbackConnections []string, customDir ...string) ([]string, error) {
	var result []string
	seen := make(map[string]bool)

	addName := func(name string) {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			return
		}
		lower := strings.ToLower(trimmed)
		if !seen[lower] {
			seen[lower] = true
			result = append(result, trimmed)
		}
	}

	// 1. Try querying sccli CLI
	if client != nil {
		if cliConns, err := client.ListConnections(); err == nil {
			for _, conn := range cliConns {
				addName(conn)
			}
		}
	}

	// 2. Scan connections directory on disk
	var dirToScan string
	if len(customDir) > 0 && customDir[0] != "" {
		dirToScan = customDir[0]
	} else {
		dirToScan = GetDefaultConnectionsDir()
	}

	if dirConns, err := ScanConnectionsDir(dirToScan); err == nil {
		for _, conn := range dirConns {
			addName(conn)
		}
	}

	// 3. Merge fallback configured connections
	for _, conn := range fallbackConnections {
		addName(conn)
	}

	return result, nil
}
