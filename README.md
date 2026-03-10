# pd

CLI for progressive discovery of project documents.

`pd` scans a docs directory in a Git repository, reads YAML frontmatter from Markdown files, and outputs structured metadata as JSON.

## Requirements

- Go 1.25+
- Git repository

## Build

```bash
task build
```

## Usage

Run `pd` from anywhere inside a Git repository. It automatically finds the repository root.

```
Usage: pd <command> [flags]

Flags:
  --root="docs"    Directory to scan, relative to repository root.

Commands:
  list    List discovery metadata from docs directory.
```

### `pd list`

Lists all valid documents under `--root` and outputs a JSON array to stdout. Invalid documents are reported as JSON to stderr.

```bash
pd list
pd list --kind adr
pd list --root docs/adr
```

Valid `--kind` values: `roadmap`, `design-doc`, `adr`, `coding`, `testing`, `tooling`, `review`, `unknown`

**stdout** (success):
```json
[
  {"path": "docs/roadmap.md", "kind": "roadmap", "title": "Roadmap", "description": "..."}
]
```

**stderr** (invalid documents, non-fatal):
```json
{"path": "docs/draft.md", "reason": "missing required field: kind"}
```

## Frontmatter Format

Each document must have a YAML frontmatter block:

```markdown
---
kind: adr
title: "Adopt goccy/go-yaml"       # optional — falls back to first H1 heading
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
