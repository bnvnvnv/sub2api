type CustomLocaleMessages = Record<string, any>

const common = {
  en: {
    apply: 'Apply',
    clear: 'Clear',
    creating: 'Creating...',
    login: 'Login',
    required: 'is required',
    sending: 'Sending...',
    tryAgain: 'Please try again'
  },
  zh: {
    apply: '应用',
    clear: '清空',
    creating: '创建中...',
    login: '登录',
    required: '不能为空',
    sending: '发送中...',
    tryAgain: '请重试'
  }
} as const

const admin = {
  en: {
    accounts: {
      cpaImport: 'Import CPA',
      cpaImportTitle: 'Import from CLIProxyAPI',
      cpaImportModeRemote: 'Remote API',
      cpaImportModeFile: 'JSON File',
      cpaImportRemoteHint: 'Connect to a CLIProxyAPI management endpoint and preview importable accounts.',
      cpaImportHint: 'Upload a CLIProxyAPI export JSON file and preview the account before importing.',
      cpaImportRemoteWarning: 'Only normal supported accounts are imported. Existing accounts are detected by source key and can be updated.',
      cpaImportWarning: 'Imported accounts may contain tokens and proxy data. Review the preview carefully before continuing.',
      cpaRemoteBaseUrl: 'CLIProxyAPI Base URL',
      cpaRemoteBaseUrlPlaceholder: 'https://cliproxy.example.com',
      cpaRemoteManagementKey: 'Management Key',
      cpaImportFile: 'CPA Export File',
      cpaImportSelectFile: 'Select JSON file',
      cpaRemotePreviewSummaryTitle: 'Remote Preview Summary',
      cpaRemotePreviewSummary: 'Total {total}, importable {importable}, skipped non-normal {skipped_non_normal}, skipped unsupported {skipped_unsupported}',
      cpaRemoteStatusFilterNote: 'Remote import only includes accounts that are currently normal and supported.',
      cpaRemoteExistingAccounts: 'Existing accounts',
      cpaRemoteNewAccounts: 'New accounts',
      cpaRemoteNoImportableAccounts: 'No importable accounts found',
      cpaImportConcurrency: 'Concurrency',
      cpaUseDefaultGroupBind: 'Bind default groups automatically',
      cpaManualGroupsHint: 'Select groups to bind to the imported accounts.',
      cpaManualGroupsMultiPlatformHint: 'Multiple platforms are selected. Import by platform or use default group binding.',
      cpaImportExistingTitle: 'Existing account found',
      cpaImportExistingDesc: 'This source matches existing account "{name}". Importing will update it.',
      cpaImportResult: 'Import Result',
      cpaImportResultSummary: 'Created {created}, updated {updated}, failed {failed}',
      cpaImportErrors: 'Failure Details',
      cpaImportPreview: 'Preview',
      cpaImportPreviewing: 'Previewing...',
      cpaImportButton: 'Start Import',
      cpaImporting: 'Importing...',
      cpaRemoteMissingFields: 'Please enter CLIProxyAPI base URL and management key',
      cpaImportPreviewFailed: 'Preview failed',
      cpaImportCompletedWithErrors: 'Import completed with errors: created {created}, updated {updated}, failed {failed}',
      cpaImportSuccess: 'Import completed: created {created}, updated {updated}',
      cpaImportFailed: 'Import failed',
      fromModel: 'Source model',
      toModel: 'Target model',
      oauth: {
        openai: {
          accessTokenAuth: 'Manual AT',
          mobileRefreshTokenAuth: 'Manual Mobile RT'
        }
      }
    },
    channels: {
      noGroupsSelected: 'No groups selected for {platform}. Select at least one group or disable this platform.',
      emptyModelsInPricing: 'Some pricing entries under {platform} have no model. Add a model or remove the entry.'
    },
    groups: {
      failedToSave: 'Failed to save group settings'
    },
    ops: {
      result: 'Result',
      timeRange: {
        custom: 'Custom'
      },
      customTimeRange: {
        startTime: 'Start time',
        endTime: 'End time'
      },
      runtime: {
        metricThresholds: 'Metric Thresholds',
        metricThresholdsHint: 'Tune alert thresholds for service-level indicators and upstream health.',
        slaMinPercent: 'Minimum SLA (%)',
        slaMinPercentHint: 'Trigger alerts when successful-request SLA drops below this percentage.',
        ttftP99MaxMs: 'P99 TTFT Max (ms)',
        ttftP99MaxMsHint: 'Trigger alerts when p99 time-to-first-token exceeds this threshold.',
        requestErrorRateMaxPercent: 'Request Error Rate Max (%)',
        requestErrorRateMaxPercentHint: 'Trigger alerts when total request error rate exceeds this percentage.',
        upstreamErrorRateMaxPercent: 'Upstream Error Rate Max (%)',
        upstreamErrorRateMaxPercentHint: 'Trigger alerts when upstream provider error rate exceeds this percentage.'
      }
    },
    proxies: {
      subscriptionImport: 'Import Subscription',
      subscriptionImportTitle: 'Import Proxy Subscription',
      subscriptionImportHint: 'Enter a proxy subscription URL to parse supported http, https, socks5, socks5h, Shadowsocks, AnyTLS, Trojan, VLESS, and Hysteria2 nodes into Proxy Management.',
      subscriptionImportNote: 'The built-in dialer imports http, https, socks5, socks5h, Shadowsocks, AnyTLS, Trojan, VLESS TCP/TLS, and Hysteria2 directly. Unsupported transports such as VLESS Reality, WS/gRPC, vmess, and tuic are detected but not imported.',
      subscriptionImportUrl: 'Subscription URL',
      subscriptionImportUrlPlaceholder: 'https://example.com/subscription',
      subscriptionImportUrlRequired: 'Please enter a subscription URL',
      subscriptionImportClientType: 'Client type',
      subscriptionImportClientTypeHint: 'Some providers return different node formats based on the client User-Agent.',
      subscriptionImportParse: 'Parse Subscription',
      subscriptionImportParsing: 'Parsing...',
      subscriptionImportParsed: 'Parsed {count} proxies',
      subscriptionImportStatsClient: 'Client: {client}',
      subscriptionImportStatsFormat: 'Format: {format}',
      subscriptionImportStatsDecoded: 'Base64 decoded',
      subscriptionImportStatsDetected: 'Detected: {protocols}',
      subscriptionImportStatsSupported: 'Supported: {protocols}',
      subscriptionImportStatsUnsupported: 'Unsupported: {protocols}',
      subscriptionImportPartialFailed: '{count} proxies failed to import. Check server logs or retry after fixing the subscription nodes.',
      subscriptionImportUnsupportedOnly: 'This subscription only returned protocols that the built-in dialer cannot use directly. Try another client type or import a subscription containing supported protocols.',
      subscriptionImportSelected: 'Selected {count}/{total} proxies',
      subscriptionImportSelectAll: 'Select Current List',
      subscriptionImportClearSelection: 'Clear Current List',
      subscriptionImportSelectHint: 'Search and select the nodes you want to import. Existing proxies will be skipped automatically.',
      subscriptionImportSearchPlaceholder: 'Search name, host, or protocol...',
      subscriptionImportNoMatch: 'No matching proxies',
      subscriptionImportButton: 'Import {count} proxies',
      subscriptionImportFailed: 'Failed to parse subscription',
      subscriptionImportEmpty: 'Parse a subscription first',
      subscriptionImportSelectionEmpty: 'Please select proxies to import',
      protocols: {
        ss: 'Shadowsocks',
        anytls: 'AnyTLS',
        trojan: 'Trojan',
        vless: 'VLESS',
        hysteria2: 'Hysteria2'
      },
      subscriptionClientTypes: {
        auto: 'Auto',
        default: 'Default',
        clash: 'Clash',
        clashMeta: 'Clash Meta',
        mihomo: 'Mihomo',
        singBox: 'sing-box',
        surge: 'Surge',
        shadowrocket: 'Shadowrocket',
        stash: 'Stash',
        quantumultX: 'Quantumult X',
        loon: 'Loon',
        v2rayn: 'v2rayN'
      },
      ssMethod: 'Cipher Method',
      ssMethodPlaceholder: 'e.g. aes-256-gcm',
      ssPassword: 'Password',
      ssPasswordPlaceholder: 'Enter Shadowsocks password',
      ssMethodRequired: 'Please enter Shadowsocks cipher method',
      ssPasswordRequired: 'Please enter Shadowsocks password',
      protocolOptions: 'Protocol options',
      protocolOptionsPlaceholder: 'e.g. sni=example.com&insecure=1',
      protocolPasswordPlaceholder: 'Enter protocol password',
      protocolSecretRequired: 'Please enter the protocol password',
      vlessUUID: 'UUID',
      vlessUUIDPlaceholder: 'Enter VLESS UUID',
      vlessUUIDRequired: 'Please enter VLESS UUID',
      qualityRegion: 'Region',
      qualityCity: 'City'
    },
    users: {
      passwordCopied: 'Password copied'
    }
  },
  zh: {
    accounts: {
      cpaImport: '导入 CPA',
      cpaImportTitle: '从 CLIProxyAPI 导入',
      cpaImportModeRemote: '远程接口',
      cpaImportModeFile: 'JSON 文件',
      cpaImportRemoteHint: '连接 CLIProxyAPI 管理接口，预览可导入账号。',
      cpaImportHint: '上传 CLIProxyAPI 导出的 JSON 文件，预览账号后再导入。',
      cpaImportRemoteWarning: '仅导入状态正常且受支持的账号；已存在账号会按来源标识识别并更新。',
      cpaImportWarning: '导入数据可能包含令牌和代理信息，请在预览后确认再继续。',
      cpaRemoteBaseUrl: 'CLIProxyAPI 地址',
      cpaRemoteBaseUrlPlaceholder: 'https://cliproxy.example.com',
      cpaRemoteManagementKey: '管理 Key',
      cpaImportFile: 'CPA 导出文件',
      cpaImportSelectFile: '请选择 JSON 文件',
      cpaRemotePreviewSummaryTitle: '远程预览摘要',
      cpaRemotePreviewSummary: '总计 {total}，可导入 {importable}，跳过非正常 {skipped_non_normal}，跳过不支持 {skipped_unsupported}',
      cpaRemoteStatusFilterNote: '远程导入仅包含当前状态正常且受支持的账号。',
      cpaRemoteExistingAccounts: '已存在账号',
      cpaRemoteNewAccounts: '新账号',
      cpaRemoteNoImportableAccounts: '没有可导入账号',
      cpaImportConcurrency: '并发数',
      cpaUseDefaultGroupBind: '自动绑定默认分组',
      cpaManualGroupsHint: '为导入账号选择要绑定的分组。',
      cpaManualGroupsMultiPlatformHint: '当前选择包含多个平台，请按平台导入或使用默认分组绑定。',
      cpaImportExistingTitle: '发现已存在账号',
      cpaImportExistingDesc: '该来源匹配已有账号“{name}”，导入后会更新该账号。',
      cpaImportResult: '导入结果',
      cpaImportResultSummary: '创建 {created}，更新 {updated}，失败 {failed}',
      cpaImportErrors: '失败详情',
      cpaImportPreview: '预览',
      cpaImportPreviewing: '预览中...',
      cpaImportButton: '开始导入',
      cpaImporting: '导入中...',
      cpaRemoteMissingFields: '请输入 CLIProxyAPI 地址和管理 Key',
      cpaImportPreviewFailed: '预览失败',
      cpaImportCompletedWithErrors: '导入完成但有错误：创建 {created}，更新 {updated}，失败 {failed}',
      cpaImportSuccess: '导入完成：创建 {created}，更新 {updated}',
      cpaImportFailed: '导入失败',
      fromModel: '源模型',
      toModel: '目标模型',
      oauth: {
        openai: {
          accessTokenAuth: '手动输入 AT',
          mobileRefreshTokenAuth: '手动输入 Mobile RT'
        }
      }
    },
    channels: {
      noGroupsSelected: '{platform} 平台未选择分组，请至少选择一个分组或禁用该平台',
      emptyModelsInPricing: '{platform} 平台下有定价条目未添加模型，请添加模型或删除该条目'
    },
    groups: {
      failedToSave: '保存分组设置失败'
    },
    ops: {
      result: '结果',
      timeRange: {
        custom: '自定义'
      },
      customTimeRange: {
        startTime: '开始时间',
        endTime: '结束时间'
      },
      runtime: {
        metricThresholds: '指标阈值',
        metricThresholdsHint: '调整服务级指标和上游健康度的告警阈值。',
        slaMinPercent: '最低 SLA（%）',
        slaMinPercentHint: '成功请求 SLA 低于该百分比时触发告警。',
        ttftP99MaxMs: 'P99 首 Token 最大耗时（ms）',
        ttftP99MaxMsHint: 'P99 首 Token 耗时超过该阈值时触发告警。',
        requestErrorRateMaxPercent: '请求错误率上限（%）',
        requestErrorRateMaxPercentHint: '总请求错误率超过该百分比时触发告警。',
        upstreamErrorRateMaxPercent: '上游错误率上限（%）',
        upstreamErrorRateMaxPercentHint: '上游供应商错误率超过该百分比时触发告警。'
      }
    },
    proxies: {
      subscriptionImport: '导入订阅',
      subscriptionImportTitle: '导入代理订阅',
      subscriptionImportHint: '输入代理订阅地址，系统会解析其中支持的 http、https、socks5、socks5h、Shadowsocks、AnyTLS、Trojan、VLESS 和 Hysteria2 节点并批量导入到 IP 管理。',
      subscriptionImportNote: '内置拨号器会直接导入 http、https、socks5、socks5h、Shadowsocks、AnyTLS、Trojan、VLESS TCP/TLS 和 Hysteria2；VLESS Reality、WS/gRPC、vmess、tuic 等暂不支持的传输只识别统计，暂不导入。',
      subscriptionImportUrl: '订阅地址',
      subscriptionImportUrlPlaceholder: 'https://example.com/subscription',
      subscriptionImportUrlRequired: '请输入订阅地址',
      subscriptionImportClientType: '客户端类型',
      subscriptionImportClientTypeHint: '部分订阅服务会根据客户端 User-Agent 返回不同协议格式。',
      subscriptionImportParse: '解析订阅',
      subscriptionImportParsing: '解析中...',
      subscriptionImportParsed: '已解析 {count} 个代理',
      subscriptionImportStatsClient: '客户端：{client}',
      subscriptionImportStatsFormat: '格式：{format}',
      subscriptionImportStatsDecoded: '已 Base64 解码',
      subscriptionImportStatsDetected: '识别到：{protocols}',
      subscriptionImportStatsSupported: '可导入：{protocols}',
      subscriptionImportStatsUnsupported: '暂不支持：{protocols}',
      subscriptionImportPartialFailed: '{count} 个代理导入失败，请检查服务端日志或修正订阅节点后重试。',
      subscriptionImportUnsupportedOnly: '这个订阅只返回了内置拨号器不能直接使用的协议。请切换客户端类型，或导入包含受支持协议的订阅。',
      subscriptionImportSelected: '已选择 {count}/{total} 个代理',
      subscriptionImportSelectAll: '选择当前列表',
      subscriptionImportClearSelection: '清空当前列表',
      subscriptionImportSelectHint: '可先搜索并勾选需要导入的节点，已存在的代理会在导入时自动跳过。',
      subscriptionImportSearchPlaceholder: '搜索名称、主机或协议...',
      subscriptionImportNoMatch: '没有匹配的代理',
      subscriptionImportButton: '导入 {count} 个代理',
      subscriptionImportFailed: '解析订阅失败',
      subscriptionImportEmpty: '请先解析订阅',
      subscriptionImportSelectionEmpty: '请选择要导入的代理',
      protocols: {
        ss: 'Shadowsocks',
        anytls: 'AnyTLS',
        trojan: 'Trojan',
        vless: 'VLESS',
        hysteria2: 'Hysteria2'
      },
      subscriptionClientTypes: {
        auto: '自动',
        default: '默认',
        clash: 'Clash',
        clashMeta: 'Clash Meta',
        mihomo: 'Mihomo',
        singBox: 'sing-box',
        surge: 'Surge',
        shadowrocket: 'Shadowrocket',
        stash: 'Stash',
        quantumultX: 'Quantumult X',
        loon: 'Loon',
        v2rayn: 'v2rayN'
      },
      ssMethod: '加密方法',
      ssMethodPlaceholder: '例如 aes-256-gcm',
      ssPassword: '密码',
      ssPasswordPlaceholder: '请输入 Shadowsocks 密码',
      ssMethodRequired: '请输入 Shadowsocks 加密方法',
      ssPasswordRequired: '请输入 Shadowsocks 密码',
      protocolOptions: '协议参数',
      protocolOptionsPlaceholder: '例如 sni=example.com&insecure=1',
      protocolPasswordPlaceholder: '请输入协议密码',
      protocolSecretRequired: '请输入协议密码',
      vlessUUID: 'UUID',
      vlessUUIDPlaceholder: '请输入 VLESS UUID',
      vlessUUIDRequired: '请输入 VLESS UUID',
      qualityRegion: '地区',
      qualityCity: '城市'
    },
    users: {
      passwordCopied: '密码已复制'
    }
  }
} as const

const messages: Record<'en' | 'zh', CustomLocaleMessages> = {
  en: {
    common: common.en,
    admin: admin.en
  },
  zh: {
    common: common.zh,
    admin: admin.zh
  }
}

export default messages
