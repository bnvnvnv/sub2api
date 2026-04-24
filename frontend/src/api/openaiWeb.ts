import { apiClient } from './client'
import type {
  OpenAIWebEntitlement,
  OpenAIWebThread,
  OpenAIWebThreadCreateRequest,
  OpenAIWebThreadMessageRequest,
  OpenAIWebThreadMessageResponse,
} from '@/types'

export async function getOpenAIWebEntitlements(): Promise<OpenAIWebEntitlement[]> {
  const response = await apiClient.get<OpenAIWebEntitlement[]>('/openai-web/entitlements')
  return response.data
}

export async function getOpenAIWebThreads(): Promise<OpenAIWebThread[]> {
  const response = await apiClient.get<OpenAIWebThread[]>('/openai-web/threads')
  return response.data
}

export async function getOpenAIWebThread(localThreadID: string): Promise<OpenAIWebThread> {
  const response = await apiClient.get<OpenAIWebThread>(`/openai-web/threads/${localThreadID}`)
  return response.data
}

export async function createOpenAIWebThread(
  payload: OpenAIWebThreadCreateRequest
): Promise<OpenAIWebThread> {
  const response = await apiClient.post<OpenAIWebThread>('/openai-web/threads', payload)
  return response.data
}

export async function archiveOpenAIWebThread(localThreadID: string): Promise<void> {
  await apiClient.post(`/openai-web/threads/${localThreadID}/archive`)
}

export async function sendOpenAIWebThreadMessage(
  localThreadID: string,
  payload: OpenAIWebThreadMessageRequest
): Promise<OpenAIWebThreadMessageResponse> {
  const response = await apiClient.post<OpenAIWebThreadMessageResponse>(`/openai-web/threads/${localThreadID}/messages`, payload)
  return response.data
}

export default {
  getOpenAIWebEntitlements,
  getOpenAIWebThreads,
  getOpenAIWebThread,
  createOpenAIWebThread,
  archiveOpenAIWebThread,
  sendOpenAIWebThreadMessage,
}
