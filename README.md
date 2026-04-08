# TrenchClaw

TrenchClaw is a lightweight AI agent focused on practical work: direct chat in the terminal, configurable model providers, workspace memory, skills, scheduled jobs, and message-based integrations such as Telegram and Discord.

This repository contains the TrenchClaw project workspace and the agent core in [`trenchclaw/`](./trenchclaw).

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
- [`trenchclaw/`](./trenchclaw): Agent source code, docs, workspace scaffold, and packaging
- [`trenchclaw/docs/`](./trenchclaw/docs): Provider, channel, and architecture docs
- [`trenchclaw/workspace/`](./trenchclaw/workspace): Default workspace identity, memory, and tool instructions

## Quick Start

### 1. Build the CLI

From the repository root:

```bash
cd trenchclaw
go build -o trenchclaw ./cmd/trenchlaw
```

### 2. Run first-time setup

```bash
./trenchclaw onboard
```

`onboard` initializes the local config and workspace. It is also available as:

```bash
./trenchclaw install
```

### 3. Add a model provider

You can authenticate and configure a default model in the setup flow, or do it manually later.

Examples:

```bash
./trenchclaw auth login
./trenchclaw auth status
./trenchclaw model
./trenchclaw model gpt-5.4
```

The project uses a `model_list` configuration, so you can add providers with entries like:

```json
{
  "model_name": "gpt-5.4",
  "model": "openai/gpt-5.4",
  "api_key": "sk-..."
}
```

Supported provider families include OpenAI, Anthropic, OpenRouter, Gemini, Groq, DeepSeek, Ollama, VLLM, Azure OpenAI, GitHub Copilot, and others documented in [`trenchclaw/docs/providers.md`](./trenchclaw/docs/providers.md).

### 4. Start using the agent

Interactive mode:

```bash
./trenchclaw
```

Direct agent chat:

```bash
./trenchclaw agent
./trenchclaw agent -m "Summarize the repository and suggest next steps"
```

Status and version:

```bash
./trenchclaw status
./trenchclaw version
```

## Main Commands

Use these commands as your default operating surface:

```bash
./trenchclaw onboard
./trenchclaw agent
./trenchclaw auth login
./trenchclaw auth status
./trenchclaw model
./trenchclaw gateway
./trenchclaw skills list
./trenchclaw cron list
./trenchclaw status
```

## Interfaces

When you run `trenchclaw` interactively, the CLI can route you into one of three entry points:

- Terminal Agent: direct chat and task execution in the terminal
- Web Console: browser-based dashboard
- TUI Dashboard: terminal UI dashboard

If setup is incomplete, the CLI redirects you to onboarding first.

## Channel Integrations

TrenchClaw can operate across external messaging platforms through the gateway.

Documented channels include:

- Telegram
- Discord
- Slack
- Matrix
- LINE
- QQ
- Weixin
- DingTalk
- Feishu
- OneBot
- MaixCam

Start the gateway with:

```bash
./trenchclaw gateway
```

Channel setup docs:

- [`trenchclaw/docs/channels/telegram/README.md`](./trenchclaw/docs/channels/telegram/README.md)
- [`trenchclaw/docs/channels/discord/README.md`](./trenchclaw/docs/channels/discord/README.md)
- [`trenchclaw/docs/channels/slack/README.md`](./trenchclaw/docs/channels/slack/README.md)

## Skills, Memory, And Workspace Customization

The workspace is part of the agent design. TrenchClaw ships with editable workspace files that define identity, memory, tools, and user-specific context.

Important files:

- [`trenchclaw/workspace/AGENT.md`](./trenchclaw/workspace/AGENT.md)
- [`trenchclaw/workspace/USER.md`](./trenchclaw/workspace/USER.md)
- [`trenchclaw/workspace/SOUL.md`](./trenchclaw/workspace/SOUL.md)
- [`trenchclaw/workspace/TOOLS.md`](./trenchclaw/workspace/TOOLS.md)
- [`trenchclaw/workspace/memory/MEMORY.md`](./trenchclaw/workspace/memory/MEMORY.md)

Skill management commands:

```bash
./trenchclaw skills list
./trenchclaw skills search git
./trenchclaw skills install <github-repo>
./trenchclaw skills remove <skill-name>
```

## Scheduling And Automation

TrenchClaw includes built-in cron support for recurring work.

Examples:

```bash
./trenchclaw cron list
./trenchclaw cron add
./trenchclaw cron enable <job-id>
./trenchclaw cron disable <job-id>
./trenchclaw cron remove <job-id>
```

## Documentation

- Provider setup: [`trenchclaw/docs/providers.md`](./trenchclaw/docs/providers.md)
- Hooks: [`trenchclaw/docs/hooks/README.md`](./trenchclaw/docs/hooks/README.md)
- Agent refactor notes: [`trenchclaw/docs/agent-refactor/README.md`](./trenchclaw/docs/agent-refactor/README.md)
- Chat app docs: [`trenchclaw/docs/chat-apps.md`](./trenchclaw/docs/chat-apps.md)

## Notes

- The package metadata inside `trenchclaw/` still references `trenchlaw` in several places. The current repo name and top-level project presentation use `TrenchClaw`.
- `.DS_Store` files are currently modified in the repository and were not changed by this README update.
