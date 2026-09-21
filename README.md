# atlas

Atlassian CLI for Jira, Confluence, Bitbucket, and JSM customer REST. JSON on stdout by default; `--human` for text.

## Install

From source (Go 1.25+):

```sh
make install
```

Homebrew (HEAD):

```sh
brew tap jacobhuemmer/atlas https://github.com/jacobhuemmer/atlas
brew install --HEAD jacobhuemmer/atlas/atlas
```

Upgrade later with `brew upgrade --fetch-HEAD jacobhuemmer/atlas/atlas`.

## Config

Write `$XDG_CONFIG_HOME/atlas/config.toml` (or `~/.config/atlas/config.toml`), or set `ATLAS_CONFIG` to a toml/json file. Sites, project keys, space keys, Bitbucket workspace, and the default JSM site live there. The binary does not ship a tenant table.

```toml
[defaults]
workspace = "WORKSPACE"
jsm_site = "ALIAS"

[[sites]]
alias = "ALIAS"
hostname = "example.atlassian.net"
uuid = "00000000-0000-0000-0000-000000000000"
role = "licensed"          # or jsm_customer

[projects]
KEY = "ALIAS"

[spaces]
SPACE = "ALIAS"
```

Missing config without `ATLAS_FAKE=1` is usage (exit 3).

## Auth

Per-site Basic `email:token` in the macOS keychain (file fallback if needed). Login stays in a terminal.

```sh
atlas auth login --site ALIAS --email EMAIL --token TOKEN
atlas auth status
```

`--from-op` is a human `login` flag only. Do not call `auth login` through MCP.

## Usage

```
atlas <namespace> <verb> [flags]
```

| Namespace | Verbs |
| --- | --- |
| `auth` | `login`, `status`, `logout` |
| `site` | `list`, `resolve` |
| `jira` | `get`, `search`, `create`, `edit`, `comment`, `transition`, `link` |
| `confluence` | `get`, `search`, `create`, `update` |
| `pr` | `get`, `list`, `create`, `comment`, `merge`, `diff` |
| `jsm` | `desks`, `types`, `list`, `get`, `create`, `comment`, `transition` |
| `mcp` | `serve` |

Every call resolves one site from config. A `jsm_customer` site refuses `jira` and `confluence` (use `atlas jsm`). Exit classes: `0` success, `3` usage/config, `4` auth, `5` service, `6` not-found.

## MCP

```sh
atlas mcp serve
```

Stdio JSON-RPC for agents. Tools: `atlas_status`, `atlas_help`, `atlas_run`. The embedded agent skill is available from `atlas_help` with `topic=atlas` and as the `atlas://skill` resource; it is not a fifth prompt. Do not pass `--human`. Writes through `atlas_run` stay dry-run unless `write_opt_in` is true. Login stays `atlas auth login` in a terminal.

## Develop

```sh
make verify
```

CI runs `make verify` and builds `./cmd/atlas` on `macos-latest` for pushes and PRs to `main`.

## License

MIT
