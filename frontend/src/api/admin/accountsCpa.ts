/**
 * Admin CPA import API endpoints.
 */

import { apiClient } from '../client'

export interface CPAPreviewAccount {
  cpa_source_key: string
  provider: string
  name: string
  email?: string
  platform: string
  type: string
  file_name: string
  warnings?: string[]
}

export interface CPAExistingAccount {
  id: number
  name: string
  platform: string
  type: string
  status: string
}

export interface PreviewFromCPAResult {
  account: CPAPreviewAccount
  existing_account?: CPAExistingAccount
}

export interface PreviewRemoteFromCPAResult {
  items: PreviewFromCPAResult[]
  total: number
  importable: number
  skipped_non_normal: number
  skipped_unsupported: number
}

export interface SyncFromCPAResult {
  created: number
  updated: number
  failed: number
  items: Array<{
    cpa_source_key: string
    provider: string
    name: string
    action: string
    error?: string
    account_id?: number
  }>
}

export async function previewFromCpa(params: {
  file_name: string
  raw_json: string
}): Promise<PreviewFromCPAResult> {
  const { data } = await apiClient.post<PreviewFromCPAResult>(
    '/admin/accounts/import/cpa/preview',
    params
  )
  return data
}

export async function previewRemoteFromCpa(params: {
  base_url: string
  management_key: string
}): Promise<PreviewRemoteFromCPAResult> {
  const { data } = await apiClient.post<PreviewRemoteFromCPAResult>(
    '/admin/accounts/import/cpa/remote/preview',
    params
  )
  return data
}

export async function importFromCpa(params: {
  file_name: string
  raw_json: string
  proxy_id?: number | null
  concurrency?: number
  use_default_group_bind?: boolean
  group_ids?: number[]
}): Promise<SyncFromCPAResult> {
  const { data } = await apiClient.post<SyncFromCPAResult>('/admin/accounts/import/cpa', params)
  return data
}

export async function importRemoteFromCpa(params: {
  base_url: string
  management_key: string
  selected_source_keys?: string[]
  proxy_id?: number | null
  concurrency?: number
  use_default_group_bind?: boolean
  group_ids?: number[]
}): Promise<SyncFromCPAResult> {
  const { data } = await apiClient.post<SyncFromCPAResult>(
    '/admin/accounts/import/cpa/remote',
    params
  )
  return data
}

export const accountsCpaAPI = {
  previewFromCpa,
  previewRemoteFromCpa,
  importFromCpa,
  importRemoteFromCpa
}

export default accountsCpaAPI
