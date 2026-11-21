// Package providerstate connects to NetBox from the provider settings. The Plugin Framework
// provider uses it to build the state its resources and data sources share.
package providerstate

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// SupportedVersions are the NetBox versions the provider is tested against.
var SupportedVersions = []string{"4.4.0", "4.4.1", "4.4.2", "4.4.4", "4.4.5", "4.4.6", "4.4.7", "4.4.8", "4.4.9", "4.4.10", "4.5.1", "4.5.2", "4.5.3", "4.5.4", "4.5.5", "4.5.6", "4.5.7", "4.5.8", "4.5.9", "4.5.10", "4.6.0", "4.6.1", "4.6.2", "4.6.3", "4.6.4", "4.6.5", "4.6.6", "4.6.7", "4.6.8", "4.6.9", "4.6.10", "4.7.0", "4.7.1", "4.7.2", "4.7.3", "4.7.4", "4.7.5"}

// Settings are the provider attributes after defaults from the environment were applied.
type Settings struct {
	Config
	SkipVersionCheck bool
	DefaultTags      []string
	// PageSize is the number of objects a plural data source asks NetBox for with each request. Zero
	// means DefaultPageSize.
	PageSize int
}

// DefaultPageSize is the page size of the plural data sources. It is the most NetBox returns for a
// page unless its administrator changed MAX_PAGE_SIZE, in which case NetBox returns less and the
// data source reads the next page.
const DefaultPageSize = 1000

// Diag is a warning or an error found while connecting.
type Diag struct {
	Error   bool
	Summary string
	Detail  string
}

// State is what resources and data sources get from the provider.
type State struct {
	Rest *netboxapi.Client
	// Version is the version of the NetBox server, for example 4.7.5. It is empty when the provider
	// does not check the version.
	Version     string
	DefaultTags map[string]bool
	// PageSize is the number of objects a plural data source asks NetBox for with each request.
	PageSize int

	// concurrent access ok, only populated on provider start
	tagCache map[string]netboxapi.TagRef

	// the lookups of the other tags, made while the provider runs
	tagMu      sync.Mutex
	tagLookups map[string]*tagLookup
}

func pageSize(n int) int {
	if n <= 0 {
		return DefaultPageSize
	}
	return n
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

	names := make([]string, 0, len(s.Headers))
	for name := range s.Headers {
		names = append(names, name)
	}
	sort.Strings(names)
	// Only the names of the headers are logged, the values may be secrets.
	tflog.Debug(ctx, "Connecting to NetBox", map[string]any{"server_url": serverURL, "custom_headers": names})
	if u, err := parseServerURL(serverURL); err == nil && u.Scheme == "http" && !isLoopback(u.Hostname()) {
		tflog.Warn(ctx, "server_url uses http, so the API token is sent to NetBox without encryption", map[string]any{"server_url": serverURL})
	}

	serverVersion := ""
	if !s.SkipVersionCheck {
		var d []Diag
		if serverVersion, d = checkVersion(ctx, rest, serverURL); d != nil {
			diags = append(diags, d...)
			for _, x := range d {
				if x.Error {
					return nil, diags
				}
			}
		}
	}

	// A web server that sends http to https answers the version check with a redirect. Reads follow
	// it, but every one then takes two requests, and a write cannot follow it at all.
	if addr := s.redirected.Load(); addr != nil {
		diags = append(diags, Diag{
			Summary: "server_url redirects to " + *addr,
			Detail:  fmt.Sprintf("Netbox answered the request to %s with a redirect to %s. Set `server_url` to that address. Reading still works, with two requests in place of one, but creating or changing an object fails when the redirect turns the request into a GET.", serverURL, *addr),
		})
	}

	state := &State{
		Rest:        rest,
		Version:     serverVersion,
		DefaultTags: make(map[string]bool, len(s.DefaultTags)),
		PageSize:    pageSize(s.PageSize),
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

// isLoopback reports whether host is this machine, where a request never leaves it.
func isLoopback(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func checkVersion(ctx context.Context, rest *netboxapi.Client, serverURL string) (string, []Diag) {
	// The client already sends the request again when the connection fails or NetBox answers 502,
	// 503 or 504, and stops when the context is cancelled. A wrong URL or token is reported at once.
	res, err := netboxapi.Request[map[string]interface{}](ctx, rest, http.MethodGet, "status/", nil, nil)
	var apiErr *netboxapi.APIError
	if errors.As(err, &apiErr) && (apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden) {
		return "", []Diag{{
			Error:   true,
			Summary: "Netbox rejected the API token",
			Detail: fmt.Sprintf(
				"Netbox answered a GET request to `%s/api/status/` with status %d.\n\n"+
					"This usually means the `api_token` is wrong, expired or disabled, or the user of the token is not active. "+
					"For a version 2 token, the value must be the whole `nbt_<key>.<token>` string. "+
					"A reverse proxy in front of Netbox can also answer this way.\n\n"+
					"Original error: %s",
				serverURL, apiErr.StatusCode, err,
			),
		}}
	}
	var certErr *tls.CertificateVerificationError
	if errors.As(err, &certErr) {
		return "", []Diag{{
			Error:   true,
			Summary: "Netbox's certificate is not trusted",
			Detail: fmt.Sprintf(
				"The provider could not verify the certificate of `%s`.\n\n"+
					"If Netbox uses a certificate from your own certificate authority, set `ca_cert_file` to the PEM file of that authority. "+
					"If the certificate is self-signed or for another host name, `allow_insecure_https = true` turns the check off, which you should only do for a test. "+
					"A certificate that has expired, or a host name that is not in the certificate, gives the same error.\n\n"+
					"Original error: %s",
				serverURL, err,
			),
		}}
	}
	if err != nil {
		return "", []Diag{{
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
		return "", []Diag{{Error: true, Summary: fmt.Sprintf("error extracting netbox version. try using the `skip_version_check` provider parameter to bypass this error. original error: %s", err)}}
	}
	if !slices.Contains(SupportedVersions, version) {
		return version, []Diag{{
			Summary: "Possibly unsupported Netbox version",
			Detail:  fmt.Sprintf("Your Netbox reports version %v. From that, the provider extracted Netbox version %v.\nThe provider was successfully tested against the following versions:\n\n  %v\n\nUnexpected errors may occur.", raw, version, strings.Join(SupportedVersions, ", ")),
		}}
	}
	return version, nil
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

// tagLookup is the lookup of one tag by name. Resources that run at the same time and use the same
// tag share it, so NetBox is asked once.
type tagLookup struct {
	mu  sync.Mutex
	ref netboxapi.TagRef
	ok  bool
}

// TagRef returns the tag with the given name. A tag is looked up once and then kept for as long as the
// provider runs, because a configuration usually uses a few tags on many objects. A lookup that fails is
// not kept. ForgetTag drops a tag that changed.
func (s *State) TagRef(ctx context.Context, name string) (netboxapi.TagRef, error) {
	if ref, ok := s.tagCache[name]; ok {
		return ref, nil
	}
	s.tagMu.Lock()
	if s.tagLookups == nil {
		s.tagLookups = map[string]*tagLookup{}
	}
	l := s.tagLookups[name]
	if l == nil {
		l = &tagLookup{}
		s.tagLookups[name] = l
	}
	s.tagMu.Unlock()

	l.mu.Lock()
	defer l.mu.Unlock()
	if l.ok {
		return l.ref, nil
	}
	ref, err := FindTagRef(ctx, s.Rest, name)
	if err != nil {
		return netboxapi.TagRef{}, err
	}
	l.ref, l.ok = ref, true
	return ref, nil
}

// ForgetTag drops the kept lookup of a tag. Call it when the tag was created, changed or deleted, so the
// next use asks NetBox again. Default tags are looked up when the provider starts and stay.
func (s *State) ForgetTag(name string) {
	s.tagMu.Lock()
	delete(s.tagLookups, name)
	s.tagMu.Unlock()
}
