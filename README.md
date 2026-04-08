<div align="center">
  <img width="1500" height="276" alt="coffeegraph-banner" src=".github/assets/coffeegraph-banner1.jpg" />
</div>

<h1 align="center">☕ CoffeeGraph ☕  </h1>

<div align="center">
  <strong>One folder. One command. Your complete AI agency running on autopilot.</strong>
  <br />
  Sales · Content · Leads · Personal OS · Modular skills · Live graph · Zero setup
  <br /><br />
  Free, offline-capable, no account required.
</div>

<br />
<p align="center">
  <a href="https://github.com/angelnicolasc/coffeegraph/actions">
    <img src="https://img.shields.io/github/actions/workflow/status/angelnicolasc/coffeegraph/ci.yml?branch=main&style=for-the-badge" alt="CI status">
  </a>
  <a href="https://github.com/angelnicolasc/coffeegraph/releases">
    <img src="https://img.shields.io/github/v/release/angelnicolasc/coffeegraph?include_prereleases&style=for-the-badge" alt="Latest Release">
  </a>
  <a href="https://github.com/angelnicolasc/coffeegraph/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge" alt="License">
  </a>
  <a href="https://goreportcard.com/report/github.com/angelnicolasc/coffeegraph">
    <img src="https://img.shields.io/badge/Go%20Report-A%2B-brightgreen?style=for-the-badge&logo=go" alt="Go Report">
  </a>
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go version">
  <img src="https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-00ADD8?style=for-the-badge" alt="Platform">
</p>

---

## 60-Second Demo

```bash
# Install (macOS/Linux/WSL)
curl -fsSL https://raw.githubusercontent.com/angelnicolasc/coffeegraph/main/scripts/install.sh | bash

# Create your agency
coffeegraph init my-agency && cd my-agency

# Add your first AI agent
coffeegraph add sales-closer

# Edit config.yaml → add your ANTHROPIC_API_KEY

# Queue a task
coffeegraph queue add --skill sales-closer --task "Follow up this week's leads" --priority 1

# Let your team work. Go get coffee.
coffeegraph coffee
```

> Your AI team is already working. Go get coffee. ☕

---

## Why CoffeeGraph?

| | CoffeeGraph | Visual builders | Custom Python scripts |
|---|---|---|---|
| Setup time | 60 seconds | Hours | Days |
| Runtime deps | Zero (single binary) | Node.js / Docker | Python + pip + venv |
| Skill authoring | Markdown | Drag-and-drop GUI | Code |
| Config format | YAML + folder structure | Proprietary DB | Code |
| Output | Local files, shareable | Cloud-locked | Terminal |
| Cost | Free + your API key | $20-200/mo platform fees | Free + your API key |

**The concept**: your AI team is a folder of Markdown files. Each file is a skill — a self-contained agent with its own identity, prompts, and output format. No DSL to learn. No SDK. No server. Just `SKILL.md`.

---

## Overview

Everything out there gives you agents to build software.
**CoffeeGraph gives you an AI team to run your business.**

No code required. No server to configure.
A folder of Markdown files and a single command is all you need.

Each **skill** is a self-contained AI agent with its own identity, prompts, n8n integration, and output format. You compose them, queue tasks, and `coffee` mode executes everything autonomously while you're away.

If something does not work, run this first:

```bash
coffeegraph doctor
```

---

## Architecture

```
coffeegraph/
├── cmd/coffeegraph/          Entry point (Cobra CLI)
├── internal/
│   ├── cli/                  Command implementations
│   ├── graph/                Graph engine (generator, watcher, patch)
│   ├── tui/                  Terminal UI (Bubble Tea + Lip Gloss)
│   ├── server/               HTTP + WebSocket server (visualizer)
│   ├── claude/               Anthropic API client (retries, jitter, backoff)
│   ├── queue/                Task queue (priority, atomic persistence)
│   ├── runner/               Task execution engine
│   ├── config/               YAML config + env override layer
│   ├── project/              Project root detection
│   ├── doctor/               Health checks
│   ├── logs/                 Execution log management
│   ├── share/                Job output sharing (Gist + HTML fallback)
│   ├── evolve/               Self-improving skill suggestions
│   ├── bot/                  Telegram bot adapter
│   ├── mcp/                  MCP server (JSON-RPC 2.0 over stdio)
│   ├── registry/             Community skill registry
│   └── fsutil/               Atomic file write utilities
├── templates/                Bundled skill templates (5 included)
│   └── <skill>/
│       ├── SKILL.md          Agent identity, inputs, output format
│       ├── prompts/          Task-specific prompt files
│       └── n8n/workflow.json n8n automation blueprint
├── web/                      Browser visualizer (Cytoscape.js, WebSocket)
└── scripts/                  install.sh — curl | bash installer
```

### Key design decisions

| Concern | Approach |
|---|---|
| State | `graph.json` as single source of truth, regenerated by debounced file watcher |
| Skills | Pure Markdown — no DSL, no SDK dependency for skill authors |
| Persistence | Atomic writes (write-then-rename) for queue and graph to prevent corruption |
| Execution | Skills run through LLM backends with exponential backoff + jitter |
| UI | Two layers: TUI (offline-first) + browser graph (real-time WebSocket) |
| Packaging | Single static binary via GoReleaser; no runtime dependencies |
| Extensibility | New skills = new folder under `templates/`; no code change required |

---

## Skills

| Skill | What it does | Integrates with |
|---|---|---|
| `sales-closer` | Qualifies leads, writes personalized emails, handles follow-ups | Gmail, Notion CRM |
| `content-engine` | Monitors trends, generates viral threads, adapts for IG and newsletters | Buffer, Hypefury |
| `lead-nurture` | 3-touch hyper-personalized outreach sequence | Instantly, Lemlist |
| `life-os` | Daily brief, expense tracking, focus blocks | Telegram, Google Calendar |
| `creator-stack` | Topic → YouTube script + X thread + IG caption + newsletter intro | Buffer, Google Docs |

```bash
coffeegraph add sales-closer
coffeegraph add content-engine
coffeegraph add creator-stack
```

---

## CLI Reference

### Core commands

```
coffeegraph init <name>           Create a new agency project
coffeegraph add <skill>           Install a skill from templates
coffeegraph run <skill> <task>    Run a single task against a skill (inline)
coffeegraph coffee                Execute all queued tasks autonomously
coffeegraph coffee --urgent       High-pressure meme mode 🔥
coffeegraph coffee --chill        Calm vibes mode ☕
coffeegraph doctor                Dependency and integration health checks
```

### Queue management

```
coffeegraph queue add             Queue a task (interactive or --flags)
coffeegraph queue list            Show pending tasks
coffeegraph queue clear           Empty the queue
```

### Monitoring & output

```
coffeegraph dashboard             Open TUI dashboard (live status)
coffeegraph visualize             Open browser graph (WebSocket live)
coffeegraph log --pretty          Pretty terminal receipts for completed jobs
coffeegraph log --pretty --last 1 Show only the most recent job
coffeegraph share [job-id]        Share last or specific completed job
```

### AI-powered

```
coffeegraph suggest               Suggest new skills from your business context
coffeegraph suggest --from-obsidian /path/to/vault
coffeegraph evolve <skill>        Improve a skill from its latest execution log
coffeegraph roast                 VC-style roast of your business plan
coffeegraph party                 Round-robin conversation between your skills
```

### Integrations

```
coffeegraph bot                   Run Telegram bot adapter (@skill task)
coffeegraph mcp                   Run as MCP server over stdio
coffeegraph skill list            List community skills
coffeegraph skill install <name>  Install community skill
```

### Advanced

```
coffeegraph deploy openclaw       Export skills to OpenClaw (SOUL.md)
coffeegraph deploy hermes         Export skills to Hermes Agent (AGENT.md)
coffeegraph nap                   Pause with playful animation
```

### Queue a task

```bash
# Interactive
coffeegraph queue add

# Non-interactive
coffeegraph queue add --skill sales-closer --task "Follow up this week's leads" --priority 1
```

---

## Coffee Mode

```bash
coffeegraph coffee
```

Picks up every task in the queue, runs them sequentially against each skill's identity and prompts, saves outputs to `.coffee/logs/`, snapshots skill state to `.coffee/snapshots/`, and sends a desktop notification when done.

**Safe to Ctrl+C mid-execution.** Current task completes, queue state is persisted, partial outputs are logged.

**Nothing to babysit.** Start it, grab a coffee, come back to results.

---

## Dashboard & Visualizer

```bash
coffeegraph dashboard    # TUI — works without network
coffeegraph visualize    # Interactive graph in browser
```

The **TUI** shows live skill status, task queue, and last output preview.
Keyboard: `↑↓` navigate · `a` add task · `c` run coffee · `v` open visualizer · `r` refresh · `q` quit

The **browser visualizer** renders your agency as an interactive Cytoscape.js graph with real-time WebSocket updates. Node color reflects skill status (idle / running / done / error).

---

## Project Layout (generated by `init`)

```
my-agency/
├── index.md          ← your business context (edit this first)
├── config.yaml       ← API keys + skill configuration
├── graph.json        ← auto-generated, do not edit
├── skills/
│   └── sales-closer/
│       └── SKILL.md  ← customize the agent
└── .coffee/
    ├── queue.json    ← task queue
    ├── logs/         ← execution outputs
    └── snapshots/    ← pre-run skill snapshots
```

The only file you must edit to get started: `config.yaml` — add your Anthropic API key.

---

## Configuration

**config.yaml**
```yaml
agency_name: my-agency
default_model: claude-sonnet-4-6
backend: anthropic                        # anthropic | ollama
auto_evolve: false
mcp: false

api_keys:
  anthropic: sk-ant-...                   # or use env var
  github: ghp_...                         # optional (share via gist)
  telegram_bot_token: 1234:abcd           # optional (coffeegraph bot)

bot:
  allowed_chat_ids: [123456789]           # required for bot

skills:
  sales-closer:
    model: claude-opus-4-6                # per-skill model override
    backend: anthropic                    # or ollama
  content-engine:
    enabled: true
```

**Environment variables** (override config.yaml)
```bash
export ANTHROPIC_API_KEY=sk-ant-...
export GITHUB_TOKEN=ghp_...
export TELEGRAM_BOT_TOKEN=1234:abcd
export COFFEEGRAPH_MODEL=claude-sonnet-4-6
export COFFEEGRAPH_BACKEND=ollama
export COFFEEGRAPH_N8N_WEBHOOK=https://your-n8n.app/webhook/coffeegraph
```

---

## Local Model Support (Ollama)

Run your entire AI agency locally — zero API bills.

```yaml
# config.yaml
backend: ollama
skills:
  sales-closer:
    backend: ollama
    model: llama3.2
```

If Ollama is not running, CoffeeGraph prints a clear error:
```
Ollama doesn't appear to be running. Start it with: ollama serve
```

---

## AI Suggestions

```bash
coffeegraph suggest
```

Reads your `index.md` business context and returns 3 skill ideas ranked by impact — with full SKILL.md drafts ready to drop in.

---

## Requirements

- [Anthropic API key](https://console.anthropic.com) — required for coffee mode and suggest (unless using Ollama)
- [n8n](https://n8n.io) — optional, for workflow automations
- macOS / Linux / Windows (WSL or native)
- Go 1.22+ — only if building from source

---

## Install

**One-liner (macOS / Linux / WSL)**
```bash
curl -fsSL https://raw.githubusercontent.com/angelnicolasc/coffeegraph/main/scripts/install.sh | bash
```

**Homebrew** *(coming soon)*
```bash
brew install angelnicolasc/tap/coffeegraph
```

**Build from source**
```bash
git clone https://github.com/angelnicolasc/coffeegraph
cd coffeegraph
make build
```

**Releases** — pre-built binaries for Linux (amd64/arm64), macOS (amd64/arm64), and Windows (amd64) available on the [Releases page](https://github.com/angelnicolasc/coffeegraph/releases).

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for full setup instructions, project structure, SKILL.md format, and code conventions.

### Quick start for contributors

```bash
git clone https://github.com/angelnicolasc/coffeegraph
cd coffeegraph
make all    # vet + lint + test + build
```

---

## Roadmap

- [ ] Parallel skill execution in coffee mode
- [ ] Web dashboard (browser-native, no TUI dependency)
- [ ] Scheduled coffee mode (`coffeegraph schedule`)
- [ ] Multi-agency workspace support
- [ ] Homebrew tap for one-line install

---

## License

[MIT](LICENSE) — free to use, fork, and extend.

---

<p align="center">
  Built for founders and solopreneurs who move fast.<br/>
  <strong>Your AI team. Your rules. One folder.</strong>
</p>
