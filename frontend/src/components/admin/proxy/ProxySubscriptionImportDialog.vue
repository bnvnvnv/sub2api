<template>
  <BaseDialog
    :show="show"
    :title="t('admin.proxies.subscriptionImportTitle')"
    width="normal"
    @close="closeDialog"
  >
    <div class="space-y-4">
      <div class="text-sm text-gray-600 dark:text-dark-300">
        {{ t('admin.proxies.subscriptionImportHint') }}
      </div>
      <div
        class="rounded-lg border border-amber-200 bg-amber-50 p-3 text-xs text-amber-600 dark:border-amber-800 dark:bg-amber-900/20 dark:text-amber-400"
      >
        {{ t('admin.proxies.subscriptionImportNote') }}
      </div>
      <div>
        <label class="input-label">{{ t('admin.proxies.subscriptionImportUrl') }}</label>
        <input
          v-model="subscriptionImportUrl"
          type="url"
          class="input"
          :placeholder="t('admin.proxies.subscriptionImportUrlPlaceholder')"
          @input="resetSubscriptionImportResult"
        />
      </div>
      <div>
        <label class="input-label">{{ t('admin.proxies.subscriptionImportClientType') }}</label>
        <select
          v-model="subscriptionClientType"
          class="input"
          :disabled="subscriptionParsing || subscriptionImporting"
          @change="resetSubscriptionImportResult"
        >
          <option
            v-for="option in subscriptionClientTypeOptions"
            :key="option.value"
            :value="option.value"
          >
            {{ t(option.labelKey) }}
          </option>
        </select>
        <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
          {{ t('admin.proxies.subscriptionImportClientTypeHint') }}
        </p>
      </div>
      <div
        v-if="subscriptionParseStats"
        class="space-y-2 rounded-xl border border-gray-200 bg-gray-50 p-3 text-xs text-gray-600 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-300"
      >
        <div class="flex flex-wrap gap-x-4 gap-y-1">
          <span>
            {{ t('admin.proxies.subscriptionImportStatsClient', { client: subscriptionParseStats.client_type }) }}
          </span>
          <span>
            {{ t('admin.proxies.subscriptionImportStatsFormat', { format: subscriptionParseStats.format || '-' }) }}
          </span>
          <span v-if="subscriptionParseStats.decoded">
            {{ t('admin.proxies.subscriptionImportStatsDecoded') }}
          </span>
        </div>
        <div class="flex flex-wrap gap-x-4 gap-y-1">
          <span>
            {{ t('admin.proxies.subscriptionImportStatsDetected', { protocols: formatProtocolCounts(subscriptionParseStats.detected_protocol_counts) }) }}
          </span>
          <span>
            {{ t('admin.proxies.subscriptionImportStatsSupported', { protocols: formatProtocolCounts(subscriptionParseStats.supported_protocol_counts) }) }}
          </span>
          <span>
            {{ t('admin.proxies.subscriptionImportStatsUnsupported', { protocols: formatProtocolCounts(subscriptionParseStats.unsupported_protocol_counts) }) }}
          </span>
        </div>
        <div
          v-if="subscriptionParseStats.importable_count === 0 && subscriptionParseStats.unsupported_count > 0"
          class="rounded-lg border border-red-200 bg-red-50 p-2 text-red-700 dark:border-red-800 dark:bg-red-900/20 dark:text-red-300"
        >
          {{ t('admin.proxies.subscriptionImportUnsupportedOnly') }}
        </div>
      </div>
      <div
        v-if="subscriptionParsedProxies.length > 0"
        class="space-y-3 rounded-xl border border-gray-200 p-4 dark:border-dark-700"
      >
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div class="space-y-1">
            <div class="text-sm font-medium text-gray-900 dark:text-white">
              {{ t('admin.proxies.subscriptionImportParsed', { count: subscriptionParsedProxies.length }) }}
            </div>
            <div class="text-xs text-gray-500 dark:text-dark-400">
              {{ t('admin.proxies.subscriptionImportSelected', { count: subscriptionSelectedCount, total: subscriptionParsedProxies.length }) }}
            </div>
          </div>
          <div class="flex flex-wrap gap-2">
            <button
              type="button"
              class="rounded-lg border border-gray-200 px-3 py-1.5 text-xs font-medium text-gray-600 transition-colors hover:bg-gray-50 dark:border-dark-600 dark:text-dark-300 dark:hover:bg-dark-800"
              :disabled="subscriptionImporting"
              @click="selectAllSubscriptionProxies"
            >
              {{ t('admin.proxies.subscriptionImportSelectAll') }}
            </button>
            <button
              type="button"
              class="rounded-lg border border-gray-200 px-3 py-1.5 text-xs font-medium text-gray-600 transition-colors hover:bg-gray-50 dark:border-dark-600 dark:text-dark-300 dark:hover:bg-dark-800"
              :disabled="subscriptionImporting"
              @click="clearSelectedSubscriptionProxies"
            >
              {{ t('admin.proxies.subscriptionImportClearSelection') }}
            </button>
          </div>
        </div>
        <div class="text-xs text-gray-500 dark:text-dark-400">
          {{ t('admin.proxies.subscriptionImportSelectHint') }}
        </div>
        <div>
          <input
            v-model="subscriptionImportSearchQuery"
            type="text"
            class="input"
            :placeholder="t('admin.proxies.subscriptionImportSearchPlaceholder')"
          />
        </div>
        <div
          class="max-h-72 overflow-auto rounded-lg border border-gray-200 bg-gray-50 dark:border-dark-700 dark:bg-dark-800"
        >
          <label
            v-for="proxy in filteredSubscriptionParsedProxies"
            :key="subscriptionProxyKey(proxy)"
            class="flex cursor-pointer items-center gap-3 border-b border-gray-200 px-3 py-2 text-sm text-gray-700 last:border-b-0 hover:bg-white dark:border-dark-700 dark:text-dark-300 dark:hover:bg-dark-900"
          >
            <input
              type="checkbox"
              class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              :checked="isSubscriptionProxySelected(proxy)"
              :disabled="subscriptionImporting"
              @change="handleSubscriptionProxySelectionChange(proxy, $event)"
            />
            <div class="min-w-0 flex-1">
              <div class="truncate font-medium text-gray-900 dark:text-white">
                {{ proxy.name }}
              </div>
              <div class="truncate font-mono text-xs text-gray-500 dark:text-dark-400">
                {{ subscriptionProxyEndpointLabel(proxy) }}
              </div>
            </div>
            <span
              class="rounded-full bg-gray-200 px-2 py-0.5 text-[11px] font-semibold uppercase tracking-wide text-gray-600 dark:bg-dark-700 dark:text-dark-300"
            >
              {{ subscriptionProxyProtocolLabel(proxy) }}
            </span>
          </label>
          <div
            v-if="filteredSubscriptionParsedProxies.length === 0"
            class="px-3 py-6 text-center text-sm text-gray-500 dark:text-dark-400"
          >
            {{ t('admin.proxies.subscriptionImportNoMatch') }}
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button
          class="btn btn-secondary"
          type="button"
          :disabled="subscriptionParsing || subscriptionImporting"
          @click="closeDialog"
        >
          {{ t('common.cancel') }}
        </button>
        <button
          class="btn btn-secondary"
          type="button"
          :disabled="subscriptionParsing || subscriptionImporting"
          @click="handleParseSubscription"
        >
          {{ subscriptionParsing ? t('admin.proxies.subscriptionImportParsing') : t('admin.proxies.subscriptionImportParse') }}
        </button>
        <button
          class="btn btn-primary"
          type="button"
          :disabled="subscriptionImporting || subscriptionSelectedCount === 0"
          @click="handleSubscriptionImport"
        >
          {{
            subscriptionImporting
              ? t('admin.proxies.importing')
              : t('admin.proxies.subscriptionImportButton', { count: subscriptionSelectedCount })
          }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import proxySubscriptionsAPI from '@/api/admin/proxySubscriptions'
import type {
  ParsedSubscriptionProxy,
  SubscriptionClientType,
  SubscriptionParseStats
} from '@/api/admin/proxySubscriptions'
import { useAppStore } from '@/stores/app'

interface Props {
  show: boolean
}

interface Emits {
  (e: 'close'): void
  (e: 'imported'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()
const appStore = useAppStore()

const subscriptionImportUrl = ref('')
const subscriptionClientType = ref<SubscriptionClientType>('auto')
const subscriptionImportSearchQuery = ref('')
const subscriptionParsedProxies = ref<ParsedSubscriptionProxy[]>([])
const subscriptionParseStats = ref<SubscriptionParseStats | null>(null)
const subscriptionSelectedProxyKeys = ref<string[]>([])
const subscriptionParsing = ref(false)
const subscriptionImporting = ref(false)

const subscriptionClientTypeOptions: Array<{ value: SubscriptionClientType; labelKey: string }> = [
  { value: 'auto', labelKey: 'admin.proxies.subscriptionClientTypes.auto' },
  { value: 'default', labelKey: 'admin.proxies.subscriptionClientTypes.default' },
  { value: 'clash', labelKey: 'admin.proxies.subscriptionClientTypes.clash' },
  { value: 'clash-meta', labelKey: 'admin.proxies.subscriptionClientTypes.clashMeta' },
  { value: 'mihomo', labelKey: 'admin.proxies.subscriptionClientTypes.mihomo' },
  { value: 'sing-box', labelKey: 'admin.proxies.subscriptionClientTypes.singBox' },
  { value: 'surge', labelKey: 'admin.proxies.subscriptionClientTypes.surge' },
  { value: 'shadowrocket', labelKey: 'admin.proxies.subscriptionClientTypes.shadowrocket' },
  { value: 'stash', labelKey: 'admin.proxies.subscriptionClientTypes.stash' },
  { value: 'quantumult-x', labelKey: 'admin.proxies.subscriptionClientTypes.quantumultX' },
  { value: 'loon', labelKey: 'admin.proxies.subscriptionClientTypes.loon' },
  { value: 'v2rayn', labelKey: 'admin.proxies.subscriptionClientTypes.v2rayn' }
]

const subscriptionProxyKey = (proxy: ParsedSubscriptionProxy) => {
  return `${proxy.import_mode || 'direct'}|${proxy.node_protocol || proxy.protocol}|${proxy.host}|${proxy.port}|${proxy.username}|${proxy.password}|${proxy.name}`
}

const formatProtocolCounts = (counts: Record<string, number>) => {
  const entries = Object.entries(counts || {})
  if (entries.length === 0) return '-'
  return entries.map(([protocol, count]) => `${protocol}:${count}`).join(', ')
}

const subscriptionProxyProtocolLabel = (proxy: ParsedSubscriptionProxy) =>
  proxy.node_protocol || proxy.protocol

const subscriptionProxyEndpointLabel = (proxy: ParsedSubscriptionProxy) =>
  `${proxy.host}:${proxy.port}`

const subscriptionSelectedProxyKeySet = computed(() => new Set(subscriptionSelectedProxyKeys.value))

const filteredSubscriptionParsedProxies = computed(() => {
  const keyword = subscriptionImportSearchQuery.value.trim().toLowerCase()
  if (!keyword) return subscriptionParsedProxies.value
  return subscriptionParsedProxies.value.filter((proxy) =>
    [proxy.name, proxy.host, proxy.protocol, proxy.username]
      .some((value) => value.toLowerCase().includes(keyword))
  )
})

const selectedSubscriptionParsedProxies = computed(() =>
  subscriptionParsedProxies.value.filter((proxy) =>
    subscriptionSelectedProxyKeySet.value.has(subscriptionProxyKey(proxy))
  )
)

const subscriptionSelectedCount = computed(() => selectedSubscriptionParsedProxies.value.length)

const resetSubscriptionImportResult = () => {
  subscriptionImportSearchQuery.value = ''
  subscriptionParsedProxies.value = []
  subscriptionParseStats.value = null
  subscriptionSelectedProxyKeys.value = []
}

const resetDialogState = () => {
  subscriptionImportUrl.value = ''
  subscriptionClientType.value = 'auto'
  resetSubscriptionImportResult()
}

watch(
  () => props.show,
  (open) => {
    if (open) {
      resetDialogState()
    }
  }
)

const selectAllSubscriptionProxies = () => {
  const next = new Set(subscriptionSelectedProxyKeys.value)
  for (const proxy of filteredSubscriptionParsedProxies.value) {
    next.add(subscriptionProxyKey(proxy))
  }
  subscriptionSelectedProxyKeys.value = [...next]
}

const clearSelectedSubscriptionProxies = () => {
  const filteredKeys = new Set(filteredSubscriptionParsedProxies.value.map(subscriptionProxyKey))
  subscriptionSelectedProxyKeys.value = subscriptionSelectedProxyKeys.value.filter((key) => !filteredKeys.has(key))
}

const isSubscriptionProxySelected = (proxy: ParsedSubscriptionProxy) =>
  subscriptionSelectedProxyKeySet.value.has(subscriptionProxyKey(proxy))

const setSubscriptionProxySelected = (proxy: ParsedSubscriptionProxy, checked: boolean) => {
  const key = subscriptionProxyKey(proxy)
  if (checked) {
    if (subscriptionSelectedProxyKeySet.value.has(key)) return
    subscriptionSelectedProxyKeys.value = [...subscriptionSelectedProxyKeys.value, key]
    return
  }
  subscriptionSelectedProxyKeys.value = subscriptionSelectedProxyKeys.value.filter((item) => item !== key)
}

const handleSubscriptionProxySelectionChange = (proxy: ParsedSubscriptionProxy, event: Event) => {
  setSubscriptionProxySelected(proxy, (event.target as HTMLInputElement).checked)
}

const closeDialog = () => {
  if (subscriptionParsing.value || subscriptionImporting.value) return
  emit('close')
}

const handleParseSubscription = async () => {
  const url = subscriptionImportUrl.value.trim()
  if (!url) {
    appStore.showError(t('admin.proxies.subscriptionImportUrlRequired'))
    return
  }

  subscriptionParsing.value = true
  try {
    const result = await proxySubscriptionsAPI.parseSubscription(url, subscriptionClientType.value)
    subscriptionParsedProxies.value = result.proxies || []
    subscriptionParseStats.value = result.stats || null
    subscriptionSelectedProxyKeys.value = subscriptionParsedProxies.value.map(subscriptionProxyKey)
    if (subscriptionParsedProxies.value.length > 0) {
      appStore.showSuccess(
        t('admin.proxies.subscriptionImportParsed', { count: subscriptionParsedProxies.value.length })
      )
    } else if (subscriptionParseStats.value?.unsupported_count) {
      appStore.showError(t('admin.proxies.subscriptionImportUnsupportedOnly'))
    } else {
      appStore.showError(t('admin.proxies.subscriptionImportFailed'))
    }
  } catch (error: any) {
    resetSubscriptionImportResult()
    appStore.showError(error.response?.data?.detail || t('admin.proxies.subscriptionImportFailed'))
    console.error('Error parsing proxy subscription:', error)
  } finally {
    subscriptionParsing.value = false
  }
}

const handleSubscriptionImport = async () => {
  if (subscriptionParsedProxies.value.length === 0) {
    appStore.showError(t('admin.proxies.subscriptionImportEmpty'))
    return
  }
  if (selectedSubscriptionParsedProxies.value.length === 0) {
    appStore.showError(t('admin.proxies.subscriptionImportSelectionEmpty'))
    return
  }

  subscriptionImporting.value = true
  try {
    const result = await proxySubscriptionsAPI.importSubscription(selectedSubscriptionParsedProxies.value)
    const created = result.created || 0
    const skipped = result.skipped || 0

    if (created > 0) {
      appStore.showSuccess(t('admin.proxies.batchImportSuccess', { created, skipped }))
    } else {
      appStore.showInfo(t('admin.proxies.batchImportAllSkipped', { skipped }))
    }
    if (result.warnings?.length) {
      appStore.showError(result.warnings.join('\n'))
    }
    if (result.failed?.length) {
      appStore.showError(
        t('admin.proxies.subscriptionImportPartialFailed', { count: result.failed.length })
      )
    }

    resetDialogState()
    emit('imported')
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.proxies.failedToImport'))
    console.error('Error importing parsed subscription proxies:', error)
  } finally {
    subscriptionImporting.value = false
  }
}
</script>
