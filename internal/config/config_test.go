package config

import (
	"os"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	content := `{"port": 8080, "backends": ["http://localhost:1234", "http://localhost:5678"]}`
	tmpFile := "test_config.json"
	os.WriteFile(tmpFile, []byte(content), 0644)
	defer os.Remove(tmpFile)

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}

	if len(cfg.Backends) != 2 {
		t.Errorf("expected 2 backends, got %d", len(cfg.Backends))
	}
}

func TestLoad_InvalidPath(t *testing.T) {
	_, err := Load("nonexistent.json")
	if err == nil {
		t.Fatalf("expected error for invalid file path")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	content := `{"port": 8080,`
	tmpFile := "invalid_config.json"
	os.WriteFile(tmpFile, []byte(content), 0644)
	defer os.Remove(tmpFile)

	_, err := Load(tmpFile)
	if err == nil {
		t.Fatalf("expected error for invalid JSON")
	}
}
