package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigValidatesRequiredFields(t *testing.T) {
	repoRoot := t.TempDir()
	configDir := filepath.Join(repoRoot, ".ai")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	content := "ai_config_repo: ../ai-governance\nversion: v0.1.0\nagent: backend-agent\noverrides: []\n"
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	config, err := LoadConfig(repoRoot)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if config.Agent != "backend-agent" {
		t.Fatalf("unexpected agent: %s", config.Agent)
	}
}

func TestLoadAgentManifestMatchesByName(t *testing.T) {
	governanceDir := t.TempDir()
	agentsDir := filepath.Join(governanceDir, "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir agents dir: %v", err)
	}

	manifest := "name: backend-agent\nincludes:\n  - policies/coding.md\n"
	if err := os.WriteFile(filepath.Join(agentsDir, "backend.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	path, loaded, err := LoadAgentManifest(governanceDir, "backend-agent")
	if err != nil {
		t.Fatalf("LoadAgentManifest returned error: %v", err)
	}

	if !strings.HasSuffix(path, filepath.Join("agents", "backend.yaml")) {
		t.Fatalf("unexpected path: %s", path)
	}
	if len(loaded.Includes) != 1 || loaded.Includes[0] != "policies/coding.md" {
		t.Fatalf("unexpected includes: %#v", loaded.Includes)
	}
}

func TestResolveOverrideFilesRejectsMissingFiles(t *testing.T) {
	_, err := resolveOverrideFiles(t.TempDir(), []string{"missing.md"})
	if err == nil {
		t.Fatal("expected missing override to fail validation")
	}
}
