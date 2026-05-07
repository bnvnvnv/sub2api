package admin

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ssutil"
	"github.com/stretchr/testify/require"
)

func TestParseSubscriptionText_SupportsStandardAndLegacySS(t *testing.T) {
	standard, err := ssutil.BuildURL("aes-256-gcm", "secret", "ss.example.com", 8388, "std")
	require.NoError(t, err)
	withPlugin, err := ssutil.BuildURL("aes-256-gcm", "secret", "plugin.example.com", 8388, "")
	require.NoError(t, err)

	legacyPayload := base64.RawURLEncoding.EncodeToString([]byte("aes-128-gcm:pass@legacy.example.com:443"))
	legacy := "ss://" + legacyPayload + "#legacy"

	result := parseSubscriptionTextDetailed(standard + "\n" + legacy + "\n" + withPlugin + "?plugin=obfs-local")
	require.Len(t, result.Proxies, 2)
	require.Equal(t, 1, result.UnsupportedProtocols["ss"])

	require.Equal(t, "ss", result.Proxies[0].Protocol)
	require.Equal(t, "std", result.Proxies[0].Name)
	require.Equal(t, "aes-256-gcm", result.Proxies[0].Username)
	require.Equal(t, "secret", result.Proxies[0].Password)

	require.Equal(t, "ss", result.Proxies[1].Protocol)
	require.Equal(t, "legacy", result.Proxies[1].Name)
	require.Equal(t, "legacy.example.com", result.Proxies[1].Host)
	require.Equal(t, 443, result.Proxies[1].Port)
	require.Equal(t, "aes-128-gcm", result.Proxies[1].Username)
	require.Equal(t, "pass", result.Proxies[1].Password)
}

func TestParseClashSubscriptionYAML_DirectOnly(t *testing.T) {
	body := []byte(`
proxies:
  - name: ss-ok
    type: ss
    server: ss.example.com
    port: 8388
    cipher: aes-256-gcm
    password: secret
  - name: ss-plugin
    type: ss
    server: plugin.example.com
    port: 443
    cipher: aes-128-gcm
    password: blocked
    plugin: obfs-local
  - name: vmess-skip
    type: vmess
    server: vmess.example.com
    port: 443
`)

	result, err := parseClashSubscriptionYAMLDetailed(body)
	require.NoError(t, err)
	require.Len(t, result.Proxies, 1)
	require.Equal(t, "ss-ok", result.Proxies[0].Name)
	require.Equal(t, "ss", result.Proxies[0].Protocol)
	require.Equal(t, "aes-256-gcm", result.Proxies[0].Username)
	require.Equal(t, "secret", result.Proxies[0].Password)
	require.Equal(t, subscriptionImportModeDirect, result.Proxies[0].ImportMode)
	require.Equal(t, 1, result.UnsupportedProtocols["ss"])
	require.Equal(t, 1, result.UnsupportedProtocols["vmess"])
}

func TestSubscriptionProfilesForClientType(t *testing.T) {
	profiles, err := subscriptionProfilesForClientType("")
	require.NoError(t, err)
	require.Len(t, profiles, 2)
	require.Equal(t, "default", profiles[0].Type)
	require.Equal(t, "clash", profiles[1].Type)

	profiles, err = subscriptionProfilesForClientType("mihomo")
	require.NoError(t, err)
	require.Len(t, profiles, 1)
	require.Equal(t, "mihomo", profiles[0].Type)
	require.Equal(t, "Mihomo/1.18.0", profiles[0].UserAgent)

	_, err = subscriptionProfilesForClientType("unknown")
	require.Error(t, err)
}

func TestParseSubscriptionBody_Base64AnyTLSDirectStats(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("anytls://secret@example.com:443?sni=example.com&fp=chrome#node"))

	result, err := parseSubscriptionBody([]byte(encoded))
	require.NoError(t, err)
	require.True(t, result.Decoded)
	require.Len(t, result.Proxies, 1)
	require.Equal(t, subscriptionImportModeDirect, result.Proxies[0].ImportMode)
	require.Equal(t, "anytls", result.Proxies[0].NodeProtocol)
	require.Contains(t, result.Proxies[0].Username, "sni=example.com")
	require.Contains(t, result.Proxies[0].Username, "insecure=0")
	require.Equal(t, 1, result.DetectedProtocolCounts["anytls"])
	require.Empty(t, result.UnsupportedProtocols)

	resp := buildParseSubscriptionResponse(subscriptionClientProfile{
		Type:      "mihomo",
		UserAgent: "Mihomo/1.18.0",
	}, result)
	require.Equal(t, 1, resp.Stats.SupportedCount)
	require.Equal(t, 1, resp.Stats.ImportableCount)
	require.Equal(t, 0, resp.Stats.UnsupportedCount)
	require.Equal(t, map[string]int{"anytls": 1}, resp.Stats.SupportedProtocolCounts)
	require.Empty(t, resp.Stats.Warnings)
}

func TestParseSubscriptionText_DirectAndUnsupportedProtocols(t *testing.T) {
	vmessConfig := map[string]any{
		"v":    "2",
		"ps":   "vmess-node",
		"add":  "vmess.example.com",
		"port": "443",
		"id":   "11111111-1111-1111-1111-111111111111",
		"net":  "ws",
		"host": "cdn.example.com",
		"path": "/ws",
		"tls":  "tls",
	}
	vmessPayload, err := json.Marshal(vmessConfig)
	require.NoError(t, err)
	vmess := "vmess://" + base64.RawURLEncoding.EncodeToString(vmessPayload)

	vlessOK := "vless://22222222-2222-2222-2222-222222222222@vless.example.com:443?security=tls&sni=vless.example.com&type=tcp#vless-ok"
	vlessReality := "vless://33333333-3333-3333-3333-333333333333@reality.example.com:443?security=reality&sni=reality.example.com&fp=chrome&pbk=public-key&sid=01&type=tcp#vless-reality"
	trojanOK := "trojan://secret@trojan.example.com:443?sni=trojan.example.com&type=tcp#trojan-ok"
	trojanGRPC := "trojan://secret@grpc.example.com:443?sni=grpc.example.com&type=grpc&serviceName=svc#trojan-grpc"
	hy2 := "hy2://hy2pass@hy2.example.com:443?sni=hy2.example.com&obfs=salamander&obfs-password=obfs#hy2-node"
	tuic := "tuic://44444444-4444-4444-4444-444444444444:tuicpass@tuic.example.com:443?sni=tuic.example.com#tuic-node"

	result := parseSubscriptionTextDetailed(vmess + "\n" + vlessOK + "\n" + vlessReality + "\n" + trojanOK + "\n" + trojanGRPC + "\n" + hy2 + "\n" + tuic)
	require.Len(t, result.Proxies, 3)
	require.Equal(t, "vless", result.Proxies[0].Protocol)
	require.Equal(t, "22222222-2222-2222-2222-222222222222", result.Proxies[0].Password)
	require.Contains(t, result.Proxies[0].Username, "security=tls")
	require.Equal(t, "trojan", result.Proxies[1].Protocol)
	require.Equal(t, "secret", result.Proxies[1].Password)
	require.Contains(t, result.Proxies[1].Username, "sni=trojan.example.com")
	require.Equal(t, "hysteria2", result.Proxies[2].Protocol)
	require.Equal(t, "hy2pass", result.Proxies[2].Password)
	require.Contains(t, result.Proxies[2].Username, "obfs=salamander")

	require.Equal(t, 1, result.UnsupportedProtocols["vmess"])
	require.Equal(t, 1, result.UnsupportedProtocols["vless"])
	require.Equal(t, 1, result.UnsupportedProtocols["trojan"])
	require.Equal(t, 1, result.UnsupportedProtocols["tuic"])
}

func TestParseClashSubscriptionYAML_MixedProtocols(t *testing.T) {
	body := []byte(`
proxies:
  - name: http-ok
    type: http
    server: http.example.com
    port: 8080
    username: user
    password: pass
  - name: socks-ok
    type: socks5
    server: socks.example.com
    port: 1080
  - name: anytls-ok
    type: anytls
    server: anytls.example.com
    port: 443
    password: anypass
    sni: anytls.example.com
  - name: trojan-ok
    type: trojan
    server: trojan.example.com
    port: 443
    password: tpass
    sni: trojan.example.com
  - name: vless-ok
    type: vless
    server: vless.example.com
    port: 443
    uuid: 55555555-5555-5555-5555-555555555555
    tls: true
    servername: vless.example.com
  - name: hy2-ok
    type: hysteria2
    server: hy2.example.com
    port: 443
    password: hy2pass
    sni: hy2.example.com
    obfs: salamander
    obfs-password: obfs
  - name: vmess-skip
    type: vmess
    server: vmess.example.com
    port: 443
`)

	result, err := parseClashSubscriptionYAMLDetailed(body)
	require.NoError(t, err)
	require.Len(t, result.Proxies, 6)
	require.Equal(t, "http", result.Proxies[0].Protocol)
	require.Equal(t, "http.example.com", result.Proxies[0].Host)
	require.Equal(t, "user", result.Proxies[0].Username)
	require.Equal(t, "socks5", result.Proxies[1].Protocol)
	require.Equal(t, "anytls", result.Proxies[2].Protocol)
	require.Equal(t, "trojan", result.Proxies[3].Protocol)
	require.Equal(t, "vless", result.Proxies[4].Protocol)
	require.Equal(t, "hysteria2", result.Proxies[5].Protocol)
	require.Equal(t, 1, result.DetectedProtocolCounts["vmess"])
	require.Equal(t, 1, result.UnsupportedProtocols["vmess"])
}
