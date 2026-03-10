# pd

CLI for progressive discovery of project documents.

`pd` scans Markdown documents under the current directory, reads YAML frontmatter, and outputs structured metadata as JSON.

## `AGENTS.md` / `CLAUDE.md` Integration

Example `AGENTS.md` / `CLAUDE.md` snippet:

```markdown
- For project document discovery, always run `pd list` first. Read document contents only with `pd show --body <path>` for paths selected from that output.
```

## Requirements

- Go 1.25+

## Installation

```bash
go install github.com/shuymn/pd@latest
```

## Usage

Run `pd` from the directory you want to inspect.

```
Usage: pd <command> [flags]

Flags:
  --root="."       Directory to scan, relative to the current directory.
  --depth=3        Limit pd list traversal depth relative to discovery root.
  --verbose        Emit list diagnostics to stderr.

Commands:
  list    List discovery metadata from docs directory.
  show    Show discovery metadata for a single document.
```

### `pd list`

Lists all valid documents under `--root` and outputs a JSON array to stdout. By default, traversal stops at depth 3 relative to the discovery root. In directories managed by Git, files and directories ignored by `.gitignore` or `.git/info/exclude` are skipped. Invalid documents are hidden by default and can be emitted as JSON to stderr with `--verbose`.

```bash
pd list
pd list --verbose
pd list --depth 1
pd list --kind adr
pd list --root docs/adr
pd list --root docs/adr --depth 1
```

Valid `--kind` values: `roadmap`, `design-doc`, `adr`, `coding`, `testing`, `tooling`, `review`, `unknown`

**stdout** (success):
```json
[
  {"path": "docs/roadmap.md", "kind": "roadmap", "title": "Roadmap", "description": "..."}
]
```

**stderr** (`--verbose` only, invalid documents, non-fatal):
```json
{"path": "docs/draft.md", "reason": "missing required field: kind"}
```

### `pd show`

Shows discovery metadata for one document.

```bash
pd show docs/adr/001.md
pd show docs/adr/001.md --body
pd show --root docs/adr 001.md
```

**stdout** (success):
```json
{"path": "001.md", "kind": "adr", "title": "Decision", "description": "..."}
```

**stderr** (invalid or missing document, fatal):
```json
{"path": "001.md", "reason": "document not found"}
```

## Frontmatter Format

Each document must have a YAML frontmatter block:

```markdown
---
kind: adr
title: "Adopt shuymn/pd" # optional — falls back to first H1 heading
description: "Decision rationale."
---
```

## Development

```bash
task          # list all tasks
task build    # build binary
task test     # run tests
task check    # lint + build + test
task fmt      # format code
```

Git hooks are managed with lefthook:

```bash
lefthook install
```
