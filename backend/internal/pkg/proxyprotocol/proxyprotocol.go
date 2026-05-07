package proxyprotocol

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/anytlsutil"
	"github.com/Wei-Shaw/sub2api/internal/pkg/hysteria2util"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ssutil"
	"github.com/Wei-Shaw/sub2api/internal/pkg/trojanutil"
	"github.com/Wei-Shaw/sub2api/internal/pkg/vlessutil"
)

type ProtocolSpec struct {
	Scheme       string
	Aliases      []string
	Canonicalize func(*url.URL) (string, error)
	BuildURL     func(username, password, host string, port int, tag string) (string, error)
	DialContext  func(context.Context, *url.URL, string, string) (net.Conn, error)
}

var extensionSpecs = []ProtocolSpec{
	{
		Scheme: "ss",
		Canonicalize: func(parsed *url.URL) (string, error) {
			node, err := ssutil.ParseParsedURL(parsed)
			if err != nil {
				return "", err
			}
			if node.Plugin != "" {
				return "", fmt.Errorf("ss plugin is not supported")
			}
			return ssutil.BuildURL(node.Method, node.Password, node.Host, node.Port, node.Tag)
		},
		BuildURL: func(username, password, host string, port int, tag string) (string, error) {
			return ssutil.BuildURL(username, password, host, port, tag)
		},
		DialContext: ssutil.DialContext,
	},
	{
		Scheme: "anytls",
		Canonicalize: func(parsed *url.URL) (string, error) {
			node, err := anytlsutil.ParseParsedURL(parsed)
			if err != nil {
				return "", err
			}
			return anytlsutil.BuildURL(node.Password, node.Host, node.Port, node.OptionsString(), node.Tag)
		},
		BuildURL: func(username, password, host string, port int, tag string) (string, error) {
			return anytlsutil.BuildURL(password, host, port, username, tag)
		},
		DialContext: anytlsutil.DialContext,
	},
	{
		Scheme: "trojan",
		Canonicalize: func(parsed *url.URL) (string, error) {
			node, err := trojanutil.ParseParsedURL(parsed)
			if err != nil {
				return "", err
			}
			return trojanutil.BuildURL(node.Password, node.Host, node.Port, node.OptionsString(), node.Tag)
		},
		BuildURL: func(username, password, host string, port int, tag string) (string, error) {
			return trojanutil.BuildURL(password, host, port, username, tag)
		},
		DialContext: trojanutil.DialContext,
	},
	{
		Scheme: "vless",
		Canonicalize: func(parsed *url.URL) (string, error) {
			node, err := vlessutil.ParseParsedURL(parsed)
			if err != nil {
				return "", err
			}
			return vlessutil.BuildURL(node.UUID, node.Host, node.Port, node.OptionsString(), node.Tag)
		},
		BuildURL: func(username, password, host string, port int, tag string) (string, error) {
			return vlessutil.BuildURL(password, host, port, username, tag)
		},
		DialContext: vlessutil.DialContext,
	},
	{
		Scheme:  "hysteria2",
		Aliases: []string{"hy2"},
		Canonicalize: func(parsed *url.URL) (string, error) {
			node, err := hysteria2util.ParseParsedURL(parsed)
			if err != nil {
				return "", err
			}
			return hysteria2util.BuildURL(node.Password, node.Host, node.Port, node.OptionsString(), node.Tag)
		},
		BuildURL: func(username, password, host string, port int, tag string) (string, error) {
			return hysteria2util.BuildURL(password, host, port, username, tag)
		},
		DialContext: hysteria2util.DialContext,
	},
}

var standardSchemes = map[string]bool{
	"http":    true,
	"https":   true,
	"socks5":  true,
	"socks5h": true,
}

var specsByScheme = buildSpecsByScheme()

func buildSpecsByScheme() map[string]ProtocolSpec {
	out := make(map[string]ProtocolSpec, len(extensionSpecs))
	for _, spec := range extensionSpecs {
		out[spec.Scheme] = spec
		for _, alias := range spec.Aliases {
			aliasSpec := spec
			aliasSpec.Scheme = alias
			out[alias] = aliasSpec
		}
	}
	return out
}

func NormalizeScheme(protocol string) string {
	protocol = strings.ToLower(strings.TrimSpace(protocol))
	switch protocol {
	case "socks":
		return "socks5"
	case "hy2":
		return "hysteria2"
	default:
		return protocol
	}
}

func IsSupported(protocol string) bool {
	scheme := NormalizeScheme(protocol)
	return standardSchemes[scheme] || specsByScheme[scheme].Canonicalize != nil
}

func SupportedSchemesForError() string {
	schemes := make([]string, 0, len(standardSchemes)+len(extensionSpecs))
	for scheme := range standardSchemes {
		schemes = append(schemes, scheme)
	}
	for _, spec := range extensionSpecs {
		schemes = append(schemes, spec.Scheme)
	}
	sort.Strings(schemes)
	return strings.Join(schemes, ", ")
}

func CanonicalizeURL(parsed *url.URL) (string, *url.URL, bool, error) {
	if parsed == nil {
		return "", nil, false, fmt.Errorf("proxy URL is nil")
	}
	spec, ok := specsByScheme[strings.ToLower(strings.TrimSpace(parsed.Scheme))]
	if !ok || spec.Canonicalize == nil {
		return "", nil, false, nil
	}
	canonical, err := spec.Canonicalize(parsed)
	if err != nil {
		return "", nil, true, err
	}
	canonicalParsed, err := url.Parse(canonical)
	if err != nil {
		return "", nil, true, fmt.Errorf("invalid proxy URL: %v", err)
	}
	return canonical, canonicalParsed, true, nil
}

func ConfigureTransportProxy(transport *http.Transport, proxyURL *url.URL) (bool, error) {
	if transport == nil || proxyURL == nil {
		return false, nil
	}
	spec, ok := specsByScheme[strings.ToLower(strings.TrimSpace(proxyURL.Scheme))]
	if !ok || spec.DialContext == nil {
		return false, nil
	}
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return spec.DialContext(ctx, proxyURL, network, addr)
	}
	return true, nil
}

func BuildURL(protocol, username, password, host string, port int, tag string) string {
	spec, ok := specsByScheme[strings.ToLower(strings.TrimSpace(protocol))]
	if ok && spec.BuildURL != nil {
		raw, err := spec.BuildURL(username, password, host, port, tag)
		if err == nil {
			return raw
		}
	}

	u := &url.URL{
		Scheme: protocol,
		Host:   net.JoinHostPort(host, strconv.Itoa(port)),
	}
	if username != "" && password != "" {
		u.User = url.UserPassword(username, password)
	}
	return u.String()
}
