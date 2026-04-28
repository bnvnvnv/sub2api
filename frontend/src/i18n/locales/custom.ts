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

const openAIWeb = {
  en: {
    title: 'OpenAI Web',
    description: 'Use eligible subscribed OpenAI accounts in a web-style chat page',
    noEligibleTitle: 'No OpenAI Web access available',
    noEligibleDesc: 'Your current groups do not have an eligible OpenAI account for web chat. Contact the administrator if you believe this is incorrect.',
    newChat: 'New Chat',
    createUnavailableDesc: 'No eligible model or account is currently available for starting a new chat.',
    historyTitle: 'History',
    historyEmptyTitle: 'No conversations yet',
    historyEmptyDesc: 'Start a new chat and your local conversation history will appear here.',
    untitledThread: 'Untitled chat',
    newChatTitle: 'New chat',
    chatModelLabel: 'Chat model',
    reasoningEffortLabel: 'Reasoning effort',
    reasoningEffortAuto: 'Auto reasoning',
    reasoningEffort: {
      low: 'Low',
      medium: 'Medium',
      high: 'High',
      xhigh: 'Extra high'
    },
    archive: 'Archive',
    archiveSuccess: 'Conversation archived',
    archiveFailed: 'Failed to archive conversation',
    emptyChatTitle: 'No messages yet',
    emptyChatDesc: 'This conversation is ready to use {model}. Send a message to continue.',
    newChatIntro: 'Ask anything, or request an image in natural language.',
    modelAutoResolved: 'Auto model',
    userLabel: 'You',
    assistantLabel: 'Assistant',
    pendingReply: 'Generating reply...',
    failedReply: 'Reply failed',
    previewImage: 'Preview image',
    downloadImage: 'Download image',
    addImages: 'Add images',
    clearImages: 'Clear images',
    removeImage: 'Remove image',
    imageAttachment: 'Image attachment',
    imageCountTag: '{count} images',
    assistantImageAlt: 'Generated image',
    imageConversationTitle: 'Image conversation',
    imageDimensionsTag: '{width}x{height}',
    composerPlaceholder: 'Message OpenAI Web...',
    failedToLoad: 'Failed to load OpenAI Web data',
    createFailed: 'Failed to create conversation',
    noMatchingGroupForModel: 'No eligible account group is available for {model}',
    emptyAssistantReply: 'The assistant returned an empty reply.',
    sendFailed: 'Failed to send message',
    attachmentLimitReached: 'You can attach at most {count} images',
    attachmentReadFailed: 'Failed to read image attachment',
    attachmentTooLarge: 'Image is too large. Maximum size is {size}',
    attachmentTypeInvalid: 'Only image files are supported',
    localCacheDisabledWarning: 'Local conversation cache is unavailable in this browser.',
    localCacheImageWarning: 'Image content is stored only in the local browser cache.'
  },
  zh: {
    title: 'OpenAI Web',
    description: '通过页面使用订阅中的 OpenAI 账号进行网页式聊天',
    noEligibleTitle: '暂无可用的 OpenAI Web 权限',
    noEligibleDesc: '当前分组没有可用于网页聊天的 OpenAI 账号。如果你认为这是配置问题，请联系管理员。',
    newChat: '新聊天',
    createUnavailableDesc: '当前没有可用于新聊天的模型或账号。',
    historyTitle: '历史记录',
    historyEmptyTitle: '暂无会话',
    historyEmptyDesc: '开始新聊天后，本地会话历史会显示在这里。',
    untitledThread: '未命名会话',
    newChatTitle: '新聊天',
    chatModelLabel: '聊天模型',
    reasoningEffortLabel: '思考强度',
    reasoningEffortAuto: '自动思考',
    reasoningEffort: {
      low: '低',
      medium: '中',
      high: '高',
      xhigh: '极高'
    },
    archive: '归档',
    archiveSuccess: '会话已归档',
    archiveFailed: '会话归档失败',
    emptyChatTitle: '暂无消息',
    emptyChatDesc: '该会话已准备使用 {model}。发送消息即可继续。',
    newChatIntro: '可以直接提问，也可以用自然语言要求生成图片。',
    modelAutoResolved: '自动模型',
    userLabel: '你',
    assistantLabel: '助手',
    pendingReply: '正在生成回复...',
    failedReply: '回复失败',
    previewImage: '预览图片',
    downloadImage: '下载图片',
    addImages: '添加图片',
    clearImages: '清空图片',
    removeImage: '移除图片',
    imageAttachment: '图片附件',
    imageCountTag: '{count} 张图片',
    assistantImageAlt: '生成图片',
    imageConversationTitle: '图片会话',
    imageDimensionsTag: '{width}x{height}',
    composerPlaceholder: '发送消息给 OpenAI Web...',
    failedToLoad: '加载 OpenAI Web 数据失败',
    createFailed: '创建会话失败',
    noMatchingGroupForModel: '没有可用于 {model} 的账号分组',
    emptyAssistantReply: '助手返回了空回复。',
    sendFailed: '发送消息失败',
    attachmentLimitReached: '最多只能添加 {count} 张图片',
    attachmentReadFailed: '读取图片失败',
    attachmentTooLarge: '图片过大，最大支持 {size}',
    attachmentTypeInvalid: '仅支持图片文件',
    localCacheDisabledWarning: '当前浏览器无法使用本地会话缓存。',
    localCacheImageWarning: '图片内容仅保存在当前浏览器本地缓存中。'
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
      subscriptionImportHint: 'Enter a proxy subscription URL to parse supported Shadowsocks nodes and import them into Proxy Management.',
      subscriptionImportNote: 'Only Shadowsocks nodes without plugins are imported for now. The subscription URL must be a public HTTP/HTTPS URL; private network URLs are rejected.',
      subscriptionImportUrl: 'Subscription URL',
      subscriptionImportUrlPlaceholder: 'https://example.com/subscription',
      subscriptionImportUrlRequired: 'Please enter a subscription URL',
      subscriptionImportParse: 'Parse Subscription',
      subscriptionImportParsing: 'Parsing...',
      subscriptionImportParsed: 'Parsed {count} proxies',
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
        ss: 'Shadowsocks'
      },
      ssMethod: 'Cipher Method',
      ssMethodPlaceholder: 'e.g. aes-256-gcm',
      ssPassword: 'Password',
      ssPasswordPlaceholder: 'Enter Shadowsocks password',
      ssMethodRequired: 'Please enter Shadowsocks cipher method',
      ssPasswordRequired: 'Please enter Shadowsocks password',
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
      subscriptionImportHint: '输入代理订阅地址，系统会解析其中支持的 Shadowsocks 节点并批量导入到 IP 管理。',
      subscriptionImportNote: '目前仅导入未启用插件的 Shadowsocks 节点；订阅地址必须是公网 HTTP/HTTPS 地址，内网地址会被拒绝。',
      subscriptionImportUrl: '订阅地址',
      subscriptionImportUrlPlaceholder: 'https://example.com/subscription',
      subscriptionImportUrlRequired: '请输入订阅地址',
      subscriptionImportParse: '解析订阅',
      subscriptionImportParsing: '解析中...',
      subscriptionImportParsed: '已解析 {count} 个代理',
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
        ss: 'Shadowsocks'
      },
      ssMethod: '加密方法',
      ssMethodPlaceholder: '例如 aes-256-gcm',
      ssPassword: '密码',
      ssPasswordPlaceholder: '请输入 Shadowsocks 密码',
      ssMethodRequired: '请输入 Shadowsocks 加密方法',
      ssPasswordRequired: '请输入 Shadowsocks 密码',
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
    nav: {
      openAIWeb: 'OpenAI Web'
    },
    openAIWeb: openAIWeb.en,
    admin: admin.en
  },
  zh: {
    common: common.zh,
    nav: {
      openAIWeb: 'OpenAI Web'
    },
    openAIWeb: openAIWeb.zh,
    admin: admin.zh
  }
}

export default messages
