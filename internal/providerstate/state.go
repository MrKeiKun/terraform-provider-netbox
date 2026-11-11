// Package providerstate connects to NetBox from the provider settings. The Plugin Framework
// provider uses it to build the state its resources and data sources share.
package providerstate

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
)

// SupportedVersions are the NetBox versions the provider is tested against.
var SupportedVersions = []string{"4.4.0", "4.4.1", "4.4.2", "4.4.4", "4.4.5", "4.4.6", "4.4.7", "4.4.8", "4.4.9", "4.4.10", "4.5.1", "4.5.2", "4.5.3", "4.5.4", "4.5.5", "4.5.6", "4.5.7", "4.5.8", "4.5.9", "4.5.10", "4.6.0", "4.6.1", "4.6.2", "4.6.3", "4.6.4", "4.6.5", "4.6.6", "4.6.7", "4.6.8", "4.6.9", "4.6.10", "4.7.0", "4.7.1", "4.7.2", "4.7.3", "4.7.4"}

// Settings are the provider attributes after defaults from the environment were applied.
type Settings struct {
	Config
	SkipVersionCheck bool
	DefaultTags      []string
}

// Diag is a warning or an error found while connecting.
type Diag struct {
	Error   bool
	Summary string
	Detail  string
}

// State is what resources and data sources get from the provider.
type State struct {
	Rest        *netboxapi.Client
	DefaultTags map[string]bool

	// concurrent access ok, only populated on provider start
	tagCache map[string]netboxapi.TagRef
}

var versionRe = regexp.MustCompile(`^v?(\d+\.\d+\.\d+)`)

// ExtractVersion returns the semantic version at the start of s, so that for example
// v4.2.8-Docker-3.2.1 becomes 4.2.8.
func ExtractVersion(s string) (string, error) {
	m := versionRe.FindStringSubmatch(s)
	if len(m) < 2 {
		return "", fmt.Errorf("no semantic version found in version string")
	}
	return m[1], nil
}

// Connect creates the client, checks the NetBox version and looks up the default tags. When it
// returns a nil state, a diagnostic with Error set tells why.
func Connect(ctx context.Context, s Settings) (*State, []Diag) {
	var diags []Diag

	serverURL := s.ServerURL
	// Trailing slashes cause errors as seen in https://github.com/MrKeiKun/terraform-provider-netbox/issues/198
	// and https://github.com/MrKeiKun/terraform-provider-netbox/issues/300
	if s.StripTrailingSlashesFromURL {
		trimmed := false
		for strings.HasSuffix(serverURL, "/") {
			serverURL = strings.TrimRight(serverURL, "/")
			trimmed = true
		}
		if trimmed {
			diags = append(diags, Diag{
				Summary: "Stripped trailing slashes from the `server_url` parameter",
				Detail:  "Trailing slashes in the `server_url` parameter lead to problems in most setups, so all trailing slashes were stripped. Use the `strip_trailing_slashes_from_url` parameter to disable this feature or remove all trailing slashes in the `server_url` to disable this warning.",
			})
		}
	}
	s.ServerURL = serverURL

	rest, err := s.Client()
	if err != nil {
		return nil, append(diags, Diag{Error: true, Summary: err.Error()})
	}

	if !s.SkipVersionCheck {
		if d := checkVersion(ctx, rest, serverURL); d != nil {
			diags = append(diags, d...)
			for _, x := range d {
				if x.Error {
					return nil, diags
				}
			}
		}
	}

	state := &State{
		Rest:        rest,
		DefaultTags: make(map[string]bool, len(s.DefaultTags)),
		tagCache:    make(map[string]netboxapi.TagRef, len(s.DefaultTags)),
	}
	for _, name := range s.DefaultTags {
		state.DefaultTags[name] = true
		ref, err := FindTagRef(ctx, rest, name)
		if err != nil {
			diags = append(diags, Diag{Summary: fmt.Sprintf("default tag not found: %s", err)})
			continue
		}
		state.tagCache[name] = ref
	}
	return state, diags
}

func checkVersion(ctx context.Context, rest *netboxapi.Client, serverURL string) []Diag {
	// The status endpoint can be transiently unavailable (e.g. Netbox still starting, a worker
	// restart, or a flaky proxy), so retry a few times with a linear backoff before giving up.
	const attempts = 5
	var res *map[string]interface{}
	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		res, err = netboxapi.Request[map[string]interface{}](ctx, rest, http.MethodGet, "status/", nil, nil)
		if err == nil {
			break
		}
		if attempt < attempts {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}
	if err != nil {
		return []Diag{{
			Error:   true,
			Summary: "Failed to determine Netbox version",
			Detail: fmt.Sprintf(
				"The provider could not get a valid JSON response from a GET request to `%s/api/status/`.\n\n"+
					"This usually means one of the following:\n"+
					"  - `server_url` is misconfigured (wrong scheme, host or port, or it already includes a path such as `/api`)\n"+
					"  - Netbox (or a reverse proxy/ingress in front of it) returned a non-JSON response, such as an HTML error page, or a response missing a `Content-Type: application/json` header\n"+
					"  - the Netbox API is temporarily unreachable or returned a server error\n\n"+
					"You can set `skip_version_check = true` on the provider to bypass this check while troubleshooting.\n\n"+
					"Original error: %s",
				serverURL, err,
			),
		}}
	}

	raw, _ := (*res)["netbox-version"].(string)
	version, err := ExtractVersion(raw)
	if err != nil {
		return []Diag{{Error: true, Summary: fmt.Sprintf("error extracting netbox version. try using the `skip_version_check` provider parameter to bypass this error. original error: %s", err)}}
	}
	if !slices.Contains(SupportedVersions, version) {
		return []Diag{{
			Summary: "Possibly unsupported Netbox version",
			Detail:  fmt.Sprintf("Your Netbox reports version %v. From that, the provider extracted Netbox version %v.\nThe provider was successfully tested against the following versions:\n\n  %v\n\nUnexpected errors may occur.", raw, version, strings.Join(SupportedVersions, ", ")),
		}}
	}
	return nil
}

// FindTagRef looks up a tag by its unique name.
func FindTagRef(ctx context.Context, c *netboxapi.Client, name string) (netboxapi.TagRef, error) {
	// We search for a unique tag. Having two hits suffices to know its not unique.
	query := url.Values{"name": {name}, "limit": {"2"}}
	tags, err := netboxapi.List[netboxapi.TagRef](ctx, c, "extras/tags/", query)
	if err != nil {
		return netboxapi.TagRef{}, fmt.Errorf("API Error trying to retrieve tag %q from netbox: %w", name, err)
	}
	switch len(tags) {
	case 0:
		return netboxapi.TagRef{}, fmt.Errorf("could not locate referenced tag %q in netbox, no results", name)
	case 1:
		return tags[0], nil
	default:
		return netboxapi.TagRef{}, fmt.Errorf("could not map tag %q to unique tag in netbox, %d results", name, len(tags))
	}
}

// TagRef returns the tag with the given name, using the cache of default tags when possible.
func (s *State) TagRef(ctx context.Context, name string) (netboxapi.TagRef, error) {
	if ref, ok := s.tagCache[name]; ok {
		return ref, nil
	}
	return FindTagRef(ctx, s.Rest, name)
}
