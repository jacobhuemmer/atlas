---
name: atlas
description: >
  Use dedicated atlas MCP for Jira, Confluence, Bitbucket, and JSM customer
  REST on one configured cloud per call. Use when calling atlas_status,
  atlas_help, or atlas_run, or when the operator names atlas, Jira,
  Confluence, Bitbucket PRs, or a JSM customer portal.
---

# Atlas MCP

Three tools on stdio. Prefer them over a shell `atlas` one-off when the
session has the atlas server. Sites, project keys, space keys, Bitbucket
workspace, and the default JSM site live in the operator catalog
(`ATLAS_CONFIG` or `$XDG_CONFIG_HOME/atlas/config.toml`). Login stays
`atlas auth login` in a terminal.

| Tool | When |
|------|------|
| `atlas_status` | Signed-in, per-site role, and per-workspace usability. No tokens. Does not open a browser. |
| `atlas_help` | Namespace/verb help, or recipe `topic`. No session. |
| `atlas_run` | One CLI namespace+verb. Returns that command's JSON. |

Do not add per-verb tools. Do not invent Atlassian REST.

## Call shape

`atlas_run` input: `namespace`, `verb`, optional `args`, `flags`,
`write_opt_in`. Flag names are long flags without dashes (`jql`, `site`,
`dry-run`). Writes stay dry-run unless `write_opt_in` is true.

Forbidden via `atlas_run`: `auth login`, `auth logout`, namespace `mcp`.
Those are terminal commands.

Failure content is `{class,message,hint}` with `usage` | `auth` | `service` | `not_found`. Site auth → tell the operator to run `atlas auth login --site ALIAS`. PR auth → tell the operator to run `atlas auth login --workspace WORKSPACE`. Do not print tokens.

Recipes (`atlas_help` `topic`, also MCP prompts): `jira-search`,
`confluence-write`, `pr-review`, `jsm-customer`. This skill is also
`atlas_help` `topic=atlas` and MCP resource `atlas://skill`.

## Site isolation

Every call resolves **one** site from the catalog. Project keys and space
keys map to an alias. JQL/CQL that names two sites is usage. A
`jsm_customer` site refuses `jira` and `confluence` (use `jsm`). A
`licensed` site refuses `jsm`.

## Workloads

| Need | `atlas_run` |
|------|-------------|
| Status | `atlas_status` (not run) |
| Help / recipe | `atlas_help` `topic=jira-search` (or namespace/verb) |
| Get issue | `namespace=jira` `verb=get` `args=["KEY-1"]` |
| Search Jira | `namespace=jira` `verb=search` `flags={jql:"project = KEY"}` |
| Create issue | `namespace=jira` `verb=create` `flags={project,type,summary}` + `write_opt_in` |
| Comment / transition / link | `jira` `comment` / `transition` / `link` |
| Confluence get/search | `confluence` `get` / `search` (`cql`) |
| Confluence create/update | writes; no delete |
| PR get/list/diff | `pr` `get` / `list` / `diff` (`repo`, `id`) |
| PR create/comment/merge | writes; no delete |
| JSM desks/list/get | `jsm` `desks` / `list` / `get` |
| JSM create/comment/transition | writes; comments are public |

## Preflight

If `atlas_status` shows the needed site or workspace `usable` false, stop and tell the operator to use the matching terminal login command. Sites use `atlas auth login --site ALIAS --email EMAIL --token TOKEN`; PR workspaces use `atlas auth login --workspace WORKSPACE --email EMAIL --token TOKEN`. Do not loop on login. A site credential never authorizes a PR operation.
