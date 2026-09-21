# MCP catalog

Public interface: `atlas mcp serve` on stdio JSON-RPC. Human CLI remains `atlas <namespace> <verb> [flags]`. This catalog is the MCP door, not an Atlassian REST dump.

`tools/list` MUST return exactly these three names.

Transport: stdio only. No SSE, no Streamable HTTP.

## Tools

| Name | Description (MUST convey) | Session |
| --- | --- | --- |
| `atlas_status` | Signed-in, session usable, per-site role. No tokens. Does not open a browser. | Optional |
| `atlas_help` | CLI help for a namespace or verb, or a recipe `topic` (`jira-search`, `confluence-write`, `pr-review`, `jsm-garda`). No session required. | None |
| `atlas_run` | Run one CLI namespace+verb with a flag map. Returns that command's JSON. Writes dry-run unless `write_opt_in` is true. Lookup examples: help topics jira-search, confluence-write, pr-review, jsm-garda. | Required for workloads |

Unknown tool name → MCP protocol error. Do not add `atlas_login`, or per-verb tools.

## `atlas_status`

**Input**: empty object.

**Output**: same JSON as `atlas auth status`. Keys: `signed_in`, `session_usable`, `sites[]` with `alias`, `hostname`, `role` (`licensed` | `jsm_customer`), `usable`. MUST NOT include token fields.

Signed out → success, `signed_in` false, no interactive login.

## `atlas_help`

**Input**:

| Field | Type | Required |
| --- | --- | --- |
| `namespace` | string | no |
| `verb` | string | no |
| `topic` | string | no (`jira-search` \| `confluence-write` \| `pr-review` \| `jsm-garda`) |

No args → overview: three tools, four recipe topics, write opt-in. `topic` set → recipe text. `namespace` / `verb` → CLI help. `topic` and `namespace` together → usage. Unknown topic → usage. No session. Text, not JSON.

## Named recipes (MCP prompts)

`prompts/list` MUST return exactly: `jira-search`, `confluence-write`, `pr-review`, `jsm-garda`. `prompts/get` returns the same body as `atlas_help` for that topic. MUST NOT add a fourth tool.

## `atlas_run`

**Input**: `namespace`, `verb`, optional `args`, `flags`, `write_opt_in`.

**Output (success)**: one text content item = CLI JSON stdout for the equivalent command (help verbs: help text).

**Output (failure)**: `isError` true; one text content item `{class,message,hint}`. Classes MUST remain `usage` | `auth` | `service` | `not_found`.

### Write gate

Write list: `jira create|edit|comment|transition|link`, `confluence create|update`, `pr create|comment|merge`, `jsm create|comment|transition`. Writes dry-run unless `write_opt_in` is true.

### Forbidden via run

`auth login`, `auth logout`, namespace `mcp` → `usage` with hint to run `atlas auth login` in a terminal.

## CLI surface

| Command | Behavior |
| --- | --- |
| `atlas --help` | Lists `auth`, `site`, `jira`, `confluence`, `pr`, `jsm`, `mcp`. JSON, exit classes 3/4/5/6. |
| `atlas mcp --help` / `atlas mcp serve --help` | Exit 0, no session. Names stdio, three tools, write opt-in default false, four recipe topics. |
| `atlas mcp serve` | JSON-RPC on stdio until stdin closes. `--human` → usage (3). |
| `atlas auth status` | Signed-out JSON with three sites, no tokens. |

## Error mapping

| CLI exit | MCP |
| --- | --- |
| 0 | `isError` false, stdout text |
| 3 | `isError` true, `class=usage` |
| 4 | `isError` true, `class=auth` |
| 5 | `isError` true, `class=service` |
| 6 | `isError` true, `class=not_found` |

Do not map these onto JSON-RPC error codes.

## Secrets

Redact access tokens, refresh tokens, authorization codes, API tokens, and client secrets in every MCP content item and log line.
