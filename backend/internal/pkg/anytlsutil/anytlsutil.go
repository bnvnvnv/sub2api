package anytlsutil

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	anytls "github.com/anytls/sing-anytls"
	M "github.com/sagernet/sing/common/metadata"
)

type Node struct {
	Password string
	Host     string
	Port     int
	SNI      string
	Insecure bool
	Tag      string
}

func (n *Node) OptionsString() string {
	values := url.Values{}
	if strings.TrimSpace(n.SNI) != "" {
		values.Set("sni", strings.TrimSpace(n.SNI))
	}
	if n.Insecure {
		values.Set("insecure", "1")
	} else {
		values.Set("insecure", "0")
	}
	return values.Encode()
}

func ParseURL(raw string) (*Node, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("anytls url is empty")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("invalid anytls url: %w", err)
	}
	return ParseParsedURL(parsed)
}

func ParseParsedURL(parsed *url.URL) (*Node, error) {
	if parsed == nil {
		return nil, fmt.Errorf("anytls url is nil")
	}
	if strings.ToLower(strings.TrimSpace(parsed.Scheme)) != "anytls" {
		return nil, fmt.Errorf("invalid anytls scheme: %s", parsed.Scheme)
	}
	if parsed.Host == "" || parsed.Hostname() == "" {
		return nil, fmt.Errorf("anytls url missing host")
	}
	port, err := strconv.Atoi(parsed.Port())
	if err != nil || port <= 0 || port > 65535 {
		return nil, fmt.Errorf("invalid anytls port: %s", parsed.Port())
	}
	password := ""
	if parsed.User != nil {
		password = strings.TrimSpace(parsed.User.Username())
		if value, ok := parsed.User.Password(); ok && strings.TrimSpace(value) != "" {
			password = strings.TrimSpace(value)
		}
	}
	if password == "" {
		return nil, fmt.Errorf("anytls password is required")
	}
	query, _ := url.ParseQuery(parsed.RawQuery)
	sni := strings.TrimSpace(firstQuery(query, "sni", "servername", "server_name"))
	if sni == "" {
		sni = parsed.Hostname()
	}
	return &Node{
		Password: password,
		Host:     parsed.Hostname(),
		Port:     port,
		SNI:      sni,
		Insecure: queryBool(query, "insecure") || queryBool(query, "allowInsecure") || queryBool(query, "skip-cert-verify"),
		Tag:      parsed.Fragment,
	}, nil
}

func BuildURL(password, host string, port int, options string, tag string) (string, error) {
	password = strings.TrimSpace(password)
	host = strings.TrimSpace(host)
	if password == "" {
		return "", fmt.Errorf("anytls password is required")
	}
	if host == "" {
		return "", fmt.Errorf("anytls host is required")
	}
	if port <= 0 || port > 65535 {
		return "", fmt.Errorf("invalid anytls port: %d", port)
	}
	values := parseOptionString(options)
	sni := strings.TrimSpace(firstQuery(values, "sni", "servername", "server_name"))
	if sni == "" {
		sni = host
	}
	insecure := queryBool(values, "insecure") || queryBool(values, "allowInsecure") || queryBool(values, "skip-cert-verify")
	values.Del("servername")
	values.Del("server_name")
	values.Del("allowInsecure")
	values.Del("skip-cert-verify")
	values.Set("sni", sni)
	if insecure {
		values.Set("insecure", "1")
	} else {
		values.Set("insecure", "0")
	}
	u := &url.URL{
		Scheme:   "anytls",
		User:     url.User(password),
		Host:     net.JoinHostPort(host, strconv.Itoa(port)),
		RawQuery: values.Encode(),
	}
	if strings.TrimSpace(tag) != "" {
		u.Fragment = tag
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
		return nil, fmt.Errorf("anytls only supports tcp, got %s", network)
	}
	destination := M.ParseSocksaddr(targetAddr)
	if !destination.IsValid() {
		return nil, fmt.Errorf("invalid target address: %s", targetAddr)
	}
	client, err := anytls.NewClient(ctx, anytls.ClientConfig{
		Password: node.Password,
		DialOut: func(ctx context.Context) (net.Conn, error) {
			dialer := tls.Dialer{
				NetDialer: &net.Dialer{},
				Config: &tls.Config{
					ServerName:         node.SNI,
					InsecureSkipVerify: node.Insecure,
					MinVersion:         tls.VersionTLS12,
				},
			}
			return dialer.DialContext(ctx, "tcp", net.JoinHostPort(node.Host, strconv.Itoa(node.Port)))
		},
		IdleSessionCheckInterval: 30 * time.Second,
		IdleSessionTimeout:       30 * time.Second,
		MinIdleSession:           0,
		Logger:                   noopLogger{},
	})
	if err != nil {
		return nil, fmt.Errorf("create anytls client: %w", err)
	}
	conn, err := client.CreateProxy(ctx, destination)
	if err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("create anytls proxy stream: %w", err)
	}
	return &clientConn{Conn: conn, closeClient: client.Close}, nil
}

type clientConn struct {
	net.Conn
	closeClient func() error
}

func (c *clientConn) Close() error {
	connErr := c.Conn.Close()
	clientErr := error(nil)
	if c.closeClient != nil {
		clientErr = c.closeClient()
	}
	if connErr != nil {
		return connErr
	}
	return clientErr
}

type noopLogger struct{}

func (noopLogger) Trace(args ...any)                             {}
func (noopLogger) Debug(args ...any)                             {}
func (noopLogger) Info(args ...any)                              {}
func (noopLogger) Warn(args ...any)                              {}
func (noopLogger) Error(args ...any)                             {}
func (noopLogger) Fatal(args ...any)                             {}
func (noopLogger) Panic(args ...any)                             {}
func (noopLogger) TraceContext(ctx context.Context, args ...any) {}
func (noopLogger) DebugContext(ctx context.Context, args ...any) {}
func (noopLogger) InfoContext(ctx context.Context, args ...any)  {}
func (noopLogger) WarnContext(ctx context.Context, args ...any)  {}
func (noopLogger) ErrorContext(ctx context.Context, args ...any) {}
func (noopLogger) FatalContext(ctx context.Context, args ...any) {}
func (noopLogger) PanicContext(ctx context.Context, args ...any) {}

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
