import { ref, type Ref } from 'vue'
import userChannelsAPI, { type UserAvailableChannel } from '@/api/channels'
import type { OpenAIWebEntitlement } from '@/types'

export type OpenAIWebComposerReasoningEffort = '' | 'low' | 'medium' | 'high' | 'xhigh'

const OPENAI_WEB_WEB_ONLY_PRO_MODELS = ['gpt-5.4-pro', 'gpt-5.5-pro']

export function normalizeRequestedModel(model: string | null | undefined): string {
  return (model ?? '').trim()
}

export function normalizeOpenAIWebComposerReasoningEffort(
  raw: string | null | undefined
): OpenAIWebComposerReasoningEffort {
  const normalized = (raw ?? '').trim().toLowerCase().replace(/[-_\s]/g, '')
  switch (normalized) {
    case 'low':
      return 'low'
    case 'medium':
      return 'medium'
    case 'high':
      return 'high'
    case 'xhigh':
    case 'extrahigh':
      return 'xhigh'
    default:
      return ''
  }
}

export function modelNeedsOpenAIWebPro(model: string | null | undefined): boolean {
  const normalized = normalizeRequestedModel(model).toLowerCase()
  return normalized.includes('gpt-5.4-pro')
    || normalized.includes('gpt-5.5-pro')
    || normalized.includes('5-4-pro')
    || normalized.includes('5-5-pro')
}

export function useOpenAIWebModelOptions(
  entitlements: Ref<OpenAIWebEntitlement[]>,
  readPreferredGroupID: () => number | null
) {
  const availableChannels = ref<UserAvailableChannel[]>([])

  async function loadAvailableChannels(): Promise<UserAvailableChannel[]> {
    try {
      return await userChannelsAPI.getAvailable()
    } catch {
      return []
    }
  }

  function openAIWebModelOptionsForEntitlement(
    entitlement?: OpenAIWebEntitlement | null,
    fallbackModel?: string | null
  ): string[] {
    const models = new Set<string>()

    addOpenAIWebChannelModels(models, entitlement?.group_id ? new Set([entitlement.group_id]) : undefined)
    addModelOption(models, entitlement?.default_model)
    addOpenAIWebOnlyProModels(models, entitlement)
    addModelOption(models, fallbackModel)

    return Array.from(models)
  }

  function openAIWebModelOptionsForDraft(
    items: OpenAIWebEntitlement[],
    fallbackModel?: string | null
  ): string[] {
    const models = new Set<string>()
    const groupIDs = new Set(
      items
        .map((entitlement) => entitlement.group_id)
        .filter((groupID) => Number.isFinite(groupID) && groupID > 0)
    )

    addOpenAIWebChannelModels(models, groupIDs)
    items.forEach((entitlement) => {
      addModelOption(models, entitlement.default_model)
      addOpenAIWebOnlyProModels(models, entitlement)
    })

    const fallback = normalizeRequestedModel(fallbackModel)
    if (fallback && (models.size === 0 || items.some((entitlement) => entitlementSupportsModel(entitlement, fallback)))) {
      addModelOption(models, fallback)
    }

    return Array.from(models)
  }

  function entitlementSupportsModel(
    entitlement: OpenAIWebEntitlement | null | undefined,
    model: string | null | undefined
  ): boolean {
    if (!entitlement) {
      return false
    }
    const requestedModel = normalizeRequestedModel(model)
    if (!requestedModel) {
      return true
    }
    const channelModels = openAIWebChannelModelSetForGroup(entitlement.group_id)
    if (channelModels.size > 0) {
      return channelModels.has(requestedModel) || entitlementSupportsOpenAIWebOnlyProModel(entitlement, requestedModel)
    }
    if (normalizeRequestedModel(entitlement.default_model) === requestedModel) {
      return true
    }
    if (hasOpenAIWebChannelModelsForEntitlements()) {
      return entitlementSupportsOpenAIWebOnlyProModel(entitlement, requestedModel)
    }
    if (!modelNeedsOpenAIWebPro(requestedModel)) {
      return true
    }
    return entitlement.has_pro_accounts
  }

  function resolveDraftEntitlement(model?: string | null): OpenAIWebEntitlement | null {
    if (entitlements.value.length === 0) {
      return null
    }

    const requestedModel = normalizeRequestedModel(model)
    const preferredGroupID = readPreferredGroupID()
    const preferred = preferredGroupID
      ? entitlements.value.find((item) => item.group_id === preferredGroupID) ?? null
      : null

    if (preferred && entitlementSupportsModel(preferred, requestedModel)) {
      return preferred
    }

    if (requestedModel) {
      const matching = entitlements.value.find((item) => entitlementSupportsModel(item, requestedModel))
      if (matching) {
        return matching
      }
    }

    return preferred ?? entitlements.value[0] ?? null
  }

  function isOpenAIPlatform(platform: string | null | undefined): boolean {
    return (platform ?? '').trim().toLowerCase() === 'openai'
  }

  function addModelOption(models: Set<string>, model: string | null | undefined) {
    const normalized = normalizeRequestedModel(model)
    if (!normalized) {
      return
    }
    models.add(normalized)
  }

  function addOpenAIWebChannelModels(models: Set<string>, allowedGroupIDs?: Set<number>) {
    if (!allowedGroupIDs || allowedGroupIDs.size === 0) {
      return
    }

    for (const channel of availableChannels.value) {
      for (const section of channel.platforms ?? []) {
        if (!isOpenAIPlatform(section.platform)) {
          continue
        }
        if (!(section.groups ?? []).some((group) => allowedGroupIDs.has(group.id))) {
          continue
        }

        for (const model of section.supported_models ?? []) {
          if (!isOpenAIPlatform(model.platform)) {
            continue
          }
          addModelOption(models, model.name)
        }
      }
    }
  }

  function addOpenAIWebOnlyProModels(models: Set<string>, entitlement?: OpenAIWebEntitlement | null) {
    if (!entitlement?.has_pro_accounts) {
      return
    }
    OPENAI_WEB_WEB_ONLY_PRO_MODELS.forEach((model) => addModelOption(models, model))
  }

  function openAIWebChannelModelSetForGroup(groupID: number | null | undefined): Set<string> {
    const models = new Set<string>()
    if (typeof groupID !== 'number' || !Number.isFinite(groupID) || groupID <= 0) {
      return models
    }
    addOpenAIWebChannelModels(models, new Set([groupID]))
    return models
  }

  function hasOpenAIWebChannelModelsForEntitlements(): boolean {
    const groupIDs = new Set(
      entitlements.value
        .map((entitlement) => entitlement.group_id)
        .filter((groupID) => Number.isFinite(groupID) && groupID > 0)
    )
    const models = new Set<string>()
    addOpenAIWebChannelModels(models, groupIDs)
    return models.size > 0
  }

  function isOpenAIWebOnlyProModel(model: string | null | undefined): boolean {
    const normalized = normalizeRequestedModel(model).toLowerCase()
    return OPENAI_WEB_WEB_ONLY_PRO_MODELS.includes(normalized)
  }

  function entitlementSupportsOpenAIWebOnlyProModel(
    entitlement: OpenAIWebEntitlement,
    model: string | null | undefined
  ): boolean {
    return entitlement.has_pro_accounts && isOpenAIWebOnlyProModel(model)
  }

  return {
    availableChannels,
    loadAvailableChannels,
    openAIWebModelOptionsForEntitlement,
    openAIWebModelOptionsForDraft,
    entitlementSupportsModel,
    resolveDraftEntitlement,
    normalizeComposerReasoningEffort: normalizeOpenAIWebComposerReasoningEffort,
    normalizeRequestedModel,
  }
}
