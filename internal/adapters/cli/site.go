package cli

import "github.com/masonhuemmer/atlas/internal/domain"

func runSite(args []string, d Deps, format string) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		return writeHelp(d.Stdout, siteHelp)
	}
	verb := args[0]
	switch verb {
	case "list":
		return success(d, format, domain.Sites)
	case "resolve":
		if len(args) < 2 || args[1] == "--help" || args[1] == "-h" {
			return writeHelp(d.Stdout, siteHelp)
		}
		site, err := domain.Lookup(args[1])
		if err != nil {
			return fail(d, err)
		}
		return success(d, format, site)
	default:
		return fail(d, domain.Usagef("unknown site verb %q", verb))
	}
}
