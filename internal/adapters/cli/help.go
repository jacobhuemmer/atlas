package cli

const rootHelp = `atlas — Atlassian CLI (Jira, Confluence, Bitbucket, JSM customer REST)

Usage: atlas <namespace> <verb> [flags]

Namespaces:
  auth         sign in, status, logout (core, not a workload)
  site         list and resolve clouds from config
  jira         get, search, create, edit, comment, transition, link on one site
  confluence   get, search, create, update (no delete)
  pr           get, list, create, comment, merge, diff (no delete)
  jsm          customer REST: desks, types, list, get, create, comment, transition
  mcp          stdio MCP for agents (serve)

Output: JSON on stdout by default. --human for text. Diagnostics on stderr.
Exit classes: 0 success; 3 usage/config; 4 auth; 5 service; 6 not-found.

Every call resolves one site from config. Sites, project keys, space keys,
Bitbucket workspace, and JSM default site live in config.toml (or ATLAS_CONFIG).
A site with role jsm_customer refuses jira and confluence (use atlas jsm).
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

const siteHelp = `atlas site — catalog from config (no HTTP)

Verbs: list, resolve
list:    aliases loaded from ATLAS_CONFIG or ~/.config/atlas/config.toml
resolve: alias, hostname, or UUID → one Site
REST later uses hostname, not UUID.
Output: JSON (default) or --human.
`

const jiraHelp = `atlas jira — licensed Jira on one cloud

Verbs: get, search, create, edit, comment, transition, link
get:        atlas jira get KEY-1
search:     atlas jira search --jql 'project = KEY' [--site ALIAS]
create:     atlas jira create --project KEY --type Task --summary '...'
edit:       atlas jira edit KEY-1 --fields '{...}'
comment:    atlas jira comment KEY-1 --body '...'
transition: atlas jira transition KEY-1 --name Done
link:       atlas jira link KEY-1 OTHER-2 [--type Relates]

Project keys map to a site in config. Combined JQL that names two sites is usage.
--site is required only when inference cannot run.
A jsm_customer site is not Jira search: use atlas jsm, not atlas jira search.
link default type is Relates. Both keys must be the same cloud.
When type is Blocks, inward is the blocker and outward is the blocked issue.
MCP writes dry-run unless write_opt_in is true.

Default search fields: summary, description, status, issuetype, priority,
labels, assignee, reporter, created, updated, project.
Browse URL: https://<hostname>/browse/<KEY>
Output: JSON (default) or --human.
`

const confluenceHelp = `atlas confluence — licensed Confluence on one cloud

Verbs: get, search, create, update
get:    atlas confluence get <pageId> --site ALIAS
search: atlas confluence search --cql 'space = KEY AND type = page'
create: atlas confluence create --space KEY --title '...' --body '...'
update: atlas confluence update <pageId> --body '...' [--site ALIAS]

Space keys map to a site in config. Body format is markdown.
CQL that names a configured space infers that site; CQL with no space and no --site is usage.
There is no delete verb.
MCP writes dry-run unless write_opt_in is true.
Live base: https://<hostname>/wiki/api/v2
Output: JSON (default) or --human.
`

const prHelp = `atlas pr — Bitbucket Cloud REST 2.0

Verbs: get, list, create, comment, merge, diff
get:     atlas pr get --repo SLUG --id 1
list:    atlas pr list --repo SLUG
create:  atlas pr create --repo SLUG --title '...' --source <branch> [--target main]
comment: atlas pr comment --repo SLUG --id 1 --body '...'
merge:   atlas pr merge --repo SLUG --id 1
diff:    atlas pr diff --repo SLUG --id 1

--workspace is optional and defaults to config defaults.workspace. Merge is a write.
There is no delete verb. SSH and git stay out of pr.
Reviewer updates are REST fields on create.
MCP writes dry-run unless write_opt_in is true.
Live base: https://api.bitbucket.org/2.0/repositories/{workspace}/{repo}/pullrequests
Output: JSON (default) or --human.
`

const jsmHelp = `atlas jsm — JSM customer REST

Verbs: desks, types, list, get, create, comment, transition
desks:      atlas jsm desks [--site ALIAS]
types:      atlas jsm types --desk ID
list:       atlas jsm list [--status open|closed|all]
get:        atlas jsm get KEY-1
create:     atlas jsm create --desk ID --type <id> --summary '...'
comment:    atlas jsm comment KEY-1 --body '...'
transition: atlas jsm transition KEY-1 --id <transition>

Missing --site uses defaults.jsm_site when set; otherwise --site is required.
jsm against a licensed site is usage. Comments are public: true only. No raiseOnBehalfOf.
Portal URL: https://<hostname>/servicedesk/customer/portal/{desk}/{KEY}
MCP writes dry-run unless write_opt_in is true.
Live base: https://<hostname>/rest/servicedeskapi
Output: JSON (default) or --human.
`
