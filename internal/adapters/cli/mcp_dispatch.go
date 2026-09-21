package cli

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/masonhuemmer/atlas/internal/domain"
)

func FlagMapToArgs(ns, verb string, pos []string, flags map[string]any) ([]string, error) {
	out := []string{"atlas", ns, verb}
	out = append(out, pos...)
	if len(flags) == 0 {
		return out, nil
	}
	keys := make([]string, 0, len(flags))
	for k := range flags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if k == "" || strings.HasPrefix(k, "-") {
			return nil, domain.Usagef("unknown flag %q", k)
		}
		name := "--" + k
		switch v := flags[k].(type) {
		case nil:
		case bool:
			if v {
				out = append(out, name)
			}
		case string:
			out = append(out, name, v)
		case float64:
			if v == float64(int64(v)) {
				out = append(out, name, strconv.FormatInt(int64(v), 10))
			} else {
				out = append(out, name, strconv.FormatFloat(v, 'f', -1, 64))
			}
		case json.Number:
			out = append(out, name, v.String())
		case []string:
			for _, s := range v {
				out = append(out, name, s)
			}
		case []any:
			for _, e := range v {
				s, ok := e.(string)
				if !ok {
					return nil, domain.Usagef("flag %s values must be strings", k)
				}
				out = append(out, name, s)
			}
		default:
			return nil, domain.Usagef("unsupported flag type for %s", k)
		}
	}
	return out, nil
}

func applyWriteGate(ns, verb string, flags map[string]any, optIn bool) map[string]any {
	if !isWrite(ns, verb) || optIn {
		return flags
	}
	n := map[string]any{"dry-run": true}
	for k, v := range flags {
		n[k] = v
	}
	n["dry-run"] = true
	return n
}

func isWrite(ns, verb string) bool {
	switch ns + " " + verb {
	case "jira create", "jira edit", "jira comment", "jira transition", "jira link",
		"confluence create", "confluence update",
		"pr create", "pr comment", "pr merge",
		"jsm create", "jsm comment", "jsm transition":
		return true
	}
	return false
}

func runForbidden(ns, verb string) error {
	if ns == "mcp" {
		return &domain.Error{Class: domain.ClassUsage, Message: "mcp is not a run namespace", Hint: "use atlas_status, atlas_help, or atlas_run"}
	}
	if ns == "auth" && (verb == "login" || verb == "logout") {
		return &domain.Error{Class: domain.ClassUsage, Message: "auth login is a human terminal command", Hint: "run atlas auth login in a terminal"}
	}
	return nil
}

func buildRunArgs(ns, verb string, pos []string, flags map[string]any, optIn bool) ([]string, error) {
	if ns == "" || verb == "" {
		return nil, domain.Usage("namespace and verb are required")
	}
	if err := runForbidden(ns, verb); err != nil {
		return nil, err
	}
	return FlagMapToArgs(ns, verb, pos, applyWriteGate(ns, verb, flags, optIn))
}
