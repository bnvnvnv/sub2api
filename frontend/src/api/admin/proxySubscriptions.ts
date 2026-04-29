/**
 * Admin proxy subscription parsing API endpoints.
 */

import { apiClient } from '../client'
import type { ProxyProtocol } from '@/extensions/proxyProtocols'

export type SubscriptionClientType =
  | 'auto'
  | 'default'
  | 'clash'
  | 'clash-meta'
  | 'mihomo'
  | 'sing-box'
  | 'surge'
  | 'shadowrocket'
  | 'stash'
  | 'quantumult-x'
  | 'loon'
  | 'v2rayn'

export interface ParsedSubscriptionProxy {
  name: string
  protocol: ProxyProtocol
  host: string
  port: number
  username: string
  password: string
  import_mode?: 'direct'
  node_protocol?: string
}

export interface SubscriptionParseStats {
  client_type: string
  user_agent: string
  format: string
  decoded: boolean
  detected_protocol_counts: Record<string, number>
  supported_protocol_counts: Record<string, number>
  unsupported_protocol_counts: Record<string, number>
  supported_count: number
  importable_count: number
  unsupported_count: number
  warnings: string[]
}

export interface ParseSubscriptionResult {
  proxies: ParsedSubscriptionProxy[]
  stats: SubscriptionParseStats
}

export async function parseSubscription(
  url: string,
  clientType: SubscriptionClientType = 'auto'
): Promise<ParseSubscriptionResult> {
  const { data } = await apiClient.post<ParseSubscriptionResult>('/admin/proxies/subscription/parse', {
    url,
    client_type: clientType
  })
  return data
}

export interface ImportSubscriptionResult {
  created: number
  skipped: number
  direct_created: number
  warnings: string[]
  failed: Array<{
    name: string
    message: string
  }>
}

export async function importSubscription(
  proxies: ParsedSubscriptionProxy[]
): Promise<ImportSubscriptionResult> {
  const { data } = await apiClient.post<ImportSubscriptionResult>('/admin/proxies/subscription/import', {
    proxies
  })
  return data
}

export const proxySubscriptionsAPI = {
  parseSubscription,
  importSubscription
}

export default proxySubscriptionsAPI
