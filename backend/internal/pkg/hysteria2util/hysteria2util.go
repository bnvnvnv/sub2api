package hysteria2util

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/sagernet/sing-quic/hysteria"
	hy2 "github.com/sagernet/sing-quic/hysteria2"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	aTLS "github.com/sagernet/sing/common/tls"
)

type Node struct {
	Password     string
	Host         string
	Port         int
	SNI          string
	Insecure     bool
	Obfs         string
	ObfsPassword string
	UpMbps       int
	DownMbps     int
	Tag          string
}

func (n *Node) OptionsString() string {
	values := url.Values{}
	if strings.TrimSpace(n.SNI) != "" {
		values.Set("sni", strings.TrimSpace(n.SNI))
	}
	if n.Insecure {
		values.Set("insecure", "1")
	}
	if strings.TrimSpace(n.Obfs) != "" {
		values.Set("obfs", strings.TrimSpace(n.Obfs))
		values.Set("obfs-password", strings.TrimSpace(n.ObfsPassword))
	}
	if n.UpMbps > 0 {
		values.Set("upmbps", strconv.Itoa(n.UpMbps))
	}
	if n.DownMbps > 0 {
		values.Set("downmbps", strconv.Itoa(n.DownMbps))
	}
	return values.Encode()
}

func ParseURL(raw string) (*Node, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("hysteria2 url is empty")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("invalid hysteria2 url: %w", err)
	}
	return ParseParsedURL(parsed)
}

func ParseParsedURL(parsed *url.URL) (*Node, error) {
	if parsed == nil {
		return nil, fmt.Errorf("hysteria2 url is nil")
	}
	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if scheme != "hysteria2" && scheme != "hy2" {
		return nil, fmt.Errorf("invalid hysteria2 scheme: %s", parsed.Scheme)
	}
	host, port, err := parseHostPort(parsed)
	if err != nil {
		return nil, err
	}
	password := ""
	if parsed.User != nil {
		password = strings.TrimSpace(parsed.User.Username())
		if value, ok := parsed.User.Password(); ok && strings.TrimSpace(value) != "" {
			password = strings.TrimSpace(value)
		}
	}
	if password == "" {
		return nil, fmt.Errorf("hysteria2 password is required")
	}
	query, _ := url.ParseQuery(parsed.RawQuery)
	if strings.TrimSpace(firstQuery(query, "pinSHA256", "pin-sha256", "pin_sha256")) != "" {
		return nil, fmt.Errorf("hysteria2 certificate pin is not supported")
	}
	obfs := strings.ToLower(strings.TrimSpace(query.Get("obfs")))
	obfsPassword := strings.TrimSpace(firstQuery(query, "obfs-password", "obfs_password"))
	if obfs != "" {
		if obfs != hy2.ObfsTypeSalamander {
			return nil, fmt.Errorf("hysteria2 obfs %q is not supported", obfs)
		}
		if obfsPassword == "" {
			return nil, fmt.Errorf("hysteria2 obfs password is required")
		}
	}
	sni := strings.TrimSpace(firstQuery(query, "sni", "servername", "server_name"))
	if sni == "" {
		sni = host
	}

	return &Node{
		Password:     password,
		Host:         host,
		Port:         port,
		SNI:          sni,
		Insecure:     queryBool(query, "insecure") || queryBool(query, "allowInsecure") || queryBool(query, "skip-cert-verify"),
		Obfs:         obfs,
		ObfsPassword: obfsPassword,
		UpMbps:       queryInt(query, "upmbps", "up"),
		DownMbps:     queryInt(query, "downmbps", "down"),
		Tag:          parsed.Fragment,
	}, nil
}

func BuildURL(password, host string, port int, options string, tag string) (string, error) {
	password = strings.TrimSpace(password)
	host = strings.TrimSpace(host)
	if password == "" {
		return "", fmt.Errorf("hysteria2 password is required")
	}
	if host == "" {
		return "", fmt.Errorf("hysteria2 host is required")
	}
	if port <= 0 || port > 65535 {
		return "", fmt.Errorf("invalid hysteria2 port: %d", port)
	}
	values := parseOptionString(options)
	if strings.TrimSpace(firstQuery(values, "sni", "servername", "server_name")) == "" {
		values.Set("sni", host)
	}
	u := &url.URL{
		Scheme:   "hysteria2",
		User:     url.User(password),
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
		return nil, fmt.Errorf("hysteria2 only supports tcp, got %s", network)
	}
	serverAddr := M.ParseSocksaddr(net.JoinHostPort(node.Host, strconv.Itoa(node.Port)))
	if !serverAddr.IsValid() {
		return nil, fmt.Errorf("invalid hysteria2 server address")
	}
	destination := M.ParseSocksaddr(targetAddr)
	if !destination.IsValid() {
		return nil, fmt.Errorf("invalid target address: %s", targetAddr)
	}

	client, err := hy2.NewClient(hy2.ClientOptions{
		Context:            ctx,
		Dialer:             N.SystemDialer,
		Logger:             noopLogger{},
		ServerAddress:      serverAddr,
		Password:           node.Password,
		SalamanderPassword: node.ObfsPassword,
		SendBPS:            uint64(node.UpMbps) * hysteria.MbpsToBps,
		ReceiveBPS:         uint64(node.DownMbps) * hysteria.MbpsToBps,
		TLSConfig:          newTLSConfig(node.SNI, node.Insecure),
		UDPDisabled:        true,
	})
	if err != nil {
		return nil, err
	}
	conn, err := client.DialConn(ctx, destination)
	if err != nil {
		_ = client.CloseWithError(err)
		return nil, err
	}
	return &clientConn{Conn: conn, closeClient: func() error {
		return client.CloseWithError(os.ErrClosed)
	}}, nil
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

type tlsConfig struct {
	config *tls.Config
}

func newTLSConfig(serverName string, insecure bool) aTLS.Config {
	return &tlsConfig{config: &tls.Config{
		ServerName:         serverName,
		InsecureSkipVerify: insecure,
		MinVersion:         tls.VersionTLS12,
	}}
}

func (c *tlsConfig) ServerName() string {
	return c.config.ServerName
}

func (c *tlsConfig) SetServerName(serverName string) {
	c.config.ServerName = serverName
}

func (c *tlsConfig) NextProtos() []string {
	return c.config.NextProtos
}

func (c *tlsConfig) SetNextProtos(nextProto []string) {
	c.config.NextProtos = nextProto
}

func (c *tlsConfig) STDConfig() (*tls.Config, error) {
	return c.config.Clone(), nil
}

func (c *tlsConfig) Client(conn net.Conn) (aTLS.Conn, error) {
	return tls.Client(conn, c.config.Clone()), nil
}

func (c *tlsConfig) Clone() aTLS.Config {
	return &tlsConfig{config: c.config.Clone()}
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

func queryInt(values url.Values, keys ...string) int {
	raw := firstQuery(values, keys...)
	if raw == "" {
		return 0
	}
	value, _ := strconv.Atoi(raw)
	if value < 0 {
		return 0
	}
	return value
}

type noopLogger struct{}

func (noopLogger) Trace(args ...any) {}
func (noopLogger) Debug(args ...any) {}
func (noopLogger) Info(args ...any)  {}
func (noopLogger) Warn(args ...any)  {}
func (noopLogger) Error(args ...any) {}
func (noopLogger) Fatal(args ...any) {}
func (noopLogger) Panic(args ...any) {}
