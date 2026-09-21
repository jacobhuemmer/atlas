package cli

const rootHelp = `atlas — Atlassian CLI for Sesami and Garda sites

Usage: atlas <namespace> <verb> [flags]

Namespaces:
  auth      sign in, status, logout (core, not a workload)
  site      list and resolve the three known clouds
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
