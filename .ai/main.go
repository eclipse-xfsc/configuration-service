package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const configPath = ".ai/config.yaml"
const syncStatePath = ".ai/context-sync.json"

type Config struct {
	AIConfigRepo string   `yaml:"ai_config_repo"`
	Version      string   `yaml:"version"`
	Agent        string   `yaml:"agent"`
	Overrides    []string `yaml:"overrides"`
}

type AgentManifest struct {
	Name     string   `yaml:"name"`
	Includes []string `yaml:"includes"`
}

type LoadedContext struct {
	Config        Config
	GovernanceDir string
	AgentFile     string
	SharedFiles   []string
	OverrideFiles []string
	Content       string
}

type SyncState struct {
	LastValidatedAt    string `json:"last_validated_at"`
	Agent              string `json:"agent"`
	Version            string `json:"version"`
	GovernanceRevision string `json:"governance_revision"`
	ConfigFingerprint  string `json:"config_fingerprint"`
}

type SyncSnapshot struct {
	Config             Config
	GovernanceRevision string
	ConfigFingerprint  string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("ai-context", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	repoRoot := fs.String("repo-root", ".", "path to the service repository root")
	syncStateFile := fs.String("sync-state-file", syncStatePath, "path to timestamp state file, relative to repo root")

	if err := fs.Parse(args); err != nil {
		return err
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		return usageError()
	}

	command := remaining[0]
	switch command {
	case "load":
		loaded, err := Load(*repoRoot)
		if err != nil {
			return err
		}
		fmt.Print(loaded.Content)
		return nil
	case "validate":
		if err := Validate(*repoRoot); err != nil {
			return err
		}
		if err := recordSuccessfulSync(*repoRoot, *syncStateFile); err != nil {
			return err
		}
		fmt.Println("AI context configuration is valid.")
		return nil
	case "ensure-sync":
		wasChecked, reason, err := EnsureSync(*repoRoot, *syncStateFile)
		if err != nil {
			return err
		}
		if wasChecked {
			fmt.Printf("AI context sync refreshed (%s).\n", reason)
		} else {
			fmt.Printf("AI context sync is fresh (%s).\n", reason)
		}
		return nil
	default:
		return usageError()
	}
}

func usageError() error {
	return fmt.Errorf("usage: go -C ./.ai run . [load|validate|ensure-sync]")
}

func EnsureSync(repoRoot, syncStateRelativePath string) (bool, string, error) {
	snapshot, err := resolveSyncSnapshot(repoRoot)
	if err != nil {
		return false, "", err
	}

	statePath := filepath.Join(repoRoot, filepath.FromSlash(syncStateRelativePath))
	stale, reason, err := isSyncStateOutdated(statePath, snapshot)
	if err != nil {
		return false, "", err
	}

	if !stale {
		return false, reason, nil
	}

	if err := Validate(repoRoot); err != nil {
		return false, "", fmt.Errorf("ai context stale (%s), validation failed: %w", reason, err)
	}

	if err := recordSuccessfulSync(repoRoot, syncStateRelativePath); err != nil {
		return false, "", err
	}

	return true, reason, nil
}

func isSyncStateOutdated(statePath string, snapshot SyncSnapshot) (bool, string, error) {
	state, err := readSyncState(statePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, "no sync state file present", nil
		}
		return true, "state file is invalid", nil
	}

	if strings.TrimSpace(state.GovernanceRevision) == "" {
		return true, "state file missing governance_revision", nil
	}

	if state.GovernanceRevision != snapshot.GovernanceRevision {
		return true, "governance revision changed", nil
	}

	if strings.TrimSpace(state.ConfigFingerprint) == "" {
		return true, "state file missing config_fingerprint", nil
	}

	if state.ConfigFingerprint != snapshot.ConfigFingerprint {
		return true, "ai config changed", nil
	}

	if strings.TrimSpace(state.LastValidatedAt) == "" {
		return false, "governance and config unchanged", nil
	}

	return false, fmt.Sprintf("governance and config unchanged since %s", state.LastValidatedAt), nil
}

func readSyncState(statePath string) (SyncState, error) {
	data, err := os.ReadFile(statePath)
	if err != nil {
		return SyncState{}, err
	}

	var state SyncState
	if err := json.Unmarshal(data, &state); err != nil {
		return SyncState{}, fmt.Errorf("parse %s: %w", statePath, err)
	}

	return state, nil
}

func recordSuccessfulSync(repoRoot, syncStateRelativePath string) error {
	snapshot, err := resolveSyncSnapshot(repoRoot)
	if err != nil {
		return err
	}

	statePath := filepath.Join(repoRoot, filepath.FromSlash(syncStateRelativePath))
	if err := os.MkdirAll(filepath.Dir(statePath), 0o755); err != nil {
		return fmt.Errorf("create sync state dir: %w", err)
	}

	state := SyncState{
		LastValidatedAt:    time.Now().UTC().Format(time.RFC3339),
		Agent:              snapshot.Config.Agent,
		Version:            snapshot.Config.Version,
		GovernanceRevision: snapshot.GovernanceRevision,
		ConfigFingerprint:  snapshot.ConfigFingerprint,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal sync state: %w", err)
	}

	data = append(data, '\n')
	if err := os.WriteFile(statePath, data, 0o644); err != nil {
		return fmt.Errorf("write sync state: %w", err)
	}

	return nil
}

func resolveSyncSnapshot(repoRoot string) (SyncSnapshot, error) {
	config, err := LoadConfig(repoRoot)
	if err != nil {
		return SyncSnapshot{}, err
	}

	governanceDir, err := ResolveGovernanceRepo(repoRoot, config)
	if err != nil {
		return SyncSnapshot{}, err
	}

	revision, err := gitRevision(governanceDir)
	if err != nil {
		return SyncSnapshot{}, err
	}

	return SyncSnapshot{
		Config:             config,
		GovernanceRevision: revision,
		ConfigFingerprint:  configFingerprint(config),
	}, nil
}

func gitRevision(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD: %w\n%s", err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

func configFingerprint(config Config) string {
	h := sha1.New()
	h.Write([]byte(config.AIConfigRepo))
	h.Write([]byte("\n"))
	h.Write([]byte(config.Version))
	h.Write([]byte("\n"))
	h.Write([]byte(config.Agent))
	for _, override := range config.Overrides {
		h.Write([]byte("\n"))
		h.Write([]byte(override))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Version returns the loader version
func Version() string {
	return "1.0.0"
}

// Load resolves governance policies and returns merged AI context
func Load(repoRoot string) (*LoadedContext, error) {
	config, err := LoadConfig(repoRoot)
	if err != nil {
		return nil, err
	}

	governanceDir, err := ResolveGovernanceRepo(repoRoot, config)
	if err != nil {
		return nil, err
	}

	agentFile, manifest, err := LoadAgentManifest(governanceDir, config.Agent)
	if err != nil {
		return nil, err
	}

	sharedFiles, err := resolveIncludedFiles(governanceDir, manifest.Includes)
	if err != nil {
		return nil, err
	}

	overrideFiles, err := resolveOverrideFiles(repoRoot, config.Overrides)
	if err != nil {
		return nil, err
	}

	content, err := buildMergedContent(config, governanceDir, agentFile, sharedFiles, overrideFiles)
	if err != nil {
		return nil, err
	}

	return &LoadedContext{
		Config:        config,
		GovernanceDir: governanceDir,
		AgentFile:     agentFile,
		SharedFiles:   sharedFiles,
		OverrideFiles: overrideFiles,
		Content:       content,
	}, nil
}

// Validate checks that governance configuration is valid
func Validate(repoRoot string) error {
	_, err := Load(repoRoot)
	return err
}

// LoadConfig reads and validates .ai/config.yaml
func LoadConfig(repoRoot string) (Config, error) {
	configFile := filepath.Join(repoRoot, configPath)
	data, err := os.ReadFile(configFile)
	if err != nil {
		return Config{}, fmt.Errorf("read %s: %w", configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", configPath, err)
	}

	config.AIConfigRepo = strings.TrimSpace(config.AIConfigRepo)
	config.Version = strings.TrimSpace(config.Version)
	config.Agent = strings.TrimSpace(config.Agent)

	if config.AIConfigRepo == "" {
		return Config{}, errors.New("ai_config_repo is required")
	}
	if config.Version == "" {
		return Config{}, errors.New("version is required")
	}
	if config.Agent == "" {
		return Config{}, errors.New("agent is required")
	}

	return config, nil
}

// ResolveGovernanceRepo resolves the governance repo from local or remote, caches at the configured version
func ResolveGovernanceRepo(repoRoot string, config Config) (string, error) {
	sourceRef := resolveSourceRef(repoRoot, config.AIConfigRepo)

	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolve user cache dir: %w", err)
	}

	repoHash := sha1.Sum([]byte(config.AIConfigRepo + ":" + config.Version))
	cacheDir := filepath.Join(cacheRoot, "ai-governance-context", hex.EncodeToString(repoHash[:8]))

	if _, err := os.Stat(cacheDir); err == nil {
		if err := runGit(repoRoot, cacheDir, "fetch", "--tags", "--prune", sourceRef); err != nil {
			return "", err
		}
	} else if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(cacheDir), 0o755); err != nil {
			return "", fmt.Errorf("create cache dir: %w", err)
		}
		if err := runGit(repoRoot, "", "clone", sourceRef, cacheDir); err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("inspect cache dir: %w", err)
	}

	if err := runGit(repoRoot, cacheDir, "checkout", "--force", config.Version); err != nil {
		return "", err
	}

	return cacheDir, nil
}

func resolveSourceRef(repoRoot, repoRef string) string {
	localRepo := repoRef
	if !filepath.IsAbs(localRepo) {
		candidate := filepath.Join(repoRoot, filepath.FromSlash(localRepo))
		if stat, err := os.Stat(candidate); err == nil && stat.IsDir() {
			return candidate
		}
	}
	if stat, err := os.Stat(localRepo); err == nil && stat.IsDir() {
		return localRepo
	}

	repoName := repoNameFromRef(repoRef)
	if repoName != "" {
		sibling := filepath.Join(filepath.Dir(repoRoot), repoName)
		if stat, err := os.Stat(sibling); err == nil && stat.IsDir() {
			return sibling
		}
	}

	return repoRef
}

func repoNameFromRef(repoRef string) string {
	normalized := strings.ReplaceAll(repoRef, "\\", "/")
	lastSlash := strings.LastIndex(normalized, "/")
	lastColon := strings.LastIndex(normalized, ":")
	cut := lastSlash
	if lastColon > cut {
		cut = lastColon
	}
	name := normalized
	if cut >= 0 && cut < len(normalized)-1 {
		name = normalized[cut+1:]
	}
	name = strings.TrimSuffix(name, ".git")
	return strings.TrimSpace(name)
}

// LoadAgentManifest loads the agent manifest and validates it
func LoadAgentManifest(governanceDir, agentName string) (string, AgentManifest, error) {
	agentsDir := filepath.Join(governanceDir, "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return "", AgentManifest{}, fmt.Errorf("read agents dir: %w", err)
	}

	var candidates []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		path := filepath.Join(agentsDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return "", AgentManifest{}, fmt.Errorf("read agent manifest %s: %w", entry.Name(), err)
		}

		var manifest AgentManifest
		if err := yaml.Unmarshal(data, &manifest); err != nil {
			return "", AgentManifest{}, fmt.Errorf("parse agent manifest %s: %w", entry.Name(), err)
		}
		if strings.TrimSpace(manifest.Name) == agentName {
			if len(manifest.Includes) == 0 {
				return "", AgentManifest{}, fmt.Errorf("agent %q does not declare includes", agentName)
			}
			return path, manifest, nil
		}
		candidates = append(candidates, manifest.Name)
	}

	sort.Strings(candidates)
	return "", AgentManifest{}, fmt.Errorf("agent %q not found in governance repo; available agents: %s", agentName, strings.Join(candidates, ", "))
}

func resolveIncludedFiles(governanceDir string, includes []string) ([]string, error) {
	files := make([]string, 0, len(includes))
	for _, include := range includes {
		path := filepath.Join(governanceDir, filepath.FromSlash(include))
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("included file %q: %w", include, err)
		}
		if info.IsDir() {
			return nil, fmt.Errorf("included path %q is a directory", include)
		}
		files = append(files, path)
	}
	return files, nil
}

func resolveOverrideFiles(repoRoot string, overrides []string) ([]string, error) {
	files := make([]string, 0, len(overrides))
	for _, override := range overrides {
		path := filepath.Join(repoRoot, filepath.FromSlash(override))
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("override file %q: %w", override, err)
		}
		if info.IsDir() {
			return nil, fmt.Errorf("override path %q is a directory", override)
		}
		files = append(files, path)
	}
	return files, nil
}

func buildMergedContent(config Config, governanceDir, agentFile string, sharedFiles, overrideFiles []string) (string, error) {
	var buffer bytes.Buffer

	buffer.WriteString("# AI Context\n\n")
	buffer.WriteString(fmt.Sprintf("- Agent: %s\n", config.Agent))
	buffer.WriteString(fmt.Sprintf("- Version: %s\n", config.Version))
	buffer.WriteString(fmt.Sprintf("- Governance repo: %s\n", config.AIConfigRepo))
	buffer.WriteString(fmt.Sprintf("- Resolved governance dir: %s\n", governanceDir))
	buffer.WriteString(fmt.Sprintf("- Agent manifest: %s\n", agentFile))

	for _, path := range sharedFiles {
		if err := appendFile(&buffer, "Shared policy", path); err != nil {
			return "", err
		}
	}

	for _, path := range overrideFiles {
		if err := appendFile(&buffer, "Local override", path); err != nil {
			return "", err
		}
	}

	return buffer.String(), nil
}

func appendFile(buffer *bytes.Buffer, sectionTitle, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	buffer.WriteString("\n")
	buffer.WriteString(fmt.Sprintf("## %s: %s\n\n", sectionTitle, path))
	buffer.Write(content)
	if len(content) == 0 || content[len(content)-1] != '\n' {
		buffer.WriteByte('\n')
	}

	return nil
}

func runGit(repoRoot, dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	} else {
		cmd.Dir = repoRoot
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		trimmed := strings.TrimSpace(string(output))
		if trimmed == "" {
			trimmed = err.Error()
		}
		return fmt.Errorf("git %s failed: %s", strings.Join(args, " "), trimmed)
	}

	return nil
}
