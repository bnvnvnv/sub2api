package repository

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/proxyurl"
	"github.com/imroc/req/v3"
	reqhttp2 "github.com/imroc/req/v3/http2"
	utls "github.com/refraction-networking/utls"
)

var (
	claudeOAuthChromeHTTP2Settings = []reqhttp2.Setting{
		{ID: reqhttp2.SettingHeaderTableSize, Val: 65536},
		{ID: reqhttp2.SettingEnablePush, Val: 0},
		{ID: reqhttp2.SettingMaxConcurrentStreams, Val: 1000},
		{ID: reqhttp2.SettingInitialWindowSize, Val: 6291456},
		{ID: reqhttp2.SettingMaxHeaderListSize, Val: 262144},
	}

	claudeOAuthChromePseudoHeaderOrder = []string{
		":method",
		":authority",
		":scheme",
		":path",
	}

	// Token endpoints are API-style requests. Keep Chrome TLS/H2 framing but do
	// not inject browser-only defaults like sec-ch-ua or navigation headers.
	claudeOAuthTokenHeaderOrder = []string{
		"user-agent",
		"accept",
		"content-type",
		"origin",
		"referer",
		"accept-language",
		"cookie",
	}

	claudeOAuthChromeHeaderPriority = reqhttp2.PriorityParam{
		StreamDep: 0,
		Exclusive: true,
		Weight:    255,
	}
)

func createClaudeOAuthBrowserClient(proxyURL string) (*req.Client, error) {
	client := req.C().
		SetTimeout(60 * time.Second).
		ImpersonateChrome().
		SetCookieJar(nil)

	return applyClaudeOAuthProxy(client, proxyURL)
}

func createClaudeOAuthTokenClient(proxyURL string) (*req.Client, error) {
	client := req.C().
		SetTimeout(60 * time.Second).
		SetCookieJar(nil).
		SetTLSFingerprint(utls.HelloChrome_120).
		SetHTTP2SettingsFrame(claudeOAuthChromeHTTP2Settings...).
		SetHTTP2ConnectionFlow(15663105).
		SetCommonPseudoHeaderOder(claudeOAuthChromePseudoHeaderOrder...).
		SetCommonHeaderOrder(claudeOAuthTokenHeaderOrder...).
		SetHTTP2HeaderPriority(claudeOAuthChromeHeaderPriority)

	return applyClaudeOAuthProxy(client, proxyURL)
}

func applyClaudeOAuthProxy(client *req.Client, proxyURL string) (*req.Client, error) {
	trimmed, _, err := proxyurl.Parse(proxyURL)
	if err != nil {
		return nil, err
	}
	if trimmed != "" {
		client.SetProxyURL(trimmed)
	}
	return client, nil
}
