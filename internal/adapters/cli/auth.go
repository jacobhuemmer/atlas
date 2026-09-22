package cli

import (
	"flag"
	"strings"

	"github.com/masonhuemmer/atlas/internal/app/auth"
	"github.com/masonhuemmer/atlas/internal/domain"
)

func runAuth(args []string, d Deps, format string, verbose bool) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		return writeHelp(d.Stdout, authHelp)
	}
	verb := args[0]
	switch verb {
	case "status":
		st, err := auth.Status(d.Store)
		if err != nil {
			return fail(d, err)
		}
		if verbose {
			fmtVerbose(d, "status ok")
		}
		return success(d, format, statusOut(st))
	case "logout":
		fsset := flag.NewFlagSet("logout", flag.ContinueOnError)
		fsset.SetOutput(d.Stderr)
		siteFlag := fsset.String("site", "", "site alias to forget")
		workspace := fsset.String("workspace", "", "Bitbucket workspace to forget")
		if err := parseMixed(fsset, args[1:]); err != nil {
			return fail(d, domain.Usage(err.Error()))
		}
		if fsset.NArg() != 0 {
			return fail(d, domain.Usage("auth logout does not accept positional arguments"))
		}
		siteSet := flagWasSet(fsset, "site")
		workspaceSet := flagWasSet(fsset, "workspace")
		if siteSet && workspaceSet {
			return fail(d, domain.Usage("use only one of --site or --workspace"))
		}
		var err error
		switch {
		case siteSet:
			if strings.TrimSpace(*siteFlag) == "" {
				return fail(d, domain.Usage("--site requires a value"))
			}
			site, lerr := domain.Lookup(*siteFlag)
			if lerr != nil {
				return fail(d, lerr)
			}
			err = auth.LogoutSite(d.Store, site.Alias)
		case workspaceSet:
			if strings.TrimSpace(*workspace) == "" {
				return fail(d, domain.Usage("--workspace requires a value"))
			}
			err = auth.LogoutWorkspace(d.Store, *workspace)
		default:
			err = auth.Logout(d.Store)
		}
		if err != nil {
			return fail(d, err)
		}
		st, err := auth.Status(d.Store)
		if err != nil {
			return fail(d, err)
		}
		return success(d, format, statusOut(st))
	case "login":
		fsset := flag.NewFlagSet("login", flag.ContinueOnError)
		fsset.SetOutput(d.Stderr)
		siteFlag := fsset.String("site", "", "site alias, hostname, or UUID")
		workspace := fsset.String("workspace", "", "Bitbucket workspace slug")
		email := fsset.String("email", "", "Atlassian email")
		token := fsset.String("token", "", "API token")
		fromOp := fsset.Bool("from-op", false, "seed from 1Password (human terminal only)")
		if err := parseMixed(fsset, args[1:]); err != nil {
			return fail(d, domain.Usage(err.Error()))
		}
		if *fromOp {
			return fail(d, domain.Usage("auth login --from-op is a human terminal command").WithHint("run atlas auth login --site ALIAS or --workspace WORKSPACE with --email EMAIL --token TOKEN in a terminal; the agent path never calls op"))
		}
		siteSet := flagWasSet(fsset, "site")
		workspaceSet := flagWasSet(fsset, "workspace")
		if siteSet && workspaceSet {
			return fail(d, domain.Usage("use only one of --site or --workspace"))
		}
		if !siteSet && !workspaceSet {
			return fail(d, domain.Usage("login requires --site or --workspace").WithHint("atlas auth login --site ALIAS --email EMAIL --token TOKEN"))
		}
		var st domain.Session
		var err error
		if workspaceSet {
			if strings.TrimSpace(*workspace) == "" {
				return fail(d, domain.Usage("--workspace requires a value"))
			}
			st, err = auth.LoginWorkspace(ctx(), d.Store, d.Login, *workspace, *email, *token)
		} else {
			if strings.TrimSpace(*siteFlag) == "" {
				return fail(d, domain.Usage("--site requires a value"))
			}
			site, lerr := domain.Lookup(*siteFlag)
			if lerr != nil {
				return fail(d, lerr)
			}
			st, err = auth.Login(ctx(), d.Store, d.Login, site, *email, *token)
		}
		if err != nil {
			return fail(d, err)
		}
		return success(d, format, statusOut(st))
	default:
		return fail(d, domain.Usagef("unknown auth verb %q", verb))
	}
}

func statusOut(st domain.Session) map[string]any {
	return map[string]any{
		"signed_in":      st.SignedIn,
		"session_usable": st.SessionUsable,
		"sites":          st.Sites,
		"workspaces":     st.Workspaces,
	}
}

func fmtVerbose(d Deps, s string) {
	_, _ = d.Stderr.Write([]byte(redact(s) + "\n"))
}

func flagWasSet(fsset *flag.FlagSet, name string) bool {
	set := false
	fsset.Visit(func(f *flag.Flag) {
		if f.Name == name {
			set = true
		}
	})
	return set
}
