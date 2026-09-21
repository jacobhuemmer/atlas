package cli

const (
	recipeJiraSearch = `jira-search

Resolve one site, then search or get. Never dual-query clouds.

  atlas jira get SDO-1
  atlas jira search --jql 'project = CAB'
  atlas_run namespace=jira verb=get args=["SDO-1"]

SDO/SDP/SES → sesamidevel. CAB → sesami-io. Combined JQL project in (SDO, CAB) is usage.
Garda is not Jira search: use atlas jsm.

Do not print tokens. Do not invent a fourth cloud.
`

	recipeConfluenceWrite = `confluence-write

CCAB lives on sesami-io. No delete verb.

  atlas confluence search --cql 'space = CCAB AND type = page AND title ~ "CAB-109"'
  atlas confluence create --space CCAB --title '…' --body '…' --dry-run
  atlas_run namespace=confluence verb=create flags space=CCAB title=… body=…

MCP writes dry-run unless write_opt_in is true.
`

	recipePRReview = `pr-review

Bitbucket Cloud REST, default workspace sesamiio. No PR delete. SSH stays out.

  atlas pr get --repo atlas --id 1
  atlas pr comment --repo atlas --id 1 --body '…' --dry-run
  atlas pr merge --repo atlas --id 1 --dry-run
  atlas_run namespace=pr verb=get flags repo=atlas id=1

Merge is a write. Reviewer updates are REST fields, not a missing Rovo action.
`

	recipeJSMGarda = `jsm-garda

Garda World is a JSM customer portal. Never Jira search on gardaworld.

  atlas jsm desks --site garda
  atlas jsm list --status open
  atlas jsm comment EOS-1 --body '…' --dry-run
  atlas_run namespace=jsm verb=desks flags site=garda

Comments are public: true only. Credentials are the Garda keyring slot.
sesami-io portal/1 is out of this pass; SDP stays atlas jira.
`
)

var recipeBodies = map[string]string{
	"jira-search":      recipeJiraSearch,
	"confluence-write": recipeConfluenceWrite,
	"pr-review":        recipePRReview,
	"jsm-garda":        recipeJSMGarda,
}

var recipeNames = []string{"jira-search", "confluence-write", "pr-review", "jsm-garda"}

func recipe(topic string) (string, bool) {
	s, ok := recipeBodies[topic]
	return s, ok
}
