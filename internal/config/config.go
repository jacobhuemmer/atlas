package config

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/masonhuemmer/atlas/internal/domain"
)

// Config is the loaded atlas config. Catalog is applied to domain on Load.
type Config struct {
	domain.Catalog
	Path string
}

type fileShape struct {
	Defaults *fileDefaults     `json:"defaults" toml:"defaults"`
	Sites    []fileSite        `json:"sites" toml:"sites"`
	Projects map[string]string `json:"projects" toml:"projects"`
	Spaces   map[string]string `json:"spaces" toml:"spaces"`
}

type fileDefaults struct {
	Workspace         string `json:"workspace" toml:"workspace"`
	DefaultSite       string `json:"default_site" toml:"default_site"`
	AssigneeAccountID string `json:"assignee_account_id" toml:"assignee_account_id"`
	JSMSite           string `json:"jsm_site" toml:"jsm_site"`
}

type fileSite struct {
	Alias    string `json:"alias" toml:"alias"`
	Hostname string `json:"hostname" toml:"hostname"`
	UUID     string `json:"uuid" toml:"uuid"`
	Role     string `json:"role" toml:"role"`
}

// Load reads ATLAS_CONFIG, then XDG toml/json. ATLAS_FAKE=1 loads testdata/fake.toml
// unless ATLAS_CONFIG is set. Missing live config is usage.
func Load() (Config, error) {
	if path := strings.TrimSpace(os.Getenv("ATLAS_CONFIG")); path != "" {
		return loadPath(path)
	}
	if os.Getenv("ATLAS_FAKE") == "1" {
		return loadFake()
	}
	for _, path := range searchPaths() {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		return loadPath(path)
	}
	return Config{}, domain.Usage("atlas config not found").WithHint("write $XDG_CONFIG_HOME/atlas/config.toml or set ATLAS_CONFIG to a toml/json file")
}

// LoadFake applies the three-cloud test fixture. Tests and ATLAS_FAKE=1 use it.
func LoadFake() (Config, error) {
	return loadFake()
}

func loadFake() (Config, error) {
	path, err := fakeConfigPath()
	if err != nil {
		return Config{}, err
	}
	return loadPath(path)
}

func fakeConfigPath() (string, error) {
	var candidates []string
	if _, file, _, ok := runtime.Caller(0); ok {
		candidates = append(candidates, filepath.Join(filepath.Dir(file), "testdata", "fake.toml"))
	}
	candidates = append(candidates,
		filepath.Join("testdata", "fake.toml"),
		filepath.Join("internal", "config", "testdata", "fake.toml"),
	)
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", domain.Usage("ATLAS_FAKE testdata/fake.toml not found").WithHint("run from the atlas module or set ATLAS_CONFIG")
}

func loadPath(path string) (Config, error) {
	b, err := os.ReadFile(path) // #nosec G304 -- operator-supplied config path
	if err != nil {
		return Config{}, domain.Usagef("cannot read config %s", path).WithHint("set ATLAS_CONFIG to an existing toml or json file")
	}
	cat, err := parse(path, b)
	if err != nil {
		return Config{}, err
	}
	cat = applyEnv(cat)
	if err := validate(cat); err != nil {
		return Config{}, err
	}
	domain.SetCatalog(cat)
	return Config{Catalog: cat, Path: path}, nil
}

func parse(name string, b []byte) (domain.Catalog, error) {
	var f fileShape
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".json":
		dec := json.NewDecoder(bytes.NewReader(b))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&f); err != nil {
			return domain.Catalog{}, domain.Usage("invalid config.json").WithHint(err.Error())
		}
	default:
		dec := toml.NewDecoder(bytes.NewReader(b))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&f); err != nil {
			return domain.Catalog{}, domain.Usage("invalid config.toml").WithHint(err.Error())
		}
	}
	return catalogFromFile(f)
}

func catalogFromFile(f fileShape) (domain.Catalog, error) {
	cat := domain.Catalog{
		Projects: map[string]string{},
		Spaces:   map[string]string{},
	}
	for _, s := range f.Sites {
		cat.Sites = append(cat.Sites, domain.Site{
			Alias:    strings.TrimSpace(s.Alias),
			Hostname: strings.TrimSpace(s.Hostname),
			UUID:     strings.TrimSpace(s.UUID),
			Role:     strings.TrimSpace(s.Role),
		})
	}
	for k, v := range f.Projects {
		cat.Projects[strings.ToUpper(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}
	for k, v := range f.Spaces {
		cat.Spaces[strings.ToUpper(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}
	if f.Defaults != nil {
		cat.Defaults = domain.Defaults{
			Workspace:         strings.TrimSpace(f.Defaults.Workspace),
			DefaultSite:       strings.TrimSpace(f.Defaults.DefaultSite),
			AssigneeAccountID: strings.TrimSpace(f.Defaults.AssigneeAccountID),
			JSMSite:           strings.TrimSpace(f.Defaults.JSMSite),
		}
	}
	return cat, nil
}

func applyEnv(cat domain.Catalog) domain.Catalog {
	if v := strings.TrimSpace(os.Getenv("ATLAS_WORKSPACE")); v != "" {
		cat.Defaults.Workspace = v
	}
	if v := strings.TrimSpace(os.Getenv("ATLAS_DEFAULT_SITE")); v != "" {
		cat.Defaults.DefaultSite = v
	}
	return cat
}

func validate(cat domain.Catalog) error {
	if len(cat.Sites) == 0 {
		return domain.Usage("config must list at least one site").WithHint("add [[sites]] alias, hostname, and role")
	}
	aliases := map[string]domain.Site{}
	hosts := map[string]string{}
	for _, s := range cat.Sites {
		if s.Alias == "" || s.Hostname == "" {
			return domain.Usage("each site needs alias and hostname")
		}
		if s.Role != domain.RoleLicensed && s.Role != domain.RoleJSMCustomer {
			return domain.Usagef("site %s role must be licensed or jsm_customer", s.Alias)
		}
		al := strings.ToLower(s.Alias)
		if _, ok := aliases[al]; ok {
			return domain.Usagef("duplicate site alias %q", s.Alias)
		}
		h := strings.ToLower(s.Hostname)
		if prev, ok := hosts[h]; ok {
			return domain.Usagef("duplicate hostname %q (%s and %s)", s.Hostname, prev, s.Alias)
		}
		aliases[al] = s
		hosts[h] = s.Alias
	}
	for proj, alias := range cat.Projects {
		if _, ok := aliases[strings.ToLower(alias)]; !ok {
			return domain.Usagef("project %s maps to unknown site %q", proj, alias)
		}
	}
	for space, alias := range cat.Spaces {
		if _, ok := aliases[strings.ToLower(alias)]; !ok {
			return domain.Usagef("space %s maps to unknown site %q", space, alias)
		}
	}
	if cat.Defaults.DefaultSite != "" {
		if _, ok := aliases[strings.ToLower(cat.Defaults.DefaultSite)]; !ok {
			return domain.Usagef("defaults.default_site unknown site %q", cat.Defaults.DefaultSite)
		}
	}
	if cat.Defaults.JSMSite != "" {
		s, ok := aliases[strings.ToLower(cat.Defaults.JSMSite)]
		if !ok {
			return domain.Usagef("defaults.jsm_site unknown site %q", cat.Defaults.JSMSite)
		}
		if s.Role != domain.RoleJSMCustomer {
			return domain.Usage("defaults.jsm_site must have role jsm_customer")
		}
	}
	return nil
}

func searchPaths() []string {
	var out []string
	xdg := os.Getenv("XDG_CONFIG_HOME")
	home, _ := os.UserHomeDir()
	if xdg != "" {
		out = append(out, filepath.Join(xdg, "atlas", "config.toml"))
	}
	if home != "" {
		out = append(out, filepath.Join(home, ".config", "atlas", "config.toml"))
	}
	if xdg != "" {
		out = append(out, filepath.Join(xdg, "atlas", "config.json"))
	}
	if home != "" {
		out = append(out, filepath.Join(home, ".config", "atlas", "config.json"))
	}
	return unique(out)
}

func unique(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range in {
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}
