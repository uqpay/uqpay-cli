package config_test

import (
	"path/filepath"
	"testing"

	"github.com/uqpay/uqpay-cli/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Env != "sandbox" {
		t.Errorf("default Env = %q, want %q", cfg.Env, "sandbox")
	}
	if cfg.Output != "table" {
		t.Errorf("default Output = %q, want %q", cfg.Output, "table")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	cfg := &config.Config{
		ClientID: "cid",
		APIKey:   "key",
		Env:      "production",
		Output:   "json",
	}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ClientID != "cid" {
		t.Errorf("ClientID = %q, want %q", loaded.ClientID, "cid")
	}
	if loaded.Env != "production" {
		t.Errorf("Env = %q, want %q", loaded.Env, "production")
	}
}

func TestApplyEnvVars(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("UQPAY_CLIENT_ID", "env_cid")
	t.Setenv("UQPAY_API_KEY", "env_key")
	t.Setenv("UQPAY_ENV", "production")
	t.Setenv("UQPAY_OUTPUT", "json")

	cfg, _ := config.Load()
	cfg.ApplyEnvVars()

	if cfg.ClientID != "env_cid" {
		t.Errorf("ClientID = %q, want env_cid", cfg.ClientID)
	}
	if cfg.Output != "json" {
		t.Errorf("Output = %q, want json", cfg.Output)
	}
}

func TestEnvVarOverridesFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	fileCfg := &config.Config{Env: "sandbox", Output: "table", ClientID: "file_cid"}
	_ = fileCfg.Save()

	t.Setenv("UQPAY_ENV", "production")
	t.Setenv("UQPAY_CLIENT_ID", "env_cid")

	loaded, _ := config.Load()
	loaded.ApplyEnvVars()

	if loaded.Env != "production" {
		t.Errorf("Env = %q, want production", loaded.Env)
	}
	if loaded.ClientID != "env_cid" {
		t.Errorf("ClientID = %q, want env_cid", loaded.ClientID)
	}
}

func TestConfigFilePath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	expected := filepath.Join(dir, ".uqpay", "config.yaml")
	if got := config.DefaultPath(); got != expected {
		t.Errorf("DefaultPath = %q, want %q", got, expected)
	}
}

func TestSetValid(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	cfg, _ := config.Load()
	if err := cfg.Set("env", "production"); err != nil {
		t.Fatal(err)
	}
	if cfg.Env != "production" {
		t.Errorf("Env = %q after Set", cfg.Env)
	}
}

func TestSetInvalidEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	cfg, _ := config.Load()
	if err := cfg.Set("env", "staging"); err == nil {
		t.Error("expected error for invalid env value")
	}
}

func TestSetUnknownKey(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	cfg, _ := config.Load()
	if err := cfg.Set("unknown-key", "value"); err == nil {
		t.Error("expected error for unknown key")
	}
}
