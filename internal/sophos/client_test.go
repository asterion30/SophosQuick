package sophos

import (
	"testing"
)

func TestParseConnectionList(t *testing.T) {
	sampleOutput := `
Available connections:
----------------------
vpn.vittal.com.ar
vpn_contingencia.vittal.com.ar
office.company.org

`
	expected := []string{
		"vpn.vittal.com.ar",
		"vpn_contingencia.vittal.com.ar",
		"office.company.org",
	}

	result := ParseConnectionList(sampleOutput)
	if len(result) != len(expected) {
		t.Fatalf("Length mismatch: got %d (%v), want %d (%v)", len(result), result, len(expected), expected)
	}

	for i, name := range expected {
		if result[i] != name {
			t.Errorf("Index %d mismatch: got %q, want %q", i, result[i], name)
		}
	}
}

func TestParseConnectionListTabular(t *testing.T) {
	sampleOutput := `
Connection Name                     Status
==============================================
vpn.vittal.com.ar                   Connected
vpn_contingencia.vittal.com.ar      Disconnected
`
	expected := []string{
		"vpn.vittal.com.ar",
		"vpn_contingencia.vittal.com.ar",
	}

	result := ParseConnectionList(sampleOutput)
	if len(result) != len(expected) {
		t.Fatalf("Length mismatch: got %d (%v), want %d (%v)", len(result), result, len(expected), expected)
	}

	for i, name := range expected {
		if result[i] != name {
			t.Errorf("Index %d mismatch: got %q, want %q", i, result[i], name)
		}
	}
}

func TestParseConnectionListDeduplication(t *testing.T) {
	sampleOutput := `
vpn.vittal.com.ar
VPN.VITTAL.COM.AR
vpn.vittal.com.ar
vpn_backup.domain.com
`
	result := ParseConnectionList(sampleOutput)
	if len(result) != 2 {
		t.Fatalf("Expected 2 deduplicated connections, got %d (%v)", len(result), result)
	}
	if result[0] != "vpn.vittal.com.ar" {
		t.Errorf("Expected first connection to be 'vpn.vittal.com.ar', got %q", result[0])
	}
	if result[1] != "vpn_backup.domain.com" {
		t.Errorf("Expected second connection to be 'vpn_backup.domain.com', got %q", result[1])
	}
}

func TestClientValidation(t *testing.T) {
	client := NewClient("non_existent_sccli_mock")

	// Empty connection name
	_, err := client.Connect("", "user", "pass")
	if err == nil {
		t.Errorf("Expected error for empty connection name, got nil")
	}

	// Empty username
	_, err = client.Connect("vpn.test", "", "pass")
	if err == nil {
		t.Errorf("Expected error for empty username, got nil")
	}

	// Empty password
	_, err = client.Connect("vpn.test", "user", "")
	if err == nil {
		t.Errorf("Expected error for empty password, got nil")
	}

	// Empty disconnect name
	_, err = client.Disconnect("")
	if err == nil {
		t.Errorf("Expected error for empty disconnect name, got nil")
	}
}

func TestClientGetSetPath(t *testing.T) {
	client := NewClient("/custom/path/sccli")
	if client.GetSccliPath() != "/custom/path/sccli" {
		t.Errorf("Expected path '/custom/path/sccli', got %q", client.GetSccliPath())
	}

	client.SetSccliPath("/updated/path/sccli")
	if client.GetSccliPath() != "/updated/path/sccli" {
		t.Errorf("Expected path '/updated/path/sccli', got %q", client.GetSccliPath())
	}
}
