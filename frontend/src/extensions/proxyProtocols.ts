import type { ProxyProtocol } from '@/types'

type TranslateFn = (key: string, ...args: unknown[]) => string

export const baseProxyProtocols = ['http', 'https', 'socks5', 'socks5h', 'ss'] as const
export const extendedProxyProtocols = ['anytls', 'trojan', 'vless', 'hysteria2'] as const
export const proxyProtocols = [...baseProxyProtocols, ...extendedProxyProtocols] as const

export type { ProxyProtocol }

const secretRequiredProtocols = new Set<ProxyProtocol>([
  'ss',
  'anytls',
  'trojan',
  'vless',
  'hysteria2'
])

export const getProxyProtocolOptions = (t: TranslateFn, includeAll = false) => {
  const options = proxyProtocols.map((protocol) => ({
    value: protocol,
    label: t(`admin.proxies.protocols.${protocol}`)
  }))
  if (!includeAll) return options
  return [{ value: '', label: t('admin.proxies.allProtocols') }, ...options]
}

export const isProxyPrimaryCredentialRequired = (protocol: ProxyProtocol) => protocol === 'ss'

export const isProxySecretRequired = (protocol: ProxyProtocol) =>
  secretRequiredProtocols.has(protocol)

export const getProxyPrimaryCredentialLabel = (protocol: ProxyProtocol, t: TranslateFn) => {
  if (protocol === 'ss') return t('admin.proxies.ssMethod')
  if (isProxySecretRequired(protocol)) return t('admin.proxies.protocolOptions')
  return t('admin.proxies.username')
}

export const getProxyPrimaryCredentialPlaceholder = (protocol: ProxyProtocol, t: TranslateFn) => {
  if (protocol === 'ss') return t('admin.proxies.ssMethodPlaceholder')
  if (isProxySecretRequired(protocol)) return t('admin.proxies.protocolOptionsPlaceholder')
  return t('admin.proxies.optionalAuth')
}

export const getProxySecretLabel = (protocol: ProxyProtocol, t: TranslateFn) => {
  if (protocol === 'vless') return t('admin.proxies.vlessUUID')
  if (protocol === 'ss') return t('admin.proxies.ssPassword')
  return t('admin.proxies.password')
}

export const getProxySecretPlaceholder = (protocol: ProxyProtocol, t: TranslateFn) => {
  if (protocol === 'vless') return t('admin.proxies.vlessUUIDPlaceholder')
  if (protocol === 'ss') return t('admin.proxies.ssPasswordPlaceholder')
  if (isProxySecretRequired(protocol)) return t('admin.proxies.protocolPasswordPlaceholder')
  return t('admin.proxies.optionalAuth')
}

export const validateProxyProtocolCredentials = (
  protocol: ProxyProtocol,
  username: string,
  password: string,
  t: TranslateFn
) => {
  if (protocol === 'ss' && !username.trim()) {
    return t('admin.proxies.ssMethodRequired')
  }
  if (isProxySecretRequired(protocol) && !password.trim()) {
    return protocol === 'vless'
      ? t('admin.proxies.vlessUUIDRequired')
      : t('admin.proxies.protocolSecretRequired')
  }
  return ''
}
