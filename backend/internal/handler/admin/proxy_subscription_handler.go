package admin

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/anytlsutil"
	"github.com/Wei-Shaw/sub2api/internal/pkg/httpclient"
	"github.com/Wei-Shaw/sub2api/internal/pkg/hysteria2util"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ssutil"
	"github.com/Wei-Shaw/sub2api/internal/pkg/trojanutil"
	"github.com/Wei-Shaw/sub2api/internal/pkg/vlessutil"
	"github.com/Wei-Shaw/sub2api/internal/util/urlvalidator"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

const (
	subscriptionFetchTimeout = 20 * time.Second
	subscriptionMaxBodyBytes = int64(8 * 1024 * 1024)

	subscriptionDefaultUserAgent = "sub2api-subscription-parser/1.0"
	subscriptionClashUserAgent   = "Clash"

	subscriptionImportModeDirect = "direct"
)

type ParseSubscriptionRequest struct {
	URL        string `json:"url" binding:"required"`
	ClientType string `json:"client_type"`
}

type ParsedSubscriptionProxy struct {
	Name         string `json:"name"`
	Protocol     string `json:"protocol"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	ImportMode   string `json:"import_mode"`
	NodeProtocol string `json:"node_protocol,omitempty"`
}

type ParseSubscriptionResponse struct {
	Proxies []ParsedSubscriptionProxy `json:"proxies"`
	Stats   SubscriptionParseStats    `json:"stats"`
}

type SubscriptionParseStats struct {
	ClientType                string         `json:"client_type"`
	UserAgent                 string         `json:"user_agent"`
	Format                    string         `json:"format"`
	Decoded                   bool           `json:"decoded"`
	DetectedProtocolCounts    map[string]int `json:"detected_protocol_counts"`
	SupportedProtocolCounts   map[string]int `json:"supported_protocol_counts"`
	UnsupportedProtocolCounts map[string]int `json:"unsupported_protocol_counts"`
	SupportedCount            int            `json:"supported_count"`
	ImportableCount           int            `json:"importable_count"`
	UnsupportedCount          int            `json:"unsupported_count"`
	Warnings                  []string       `json:"warnings"`
}

type subscriptionClientProfile struct {
	Type      string
	UserAgent string
	Accept    string
}

type subscriptionParseResult struct {
	Proxies                []ParsedSubscriptionProxy
	Format                 string
	Decoded                bool
	DetectedProtocolCounts map[string]int
	UnsupportedProtocols   map[string]int
}

type ProxySubscriptionHandler struct{}

func NewProxySubscriptionHandler() *ProxySubscriptionHandler {
	return &ProxySubscriptionHandler{}
}

func (h *ProxySubscriptionHandler) ParseSubscription(c *gin.Context) {
	var req ParseSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	validatedURL, err := urlvalidator.ValidateHTTPURL(req.URL, false, urlvalidator.ValidationOptions{
		AllowPrivate: false,
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := fetchAndParseSubscription(c.Request.Context(), validatedURL, req.ClientType)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, result)
}

func fetchAndParseSubscription(ctx context.Context, rawURL, clientType string) (*ParseSubscriptionResponse, error) {
	profiles, err := subscriptionProfilesForClientType(clientType)
	if err != nil {
		return nil, err
	}

	var firstParseErr error
	var firstUnsupported *ParseSubscriptionResponse

	for _, profile := range profiles {
		body, err := fetchSubscriptionBody(ctx, rawURL, profile)
		if err != nil {
			if firstParseErr != nil {
				return nil, firstParseErr
			}
			return nil, err
		}

		parsed, parseErr := parseSubscriptionBody(body)
		if parseErr != nil {
			if firstParseErr == nil {
				firstParseErr = parseErr
			}
			continue
		}

		resp := buildParseSubscriptionResponse(profile, parsed)
		if len(resp.Proxies) > 0 {
			return resp, nil
		}
		if resp.Stats.UnsupportedCount > 0 && firstUnsupported == nil {
			firstUnsupported = resp
		}
	}

	if firstUnsupported != nil {
		return firstUnsupported, nil
	}
	if firstParseErr != nil {
		return nil, firstParseErr
	}
	return nil, errors.New("no supported proxies found in subscription")
}

func subscriptionProfilesForClientType(raw string) ([]subscriptionClientProfile, error) {
	clientType := strings.ToLower(strings.TrimSpace(raw))
	if clientType == "" {
		clientType = "auto"
	}

	profiles := map[string]subscriptionClientProfile{
		"default": {
			Type:      "default",
			UserAgent: subscriptionDefaultUserAgent,
			Accept:    "text/plain, application/yaml, text/yaml, */*",
		},
		"clash": {
			Type:      "clash",
			UserAgent: subscriptionClashUserAgent,
			Accept:    "text/plain, application/yaml, text/yaml, */*",
		},
		"clash-meta": {
			Type:      "clash-meta",
			UserAgent: "clash.meta",
			Accept:    "*/*",
		},
		"mihomo": {
			Type:      "mihomo",
			UserAgent: "Mihomo/1.18.0",
			Accept:    "*/*",
		},
		"sing-box": {
			Type:      "sing-box",
			UserAgent: "sing-box/1.10.0",
			Accept:    "*/*",
		},
		"surge": {
			Type:      "surge",
			UserAgent: "Surge Mac/5.0",
			Accept:    "*/*",
		},
		"shadowrocket": {
			Type:      "shadowrocket",
			UserAgent: "Shadowrocket/1995 CFNetwork/1408.0.4 Darwin/22.5.0",
			Accept:    "*/*",
		},
		"stash": {
			Type:      "stash",
			UserAgent: "Stash/2.0.0",
			Accept:    "*/*",
		},
		"quantumult-x": {
			Type:      "quantumult-x",
			UserAgent: "Quantumult X/1.0.30",
			Accept:    "*/*",
		},
		"loon": {
			Type:      "loon",
			UserAgent: "Loon/3.2.0",
			Accept:    "*/*",
		},
		"v2rayn": {
			Type:      "v2rayn",
			UserAgent: "v2rayN/6.0",
			Accept:    "*/*",
		},
	}

	if clientType == "auto" {
		return []subscriptionClientProfile{profiles["default"], profiles["clash"]}, nil
	}

	profile, ok := profiles[clientType]
	if !ok {
		return nil, fmt.Errorf("unsupported subscription client type %q", clientType)
	}
	return []subscriptionClientProfile{profile}, nil
}

func fetchSubscriptionBody(ctx context.Context, rawURL string, profile subscriptionClientProfile) ([]byte, error) {
	client, err := httpclient.GetClient(httpclient.Options{
		Timeout:            subscriptionFetchTimeout,
		ValidateResolvedIP: true,
		AllowPrivateHosts:  false,
	})
	if err != nil {
		return nil, fmt.Errorf("build subscription client: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build subscription request: %w", err)
	}
	req.Header.Set("Accept", profile.Accept)
	req.Header.Set("User-Agent", profile.UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch subscription failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subscription request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, subscriptionMaxBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("read subscription body failed: %w", err)
	}
	if len(body) == 0 {
		return nil, errors.New("subscription response is empty")
	}
	return body, nil
}

func parseSubscriptionBody(body []byte) (*subscriptionParseResult, error) {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return nil, errors.New("subscription response is empty")
	}

	if result := parseSubscriptionTextDetailed(trimmed); result.hasSignals() {
		return result, nil
	}

	if result, err := parseClashSubscriptionYAMLDetailed([]byte(trimmed)); err == nil && result.hasSignals() {
		return result, nil
	}

	decoded, err := decodeSubscriptionBase64(trimmed)
	if err == nil && !bytes.Equal(decoded, body) {
		if result, err := parseSubscriptionBody(decoded); err == nil && result.hasSignals() {
			result.Decoded = true
			return result, nil
		}
	}

	return nil, errors.New("no supported proxies found in subscription")
}

func parseSubscriptionText(raw string) ([]ParsedSubscriptionProxy, error) {
	result := parseSubscriptionTextDetailed(raw)
	if len(result.Proxies) == 0 {
		return nil, errors.New("no supported proxies found")
	}
	return result.Proxies, nil
}

func parseSubscriptionTextDetailed(raw string) *subscriptionParseResult {
	lines := strings.FieldsFunc(raw, func(r rune) bool {
		return r == '\n' || r == '\r'
	})
	if len(lines) == 0 {
		return newSubscriptionParseResult("text")
	}

	result := newSubscriptionParseResult("text")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parsedURL, err := url.Parse(line)
		if err != nil {
			continue
		}
		scheme := strings.ToLower(strings.TrimSpace(parsedURL.Scheme))
		if !isSubscriptionProxyProtocol(scheme) {
			continue
		}
		incrementCount(result.DetectedProtocolCounts, normalizeDetectedProtocol(scheme))

		proxy, supported := parsedSubscriptionProxyFromURL(parsedURL)
		if supported {
			result.Proxies = append(result.Proxies, proxy)
			continue
		}
		incrementCount(result.UnsupportedProtocols, normalizeDetectedProtocol(scheme))
	}

	return result
}

func parseClashSubscriptionYAML(body []byte) ([]ParsedSubscriptionProxy, error) {
	result, err := parseClashSubscriptionYAMLDetailed(body)
	if err != nil {
		return nil, err
	}
	if len(result.Proxies) == 0 {
		return nil, errors.New("clash subscription contains no supported proxies")
	}
	return result.Proxies, nil
}

func parseClashSubscriptionYAMLDetailed(body []byte) (*subscriptionParseResult, error) {
	type clashConfig struct {
		Proxies []map[string]any `yaml:"proxies"`
	}

	var cfg clashConfig
	if err := yaml.Unmarshal(body, &cfg); err != nil {
		return nil, err
	}
	if len(cfg.Proxies) == 0 {
		return newSubscriptionParseResult("clash"), nil
	}

	result := newSubscriptionParseResult("clash")
	for _, item := range cfg.Proxies {
		protocol := normalizeDetectedProtocol(yamlString(item, "type"))
		if protocol == "" {
			continue
		}
		incrementCount(result.DetectedProtocolCounts, protocol)

		host := strings.TrimSpace(yamlString(item, "server"))
		port := yamlInt(item, "port")
		if host == "" || port <= 0 || port > 65535 {
			incrementCount(result.UnsupportedProtocols, protocol)
			continue
		}

		switch protocol {
		case "ss":
			method := strings.TrimSpace(yamlString(item, "cipher"))
			password := strings.TrimSpace(yamlString(item, "password"))
			if method != "" && password != "" && strings.TrimSpace(yamlString(item, "plugin")) == "" {
				result.Proxies = append(result.Proxies, ParsedSubscriptionProxy{
					Name:         subscriptionProxyName(yamlString(item, "name"), host, port),
					Protocol:     "ss",
					Host:         host,
					Port:         port,
					Username:     method,
					Password:     password,
					ImportMode:   subscriptionImportModeDirect,
					NodeProtocol: "ss",
				})
				continue
			}
		case "anytls":
			password := strings.TrimSpace(yamlString(item, "password"))
			sni := strings.TrimSpace(firstYAMLString(item, "sni", "servername", "server_name"))
			if password != "" {
				if sni == "" {
					sni = host
				}
				node := &anytlsutil.Node{
					Password: password,
					Host:     host,
					Port:     port,
					SNI:      sni,
					Insecure: yamlBool(item, "skip-cert-verify") || yamlBool(item, "insecure"),
				}
				result.Proxies = append(result.Proxies, ParsedSubscriptionProxy{
					Name:         subscriptionProxyName(yamlString(item, "name"), host, port),
					Protocol:     "anytls",
					Host:         host,
					Port:         port,
					Username:     node.OptionsString(),
					Password:     password,
					ImportMode:   subscriptionImportModeDirect,
					NodeProtocol: "anytls",
				})
				continue
			}
		case "trojan":
			password := strings.TrimSpace(yamlString(item, "password"))
			transport := strings.ToLower(strings.TrimSpace(firstYAMLString(item, "network", "transport")))
			if password != "" && (transport == "" || transport == "tcp") {
				node := &trojanutil.Node{
					Password: password,
					Host:     host,
					Port:     port,
					SNI:      defaultString(firstYAMLString(item, "sni", "servername", "server_name", "peer"), host),
					Insecure: yamlBool(item, "skip-cert-verify") || yamlBool(item, "insecure"),
				}
				result.Proxies = append(result.Proxies, ParsedSubscriptionProxy{
					Name:         subscriptionProxyName(yamlString(item, "name"), host, port),
					Protocol:     "trojan",
					Host:         host,
					Port:         port,
					Username:     node.OptionsString(),
					Password:     password,
					ImportMode:   subscriptionImportModeDirect,
					NodeProtocol: "trojan",
				})
				continue
			}
		case "vless":
			uuid := strings.TrimSpace(firstYAMLString(item, "uuid", "id"))
			transport := strings.ToLower(strings.TrimSpace(firstYAMLString(item, "network", "transport")))
			security := "none"
			if yamlBool(item, "tls") {
				security = "tls"
			}
			if value := strings.ToLower(strings.TrimSpace(yamlString(item, "security"))); value != "" {
				security = value
			}
			if _, hasReality := item["reality-opts"]; uuid != "" && !hasReality && (transport == "" || transport == "tcp") && (security == "none" || security == "tls") {
				node := &vlessutil.Node{
					UUID:     uuid,
					Host:     host,
					Port:     port,
					Security: security,
					SNI:      defaultString(firstYAMLString(item, "sni", "servername", "server_name"), host),
					Flow:     strings.TrimSpace(yamlString(item, "flow")),
					Insecure: yamlBool(item, "skip-cert-verify") || yamlBool(item, "insecure"),
				}
				if _, err := vlessutil.ParseURL(mustBuildVLESSURL(node)); err == nil {
					result.Proxies = append(result.Proxies, ParsedSubscriptionProxy{
						Name:         subscriptionProxyName(yamlString(item, "name"), host, port),
						Protocol:     "vless",
						Host:         host,
						Port:         port,
						Username:     node.OptionsString(),
						Password:     uuid,
						ImportMode:   subscriptionImportModeDirect,
						NodeProtocol: "vless",
					})
					continue
				}
			}
		case "hysteria2":
			password := strings.TrimSpace(yamlString(item, "password"))
			obfs := strings.ToLower(strings.TrimSpace(yamlString(item, "obfs")))
			obfsPassword := strings.TrimSpace(firstYAMLString(item, "obfs-password", "obfs_password"))
			if password != "" && (obfs == "" || (obfs == "salamander" && obfsPassword != "")) {
				node := &hysteria2util.Node{
					Password:     password,
					Host:         host,
					Port:         port,
					SNI:          defaultString(firstYAMLString(item, "sni", "servername", "server_name"), host),
					Insecure:     yamlBool(item, "skip-cert-verify") || yamlBool(item, "insecure"),
					Obfs:         obfs,
					ObfsPassword: obfsPassword,
					UpMbps:       firstYAMLInt(item, "upmbps", "up"),
					DownMbps:     firstYAMLInt(item, "downmbps", "down"),
				}
				result.Proxies = append(result.Proxies, ParsedSubscriptionProxy{
					Name:         subscriptionProxyName(yamlString(item, "name"), host, port),
					Protocol:     "hysteria2",
					Host:         host,
					Port:         port,
					Username:     node.OptionsString(),
					Password:     password,
					ImportMode:   subscriptionImportModeDirect,
					NodeProtocol: "hysteria2",
				})
				continue
			}
		case "http", "https", "socks5", "socks5h":
			result.Proxies = append(result.Proxies, ParsedSubscriptionProxy{
				Name:         subscriptionProxyName(yamlString(item, "name"), host, port),
				Protocol:     protocol,
				Host:         host,
				Port:         port,
				Username:     strings.TrimSpace(yamlString(item, "username")),
				Password:     strings.TrimSpace(yamlString(item, "password")),
				ImportMode:   subscriptionImportModeDirect,
				NodeProtocol: protocol,
			})
			continue
		}

		incrementCount(result.UnsupportedProtocols, protocol)
	}

	return result, nil
}

func decodeSubscriptionBase64(raw string) ([]byte, error) {
	compact := strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\r', '\n':
			return -1
		default:
			return r
		}
	}, raw)
	if compact == "" {
		return nil, errors.New("empty base64 payload")
	}

	encodings := []*base64.Encoding{
		base64.RawURLEncoding,
		base64.URLEncoding,
		base64.RawStdEncoding,
		base64.StdEncoding,
	}
	for _, encoding := range encodings {
		decoded, err := encoding.DecodeString(compact)
		if err == nil {
			return decoded, nil
		}
	}

	padded := compact + strings.Repeat("=", (4-len(compact)%4)%4)
	for _, encoding := range encodings {
		decoded, err := encoding.DecodeString(padded)
		if err == nil {
			return decoded, nil
		}
	}
	return nil, errors.New("invalid base64 payload")
}

func subscriptionProxyName(name, host string, port int) string {
	name = strings.TrimSpace(name)
	if name != "" {
		return name
	}
	return fmt.Sprintf("%s:%d", strings.TrimSpace(host), port)
}

func newSubscriptionParseResult(format string) *subscriptionParseResult {
	return &subscriptionParseResult{
		Format:                 format,
		DetectedProtocolCounts: make(map[string]int),
		UnsupportedProtocols:   make(map[string]int),
	}
}

func (r *subscriptionParseResult) hasSignals() bool {
	return r != nil && (len(r.Proxies) > 0 || len(r.DetectedProtocolCounts) > 0 || len(r.UnsupportedProtocols) > 0)
}

func buildParseSubscriptionResponse(profile subscriptionClientProfile, result *subscriptionParseResult) *ParseSubscriptionResponse {
	if result == nil {
		result = newSubscriptionParseResult("")
	}

	supportedCounts := make(map[string]int)
	for _, proxy := range result.Proxies {
		incrementCount(supportedCounts, proxy.Protocol)
	}

	warnings := make([]string, 0, 1)
	if len(result.UnsupportedProtocols) > 0 {
		warnings = append(warnings, "Some detected proxy protocols are not supported by the built-in outbound dialer and were not imported.")
	}

	return &ParseSubscriptionResponse{
		Proxies: result.Proxies,
		Stats: SubscriptionParseStats{
			ClientType:                profile.Type,
			UserAgent:                 profile.UserAgent,
			Format:                    result.Format,
			Decoded:                   result.Decoded,
			DetectedProtocolCounts:    sortedCopyCounts(result.DetectedProtocolCounts),
			SupportedProtocolCounts:   sortedCopyCounts(supportedCounts),
			UnsupportedProtocolCounts: sortedCopyCounts(result.UnsupportedProtocols),
			SupportedCount:            sumCounts(supportedCounts),
			ImportableCount:           len(result.Proxies),
			UnsupportedCount:          sumCounts(result.UnsupportedProtocols),
			Warnings:                  warnings,
		},
	}
}

func parsedSubscriptionProxyFromURL(parsed *url.URL) (ParsedSubscriptionProxy, bool) {
	if parsed == nil {
		return ParsedSubscriptionProxy{}, false
	}
	scheme := normalizeDetectedProtocol(parsed.Scheme)
	if !isSupportedImportProtocol(scheme) {
		return ParsedSubscriptionProxy{}, false
	}

	if scheme == "ss" {
		node, err := ssutil.ParseParsedURL(parsed)
		if err != nil || node == nil || node.Plugin != "" {
			return ParsedSubscriptionProxy{}, false
		}
		return ParsedSubscriptionProxy{
			Name:         subscriptionProxyName(node.Tag, node.Host, node.Port),
			Protocol:     "ss",
			Host:         node.Host,
			Port:         node.Port,
			Username:     node.Method,
			Password:     node.Password,
			ImportMode:   subscriptionImportModeDirect,
			NodeProtocol: "ss",
		}, true
	}

	if scheme == "anytls" {
		node, err := anytlsutil.ParseParsedURL(parsed)
		if err != nil || node == nil {
			return ParsedSubscriptionProxy{}, false
		}
		return ParsedSubscriptionProxy{
			Name:         subscriptionProxyName(node.Tag, node.Host, node.Port),
			Protocol:     "anytls",
			Host:         node.Host,
			Port:         node.Port,
			Username:     node.OptionsString(),
			Password:     node.Password,
			ImportMode:   subscriptionImportModeDirect,
			NodeProtocol: "anytls",
		}, true
	}

	if scheme == "trojan" {
		node, err := trojanutil.ParseParsedURL(parsed)
		if err != nil || node == nil {
			return ParsedSubscriptionProxy{}, false
		}
		return ParsedSubscriptionProxy{
			Name:         subscriptionProxyName(node.Tag, node.Host, node.Port),
			Protocol:     "trojan",
			Host:         node.Host,
			Port:         node.Port,
			Username:     node.OptionsString(),
			Password:     node.Password,
			ImportMode:   subscriptionImportModeDirect,
			NodeProtocol: "trojan",
		}, true
	}

	if scheme == "vless" {
		node, err := vlessutil.ParseParsedURL(parsed)
		if err != nil || node == nil {
			return ParsedSubscriptionProxy{}, false
		}
		return ParsedSubscriptionProxy{
			Name:         subscriptionProxyName(node.Tag, node.Host, node.Port),
			Protocol:     "vless",
			Host:         node.Host,
			Port:         node.Port,
			Username:     node.OptionsString(),
			Password:     node.UUID,
			ImportMode:   subscriptionImportModeDirect,
			NodeProtocol: "vless",
		}, true
	}

	if scheme == "hysteria2" {
		node, err := hysteria2util.ParseParsedURL(parsed)
		if err != nil || node == nil {
			return ParsedSubscriptionProxy{}, false
		}
		return ParsedSubscriptionProxy{
			Name:         subscriptionProxyName(node.Tag, node.Host, node.Port),
			Protocol:     "hysteria2",
			Host:         node.Host,
			Port:         node.Port,
			Username:     node.OptionsString(),
			Password:     node.Password,
			ImportMode:   subscriptionImportModeDirect,
			NodeProtocol: "hysteria2",
		}, true
	}

	host := strings.TrimSpace(parsed.Hostname())
	port, err := strconv.Atoi(parsed.Port())
	if host == "" || err != nil || port <= 0 || port > 65535 {
		return ParsedSubscriptionProxy{}, false
	}

	username := ""
	password := ""
	if parsed.User != nil {
		username = parsed.User.Username()
		password, _ = parsed.User.Password()
	}

	return ParsedSubscriptionProxy{
		Name:         subscriptionProxyName(parsed.Fragment, host, port),
		Protocol:     scheme,
		Host:         host,
		Port:         port,
		Username:     username,
		Password:     password,
		ImportMode:   subscriptionImportModeDirect,
		NodeProtocol: scheme,
	}, true
}

func isSubscriptionProxyProtocol(protocol string) bool {
	switch normalizeDetectedProtocol(protocol) {
	case "http", "https", "socks5", "socks5h", "ss", "ssr", "vmess", "vless", "trojan", "hysteria", "hysteria2", "hy2", "tuic", "anytls":
		return true
	default:
		return false
	}
}

func isSupportedImportProtocol(protocol string) bool {
	return isAdminProxyProtocolSupported(protocol)
}

func normalizeDetectedProtocol(protocol string) string {
	return normalizeAdminProxyProtocol(protocol)
}

func yamlString(item map[string]any, key string) string {
	if item == nil {
		return ""
	}
	value, ok := item[key]
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func firstYAMLString(item map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := yamlString(item, key); value != "" {
			return value
		}
	}
	return ""
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return strings.TrimSpace(fallback)
}

func firstYAMLInt(item map[string]any, keys ...string) int {
	for _, key := range keys {
		if value := yamlInt(item, key); value != 0 {
			return value
		}
	}
	return 0
}

func mustBuildVLESSURL(node *vlessutil.Node) string {
	if node == nil {
		return ""
	}
	raw, _ := vlessutil.BuildURL(node.UUID, node.Host, node.Port, node.OptionsString(), node.Tag)
	return raw
}

func yamlBool(item map[string]any, key string) bool {
	if item == nil {
		return false
	}
	value, ok := item[key]
	if !ok {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "1", "true", "yes", "on":
			return true
		default:
			return false
		}
	default:
		return strings.EqualFold(strings.TrimSpace(fmt.Sprint(typed)), "true")
	}
}

func yamlInt(item map[string]any, key string) int {
	if item == nil {
		return 0
	}
	value, ok := item[key]
	if !ok {
		return 0
	}
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		port, _ := strconv.Atoi(strings.TrimSpace(typed))
		return port
	default:
		port, _ := strconv.Atoi(strings.TrimSpace(fmt.Sprint(typed)))
		return port
	}
}

func normalizeYAMLStringMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = normalizeYAMLValue(value)
	}
	return out
}

func normalizeYAMLValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return normalizeYAMLStringMap(typed)
	case map[any]any:
		out := make(map[string]any, len(typed))
		for key, value := range typed {
			out[fmt.Sprint(key)] = normalizeYAMLValue(value)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i := range typed {
			out[i] = normalizeYAMLValue(typed[i])
		}
		return out
	default:
		return typed
	}
}

func incrementCount(counts map[string]int, protocol string) {
	protocol = normalizeDetectedProtocol(protocol)
	if protocol == "" {
		return
	}
	counts[protocol]++
}

func sumCounts(counts map[string]int) int {
	total := 0
	for _, count := range counts {
		total += count
	}
	return total
}

func sortedCopyCounts(counts map[string]int) map[string]int {
	if len(counts) == 0 {
		return map[string]int{}
	}
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make(map[string]int, len(counts))
	for _, key := range keys {
		out[key] = counts[key]
	}
	return out
}
