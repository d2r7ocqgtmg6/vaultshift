package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_MissingRequiredFields(t *testing.T) {
	_, err := Load("")
	if err == nil {
		t.Fatal("expected error for missing required fields, got nil")
	}
}

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("VAULTSHIFT_SOURCE_ADDR", "http://source:8200")
	t.Setenv("VAULTSHIFT_SOURCE_TOKEN", "s.sourcetoken")
	t.Setenv("VAULTSHIFT_DEST_ADDR", "http://dest:8200")
	t.Setenv("VAULTSHIFT_DEST_TOKEN", "s.desttoken")
	t.Setenv("VAULTSHIFT_DRY_RUN", "true")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceAddr != "http://source:8200" {
		t.Errorf("expected source_addr %q, got %q", "http://source:8200", cfg.SourceAddr)
	}
	if !cfg.DryRun {
		t.Error("expected dry_run to be true")
	}
	if cfg.AuditLogPath != "vaultshift-audit.log" {
		t.Errorf("expected default audit_log_path, got %q", cfg.AuditLogPath)
	}
}

func TestLoad_FromFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := []byte(`
source_addr: "http://vault-src:8200"
source_token: "s.abc123"
source_namespace: "team-a"
dest_addr: "http://vault-dst:8200"
dest_token: "s.xyz789"
dest_namespace: "team-b"
dry_run: false
audit_log_path: "/tmp/audit.log"
`)
	if err := os.WriteFile(cfgPath, content, 0600); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceNamespace != "team-a" {
		t.Errorf("expected source_namespace %q, got %q", "team-a", cfg.SourceNamespace)
	}
	if cfg.AuditLogPath != "/tmp/audit.log" {
		t.Errorf("expected audit_log_path %q, got %q", "/tmp/audit.log", cfg.AuditLogPath)
	}
}

func TestValidate_MissingDestToken(t *testing.T) {
	cfg := &Config{
		SourceAddr:  "http://src:8200",
		SourceToken: "s.token",
		DestAddr:    "http://dst:8200",
	}
	if err := cfg.validate(); err == nil {
		t.Error("expected validation error for missing dest_token")
	}
}
