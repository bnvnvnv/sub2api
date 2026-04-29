package vlessutil

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/sagernet/sing-vmess/vless"
	M "github.com/sagernet/sing/common/metadata"
)

type Node struct {
	UUID     string
	Host     string
	Port     int
	Security string
	SNI      string
	Flow     string
	Insecure bool
	Tag      string
}

func (n *Node) OptionsString() string {
	values := url.Values{}
	if strings.TrimSpace(n.Security) != "" {
		values.Set("security", strings.TrimSpace(n.Security))
	}
	if strings.TrimSpace(n.SNI) != "" {
		values.Set("sni", strings.TrimSpace(n.SNI))
	}
	if strings.TrimSpace(n.Flow) != "" {
		values.Set("flow", strings.TrimSpace(n.Flow))
	}
	if n.Insecure {
		values.Set("insecure", "1")
	}
	return values.Encode()
}

func ParseURL(raw string) (*Node, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("vless url is empty")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("invalid vless url: %w", err)
	}
	return ParseParsedURL(parsed)
}

func ParseParsedURL(parsed *url.URL) (*Node, error) {
	if parsed == nil {
		return nil, fmt.Errorf("vless url is nil")
	}
	if strings.ToLower(strings.TrimSpace(parsed.Scheme)) != "vless" {
		return nil, fmt.Errorf("invalid vless scheme: %s", parsed.Scheme)
	}
	host, port, err := parseHostPort(parsed)
	if err != nil {
		return nil, err
	}
	uuid := ""
	if parsed.User != nil {
		uuid = strings.TrimSpace(parsed.User.Username())
	}
	if uuid == "" {
		return nil, fmt.Errorf("vless uuid is required")
	}

	query, _ := url.ParseQuery(parsed.RawQuery)
	transport := strings.ToLower(strings.TrimSpace(firstQuery(query, "type", "network")))
	if transport != "" && transport != "tcp" {
		return nil, fmt.Errorf("vless transport %q is not supported", transport)
	}
	security := strings.ToLower(strings.TrimSpace(query.Get("security")))
	if security == "" {
		security = tlsQuerySecurity(query)
	}
	switch security {
	case "", "none":
		security = "none"
	case "tls":
	case "reality":
		return nil, fmt.Errorf("vless reality security is not supported")
	default:
		return nil, fmt.Errorf("vless security %q is not supported", security)
	}
	flow := strings.TrimSpace(query.Get("flow"))
	if flow != "" && security != "tls" {
		return nil, fmt.Errorf("vless flow requires tls security")
	}
	sni := strings.TrimSpace(firstQuery(query, "sni", "servername", "server_name", "peer"))
	if sni == "" && security == "tls" {
		sni = host
	}

	node := &Node{
		UUID:     uuid,
		Host:     host,
		Port:     port,
		Security: security,
		SNI:      sni,
		Flow:     flow,
		Insecure: queryBool(query, "insecure") || queryBool(query, "allowInsecure") || queryBool(query, "skip-cert-verify"),
		Tag:      parsed.Fragment,
	}
	if _, err := vless.NewClient(node.UUID, node.Flow, noopLogger{}); err != nil {
		return nil, err
	}
	return node, nil
}

func BuildURL(uuid, host string, port int, options string, tag string) (string, error) {
	uuid = strings.TrimSpace(uuid)
	host = strings.TrimSpace(host)
	if uuid == "" {
		return "", fmt.Errorf("vless uuid is required")
	}
	if host == "" {
		return "", fmt.Errorf("vless host is required")
	}
	if port <= 0 || port > 65535 {
		return "", fmt.Errorf("invalid vless port: %d", port)
	}
	values := parseOptionString(options)
	security := strings.ToLower(strings.TrimSpace(values.Get("security")))
	if security == "" {
		security = tlsQuerySecurity(values)
		if security == "" {
			security = "none"
		}
	}
	values.Del("tls")
	values.Set("security", security)
	if security == "tls" && strings.TrimSpace(firstQuery(values, "sni", "servername", "server_name", "peer")) == "" {
		values.Set("sni", host)
	}
	u := &url.URL{
		Scheme:   "vless",
		User:     url.User(uuid),
		Host:     net.JoinHostPort(host, strconv.Itoa(port)),
		RawQuery: values.Encode(),
	}
	if strings.TrimSpace(tag) != "" {
		u.Fragment = strings.TrimSpace(tag)
	}
	return u.String(), nil
}

func DialContext(ctx context.Context, proxyURL *url.URL, network, targetAddr string) (net.Conn, error) {
	node, err := ParseParsedURL(proxyURL)
	if err != nil {
		return nil, err
	}
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return nil, fmt.Errorf("vless only supports tcp, got %s", network)
	}
	destination := M.ParseSocksaddr(targetAddr)
	if !destination.IsValid() {
		return nil, fmt.Errorf("invalid target address: %s", targetAddr)
	}

	var conn net.Conn
	if node.Security == "tls" {
		dialer := tls.Dialer{
			NetDialer: &net.Dialer{},
			Config: &tls.Config{
				ServerName:         node.SNI,
				InsecureSkipVerify: node.Insecure,
				MinVersion:         tls.VersionTLS12,
			},
		}
		conn, err = dialer.DialContext(ctx, "tcp", net.JoinHostPort(node.Host, strconv.Itoa(node.Port)))
	} else {
		var dialer net.Dialer
		conn, err = dialer.DialContext(ctx, "tcp", net.JoinHostPort(node.Host, strconv.Itoa(node.Port)))
	}
	if err != nil {
		return nil, err
	}
	client, err := vless.NewClient(node.UUID, node.Flow, noopLogger{})
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return client.DialEarlyConn(conn, destination)
}

func parseHostPort(parsed *url.URL) (string, int, error) {
	if parsed.Host == "" || parsed.Hostname() == "" {
		return "", 0, fmt.Errorf("proxy url missing host")
	}
	port, err := strconv.Atoi(parsed.Port())
	if err != nil || port <= 0 || port > 65535 {
		return "", 0, fmt.Errorf("invalid proxy port: %s", parsed.Port())
	}
	return strings.TrimSpace(parsed.Hostname()), port, nil
}

func parseOptionString(raw string) url.Values {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return url.Values{}
	}
	values, err := url.ParseQuery(raw)
	if err == nil && len(values) > 0 {
		return values
	}
	return url.Values{"sni": []string{raw}}
}

func firstQuery(values url.Values, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(values.Get(key)); value != "" {
			return value
		}
	}
	return ""
}

func queryBool(values url.Values, key string) bool {
	switch strings.ToLower(strings.TrimSpace(values.Get(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func tlsQuerySecurity(values url.Values) string {
	switch strings.ToLower(strings.TrimSpace(values.Get("tls"))) {
	case "":
		return ""
	case "1", "true", "yes", "on", "tls":
		return "tls"
	case "0", "false", "no", "off", "none":
		return "none"
	default:
		return strings.ToLower(strings.TrimSpace(values.Get("tls")))
	}
}

type noopLogger struct{}

func (noopLogger) Trace(args ...any) {}
func (noopLogger) Debug(args ...any) {}
func (noopLogger) Info(args ...any)  {}
func (noopLogger) Warn(args ...any)  {}
func (noopLogger) Error(args ...any) {}
func (noopLogger) Fatal(args ...any) {}
func (noopLogger) Panic(args ...any) {}
