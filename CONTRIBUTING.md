# Contributing to CoffeeGraph

Thanks for your interest in contributing! This guide will help you set up the project and start hacking.

---

## Prerequisites

- **Go 1.22+** — [install instructions](https://go.dev/dl/)
- **golangci-lint** (optional) — `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- **Git**

---

## Getting Started

```bash
git clone https://github.com/angelnicolasc/coffeegraph
cd coffeegraph
go mod tidy
```

### Build the binary

```bash
make build
# or manually:
go build -o coffeegraph ./cmd/coffeegraph
```

### What happens when you build

The binary embeds all templates (`templates/`) and web assets (`web/`) at compile time via `go:embed`. This means:

- **No external files needed at runtime** — the binary is fully self-contained
- **Changes to templates require a rebuild** — if you edit files in `templates/`, you must run `make build` again
- **The `templates` package** (`templates/embed.go`) exposes an `embed.FS` that the init and add commands use to copy skill files into user projects
- **The `web` package** (`web/embed.go`) exposes the `graph.html` visualizer that the server serves

If `go build` fails with template-related errors, run `go mod tidy` first — the embed directives require the files to exist at compile time.

### Run the binary

```bash
./coffeegraph init my-test-agency
cd my-test-agency
../coffeegraph doctor        # check dependencies
../coffeegraph add sales-closer
../coffeegraph dashboard
```

> **Note:** `coffeegraph init` works without an API key. Commands that call the AI (`coffee`, `run`, `suggest`, `evolve`, `roast`, `party`) require an Anthropic API key in `config.yaml` or the `ANTHROPIC_API_KEY` environment variable.

### Run tests

```bash
make test
# or manually:
go test -race -count=1 ./...
```

### Lint

```bash
make lint
# or manually:
golangci-lint run ./...
```

### Generate coverage report

```bash
make coverage
# Opens .coverage/index.html
```

### Run all checks

```bash
make all    # vet + lint + test + build
```

---

## Project Structure

```
cmd/coffeegraph/         Entry point (Cobra CLI commands)
internal/
  cli/                   Command implementations (one file per command)
  graph/                 Graph engine (generator, watcher, patch, validation)
  tui/                   Terminal UI (Bubble Tea + Lip Gloss)
  server/                HTTP + WebSocket server (browser visualizer)
  claude/                Anthropic API client (retries, jitter, exp. backoff)
  queue/                 Task queue (priority, persistence, atomic writes)
  config/                YAML config + env override layer
  project/               Project root detection
  runner/                Task execution engine
  fsutil/                Atomic file write utilities
  logs/                  Execution log management
  doctor/                Health checks
  share/                 Job sharing (GitHub Gist / local HTML)
  evolve/                Self-improving skill suggestions
  bot/                   Telegram bot adapter
  mcp/                   MCP server (JSON-RPC 2.0 over stdio)
  registry/              Community skill registry client
  types/                 Shared interfaces (LLMClient, LLMResult)
templates/               Bundled skill templates (embedded at compile time)
web/                     Browser visualizer (Cytoscape.js, WebSocket)
scripts/                 install.sh — curl | bash installer
```

### Architecture principles

- **Every feature lives in its own package** under `internal/`
- **No global state** — inject dependencies through constructors
- **Context propagation** — every blocking call takes `context.Context`
- **Atomic file writes** — use `fsutil.AtomicWriteFile` for any persistent state (queue, graph, logs)
- **Errors wrap with context** — `fmt.Errorf("component: %w", err)` at every layer

---

## Adding a Skill

1. Create `templates/<your-skill>/`
2. Add `SKILL.md` following the required format:
   - Must contain `# Skill: <name>` header or `## Identity` section
   - Must contain `## Workflow` or `## Output Format` section
   - See any existing skill in `templates/` for reference
3. Optionally add `prompts/` and `n8n/workflow.json`
4. Open a PR — include a sample output in the description

### SKILL.md Validation

The CLI validates SKILL.md files before execution. If a skill is malformed, you'll get a clear error:

```
SKILL.md at skills/my-skill/SKILL.md is missing required structure.
  Expected at least one of:
    • '# Skill: <name>' header
    • '## Identity' section
  See templates/ for examples of valid SKILL.md files
```

---

## Code Conventions

- Use `context.Context` for cancellation in long-running operations
- Use `fsutil.AtomicWriteFile` for any file that could be read by other processes
- Wrap errors with `fmt.Errorf("context: %w", err)` for traceability
- Keep CLI commands thin — logic lives in `internal/` packages
- Every exported function and type should have a godoc comment
- New packages should have at least one `_test.go` file

---

## Troubleshooting

### `go build` fails with "pattern templates: no matching files"

The `go:embed` directives require template files to exist. Make sure you haven't deleted files from `templates/` or `web/`.

```bash
git checkout -- templates/ web/
go build ./cmd/coffeegraph
```

### Tests fail with "no test files" warnings

These warnings are informational, not errors. Packages without `_test.go` files print this message. The test suite still passes if `exit code: 0`.

### `coffeegraph doctor` says "Anthropic API key: WARN"

Set your API key either in `config.yaml` or as an environment variable:

```bash
export ANTHROPIC_API_KEY=sk-ant-...
```

### Binary can't find config.yaml

CoffeeGraph searches upward from the current directory for a `config.yaml` file. Make sure you're running commands from inside a project directory (one created by `coffeegraph init`).

---

## Submitting a PR

1. Fork the repo
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Run `make all` to verify everything passes
4. Commit with descriptive messages (conventional commits preferred)
5. Open a PR against `main`

---

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
