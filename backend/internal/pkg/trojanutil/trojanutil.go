package trojanutil

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
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
	}
	return values.Encode()
}

func ParseURL(raw string) (*Node, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("trojan url is empty")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("invalid trojan url: %w", err)
	}
	return ParseParsedURL(parsed)
}

func ParseParsedURL(parsed *url.URL) (*Node, error) {
	if parsed == nil {
		return nil, fmt.Errorf("trojan url is nil")
	}
	if strings.ToLower(strings.TrimSpace(parsed.Scheme)) != "trojan" {
		return nil, fmt.Errorf("invalid trojan scheme: %s", parsed.Scheme)
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
		return nil, fmt.Errorf("trojan password is required")
	}

	query, _ := url.ParseQuery(parsed.RawQuery)
	security := strings.ToLower(strings.TrimSpace(query.Get("security")))
	if security != "" && security != "tls" {
		return nil, fmt.Errorf("trojan only supports tls security")
	}
	transport := strings.ToLower(strings.TrimSpace(firstQuery(query, "type", "network")))
	if transport != "" && transport != "tcp" {
		return nil, fmt.Errorf("trojan transport %q is not supported", transport)
	}
	sni := strings.TrimSpace(firstQuery(query, "sni", "peer", "servername", "server_name"))
	if sni == "" {
		sni = host
	}

	return &Node{
		Password: password,
		Host:     host,
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
		return "", fmt.Errorf("trojan password is required")
	}
	if host == "" {
		return "", fmt.Errorf("trojan host is required")
	}
	if port <= 0 || port > 65535 {
		return "", fmt.Errorf("invalid trojan port: %d", port)
	}
	values := parseOptionString(options)
	if strings.TrimSpace(firstQuery(values, "sni", "peer", "servername", "server_name")) == "" {
		values.Set("sni", host)
	}
	u := &url.URL{
		Scheme:   "trojan",
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
		return nil, fmt.Errorf("trojan only supports tcp, got %s", network)
	}
	dialer := tls.Dialer{
		NetDialer: &net.Dialer{},
		Config: &tls.Config{
			ServerName:         node.SNI,
			InsecureSkipVerify: node.Insecure,
			MinVersion:         tls.VersionTLS12,
		},
	}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(node.Host, strconv.Itoa(node.Port)))
	if err != nil {
		return nil, err
	}
	if err := writeTrojanConnect(conn, node.Password, targetAddr); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

func writeTrojanConnect(conn net.Conn, password, targetAddr string) error {
	sum := sha256.Sum224([]byte(password))
	handshake := make([]byte, 0, 64+len(targetAddr))
	handshake = append(handshake, []byte(hex.EncodeToString(sum[:]))...)
	handshake = append(handshake, '\r', '\n', 0x01)
	var err error
	handshake, err = appendSocksAddr(handshake, targetAddr)
	if err != nil {
		return err
	}
	handshake = append(handshake, '\r', '\n')
	_, err = conn.Write(handshake)
	return err
}

func appendSocksAddr(out []byte, addr string) ([]byte, error) {
	host, portString, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid target address: %s", addr)
	}
	port, err := strconv.Atoi(portString)
	if err != nil || port <= 0 || port > 65535 {
		return nil, fmt.Errorf("invalid target port: %s", portString)
	}
	if ip, err := netip.ParseAddr(host); err == nil {
		if ip.Is4() {
			v := ip.As4()
			out = append(out, 0x01)
			out = append(out, v[:]...)
		} else {
			v := ip.As16()
			out = append(out, 0x04)
			out = append(out, v[:]...)
		}
	} else {
		if len(host) == 0 || len(host) > 255 {
			return nil, fmt.Errorf("invalid target host: %s", host)
		}
		out = append(out, 0x03, byte(len(host)))
		out = append(out, []byte(host)...)
	}
	out = binary.BigEndian.AppendUint16(out, uint16(port))
	return out, nil
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
