/**
 * Admin proxy subscription parsing API endpoints.
 */

import { apiClient } from '../client'

export interface ParsedSubscriptionProxy {
  name: string
  protocol: 'ss'
  host: string
  port: number
  username: string
  password: string
}

export async function parseSubscription(url: string): Promise<{
  proxies: ParsedSubscriptionProxy[]
}> {
  const { data } = await apiClient.post<{
    proxies: ParsedSubscriptionProxy[]
  }>('/admin/proxies/subscription/parse', { url })
  return data
}

export const proxySubscriptionsAPI = {
  parseSubscription
}

export default proxySubscriptionsAPI
