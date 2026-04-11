// TrenchLaw - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 TrenchLaw contributors

package config

// Runtime environment variable keys for the trenchlaw process.
// These control the location of files and binaries at runtime and are read
// directly via os.Getenv / os.LookupEnv. All trenchlaw-specific keys use the
// TRENCHLAW_ prefix. Reference these constants instead of inline string
// literals to keep all supported knobs visible in one place and to prevent
// typos.
const (
	// EnvHome overrides the base directory for all trenchlaw data
	// (config, workspace, skills, auth store, …).
	// Default: ~/.trenchlaw
	EnvHome = "TRENCHLAW_HOME"

	// EnvConfig overrides the full path to the JSON config file.
	// Default: $TRENCHLAW_HOME/config.json
	EnvConfig = "TRENCHLAW_CONFIG"

	// EnvBuiltinSkills overrides the directory from which built-in
	// skills are loaded.
	// Default: <cwd>/skills
	EnvBuiltinSkills = "TRENCHLAW_BUILTIN_SKILLS"

	// EnvBinary overrides the path to the trenchlaw executable.
	// Used by the web launcher when spawning the gateway subprocess.
	// Default: resolved from the same directory as the current executable.
	EnvBinary = "TRENCHLAW_BINARY"

	// EnvGatewayHost overrides the host address for the gateway server.
	// Default: "127.0.0.1"
	EnvGatewayHost = "TRENCHLAW_GATEWAY_HOST"
)
