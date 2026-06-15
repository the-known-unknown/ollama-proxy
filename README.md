# ollama-proxy

A lightweight, single-binary security proxy for a local LLM platform (Ollama).
It sits in front of Ollama and requires an API key on every incoming request,
forwarding authorized traffic upstream and rejecting everything else with `401`.

## Quickstart

```sh
# 1. Clone the repo
git clone https://github.com/asimmittal/ollama-proxy.git
cd ollama-proxy

# 2. Build the binary (requires Go 1.22+)
make build        # produces ./bin/ollama-proxy

# 3. Symlink it into your user bin folder so `ollama-proxy` is on your PATH
mkdir -p "$HOME/bin"
ln -sf "$(pwd)/bin/ollama-proxy" "$HOME/bin/ollama-proxy"
# (ensure $HOME/bin is on your PATH, e.g. add `export PATH="$HOME/bin:$PATH"` to your shell rc)

# 4. Run it with an API key
ollama-proxy --api-key "my-secret-key" --model llama3

# 5. Call it from another terminal
curl http://localhost:11435/v1/models \
  -H "Authorization: Bearer my-secret-key"
```

That's it. The proxy verifies Ollama is reachable (starting `ollama serve` if
needed), checks the model exists, and then forwards authorized requests upstream.

## Features

- API key enforcement via `Authorization: Bearer <key>` or `X-API-Key`
- Transparent reverse proxy with streaming (SSE) support
- Upstream health check before serving
- Optional model validation against `GET /v1/models`
- Auto-starts Ollama (`ollama serve`) if it is not already running
- Zero runtime dependencies — a single static binary

## Requirements

- To build: Go 1.22+ (https://go.dev/dl)
- To run: nothing (the compiled binary is self-contained). For the auto-start
  feature, `ollama` must be installed and on your `PATH`.

## Build

    make build        # produces ./bin/ollama-proxy

Or directly:

    go build -o bin/ollama-proxy ./cmd/ollama-proxy

## Install

    sudo install -m 0755 bin/ollama-proxy /usr/local/bin/ollama-proxy

## Usage

    ollama-proxy --api-key "my-secret-key" --model llama3

Point any OpenAI-compatible client at the proxy instead of Ollama:

    curl http://localhost:11435/v1/models \
      -H "Authorization: Bearer my-secret-key"

### Flags

| Flag | Env var | Default | Description |
|------|---------|---------|-------------|
| `--api-key` | `OLLAMA_PROXY_API_KEY` | _(none)_ | Key required on every request. If omitted you are prompted to confirm insecure mode. |
| `--host` | `OLLAMA_PROXY_HOST` | `http://localhost:11434` | Upstream Ollama base URL. |
| `--model` | `OLLAMA_PROXY_MODEL` | _(none)_ | If set, verified against `GET /v1/models` at startup. If unset, available models are listed. |
| `--platform` | `OLLAMA_PROXY_PLATFORM` | `ollama` | Upstream platform (only `ollama` is supported today). |
| `--port` | `OLLAMA_PROXY_PORT` | `11435` | Port the proxy listens on. |
| `--insecure` | — | `false` | Run without an API key without the interactive prompt. |

Flags take precedence over environment variables, which take precedence over defaults.

### Startup behavior

1. If no API key is set and `--insecure` is not passed, you are asked to confirm.
2. If the upstream platform is not responding, the proxy tries `ollama serve`
   in the background and waits up to 15s for it to become ready.
3. The upstream `/` endpoint must return `200`.
4. If `--model` is set it must exist upstream; otherwise the available models are listed.

## Testing

    make test     # go test ./...
    make vet      # go vet ./...

## Cross-compiling release binaries

    make release                  # builds darwin/linux for amd64 + arm64 into ./bin
    make build VERSION=v1.2.3     # stamp a version into a local build

Check the version baked into a binary:

    ollama-proxy --version

## Continuous integration & releases

Two GitHub Actions workflows live in `.github/workflows/`:

- **`ci.yml`** runs `go vet`, `go test`, and `go build` on every push to `main`
  and on every pull request, across Linux and macOS.
- **`release.yml`** runs [GoReleaser](https://goreleaser.com) whenever a
  `v*` tag is pushed. It cross-compiles binaries for linux/macOS (amd64 + arm64),
  builds `.tar.gz` archives plus a `checksums.txt`, generates a changelog, and
  publishes them to a GitHub Release. The version is taken from the tag and
  stamped into the binary via `-X main.version`.

To cut a release, push a semver tag:

    git tag v0.1.0
    git push origin v0.1.0

The workflow uses the automatically provided `GITHUB_TOKEN`, so no extra secrets
are required. You can validate the GoReleaser config locally with
`goreleaser check` and dry-run a build with `goreleaser release --snapshot --clean`.

## Run as a systemd service (Ubuntu)

`/etc/systemd/system/ollama-proxy.service`:

    [Unit]
    Description=Ollama Proxy
    After=network.target

    [Service]
    Environment=OLLAMA_PROXY_API_KEY=change-me
    ExecStart=/usr/local/bin/ollama-proxy --model llama3
    Restart=on-failure

    [Install]
    WantedBy=multi-user.target

    sudo systemctl daemon-reload
    sudo systemctl enable --now ollama-proxy

## Project layout

    cmd/ollama-proxy/   entry point and orchestration
    internal/config/    CLI/env parsing, validation, prompts
    internal/upstream/  health checks, model listing, platform lifecycle
    internal/auth/      API key middleware
    internal/proxy/     reverse proxy setup

## Notes

The module path is `github.com/asimmittal/ollama-proxy`. If you fork this,
update the path in `go.mod` and the imports in `cmd/ollama-proxy/main.go`.
