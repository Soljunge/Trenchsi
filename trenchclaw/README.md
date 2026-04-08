# TrenchClaw Core

This directory contains the TrenchClaw agent core: the Go CLI, workspace scaffold, provider configuration, channel integrations, and supporting docs.

For the main project overview and quick start, go to [the repository README](../README.md).

## Start Here

```bash
go build -o trenchclaw ./cmd/trenchlaw
./trenchclaw onboard
./trenchclaw agent
```

## Core Features

- Terminal-first AI agent workflow
- Configurable provider and model selection
- Web, terminal, and gateway entry points
- Skills and workspace-based customization
- Multi-channel chat integrations
- Built-in scheduling with cron commands

## Documentation

- Providers: [`docs/providers.md`](./docs/providers.md)
- Telegram: [`docs/channels/telegram/README.md`](./docs/channels/telegram/README.md)
- Discord: [`docs/channels/discord/README.md`](./docs/channels/discord/README.md)
- Hooks: [`docs/hooks/README.md`](./docs/hooks/README.md)
- Agent refactor notes: [`docs/agent-refactor/README.md`](./docs/agent-refactor/README.md)
