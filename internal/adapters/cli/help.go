package cli

const rootHelp = `atlas — Atlassian CLI for Sesami and Garda sites

Usage: atlas <namespace> <verb> [flags]

Namespaces:
  auth         sign in, status, logout (core, not a workload)
  site         list and resolve the three known clouds
  jira         get, search, create, edit, comment, transition, link on one site
  confluence   get, search, create, update (no delete). CCAB on sesami-io
  pr           get, list, create, comment, merge, diff (no delete). workspace sesamiio
  jsm          Garda customer REST: desks, types, list, get, create, comment, transition
  mcp          stdio MCP for agents (serve)

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

const confluenceHelp = `atlas confluence — licensed Confluence on one cloud

Verbs: get, search, create, update
get:    atlas confluence get <pageId> --site sesami-io
search: atlas confluence search --cql 'space = CCAB AND type = page AND title ~ "CAB-109"'
create: atlas confluence create --space CCAB --title '...' --body '...'
update: atlas confluence update <pageId> --body '...' [--site sesami-io]

CCAB lives on sesami-io. Space key CCAB infers that site. Body format is markdown.
CQL that names CCAB infers sesami-io; CQL with no space and no --site is usage.
There is no delete verb.
MCP writes dry-run unless write_opt_in is true.
Live base: https://sesami-io.atlassian.net/wiki/api/v2
Output: JSON (default) or --human.
`

const prHelp = `atlas pr — Bitbucket Cloud REST 2.0

Verbs: get, list, create, comment, merge, diff
get:     atlas pr get --repo atlas --id 1
list:    atlas pr list --repo atlas
create:  atlas pr create --repo atlas --title '...' --source <branch> [--target main]
comment: atlas pr comment --repo atlas --id 1 --body '...'
merge:   atlas pr merge --repo atlas --id 1
diff:    atlas pr diff --repo atlas --id 1

--workspace is optional and defaults to sesamiio. Merge is a write.
There is no delete verb. SSH and git stay out of pr.
Reviewer updates are REST fields on create, not a missing Rovo action.
MCP writes dry-run unless write_opt_in is true.
Live base: https://api.bitbucket.org/2.0/repositories/{workspace}/{repo}/pullrequests
Output: JSON (default) or --human.
`

const jsmHelp = `atlas jsm — Garda JSM customer REST (gardaworld.atlassian.net)

Verbs: desks, types, list, get, create, comment, transition
desks:      atlas jsm desks [--site garda]
types:      atlas jsm types --desk 3
list:       atlas jsm list [--status open|closed|all]
get:        atlas jsm get EOS-1
create:     atlas jsm create --desk 3 --type <id> --summary '...'
comment:    atlas jsm comment EOS-1 --body '...'
transition: atlas jsm transition EOS-1 --id <transition>

--site garda is the only JSM site in v1. Missing --site infers garda.
jsm against sesamidevel or sesami-io is usage (portal/1 deferred; SDP stays atlas jira).
Never Jira search on gardaworld. Comments are public: true only. No raiseOnBehalfOf.
Credentials are the Garda keyring slot, not the licensed Jira token.
Portal URL: https://gardaworld.atlassian.net/servicedesk/customer/portal/{desk}/{KEY}
MCP writes dry-run unless write_opt_in is true.
Live base: https://gardaworld.atlassian.net/rest/servicedeskapi
Output: JSON (default) or --human.
`
