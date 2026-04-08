package uninstall

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sipeed/trenchlaw/pkg/config"
)

func TestRunRemovesLocalState(t *testing.T) {
	originalInput := uninstallInput
	originalOutput := uninstallOutput
	t.Cleanup(func() {
		uninstallInput = originalInput
		uninstallOutput = originalOutput
	})

	tmpDir := t.TempDir()
	userHome := filepath.Join(tmpDir, "user")
	trenchlawHome := filepath.Join(tmpDir, ".trenchlaw")
	sshKeyPath := filepath.Join(userHome, ".ssh", "trenchlaw_ed25519.key")

	t.Setenv("HOME", userHome)
	t.Setenv(config.EnvHome, trenchlawHome)

	writeTestFile(t, filepath.Join(trenchlawHome, "config.json"))
	writeTestFile(t, filepath.Join(trenchlawHome, "launcher-config.json"))
	writeTestFile(t, filepath.Join(trenchlawHome, "workspace", "AGENT.md"))
	writeTestFile(t, sshKeyPath)
	writeTestFile(t, sshKeyPath+".pub")

	var output bytes.Buffer
	uninstallOutput = &output

	if err := run(options{yes: true}); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	assertMissing(t, trenchlawHome)
	assertMissing(t, sshKeyPath)
	assertMissing(t, sshKeyPath+".pub")

	if !strings.Contains(output.String(), "trenchclaw onboard") {
		t.Fatalf("run() output = %q, want install guidance", output.String())
	}
}

func TestRunCanceledKeepsState(t *testing.T) {
	originalInput := uninstallInput
	originalOutput := uninstallOutput
	t.Cleanup(func() {
		uninstallInput = originalInput
		uninstallOutput = originalOutput
	})

	tmpDir := t.TempDir()
	userHome := filepath.Join(tmpDir, "user")
	trenchlawHome := filepath.Join(tmpDir, ".trenchlaw")
	configPath := filepath.Join(trenchlawHome, "config.json")

	t.Setenv("HOME", userHome)
	t.Setenv(config.EnvHome, trenchlawHome)

	writeTestFile(t, configPath)

	uninstallInput = strings.NewReader("n\n")
	var output bytes.Buffer
	uninstallOutput = &output

	if err := run(options{}); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected config to remain after cancel, stat error = %v", err)
	}
	if !strings.Contains(output.String(), "uninstall canceled") {
		t.Fatalf("run() output = %q, want cancel message", output.String())
	}
}

func TestRunRemovesExternalConfigPath(t *testing.T) {
	originalInput := uninstallInput
	originalOutput := uninstallOutput
	t.Cleanup(func() {
		uninstallInput = originalInput
		uninstallOutput = originalOutput
	})

	tmpDir := t.TempDir()
	userHome := filepath.Join(tmpDir, "user")
	trenchlawHome := filepath.Join(tmpDir, ".trenchlaw")
	externalConfigDir := filepath.Join(tmpDir, "custom-config")
	configPath := filepath.Join(externalConfigDir, "config.json")
	launcherConfigPath := filepath.Join(externalConfigDir, "launcher-config.json")

	t.Setenv("HOME", userHome)
	t.Setenv(config.EnvHome, trenchlawHome)
	t.Setenv(config.EnvConfig, configPath)

	writeTestFile(t, configPath)
	writeTestFile(t, launcherConfigPath)
	writeTestFile(t, filepath.Join(trenchlawHome, "workspace", "AGENT.md"))

	uninstallOutput = &bytes.Buffer{}

	if err := run(options{yes: true}); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	assertMissing(t, trenchlawHome)
	assertMissing(t, configPath)
	assertMissing(t, launcherConfigPath)
}

func writeTestFile(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %q to be removed, stat err = %v", path, err)
	}
}
