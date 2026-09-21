package cli

import (
	"flag"
	"strings"

	"github.com/masonhuemmer/atlas/internal/domain"
)

func runJSM(args []string, d Deps, format string) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		return writeHelp(d.Stdout, jsmHelp)
	}
	verb, args := args[0], args[1:]
	switch verb {
	case "desks":
		return jsmDesks(args, d, format)
	case "types":
		return jsmTypes(args, d, format)
	case "list":
		return jsmList(args, d, format)
	case "get":
		return jsmGet(args, d, format)
	case "create":
		return jsmCreate(args, d, format)
	case "comment":
		return jsmComment(args, d, format)
	case "transition":
		return jsmTransition(args, d, format)
	default:
		return fail(d, domain.Usagef("unknown jsm verb %q", verb).WithHint("atlas jsm desks|types|list|get|create|comment|transition"))
	}
}

func jsmDesks(args []string, d Deps, format string) int {
	if hasHelp(args) {
		return writeHelp(d.Stdout, jsmHelp)
	}
	fsset := flag.NewFlagSet("jsm desks", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	siteFlag := fsset.String("site", "", "site alias (defaults.jsm_site when set)")
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	site, err := resolveJSMSite(*siteFlag)
	if err != nil {
		return fail(d, err)
	}
	if d.JSM == nil {
		return fail(d, domain.Service("jsm adapter not configured"))
	}
	page, err := d.JSM.Desks(ctx(), site.Hostname)
	if err != nil {
		return fail(d, err)
	}
	return success(d, format, page)
}

func jsmTypes(args []string, d Deps, format string) int {
	if hasHelp(args) {
		return writeHelp(d.Stdout, jsmHelp)
	}
	fsset := flag.NewFlagSet("jsm types", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	siteFlag := fsset.String("site", "", "site alias (defaults.jsm_site when set)")
	desk := fsset.String("desk", "", "serviceDeskId")
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	deskID := strings.TrimSpace(*desk)
	if deskID == "" {
		deskID = strings.TrimSpace(fsset.Arg(0))
	}
	if deskID == "" {
		return fail(d, domain.Usage("types requires --desk").WithHint("atlas jsm types --desk 3"))
	}
	site, err := resolveJSMSite(*siteFlag)
	if err != nil {
		return fail(d, err)
	}
	if d.JSM == nil {
		return fail(d, domain.Service("jsm adapter not configured"))
	}
	page, err := d.JSM.Types(ctx(), site.Hostname, deskID)
	if err != nil {
		return fail(d, err)
	}
	return success(d, format, page)
}

func jsmList(args []string, d Deps, format string) int {
	if hasHelp(args) {
		return writeHelp(d.Stdout, jsmHelp)
	}
	fsset := flag.NewFlagSet("jsm list", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	siteFlag := fsset.String("site", "", "site alias (defaults.jsm_site when set)")
	status := fsset.String("status", "", "open, closed, or all (default all)")
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	site, err := resolveJSMSite(*siteFlag)
	if err != nil {
		return fail(d, err)
	}
	st, err := domain.NormalizeJSMStatus(*status)
	if err != nil {
		return fail(d, err)
	}
	if d.JSM == nil {
		return fail(d, domain.Service("jsm adapter not configured"))
	}
	page, err := d.JSM.List(ctx(), site.Hostname, st)
	if err != nil {
		return fail(d, err)
	}
	return success(d, format, page)
}

func jsmGet(args []string, d Deps, format string) int {
	if hasHelp(args) {
		return writeHelp(d.Stdout, jsmHelp)
	}
	fsset := flag.NewFlagSet("jsm get", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	siteFlag := fsset.String("site", "", "site alias (defaults.jsm_site when set)")
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	key := fsset.Arg(0)
	if strings.TrimSpace(key) == "" {
		return fail(d, domain.Usage("request key is required").WithHint("atlas jsm get KEY-1"))
	}
	site, err := resolveJSMSite(*siteFlag)
	if err != nil {
		return fail(d, err)
	}
	if d.JSM == nil {
		return fail(d, domain.Service("jsm adapter not configured"))
	}
	req, err := d.JSM.Get(ctx(), site.Hostname, key)
	if err != nil {
		return fail(d, err)
	}
	return success(d, format, req)
}

func jsmCreate(args []string, d Deps, format string) int {
	if hasHelp(args) {
		return writeHelp(d.Stdout, jsmHelp)
	}
	fsset := flag.NewFlagSet("jsm create", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	siteFlag := fsset.String("site", "", "site alias (defaults.jsm_site when set)")
	desk := fsset.String("desk", "", "serviceDeskId")
	typ := fsset.String("type", "", "requestTypeId")
	summary := fsset.String("summary", "", "summary")
	description := fsset.String("description", "", "description")
	dry := fsset.Bool("dry-run", false, "")
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	if strings.TrimSpace(*desk) == "" || strings.TrimSpace(*typ) == "" || strings.TrimSpace(*summary) == "" {
		return fail(d, domain.Usage("create requires --desk, --type, and --summary").WithHint("atlas jsm create --desk 3 --type 40 --summary '…'"))
	}
	site, err := resolveJSMSite(*siteFlag)
	if err != nil {
		return fail(d, err)
	}
	if d.JSM == nil {
		return fail(d, domain.Service("jsm adapter not configured"))
	}
	in := domain.CreateRequest{
		DeskID:      strings.TrimSpace(*desk),
		TypeID:      strings.TrimSpace(*typ),
		Summary:     strings.TrimSpace(*summary),
		Description: *description,
	}
	req, err := d.JSM.Create(ctx(), site.Hostname, in, *dry)
	if err != nil {
		return fail(d, err)
	}
	if *dry {
		return success(d, format, jsmDryRun{
			DryRun:    true,
			Namespace: "jsm",
			Verb:      "create",
			Desk:      in.DeskID,
			Type:      in.TypeID,
			Summary:   in.Summary,
		})
	}
	return success(d, format, req)
}

func jsmComment(args []string, d Deps, format string) int {
	if hasHelp(args) {
		return writeHelp(d.Stdout, jsmHelp)
	}
	fsset := flag.NewFlagSet("jsm comment", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	siteFlag := fsset.String("site", "", "site alias (defaults.jsm_site when set)")
	body := fsset.String("body", "", "public comment body")
	dry := fsset.Bool("dry-run", false, "")
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	key := fsset.Arg(0)
	if strings.TrimSpace(key) == "" {
		return fail(d, domain.Usage("request key is required").WithHint("atlas jsm comment KEY-1 --body '…'"))
	}
	if strings.TrimSpace(*body) == "" {
		return fail(d, domain.Usage("comment requires --body").WithHint("atlas jsm comment KEY-1 --body '…'"))
	}
	site, err := resolveJSMSite(*siteFlag)
	if err != nil {
		return fail(d, err)
	}
	if d.JSM == nil {
		return fail(d, domain.Service("jsm adapter not configured"))
	}
	if err := d.JSM.Comment(ctx(), site.Hostname, key, *body, *dry); err != nil {
		return fail(d, err)
	}
	if *dry {
		return success(d, format, jsmDryRun{
			DryRun:    true,
			Namespace: "jsm",
			Verb:      "comment",
			Key:       strings.ToUpper(strings.TrimSpace(key)),
			Body:      *body,
			Public:    true,
		})
	}
	return success(d, format, map[string]any{
		"key":    strings.ToUpper(strings.TrimSpace(key)),
		"body":   *body,
		"public": true,
	})
}

func jsmTransition(args []string, d Deps, format string) int {
	if hasHelp(args) {
		return writeHelp(d.Stdout, jsmHelp)
	}
	fsset := flag.NewFlagSet("jsm transition", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	siteFlag := fsset.String("site", "", "site alias (defaults.jsm_site when set)")
	id := fsset.String("id", "", "transition id")
	dry := fsset.Bool("dry-run", false, "")
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	key := fsset.Arg(0)
	if strings.TrimSpace(key) == "" {
		return fail(d, domain.Usage("request key is required").WithHint("atlas jsm transition KEY-1 --id 21"))
	}
	if strings.TrimSpace(*id) == "" {
		return fail(d, domain.Usage("transition requires --id").WithHint("atlas jsm transition KEY-1 --id 21"))
	}
	site, err := resolveJSMSite(*siteFlag)
	if err != nil {
		return fail(d, err)
	}
	if d.JSM == nil {
		return fail(d, domain.Service("jsm adapter not configured"))
	}
	req, err := d.JSM.Transition(ctx(), site.Hostname, key, *id, *dry)
	if err != nil {
		return fail(d, err)
	}
	if *dry {
		return success(d, format, jsmDryRun{
			DryRun:    true,
			Namespace: "jsm",
			Verb:      "transition",
			Key:       strings.ToUpper(strings.TrimSpace(key)),
			ID:        strings.TrimSpace(*id),
		})
	}
	return success(d, format, req)
}

type jsmDryRun struct {
	DryRun    bool   `json:"dry_run"`
	Namespace string `json:"namespace"`
	Verb      string `json:"verb"`
	Desk      string `json:"desk,omitempty"`
	Type      string `json:"type,omitempty"`
	Summary   string `json:"summary,omitempty"`
	Key       string `json:"key,omitempty"`
	Body      string `json:"body,omitempty"`
	Public    bool   `json:"public,omitempty"`
	ID        string `json:"id,omitempty"`
}

func resolveJSMSite(flag string) (domain.Site, error) {
	flag = strings.TrimSpace(flag)
	if flag == "" {
		flag = domain.DefaultJSMSite()
		if flag == "" {
			return domain.Site{}, domain.Usage("jsm requires --site").WithHint("set defaults.jsm_site or pass --site ALIAS")
		}
	}
	return domain.LookupJSM(flag)
}
