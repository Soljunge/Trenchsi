# Trenchsi Core

<p align="center">
  <img src="./assets/agent-avatar.png" alt="Trenchsi shrimp avatar" width="220" />
</p>

This directory contains the Trenchsi agent core: the Go CLI, startup launcher, web console backend, workspace scaffold, provider stack, tool system, automation features, and multi-channel integrations, with a strong focus on memecoin trenching and trading workflows.

For the main project overview and quick start, go to [the repository README](../README.md).

## Start Here

```bash
go build -o trenchsi ./cmd/trenchlaw
./trenchsi onboard
./trenchsi
```

If you want the terminal agent directly, use:

```bash
./trenchsi agent
```

Typical first-run flow:

```bash
./trenchsi auth login
./trenchsi model
./trenchsi status
```

## Core Features

- Built for memecoin trenching, trade support, and fast operator workflows
- Interactive startup launcher with terminal, web, and TUI entry points
- Configurable providers and models with auth flows and provider fallback support
- Workspace-driven identity, memory, tools, and user customization
- Skill discovery, installation, and per-workspace skill management
- Web search, file operations, shell execution, and extensible tool registry
- MCP integration with deferred tool discovery
- Session history and persistent memory storage
- Gateway-based chat integrations across Telegram, Discord, Slack, Matrix, WhatsApp, QQ, WeCom, Feishu, OneBot, and more
- Agent bindings for routing different channels or contexts to different agents
- Built-in scheduling and recurring jobs with cron commands
- Sandbox controls and sensitive data filtering for safer tool execution

## Documentation

- Configuration: [`docs/configuration.md`](./docs/configuration.md)
- Providers: [`docs/providers.md`](./docs/providers.md)
- Chat apps: [`docs/chat-apps.md`](./docs/chat-apps.md)
- Tools: [`docs/tools_configuration.md`](./docs/tools_configuration.md)
- Hooks: [`docs/hooks/README.md`](./docs/hooks/README.md)
- Telegram: [`docs/channels/telegram/README.md`](./docs/channels/telegram/README.md)
- Discord: [`docs/channels/discord/README.md`](./docs/channels/discord/README.md)
- Agent refactor notes: [`docs/agent-refactor/README.md`](./docs/agent-refactor/README.md)
