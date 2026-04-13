# Trenchsi

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Built With: Go](https://img.shields.io/badge/Built%20With-Go-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Interface: CLI Agent](https://img.shields.io/badge/Interface-CLI%20Agent-F97316)](#quick-start)
[![Channel: Telegram](https://img.shields.io/badge/Channel-Telegram-26A5E4?logo=telegram&logoColor=white)](./trenchsi/docs/channels/telegram/README.md)
[![Gateway: Enabled](https://img.shields.io/badge/Gateway-Multi--Channel-111827)](#channel-integrations)

<p align="center">
  <img src="./trenchsi/assets/agent-avatar.png" alt="Trenchsi shrimp avatar" width="260" />
</p>

Trenchsi is a lightweight AI agent focused on practical work: direct chat in the terminal, configurable model providers, workspace memory, skills, scheduled jobs, and message-based integrations such as Telegram and Discord.

This repository contains the Trenchsi project workspace and the agent core in [`trenchsi/`](./trenchsi).

## What The Agent Can Do

- Chat directly in the terminal with a configurable default model
- Run in multiple interfaces: terminal agent, web console, and TUI dashboard
- Connect to external channels such as Telegram, Discord, Slack, Matrix, LINE, QQ, Weixin, DingTalk, Feishu, and more
- Use multiple model providers through a model-centric configuration
- Manage reusable skills for specialized workflows
- Keep workspace context with files such as `AGENT.md`, `USER.md`, `SOUL.md`, and `MEMORY.md`
- Schedule recurring work with built-in cron commands
- Expose a gateway process for always-on or multi-channel agent operation

## Repository Layout

- [`README.md`](./README.md): GitHub landing page and quick start
- [`trenchsi/`](./trenchsi): Agent source code, docs, workspace scaffold, and packaging
- [`trenchsi/docs/`](./trenchsi/docs): Provider, channel, and architecture docs
- [`trenchsi/workspace/`](./trenchsi/workspace): Default workspace identity, memory, and tool instructions

## Quick Start

The shortest path is: build once, run onboarding once, set a default model, then use `trenchsi`.

### 1. Build the CLI

From the repository root:

```bash
cd trenchsi
go build -o trenchsi ./cmd/trenchlaw
```

### 2. Run onboarding

```bash
./trenchsi onboard
```

This creates the local config and workspace files. `install` is an alias:

```bash
./trenchsi install
```

### 3. Set a default model

You can do this during onboarding or later with the CLI:

```bash
./trenchsi auth login
./trenchsi auth status
./trenchsi model
./trenchsi model gpt-5.4
```

If you prefer editing config directly, add a model entry such as:

```json
{
  "model_name": "gpt-5.4",
  "model": "openai/gpt-5.4",
  "api_key": "sk-..."
}
```

Provider setup details are in [`trenchsi/docs/providers.md`](./trenchsi/docs/providers.md).

### 4. Start the agent

For normal use:

```bash
./trenchsi
```

If onboarding is complete, the interactive launcher lets you choose:

- Terminal Agent
- Web Console
- TUI Dashboard

If you want the terminal agent directly:

```bash
./trenchsi agent
./trenchsi agent -m "Summarize the repository and suggest next steps"
```

Useful checks:

```bash
./trenchsi status
./trenchsi version
```

## Most Used Commands

Use these first before exploring the rest of the CLI:

```bash
./trenchsi onboard      # first-time setup
./trenchsi              # launcher: agent, web, or tui
./trenchsi agent        # terminal agent directly
./trenchsi auth login   # connect a provider
./trenchsi model        # list or set the default model
./trenchsi status       # confirm config and runtime state
./trenchsi gateway      # run integrations and long-lived channels
./trenchsi skills list  # inspect installed skills
./trenchsi cron list    # inspect scheduled jobs
```

## Interfaces

`trenchsi` is the default entry point.

- `trenchsi`: opens the startup selector in an interactive terminal
- `trenchsi agent`: skips the selector and starts terminal chat
- `trenchsi gateway`: runs the always-on gateway for channels and automations

If setup is incomplete, `trenchsi` sends you to onboarding first. If no default model is configured, `trenchsi agent` tells you to finish setup or choose a model.

## Channel Integrations

Trenchsi can operate across external messaging platforms through the gateway.

Documented channel:

- Telegram

Start the gateway with:

```bash
./trenchsi gateway
```

Channel setup docs:

- [`trenchsi/docs/channels/telegram/README.md`](./trenchsi/docs/channels/telegram/README.md)
- [`trenchsi/docs/channels/discord/README.md`](./trenchsi/docs/channels/discord/README.md)
- [`trenchsi/docs/channels/slack/README.md`](./trenchsi/docs/channels/slack/README.md)

## Skills, Memory, And Workspace Customization

The workspace is part of the agent design. Trenchsi ships with editable workspace files that define identity, memory, tools, and user-specific context.

Important files:

- [`trenchsi/workspace/AGENT.md`](./trenchsi/workspace/AGENT.md)
- [`trenchsi/workspace/USER.md`](./trenchsi/workspace/USER.md)
- [`trenchsi/workspace/SOUL.md`](./trenchsi/workspace/SOUL.md)
- [`trenchsi/workspace/TOOLS.md`](./trenchsi/workspace/TOOLS.md)
- [`trenchsi/workspace/memory/MEMORY.md`](./trenchsi/workspace/memory/MEMORY.md)

Skill management commands:

```bash
./trenchsi skills list
./trenchsi skills search git
./trenchsi skills install <github-repo>
./trenchsi skills remove <skill-name>
```

## Scheduling And Automation

Trenchsi includes built-in cron support for recurring work.

Examples:

```bash
./trenchsi cron list
./trenchsi cron add
./trenchsi cron enable <job-id>
./trenchsi cron disable <job-id>
./trenchsi cron remove <job-id>
```

## Documentation

- Provider setup: [`trenchsi/docs/providers.md`](./trenchsi/docs/providers.md)
- Hooks: [`trenchsi/docs/hooks/README.md`](./trenchsi/docs/hooks/README.md)
- Agent refactor notes: [`trenchsi/docs/agent-refactor/README.md`](./trenchsi/docs/agent-refactor/README.md)
- Chat app docs: [`trenchsi/docs/chat-apps.md`](./trenchsi/docs/chat-apps.md)

## Notes

- The package metadata inside `trenchsi/` still references `trenchlaw` in several places. The current repo name and top-level project presentation use `Trenchsi`.
- `.DS_Store` files are currently modified in the repository and were not changed by this README update.
