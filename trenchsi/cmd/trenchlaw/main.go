// Trenchsi - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 Trenchsi contributors

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/sipeed/trenchlaw/cmd/trenchlaw/internal"
	"github.com/sipeed/trenchlaw/cmd/trenchlaw/internal/agent"
	"github.com/sipeed/trenchlaw/cmd/trenchlaw/internal/auth"
	"github.com/sipeed/trenchlaw/cmd/trenchlaw/internal/cron"
	"github.com/sipeed/trenchlaw/cmd/trenchlaw/internal/gateway"
	"github.com/sipeed/trenchlaw/cmd/trenchlaw/internal/migrate"
	"github.com/sipeed/trenchlaw/cmd/trenchlaw/internal/model"
	"github.com/sipeed/trenchlaw/cmd/trenchlaw/internal/onboard"
	"github.com/sipeed/trenchlaw/cmd/trenchlaw/internal/skills"
	"github.com/sipeed/trenchlaw/cmd/trenchlaw/internal/status"
	"github.com/sipeed/trenchlaw/cmd/trenchlaw/internal/uninstall"
	"github.com/sipeed/trenchlaw/cmd/trenchlaw/internal/version"
	"github.com/sipeed/trenchlaw/pkg/config"
)

var runDefaultWebCommand = runWebCommand
var runDefaultAgentCommand = runAgentChatCommand
var runDefaultTUICommand = runTUICommand
var runDefaultOnboardCommand = func() error {
	onboard.Run(false)
	return nil
}
var startupOnboardComplete = func() bool {
	return onboard.IsComplete()
}
var startupConfigExists = func() bool {
	_, err := os.Stat(internal.GetConfigPath())
	return err == nil
}
var defaultCommandOutput io.Writer = os.Stdout
var defaultCommandIsInteractive = isDefaultCommandInteractive
var defaultCommandSelector = promptStartupChoice

type startupOption struct {
	key         string
	label       string
	description string
}

const (
	startupANSIReset    = "\033[0m"
	startupANSIDim      = "\033[38;2;214;122;76m"
	startupANSITitle    = "\033[1;38;2;249;115;22m"
	startupANSIOption   = "\033[1;38;2;251;146;60m"
	startupANSIActive   = "\033[1;38;2;234;88;12m"
	startupANSIInactive = "\033[1;38;2;249;115;22m"
	startupANSIRail     = "\033[38;2;251;146;60m"
	startupANSIPrompt   = "\033[1;38;2;234;88;12m"
	startupANSIBgClear  = "\033[H\033[2J"
)

var startupOptions = []startupOption{
	{
		key:         "agent",
		label:       "Terminal Agent",
		description: "Chat or ask for file changes in terminal",
	},
	{
		key:         "web",
		label:       "Web Console",
		description: "Open the browser dashboard",
	},
	{
		key:         "tui",
		label:       "TUI Dashboard",
		description: "Open the terminal dashboard",
	},
}

func NewJameclawCommand() *cobra.Command {
	short := fmt.Sprintf("%s trenchsi - Personal AI Assistant v%s\n\n", internal.Logo, config.GetVersion())

	cmd := &cobra.Command{
		Use:          "trenchsi",
		Short:        short,
		Example:      "trenchsi\ntrenchsi gateway\ntrenchsi version",
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runInteractiveDefaultCommand()
		},
	}

	cmd.AddCommand(
		onboard.NewOnboardCommand(),
		agent.NewAgentCommand(),
		auth.NewAuthCommand(),
		gateway.NewGatewayCommand(),
		status.NewStatusCommand(),
		cron.NewCronCommand(),
		migrate.NewMigrateCommand(),
		skills.NewSkillsCommand(),
		model.NewModelCommand(),
		version.NewVersionCommand(),
		uninstall.NewUninstallCommand(),
	)

	return cmd
}

const (
	colorOrange = "\033[1;38;2;249;115;22m"
	colorReset = "\033[0m"
	banner     = "\r\n" +
		colorOrange + "████████╗██████╗ ███████╗███╗   ██╗ ██████╗██╗  ██╗ ██████╗██╗      █████╗ ██╗    ██╗\n" +
		colorOrange + "╚══██╔══╝██╔══██╗██╔════╝████╗  ██║██╔════╝██║  ██║██╔════╝██║     ██╔══██╗██║    ██║\n" +
		colorOrange + "   ██║   ██████╔╝█████╗  ██╔██╗ ██║██║     ███████║██║     ██║     ███████║██║ █╗ ██║\n" +
		colorOrange + "   ██║   ██╔══██╗██╔══╝  ██║╚██╗██║██║     ██╔══██║██║     ██║     ██╔══██║██║███╗██║\n" +
		colorOrange + "   ██║   ██║  ██║███████╗██║ ╚████║╚██████╗██║  ██║╚██████╗███████╗██║  ██║╚███╔███╔╝\n" +
		colorOrange + "   ╚═╝   ╚═╝  ╚═╝╚══════╝╚═╝  ╚═══╝ ╚═════╝╚═╝  ╚═╝ ╚═════╝╚══════╝╚═╝  ╚═╝ ╚══╝╚══╝\n" +
		colorReset + "\r\n"
)

func runInteractiveDefaultCommand() error {
	if !startupOnboardComplete() {
		return runDefaultOnboardCommand()
	}

	if !defaultCommandIsInteractive() {
		return runDefaultAgentCommand()
	}

	for {
		choice, err := defaultCommandSelector()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch choice {
		case "agent":
			return runDefaultAgentCommand()
		case "web":
			return runDefaultWebCommand()
		case "tui":
			return runDefaultTUICommand()
		case "":
			fmt.Fprintln(defaultCommandOutput, "Choose one of: agent, web, or tui.")
		default:
			fmt.Fprintf(defaultCommandOutput, "Unknown startup option %q. Choose agent, web, or tui.\n", choice)
		}
	}
}

func renderSetupRequired() {
	fmt.Fprintln(defaultCommandOutput)
	fmt.Fprintf(defaultCommandOutput, "%sTrenchsi is not installed yet.%s\n", startupANSITitle, startupANSIReset)
	fmt.Fprintf(defaultCommandOutput, "%sRun the setup command first, then come back to `trenchsi`.%s\n", startupANSIDim, startupANSIReset)
	fmt.Fprintln(defaultCommandOutput)
	fmt.Fprintf(defaultCommandOutput, "%sOnboard:%s trenchsi onboard\n", startupANSIOption, startupANSIReset)
	fmt.Fprintf(defaultCommandOutput, "%sAlias:%s trenchsi install\n", startupANSIOption, startupANSIReset)
}

func runAgentChatCommand() error {
	cfg, err := internal.LoadConfig()
	if err != nil {
		return err
	}

	if cfg.Agents.Defaults.ModelName == "" {
		fmt.Fprintln(defaultCommandOutput)
		fmt.Fprintf(defaultCommandOutput, "%sNo default model configured.%s\n", startupANSITitle, startupANSIReset)
		fmt.Fprintf(defaultCommandOutput, "%sAdd a model API key or set a default model first.%s\n", startupANSIDim, startupANSIReset)
		fmt.Fprintln(defaultCommandOutput)
		fmt.Fprintf(defaultCommandOutput, "%sTry next:%s trenchsi onboard\n", startupANSIOption, startupANSIReset)
		fmt.Fprintf(defaultCommandOutput, "%sOr set a model:%s trenchsi model <model-name>\n", startupANSIOption, startupANSIReset)
		return nil
	}

	return agent.NewAgentCommand().RunE(nil, nil)
}

func promptStartupChoice() (string, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return promptStartupChoiceFallback()
	}
	defer func() {
		_ = term.Restore(fd, oldState)
	}()

	selected := 0
	buf := make([]byte, 3)
	for {
		renderStartupChoice(selected)

		n, readErr := os.Stdin.Read(buf)
		if readErr != nil {
			return "", readErr
		}
		if n == 0 {
			continue
		}

		key := normalizeStartupKey(buf[:n])
		switch key {
		case "up":
			selected = (selected + len(startupOptions) - 1) % len(startupOptions)
		case "down":
			selected = (selected + 1) % len(startupOptions)
		case "select":
			fmt.Fprint(defaultCommandOutput, startupANSIBgClear)
			return startupOptions[selected].key, nil
		case "cancel":
			fmt.Fprint(defaultCommandOutput, startupANSIBgClear)
			return "", io.EOF
		case "web", "tui", "agent":
			fmt.Fprint(defaultCommandOutput, startupANSIBgClear)
			return key, nil
		}
	}
}

func promptStartupChoiceFallback() (string, error) {
	fmt.Fprintln(defaultCommandOutput, "Choose how to start Trenchsi:")
	for index, option := range startupOptions {
		fmt.Fprintf(defaultCommandOutput, "  %d. %s\n", index+1, option.label)
	}
	fmt.Fprint(defaultCommandOutput, "Select 1, 2, or 3: ")

	var line string
	if _, err := fmt.Fscanln(os.Stdin, &line); err != nil {
		if err == io.EOF {
			return "", io.EOF
		}
		return "", err
	}
	return normalizeStartupChoice(line), nil
}

func renderStartupChoice(selected int) {
	fmt.Fprint(defaultCommandOutput, startupANSIBgClear)
	startupWriteLine("%sTrenchsi Startup%s", startupANSITitle, startupANSIReset)
	startupWriteLine("%sUse arrows, Enter, or Space.%s", startupANSIDim, startupANSIReset)
	startupWriteLine("")
	for index, option := range startupOptions {
		isSelected := index == selected
		titleColor := startupANSIOption
		markerColor := startupANSIInactive
		railColor := startupANSIRail
		marker := "◇"
		if isSelected {
			markerColor = startupANSIActive
			railColor = startupANSIActive
			titleColor = startupANSIOption
			marker = "◆"
		}

		startupWriteLine("  %s│%s", railColor, startupANSIReset)
		startupWriteLine(
			"  %s%s%s %s%s%s",
			markerColor,
			marker,
			startupANSIReset,
			titleColor,
			option.label,
			startupANSIReset,
		)
		startupWriteLine(
			"  %s│%s %s%s%s",
			railColor,
			startupANSIReset,
			startupANSIDim,
			option.description,
			startupANSIReset,
		)
	}
	startupWriteLine("  %s│%s", startupANSIRail, startupANSIReset)
	startupWriteLine("")
	startupWriteLine("%sSelect%s", startupANSIPrompt, startupANSIReset)
}

func startupWriteLine(format string, args ...any) {
	fmt.Fprintf(defaultCommandOutput, format+"\r\n", args...)
}

func normalizeStartupKey(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	if len(raw) >= 3 && raw[0] == 27 && raw[1] == '[' {
		switch raw[2] {
		case 'A':
			return "up"
		case 'B':
			return "down"
		}
	}

	switch strings.ToLower(string(raw)) {
	case "\r", "\n", " ":
		return "select"
	case "\x03", "\x1b":
		return "cancel"
	case "k":
		return "up"
	case "j":
		return "down"
	default:
		return normalizeStartupChoice(string(raw))
	}
}

func normalizeStartupChoice(raw string) string {
	choice := strings.ToLower(strings.TrimSpace(raw))
	switch choice {
	case "1", "agent", "chat", "terminal", "terminal-agent", "terminal chat", "terminal-chat", "edit", "modify":
		return "agent"
	case "2", "web", "webconsole", "web-console", "dashboard", "browser", "web console":
		return "web"
	case "3", "tui", "terminal-ui", "terminal dashboard", "dashboard-tui":
		return "tui"
	default:
		return ""
	}
}

func isDefaultCommandInteractive() bool {
	stdinInfo, err := os.Stdin.Stat()
	if err != nil || (stdinInfo.Mode()&os.ModeCharDevice) == 0 {
		return false
	}
	stdoutInfo, err := os.Stdout.Stat()
	if err != nil || (stdoutInfo.Mode()&os.ModeCharDevice) == 0 {
		return false
	}
	return true
}

func runTUICommand() error {
	binary, err := resolveLauncherBinary("TRENCHSI_TUI_BINARY", "trenchsi-launcher-tui", "trenchlaw-launcher-tui")
	if err != nil {
		return fmt.Errorf("%w. Build it with `go build -o /Users/horatiubudai/.local/bin/trenchsi-launcher-tui ./cmd/trenchlaw-launcher-tui`", err)
	}

	cmd := exec.Command(binary)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runWebCommand() error {
	binary, err := resolveLauncherBinary("TRENCHSI_WEB_BINARY", "trenchsi-web", "trenchlaw-web")
	if err != nil {
		return fmt.Errorf("%w. Build it with `CGO_ENABLED=0 go build -o /Users/horatiubudai/.local/bin/trenchsi-web ./web/backend`", err)
	}

	cmd := exec.Command(binary)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func resolveLauncherBinary(envVar string, binaryNames ...string) (string, error) {
	for _, candidateEnv := range []string{envVar, strings.ReplaceAll(envVar, "TRENCHSI", "TRENCHLAW")} {
		if custom := os.Getenv(candidateEnv); custom != "" {
			if info, err := os.Stat(custom); err == nil && !info.IsDir() {
				return custom, nil
			}
		}
	}

	names := append([]string(nil), binaryNames...)
	if runtime.GOOS == "windows" {
		for i := range names {
			names[i] += ".exe"
		}
	}

	candidates := make([]string, 0, 12)
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		for _, name := range names {
			candidates = append(candidates,
				filepath.Join(exeDir, name),
				filepath.Join(exeDir, "build", name),
			)
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		for _, name := range names {
			candidates = append(candidates,
				filepath.Join(cwd, "build", name),
				filepath.Join(cwd, "cmd", name, name),
			)
		}
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}

	for _, name := range names {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("%s binary not found", strings.Join(binaryNames, " or "))
}

func main() {
	fmt.Printf("%s", banner)
	cmd := NewJameclawCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
