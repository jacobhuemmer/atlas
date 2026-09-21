package cli

const rootHelp = `atlas — Atlassian CLI for Sesami and Garda sites

Usage: atlas <namespace> <verb> [flags]

Namespaces:
  auth      sign in, status, logout (core, not a workload)
  site      list and resolve the three known clouds
  jira      get, search, create, edit, comment, transition, link on one site
  mcp       stdio MCP for agents (serve)

Output: JSON on stdout by default. --human for text. Diagnostics on stderr.
Exit classes: 0 success; 3 usage/config; 4 auth; 5 service; 6 not-found.

Every call resolves one site. SDO/SDP/SES stay on sesamidevel; CAB/CCAB stay
on sesami-io; Garda customer traffic never hits Jira search.
Secrets never appear in stdout, MCP text, or logs.
`

const authHelp = `atlas auth — per-site Basic email:token (keyring service atlas-cli)

Verbs: status, login, logout
status: signed_in, session_usable, sites[] with hostname and role. No tokens.
login:  --site ALIAS --email EMAIL --token TOKEN
        --from-op is a human terminal flag only (not via atlas_run).
logout: optional --site ALIAS (all sites if omitted)
Login and logout stay terminal-only. Do not call them through atlas_run.
Output: JSON (default) or --human.
`

const siteHelp = `atlas site — three-cloud table (no HTTP)

Verbs: list, resolve
list:    sesamidevel, sesami-io, garda (UUID is display-only)
resolve: alias, hostname, or UUID → one Site
REST later uses hostname, not UUID.
Output: JSON (default) or --human.
`

const jiraHelp = `atlas jira — licensed Jira on one cloud

Verbs: get, search, create, edit, comment, transition, link
get:        atlas jira get SDO-1
search:     atlas jira search --jql 'project = CAB' [--site sesami-io]
create:     atlas jira create --project SDO --type Task --summary '...'
edit:       atlas jira edit SDO-1 --fields '{...}'
comment:    atlas jira comment SDO-1 --body '...'
transition: atlas jira transition SDO-1 --name Done
link:       atlas jira link SDO-1 SDP-2 [--type Relates]

SDO/SDP/SES → sesamidevel. CAB → sesami-io. Combined JQL project in (SDO, CAB)
is usage. --site is required only when inference cannot run.
Garda is not Jira search: use atlas jsm, not atlas jira search.
SDP has no Story type. SDP comments are licensed Jira comments (no public flag).
Transitions match by name: SDO/SES Done; SDP Task Mark as done; SDP Incident Resolve.
link default type is Relates. Both keys must be the same cloud (SDO↔SDP ok; SDO↔CAB usage).
When type is Blocks, inward is the blocker and outward is the blocked issue.
MCP writes dry-run unless write_opt_in is true.

Default search fields: summary, description, status, issuetype, priority,
labels, assignee, reporter, created, updated, project.
Browse URL: https://<hostname>/browse/<KEY>
Output: JSON (default) or --human.
`
