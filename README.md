# Multi Commands (`multi-cmd`)

`multi-cmd` is a Bubble Tea TUI that scans a directory for git repos (or any folders), lets you filter and select them, then runs the commands you chose from `commands.yaml` across each selection. It streams progress in the UI and saves a combined log/report to the output path you provide.

## Run It

```bash
# Build once (optional)
./build.sh

# Run with defaults (scan current dir, use commands.yaml, auto output file)
./multi-cmd

# Or specify paths: scan dir, config file, output file
./multi-cmd ../ commands.yaml results.md
```

## Recommended CLI Tools

The bundled `commands.yaml` expects these binaries to be on your PATH:

- `git` – used for branch, status, commit, and remote checks
- `rg` (ripgrep) – used for text scans
- `cloc` – used for LOC metrics
- `jq` – for JSON parsing/manipulation


