<template>
  <AppLayout>
    <div v-if="loading" class="flex justify-center py-20">
      <div class="h-10 w-10 animate-spin rounded-full border-2 border-cyan-500 border-t-transparent" />
    </div>

    <section
      v-else-if="!hasWorkspace"
      class="rounded-3xl border border-dashed border-slate-300 bg-white p-10 text-center shadow-sm"
    >
      <div class="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-slate-100 text-slate-500">
        <Icon name="chat" size="xl" />
      </div>
      <h2 class="mt-5 text-xl font-semibold text-slate-900">
        {{ t('openAIWeb.noEligibleTitle') }}
      </h2>
      <p class="mx-auto mt-3 max-w-2xl text-sm leading-6 text-slate-500">
        {{ t('openAIWeb.noEligibleDesc') }}
      </p>
    </section>

    <div
      v-else
      class="grid min-h-[calc(100vh-10rem)] gap-6 xl:grid-cols-[320px_minmax(0,1fr)]"
    >
      <aside class="flex min-h-0 flex-col overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
        <div class="border-b border-slate-200 p-4">
          <button
            type="button"
            class="inline-flex w-full items-center justify-center gap-2 rounded-2xl bg-slate-950 px-4 py-3 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-300"
            :disabled="!canStartNewChat"
            @click="handleStartNewChat"
          >
            <Icon name="plus" size="sm" />
            <span>{{ t('openAIWeb.newChat') }}</span>
          </button>

          <p v-if="!canStartNewChat" class="mt-3 text-xs leading-5 text-slate-500">
            {{ t('openAIWeb.createUnavailableDesc') }}
          </p>
        </div>

        <div class="flex-1 overflow-y-auto p-3">
          <div class="mb-3 px-2 text-sm font-semibold text-slate-500">
            {{ t('openAIWeb.historyTitle') }}
          </div>

          <div
            v-if="sortedThreads.length === 0"
            class="rounded-2xl border border-dashed border-slate-200 bg-slate-50 p-5 text-center"
          >
            <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-2xl bg-white text-slate-400 shadow-sm">
              <Icon name="inbox" size="lg" />
            </div>
            <h2 class="mt-4 text-sm font-semibold text-slate-900">
              {{ t('openAIWeb.historyEmptyTitle') }}
            </h2>
            <p class="mt-2 text-xs leading-6 text-slate-500">
              {{ t('openAIWeb.historyEmptyDesc') }}
            </p>
          </div>

          <div v-else class="space-y-2">
            <button
              v-for="thread in sortedThreads"
              :key="thread.local_thread_id"
              type="button"
              class="w-full rounded-2xl border px-4 py-3 text-left transition"
              :class="selectedThreadId === thread.local_thread_id
                ? 'border-slate-900 bg-slate-950 text-white shadow-lg shadow-slate-900/10'
                : 'border-slate-200 bg-slate-50 hover:border-slate-300 hover:bg-white'"
              @click="selectedThreadId = thread.local_thread_id"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <p class="truncate text-sm font-semibold">
                    {{ thread.title || t('openAIWeb.untitledThread') }}
                  </p>
                  <p
                    class="mt-1 truncate text-xs"
                    :class="selectedThreadId === thread.local_thread_id ? 'text-slate-300' : 'text-slate-500'"
                  >
                    {{ formatRelativeWithDateTime(thread.updated_at) }}
                  </p>
                </div>
              </div>
            </button>
          </div>
        </div>
      </aside>

      <section class="flex min-h-0 flex-col overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
        <div class="border-b border-slate-200 px-5 py-4 sm:px-6">
          <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
            <h1 class="truncate text-xl font-semibold tracking-tight text-slate-900">
              {{
                selectedThread
                  ? (selectedThread.title || t('openAIWeb.untitledThread'))
                  : t('openAIWeb.newChatTitle')
              }}
            </h1>

            <div class="flex flex-wrap items-center gap-2">
              <select
                v-model="composerRequestedModel"
                class="rounded-full border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-900 outline-none transition focus:border-cyan-400 focus:bg-white focus:ring-4 focus:ring-cyan-100 disabled:cursor-not-allowed disabled:opacity-60"
                :aria-label="t('openAIWeb.chatModelLabel')"
                :disabled="composerModelOptions.length === 0 || !canComposeCurrentChat"
              >
                <option
                  v-for="option in composerModelOptions"
                  :key="option"
                  :value="option"
                >
                  {{ option }}
                </option>
              </select>

              <select
                v-model="composerReasoningEffort"
                class="rounded-full border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-900 outline-none transition focus:border-cyan-400 focus:bg-white focus:ring-4 focus:ring-cyan-100 disabled:cursor-not-allowed disabled:opacity-60"
                :aria-label="t('openAIWeb.reasoningEffortLabel')"
                :disabled="!canComposeCurrentChat"
              >
                <option
                  v-for="option in reasoningEffortOptions"
                  :key="option.value || 'auto'"
                  :value="option.value"
                >
                  {{ option.label }}
                </option>
              </select>

              <button
                v-if="selectedThread"
                type="button"
                class="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-white px-3 py-2 text-sm font-medium text-slate-500 transition hover:border-rose-200 hover:bg-rose-50 hover:text-rose-600 disabled:cursor-not-allowed disabled:opacity-60"
                :aria-label="t('openAIWeb.archive')"
                :disabled="archivingIds.has(selectedThread.local_thread_id)"
                :title="t('openAIWeb.archive')"
                @click="handleArchiveThread(selectedThread)"
              >
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </div>
        </div>

        <div
          ref="messageViewport"
          class="flex-1 overflow-y-auto bg-[radial-gradient(circle_at_top,_rgba(14,165,233,0.08),_transparent_38%),linear-gradient(180deg,_#f8fafc_0%,_#ffffff_34%)] px-5 py-5 sm:px-6"
        >
          <div v-if="selectedMessages.length === 0" class="flex min-h-full items-center justify-center py-10">
            <div class="w-full max-w-2xl rounded-3xl border border-dashed border-slate-200 bg-white/90 p-8 text-center shadow-sm">
              <div class="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-slate-100 text-slate-500">
                <Icon :name="selectedThread ? 'chatBubble' : 'sparkles'" size="xl" />
              </div>
              <h3 class="mt-5 text-xl font-semibold text-slate-900">
                {{
                  selectedThread
                    ? t('openAIWeb.emptyChatTitle')
                    : t('openAIWeb.newChatTitle')
                }}
              </h3>
              <p class="mx-auto mt-3 max-w-xl text-sm leading-6 text-slate-500">
                <template v-if="selectedThread">
                  {{
                    t('openAIWeb.emptyChatDesc', {
                      model: selectedThread.requested_model || t('openAIWeb.modelAutoResolved'),
                    })
                  }}
                </template>
                <template v-else-if="canComposeCurrentChat">
                  {{ t('openAIWeb.newChatIntro') }}
                </template>
                <template v-else>
                  {{ t('openAIWeb.createUnavailableDesc') }}
                </template>
              </p>

              <div class="mt-5 flex flex-wrap justify-center gap-2 text-xs">
                <span class="rounded-full bg-slate-100 px-3 py-1 font-medium text-slate-600">
                  {{
                    composerRequestedModel
                      || selectedThread?.requested_model
                      || draftEntitlement?.default_model
                      || t('openAIWeb.modelAutoResolved')
                  }}
                </span>
                <span
                  v-if="composerReasoningEffort"
                  class="rounded-full bg-violet-50 px-3 py-1 font-medium text-violet-700"
                >
                  {{ t(`openAIWeb.reasoningEffort.${composerReasoningEffort}`) }}
                </span>
              </div>
            </div>
          </div>

          <div v-else class="space-y-4">
            <div
              v-for="message in selectedMessages"
              :key="message.id"
              class="flex"
              :class="message.role === 'user' ? 'justify-end' : 'justify-start'"
            >
              <div
                class="max-w-3xl rounded-[28px] border px-5 py-4 shadow-sm"
                :class="message.role === 'user'
                  ? 'border-slate-900 bg-slate-950 text-white shadow-slate-900/10'
                  : message.status === 'error'
                    ? 'border-rose-200 bg-rose-50 text-rose-900'
                    : 'border-slate-200 bg-white text-slate-900'"
              >
                <div class="mb-3 flex items-center justify-between gap-4 text-xs">
                  <span
                    class="font-semibold uppercase tracking-[0.18em]"
                    :class="message.role === 'user'
                      ? 'text-slate-300'
                      : message.status === 'error'
                        ? 'text-rose-500'
                        : 'text-slate-400'"
                  >
                    {{ message.role === 'user' ? t('openAIWeb.userLabel') : t('openAIWeb.assistantLabel') }}
                  </span>
                  <span :class="message.role === 'user' ? 'text-slate-400' : 'text-slate-400'">
                    {{ formatDateTime(message.created_at) }}
                  </span>
                </div>

                <div
                  v-if="message.role === 'user' && message.content?.trim()"
                  class="whitespace-pre-wrap break-words text-sm leading-7"
                >
                  {{ message.content }}
                </div>
                <div
                  v-else-if="message.status === 'pending'"
                  class="space-y-3 text-sm leading-7 text-slate-600"
                >
                  <div class="inline-flex items-center gap-2 rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-500">
                    <span class="h-2 w-2 animate-pulse rounded-full bg-cyan-500" />
                    {{ t('openAIWeb.pendingReply') }}
                  </div>
                </div>
                <div v-else-if="message.status === 'error'" class="space-y-3 text-sm leading-7">
                  <p class="font-medium">{{ t('openAIWeb.failedReply') }}</p>
                  <p class="whitespace-pre-wrap break-words">{{ message.error || message.content }}</p>
                </div>
                <div
                  v-else-if="message.content?.trim()"
                  class="openai-web-markdown prose prose-sm max-w-none text-sm leading-7 text-slate-700"
                  v-html="renderMessageHtml(message.content)"
                />

                <div
                  v-if="messageDisplayImages(message).length > 0"
                  class="mt-4 grid gap-3 sm:grid-cols-2"
                >
                  <div
                    v-for="(image, index) in messageDisplayImages(message)"
                    :key="`${message.id}-image-${index}`"
                    class="group overflow-hidden rounded-2xl border border-slate-200 bg-slate-50 transition hover:border-cyan-300 hover:bg-white"
                  >
                    <a
                      :href="image.data_url"
                      target="_blank"
                      rel="noreferrer"
                      class="block"
                    >
                      <img
                        :src="image.data_url"
                        :alt="imageAltText(image)"
                        class="max-h-[340px] w-full object-cover"
                      >
                    </a>
                    <div class="space-y-3 border-t border-slate-200 px-3 py-3 text-xs text-slate-500">
                      <p
                        v-if="imageDisplayCaption(image)"
                        class="line-clamp-3 leading-5"
                      >
                        {{ imageDisplayCaption(image) }}
                      </p>
                      <div class="flex flex-wrap gap-2">
                        <span
                          v-if="image.file_name"
                          class="rounded-full bg-white px-2.5 py-1 font-medium text-slate-600 shadow-sm"
                        >
                          {{ image.file_name }}
                        </span>
                        <span
                          v-if="imageDimensionsLabel(image)"
                          class="rounded-full bg-white px-2.5 py-1 font-medium text-slate-600 shadow-sm"
                        >
                          {{ imageDimensionsLabel(image) }}
                        </span>
                        <span
                          v-if="imageFileSizeLabel(image)"
                          class="rounded-full bg-white px-2.5 py-1 font-medium text-slate-600 shadow-sm"
                        >
                          {{ imageFileSizeLabel(image) }}
                        </span>
                      </div>
                      <div class="flex flex-wrap gap-2">
                        <a
                          :href="image.data_url"
                          target="_blank"
                          rel="noreferrer"
                          class="inline-flex items-center gap-1 rounded-full border border-slate-200 bg-white px-2.5 py-1 font-medium text-slate-600 transition hover:border-slate-300 hover:text-slate-900"
                        >
                          <Icon name="externalLink" size="xs" />
                          <span>{{ t('openAIWeb.previewImage') }}</span>
                        </a>
                        <a
                          :href="image.data_url"
                          :download="downloadFileNameForImage(image, index, message.role)"
                          class="inline-flex items-center gap-1 rounded-full border border-slate-200 bg-white px-2.5 py-1 font-medium text-slate-600 transition hover:border-slate-300 hover:text-slate-900"
                        >
                          <Icon name="download" size="xs" />
                          <span>{{ t('openAIWeb.downloadImage') }}</span>
                        </a>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="border-t border-slate-200 bg-white px-5 py-5 sm:px-6">
          <input
            ref="composerFileInput"
            type="file"
            accept="image/*"
            multiple
            class="hidden"
            @change="handleComposerFilesSelected"
          >

          <div
            v-if="composerAttachments.length > 0"
            class="mb-4 flex flex-wrap items-center gap-2 text-xs"
          >
            <span class="rounded-full bg-amber-50 px-3 py-1 font-medium text-amber-700">
              {{ t('openAIWeb.imageCountTag', { count: composerAttachments.length }) }}
            </span>
            <button
              type="button"
              class="inline-flex items-center gap-2 rounded-full border border-rose-200 bg-white px-3 py-1 font-medium text-rose-600 transition hover:border-rose-300 hover:text-rose-700"
              @click="clearComposerAttachments"
            >
              <Icon name="trash" size="sm" />
              <span>{{ t('openAIWeb.clearImages') }}</span>
            </button>
          </div>

          <div
            v-if="composerAttachments.length > 0"
            class="mb-4 grid gap-3 sm:grid-cols-2"
          >
            <div
              v-for="attachment in composerAttachments"
              :key="attachment.id"
              class="overflow-hidden rounded-2xl border border-slate-200 bg-slate-50"
            >
              <img
                :src="attachment.data_url"
                :alt="attachment.file_name || t('openAIWeb.assistantImageAlt')"
                class="max-h-48 w-full object-cover"
              >
              <div class="flex items-center justify-between gap-3 border-t border-slate-200 px-3 py-2">
                <p class="min-w-0 truncate text-xs text-slate-500">
                  {{ attachment.file_name || t('openAIWeb.imageAttachment') }}
                </p>
                <button
                  type="button"
                  class="inline-flex items-center gap-1 rounded-full border border-slate-200 bg-white px-2 py-1 text-xs font-medium text-slate-600 transition hover:border-slate-300 hover:text-slate-900"
                  @click="removeComposerAttachment(attachment.id)"
                >
                  <Icon name="x" size="xs" />
                  <span>{{ t('openAIWeb.removeImage') }}</span>
                </button>
              </div>
            </div>
          </div>

          <div class="flex items-end gap-3 rounded-3xl border border-slate-200 bg-slate-50 px-3 py-3 transition focus-within:border-cyan-400 focus-within:bg-white focus-within:ring-4 focus-within:ring-cyan-100">
            <button
              type="button"
              class="mb-1 inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl text-slate-500 transition hover:bg-white hover:text-slate-900 disabled:cursor-not-allowed disabled:opacity-50"
              :aria-label="t('openAIWeb.addImages')"
              :disabled="!canComposeCurrentChat"
              @click="openComposerFilePicker"
            >
              <Icon name="upload" size="sm" />
            </button>

            <textarea
              v-model="composer"
              rows="1"
              :placeholder="t('openAIWeb.composerPlaceholder')"
              class="max-h-44 min-h-10 flex-1 resize-none border-0 bg-transparent px-1 py-2 text-sm text-slate-900 outline-none disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="!canComposeCurrentChat"
              @keydown.enter.exact.prevent="handleSendMessage"
            />

            <button
              type="button"
              class="inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-slate-950 text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
              :disabled="sending || !canComposeCurrentChat || (!composer.trim() && composerAttachments.length === 0)"
              @click="handleSendMessage"
            >
              <Icon :name="sending ? 'refresh' : 'arrowUp'" size="sm" />
            </button>
          </div>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import DOMPurify from 'dompurify'
import { marked } from 'marked'
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import {
  useOpenAIWebModelOptions,
  type OpenAIWebComposerReasoningEffort,
} from '@/composables/useOpenAIWebModelOptions'
import openAIWebAPI from '@/api/openaiWeb'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDateTime, formatRelativeWithDateTime } from '@/utils/format'
import type {
  OpenAIWebEntitlement,
  OpenAIWebThread,
  OpenAIWebThreadMessageAttachment,
  OpenAIWebThreadMessageImage,
  OpenAIWebThreadMessageUsage,
} from '@/types'

type LocalMessageStatus = 'done' | 'pending' | 'error'
type LocalMessageRole = 'user' | 'assistant'

interface OpenAIWebLocalMessageImage extends OpenAIWebThreadMessageImage {
  file_name?: string
  byte_size?: number
}

interface OpenAIWebLocalMessage {
  id: string
  role: LocalMessageRole
  content: string
  images?: OpenAIWebLocalMessageImage[] | null
  created_at: string
  status: LocalMessageStatus
  error?: string
  request_id?: string
  response_id?: string | null
  usage?: OpenAIWebThreadMessageUsage | null
}

interface ComposerAttachment extends OpenAIWebThreadMessageAttachment {
  id: string
  byte_size: number
}

interface OpenAIWebComposerSettings {
  requested_model?: string
  reasoning_effort?: OpenAIWebComposerReasoningEffort
}

const OPENAI_WEB_MESSAGE_STORAGE_PREFIX = 'openai-web-thread-messages:'
const OPENAI_WEB_COMPOSER_SETTINGS_STORAGE_PREFIX = 'openai-web-thread-composer-settings:'
const OPENAI_WEB_DRAFT_COMPOSER_SETTINGS_STORAGE_KEY = 'openai-web-draft-composer-settings'
const OPENAI_WEB_PREFERRED_GROUP_STORAGE_KEY = 'openai-web-preferred-group-id'
const OPENAI_WEB_MAX_ATTACHMENT_COUNT = 4
const OPENAI_WEB_MAX_ATTACHMENT_BYTES = 8 * 1024 * 1024

marked.setOptions({
  gfm: true,
  breaks: true,
})

const { t } = useI18n()
const appStore = useAppStore()
const route = useRoute()
const router = useRouter()

const loading = ref(true)
const sending = ref(false)
const entitlements = ref<OpenAIWebEntitlement[]>([])
const {
  availableChannels,
  loadAvailableChannels,
  openAIWebModelOptionsForEntitlement,
  openAIWebModelOptionsForDraft,
  resolveDraftEntitlement,
  normalizeComposerReasoningEffort,
  normalizeRequestedModel,
} = useOpenAIWebModelOptions(entitlements, readPreferredGroupID)
const threads = ref<OpenAIWebThread[]>([])
const threadMessages = ref<Record<string, OpenAIWebLocalMessage[]>>({})
const archivingIds = ref<Set<string>>(new Set())
const selectedThreadId = ref(typeof route.query.thread === 'string' ? route.query.thread : '')
const composer = ref('')
const composerRequestedModel = ref('')
const composerReasoningEffort = ref<OpenAIWebComposerReasoningEffort>('')
const composerFileInput = ref<HTMLInputElement | null>(null)
const composerAttachments = ref<ComposerAttachment[]>([])
const messageViewport = ref<HTMLElement | null>(null)
const localCacheWarningState = ref<'none' | 'trimmed' | 'disabled'>('none')

const hasWorkspace = computed(() => entitlements.value.length > 0 || threads.value.length > 0)
const canStartNewChat = computed(() => entitlements.value.length > 0)

const sortedThreads = computed(() => {
  return [...threads.value].sort((left, right) => {
    return new Date(right.updated_at).getTime() - new Date(left.updated_at).getTime()
  })
})

const selectedThread = computed(() => {
  return sortedThreads.value.find((item) => item.local_thread_id === selectedThreadId.value) ?? null
})

const selectedThreadEntitlement = computed(() => {
  if (!selectedThread.value) {
    return null
  }
  return entitlements.value.find((item) => item.group_id === selectedThread.value?.group_id) ?? null
})

const selectedThreadModelOptions = computed(() => {
  return openAIWebModelOptionsForEntitlement(
    selectedThreadEntitlement.value,
    selectedThread.value?.requested_model || composerRequestedModel.value
  )
})

const draftEntitlement = computed(() => {
  return resolveDraftEntitlement(composerRequestedModel.value)
})

const draftModelOptions = computed(() => {
  if (entitlements.value.length === 0) {
    return []
  }
  return openAIWebModelOptionsForDraft(entitlements.value, composerRequestedModel.value)
})

const composerModelOptions = computed(() => {
  return selectedThread.value ? selectedThreadModelOptions.value : draftModelOptions.value
})

const reasoningEffortOptions = computed(() => ([
  { value: '' as OpenAIWebComposerReasoningEffort, label: t('openAIWeb.reasoningEffortAuto') },
  { value: 'low' as OpenAIWebComposerReasoningEffort, label: t('openAIWeb.reasoningEffort.low') },
  { value: 'medium' as OpenAIWebComposerReasoningEffort, label: t('openAIWeb.reasoningEffort.medium') },
  { value: 'high' as OpenAIWebComposerReasoningEffort, label: t('openAIWeb.reasoningEffort.high') },
  { value: 'xhigh' as OpenAIWebComposerReasoningEffort, label: t('openAIWeb.reasoningEffort.xhigh') },
]))

const selectedMessages = computed(() => {
  const threadID = selectedThread.value?.local_thread_id
  if (!threadID) {
    return []
  }
  return threadMessages.value[threadID] ?? []
})

const canComposeCurrentChat = computed(() => {
  return !!selectedThread.value || canStartNewChat.value
})

watch(
  () => route.query.thread,
  (value) => {
    const next = typeof value === 'string' ? value : ''
    if (next !== selectedThreadId.value) {
      selectedThreadId.value = next
    }
  }
)

watch(
  () => sortedThreads.value.map((item) => item.local_thread_id),
  (items) => {
    if (selectedThreadId.value && !items.includes(selectedThreadId.value)) {
      selectedThreadId.value = ''
    }
  },
  { immediate: true }
)

watch(
  () => selectedThreadId.value,
  async (value) => {
    const current = typeof route.query.thread === 'string' ? route.query.thread : ''
    if (value === current) {
      return
    }

    const query = { ...route.query }
    if (value) {
      query.thread = value
    } else {
      delete query.thread
    }

    try {
      await router.replace({ query })
    } catch {
      // noop
    }
  }
)

watch(
  () => selectedThread.value?.local_thread_id,
  (threadID) => {
    if (threadID && selectedThread.value) {
      ensureThreadMessagesLoaded(threadID)
      applyComposerSettingsForThread(selectedThread.value)
      persistPreferredGroupID(selectedThread.value.group_id)
    } else {
      applyDraftComposerSettings()
    }
    composer.value = ''
    composerAttachments.value = []
    void scrollMessagesToBottom()
  },
  { immediate: true }
)

watch(
  () => [selectedThread.value?.local_thread_id, composerRequestedModel.value, composerReasoningEffort.value] as const,
  ([threadID]) => {
    if (threadID) {
      persistComposerSettings(threadID, {
        requested_model: composerRequestedModel.value,
        reasoning_effort: composerReasoningEffort.value,
      })
      return
    }

    persistDraftComposerSettings({
      requested_model: composerRequestedModel.value,
      reasoning_effort: composerReasoningEffort.value,
    })
  }
)

watch(
  () => entitlements.value,
  () => {
    if (!selectedThread.value) {
      applyDraftComposerSettings()
    }
  },
  { deep: true }
)

async function loadAll() {
  try {
    const [loadedEntitlements, loadedThreads, loadedChannels] = await Promise.all([
      openAIWebAPI.getOpenAIWebEntitlements(),
      openAIWebAPI.getOpenAIWebThreads(),
      loadAvailableChannels(),
    ])

    entitlements.value = loadedEntitlements
    threads.value = loadedThreads
    availableChannels.value = loadedChannels
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('openAIWeb.failedToLoad')))
  } finally {
    loading.value = false
  }
}

function handleStartNewChat() {
  if (!canStartNewChat.value) {
    return
  }
  selectedThreadId.value = ''
  composer.value = ''
  composerAttachments.value = []
  void scrollMessagesToBottom()
}

async function handleArchiveThread(thread: OpenAIWebThread) {
  const next = new Set(archivingIds.value)
  next.add(thread.local_thread_id)
  archivingIds.value = next

  try {
    await openAIWebAPI.archiveOpenAIWebThread(thread.local_thread_id)
    threads.value = threads.value.filter((item) => item.local_thread_id !== thread.local_thread_id)
    clearThreadMessages(thread.local_thread_id)
    if (selectedThreadId.value === thread.local_thread_id) {
      selectedThreadId.value = ''
    }
    appStore.showSuccess(t('openAIWeb.archiveSuccess'))
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('openAIWeb.archiveFailed')))
  } finally {
    const updated = new Set(archivingIds.value)
    updated.delete(thread.local_thread_id)
    archivingIds.value = updated
  }
}

function readPreferredGroupID(): number | null {
  if (typeof window === 'undefined') {
    return null
  }

  const raw = window.localStorage.getItem(OPENAI_WEB_PREFERRED_GROUP_STORAGE_KEY)
  if (!raw) {
    return null
  }

  const parsed = Number.parseInt(raw, 10)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null
}

function persistPreferredGroupID(groupID: number) {
  if (typeof window === 'undefined' || !Number.isFinite(groupID) || groupID <= 0) {
    return
  }

  try {
    window.localStorage.setItem(OPENAI_WEB_PREFERRED_GROUP_STORAGE_KEY, String(groupID))
  } catch {
    // noop
  }
}

function resolveComposerRequestedModel(thread: OpenAIWebThread | null): string {
  const explicit = normalizeRequestedModel(composerRequestedModel.value)
  if (explicit) {
    return explicit
  }
  if (thread?.requested_model?.trim()) {
    return thread.requested_model.trim()
  }
  if (selectedThreadEntitlement.value?.default_model?.trim()) {
    return selectedThreadEntitlement.value.default_model.trim()
  }
  if (draftEntitlement.value?.default_model?.trim()) {
    return draftEntitlement.value.default_model.trim()
  }
  return composerModelOptions.value[0] || ''
}

function applyComposerSettingsForThread(thread: OpenAIWebThread | null) {
  if (!thread) {
    composerRequestedModel.value = ''
    composerReasoningEffort.value = ''
    return
  }

  const saved = readComposerSettings(thread.local_thread_id)
  const defaultModel = thread.requested_model?.trim()
    || selectedThreadEntitlement.value?.default_model?.trim()
    || selectedThreadModelOptions.value[0]
    || ''
  const requestedModel = saved.requested_model?.trim() || defaultModel
  const availableModels = openAIWebModelOptionsForEntitlement(selectedThreadEntitlement.value, requestedModel)

  composerRequestedModel.value = availableModels.includes(requestedModel) ? requestedModel : defaultModel
  composerReasoningEffort.value = normalizeComposerReasoningEffort(saved.reasoning_effort)
}

function applyDraftComposerSettings() {
  const saved = readDraftComposerSettings()
  const savedModel = saved.requested_model?.trim() || ''
  const availableModels = openAIWebModelOptionsForDraft(entitlements.value, savedModel)
  const defaultEntitlement = resolveDraftEntitlement(savedModel)
  const defaultModel = defaultEntitlement?.default_model?.trim()
    || availableModels[0]
    || ''
  const requestedModel = savedModel || defaultModel

  composerRequestedModel.value = availableModels.includes(requestedModel) ? requestedModel : defaultModel
  composerReasoningEffort.value = normalizeComposerReasoningEffort(saved.reasoning_effort)
}

async function createThreadForFirstMessage(
  content: string,
  requestedModel: string,
  attachments: ComposerAttachment[]
): Promise<OpenAIWebThread | null> {
  const entitlement = resolveDraftEntitlement(requestedModel)
  if (!entitlement) {
    appStore.showError(t('openAIWeb.noMatchingGroupForModel', {
      model: requestedModel || t('openAIWeb.modelAutoResolved'),
    }))
    return null
  }

  try {
    const created = await openAIWebAPI.createOpenAIWebThread({
      group_id: entitlement.group_id,
      requested_model: requestedModel || undefined,
      title: buildImplicitThreadTitle(content, attachments.length),
      cache_policy: 'local_only',
    })

    threads.value = [created, ...threads.value.filter((item) => item.local_thread_id !== created.local_thread_id)]
    selectedThreadId.value = created.local_thread_id
    persistPreferredGroupID(created.group_id)
    ensureThreadMessagesLoaded(created.local_thread_id)
    return created
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('openAIWeb.createFailed')))
    return null
  }
}

function buildImplicitThreadTitle(content: string, attachmentCount: number): string {
  const normalized = content.replace(/\s+/g, ' ').trim()
  if (normalized) {
    return normalized.slice(0, 80)
  }
  if (attachmentCount > 0) {
    return t('openAIWeb.imageConversationTitle')
  }
  return t('openAIWeb.newChatTitle')
}

async function handleSendMessage() {
  const content = composer.value.trim()
  const rawAttachments = [...composerAttachments.value]
  if (sending.value || (!content && rawAttachments.length === 0)) {
    return
  }

  let thread = selectedThread.value
  const requestedModel = resolveComposerRequestedModel(thread)
  const reasoningEffort = normalizeComposerReasoningEffort(composerReasoningEffort.value)
  let pendingAssistantID = ''

  sending.value = true
  try {
    if (!thread) {
      if (!canStartNewChat.value) {
        appStore.showError(t('openAIWeb.createUnavailableDesc'))
        return
      }

      thread = await createThreadForFirstMessage(content, requestedModel, rawAttachments)
      if (!thread) {
        return
      }
    }

    composer.value = ''
    composerAttachments.value = []

    const attachments = rawAttachments.map(composerAttachmentToRequest)
    const existingMessages = threadMessages.value[thread.local_thread_id] ?? []
    const userMessage = buildLocalMessage('user', content, 'done', rawAttachments.map(composerAttachmentToMessageImage))
    const pendingAssistant = buildLocalMessage('assistant', '', 'pending')
    pendingAssistantID = pendingAssistant.id
    const draftMessages = [...existingMessages, userMessage, pendingAssistant]
    persistThreadMessages(thread.local_thread_id, draftMessages)

    const response = await openAIWebAPI.sendOpenAIWebThreadMessage(thread.local_thread_id, {
      content,
      requested_model: requestedModel || undefined,
      reasoning_effort: reasoningEffort || undefined,
      attachments,
    })

    threads.value = [
      response.thread,
      ...threads.value.filter((item) => item.local_thread_id !== response.thread.local_thread_id),
    ]
    if (selectedThreadId.value === response.thread.local_thread_id) {
      composerRequestedModel.value = response.thread.requested_model || requestedModel
    }
    persistPreferredGroupID(response.thread.group_id)

    persistThreadMessages(
      thread.local_thread_id,
      draftMessages.map((message) => {
        if (message.id !== pendingAssistant.id) {
          return message
        }
        return {
          ...message,
          content: response.assistant_text?.trim() || ((response.assistant_images?.length ?? 0) > 0 ? '' : t('openAIWeb.emptyAssistantReply')),
          images: response.assistant_images ?? [],
          status: 'done',
          request_id: response.request_id,
          response_id: response.response_id,
          usage: response.usage,
        }
      })
    )
  } catch (err) {
    const errorMessage = extractApiErrorMessage(err, t('openAIWeb.sendFailed'))
    const threadID = thread?.local_thread_id
    if (threadID) {
      persistThreadMessages(
        threadID,
        (threadMessages.value[threadID] ?? []).map((message) => {
          if (message.id !== pendingAssistantID) {
            return message
          }
          return {
            ...message,
            content: errorMessage,
            status: 'error',
            error: errorMessage,
          }
        })
      )
    }
    appStore.showError(errorMessage)
  } finally {
    sending.value = false
    void scrollMessagesToBottom()
  }
}

function openComposerFilePicker() {
  composerFileInput.value?.click()
}

async function handleComposerFilesSelected(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files ?? [])
  input.value = ''

  if (files.length === 0) {
    return
  }

  const remaining = OPENAI_WEB_MAX_ATTACHMENT_COUNT - composerAttachments.value.length
  if (remaining <= 0) {
    appStore.showWarning(t('openAIWeb.attachmentLimitReached', { count: OPENAI_WEB_MAX_ATTACHMENT_COUNT }))
    return
  }

  if (files.length > remaining) {
    appStore.showWarning(t('openAIWeb.attachmentLimitReached', { count: OPENAI_WEB_MAX_ATTACHMENT_COUNT }))
  }

  const accepted = files.slice(0, Math.max(0, remaining))
  const nextItems: ComposerAttachment[] = []

  for (const file of accepted) {
    if (!file.type.startsWith('image/')) {
      appStore.showWarning(t('openAIWeb.attachmentTypeInvalid'))
      continue
    }
    if (file.size > OPENAI_WEB_MAX_ATTACHMENT_BYTES) {
      appStore.showWarning(t('openAIWeb.attachmentTooLarge', { size: formatAttachmentSizeLimit(OPENAI_WEB_MAX_ATTACHMENT_BYTES) }))
      continue
    }

    try {
      nextItems.push(await readComposerAttachment(file))
    } catch {
      appStore.showError(t('openAIWeb.attachmentReadFailed'))
    }
  }

  if (nextItems.length > 0) {
    composerAttachments.value = [...composerAttachments.value, ...nextItems]
  }
}

function removeComposerAttachment(attachmentID: string) {
  composerAttachments.value = composerAttachments.value.filter((item) => item.id !== attachmentID)
}

function clearComposerAttachments() {
  composerAttachments.value = []
}

function renderMessageHtml(content: string): string {
  const html = marked.parse(content || '') as string
  return DOMPurify.sanitize(html)
}

function messageDisplayImages(message: OpenAIWebLocalMessage): OpenAIWebLocalMessageImage[] {
  return (message.images ?? []).filter((item) => !!item?.data_url)
}

function imageAltText(image: OpenAIWebLocalMessageImage): string {
  return image.revised_prompt?.trim() || image.file_name?.trim() || t('openAIWeb.assistantImageAlt')
}

function imageDisplayCaption(image: OpenAIWebLocalMessageImage): string {
  return image.revised_prompt?.trim() || image.file_name?.trim() || ''
}

function imageDimensionsLabel(image: OpenAIWebLocalMessageImage): string {
  if (!image.width || !image.height) {
    return ''
  }
  return t('openAIWeb.imageDimensionsTag', { width: image.width, height: image.height })
}

function imageFileSizeLabel(image: OpenAIWebLocalMessageImage): string {
  if (!image.byte_size) {
    return ''
  }
  return formatAttachmentByteSize(image.byte_size)
}

function downloadFileNameForImage(
  image: OpenAIWebLocalMessageImage,
  index: number,
  role: LocalMessageRole
): string {
  const explicitName = image.file_name?.trim()
  if (explicitName) {
    return explicitName
  }
  const prefix = role === 'assistant' ? 'openai-web-output' : 'openai-web-input'
  return `${prefix}-${index + 1}.${imageFileExtension(image.mime_type)}`
}

function imageFileExtension(mimeType?: string): string {
  switch ((mimeType || '').toLowerCase()) {
    case 'image/jpeg':
    case 'image/jpg':
      return 'jpg'
    case 'image/webp':
      return 'webp'
    case 'image/gif':
      return 'gif'
    default:
      return 'png'
  }
}

function buildLocalMessage(
  role: LocalMessageRole,
  content: string,
  status: LocalMessageStatus,
  images?: OpenAIWebLocalMessageImage[]
): OpenAIWebLocalMessage {
  return {
    id: createLocalMessageID(),
    role,
    content,
    images,
    status,
    created_at: new Date().toISOString(),
  }
}

function composerAttachmentToRequest(item: ComposerAttachment): OpenAIWebThreadMessageAttachment {
  return {
    file_name: item.file_name,
    content_type: item.content_type,
    data_url: item.data_url,
    width: item.width,
    height: item.height,
  }
}

function composerAttachmentToMessageImage(item: ComposerAttachment): OpenAIWebLocalMessageImage {
  return {
    data_url: item.data_url,
    mime_type: item.content_type,
    file_name: item.file_name,
    width: item.width,
    height: item.height,
    byte_size: item.byte_size,
  }
}

function createLocalMessageID(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`
}

async function readComposerAttachment(file: File): Promise<ComposerAttachment> {
  const dataURL = await readFileAsDataURL(file)
  const dimensions = await readImageDimensions(dataURL)

  return {
    id: createLocalMessageID(),
    file_name: file.name,
    content_type: file.type,
    data_url: dataURL,
    width: dimensions.width,
    height: dimensions.height,
    byte_size: file.size,
  }
}

function readFileAsDataURL(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(new Error('read data url failed'))
    reader.readAsDataURL(file)
  })
}

function readImageDimensions(dataURL: string): Promise<{ width?: number; height?: number }> {
  return new Promise((resolve) => {
    const image = new Image()
    image.onload = () => resolve({
      width: image.naturalWidth || undefined,
      height: image.naturalHeight || undefined,
    })
    image.onerror = () => resolve({})
    image.src = dataURL
  })
}

function formatAttachmentSizeLimit(bytes: number): string {
  return formatAttachmentByteSize(bytes)
}

function formatAttachmentByteSize(bytes: number): string {
  if (bytes < 1024 * 1024) {
    return `${Math.max(1, Math.round(bytes / 1024))} KB`
  }
  return `${(bytes / (1024 * 1024)).toFixed(bytes >= 10 * 1024 * 1024 ? 0 : 1)} MB`
}

function ensureThreadMessagesLoaded(threadID: string) {
  if (!threadID || Object.prototype.hasOwnProperty.call(threadMessages.value, threadID)) {
    return
  }
  const next = { ...threadMessages.value }
  next[threadID] = readThreadMessages(threadID)
  threadMessages.value = next
}

function persistThreadMessages(threadID: string, messages: OpenAIWebLocalMessage[]) {
  const next = { ...threadMessages.value, [threadID]: messages }
  threadMessages.value = next
  if (typeof window !== 'undefined') {
    try {
      window.localStorage.setItem(storageKey(threadID), JSON.stringify(messages))
      void scrollMessagesToBottom()
      return
    } catch {
      try {
        window.localStorage.setItem(storageKey(threadID), JSON.stringify(messages.map(compactLocalMessageForStorage)))
        if (localCacheWarningState.value === 'none') {
          localCacheWarningState.value = 'trimmed'
          appStore.showWarning(t('openAIWeb.localCacheImageWarning'))
        }
      } catch {
        if (localCacheWarningState.value !== 'disabled') {
          localCacheWarningState.value = 'disabled'
          appStore.showWarning(t('openAIWeb.localCacheDisabledWarning'))
        }
      }
    }
  }
  void scrollMessagesToBottom()
}

function clearThreadMessages(threadID: string) {
  const next = { ...threadMessages.value }
  delete next[threadID]
  threadMessages.value = next
  if (typeof window !== 'undefined') {
    window.localStorage.removeItem(storageKey(threadID))
    window.localStorage.removeItem(composerSettingsStorageKey(threadID))
  }
}

function readThreadMessages(threadID: string): OpenAIWebLocalMessage[] {
  if (typeof window === 'undefined') {
    return []
  }

  try {
    const raw = window.localStorage.getItem(storageKey(threadID))
    if (!raw) {
      return []
    }
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) {
      return []
    }
    return parsed.filter((item): item is OpenAIWebLocalMessage => {
      return !!item
        && typeof item.id === 'string'
        && typeof item.role === 'string'
        && typeof item.created_at === 'string'
        && typeof item.content === 'string'
    })
  } catch {
    return []
  }
}

function compactLocalMessageForStorage(message: OpenAIWebLocalMessage): OpenAIWebLocalMessage {
  if (!message.images?.length) {
    return message
  }
  return {
    ...message,
    images: [],
  }
}

function storageKey(threadID: string): string {
  return `${OPENAI_WEB_MESSAGE_STORAGE_PREFIX}${threadID}`
}

function composerSettingsStorageKey(threadID: string): string {
  return `${OPENAI_WEB_COMPOSER_SETTINGS_STORAGE_PREFIX}${threadID}`
}

function readComposerSettings(threadID: string): OpenAIWebComposerSettings {
  if (typeof window === 'undefined' || !threadID) {
    return {}
  }

  try {
    const raw = window.localStorage.getItem(composerSettingsStorageKey(threadID))
    if (!raw) {
      return {}
    }
    const parsed = JSON.parse(raw) as OpenAIWebComposerSettings | null
    if (!parsed || typeof parsed !== 'object') {
      return {}
    }
    return {
      requested_model: typeof parsed.requested_model === 'string' ? parsed.requested_model : undefined,
      reasoning_effort: normalizeComposerReasoningEffort(parsed.reasoning_effort),
    }
  } catch {
    return {}
  }
}

function persistComposerSettings(threadID: string, settings: OpenAIWebComposerSettings) {
  if (typeof window === 'undefined' || !threadID) {
    return
  }

  try {
    window.localStorage.setItem(composerSettingsStorageKey(threadID), JSON.stringify({
      requested_model: settings.requested_model?.trim() || undefined,
      reasoning_effort: normalizeComposerReasoningEffort(settings.reasoning_effort),
    }))
  } catch {
    // noop
  }
}

function readDraftComposerSettings(): OpenAIWebComposerSettings {
  if (typeof window === 'undefined') {
    return {}
  }

  try {
    const raw = window.localStorage.getItem(OPENAI_WEB_DRAFT_COMPOSER_SETTINGS_STORAGE_KEY)
    if (!raw) {
      return {}
    }
    const parsed = JSON.parse(raw) as OpenAIWebComposerSettings | null
    if (!parsed || typeof parsed !== 'object') {
      return {}
    }
    return {
      requested_model: typeof parsed.requested_model === 'string' ? parsed.requested_model : undefined,
      reasoning_effort: normalizeComposerReasoningEffort(parsed.reasoning_effort),
    }
  } catch {
    return {}
  }
}

function persistDraftComposerSettings(settings: OpenAIWebComposerSettings) {
  if (typeof window === 'undefined') {
    return
  }

  try {
    window.localStorage.setItem(OPENAI_WEB_DRAFT_COMPOSER_SETTINGS_STORAGE_KEY, JSON.stringify({
      requested_model: settings.requested_model?.trim() || undefined,
      reasoning_effort: normalizeComposerReasoningEffort(settings.reasoning_effort),
    }))
  } catch {
    // noop
  }
}

async function scrollMessagesToBottom() {
  await nextTick()
  const viewport = messageViewport.value
  if (viewport) {
    viewport.scrollTop = viewport.scrollHeight
  }
}

onMounted(() => {
  void loadAll()
})
</script>

<style scoped>
.openai-web-markdown :deep(p) {
  margin: 0 0 0.9rem;
}

.openai-web-markdown :deep(p:last-child) {
  margin-bottom: 0;
}

.openai-web-markdown :deep(pre) {
  overflow-x: auto;
  border-radius: 1rem;
  background: #0f172a;
  padding: 1rem;
  color: #e2e8f0;
}

.openai-web-markdown :deep(code) {
  border-radius: 0.45rem;
  background: rgba(15, 23, 42, 0.06);
  padding: 0.12rem 0.35rem;
  font-size: 0.9em;
}

.openai-web-markdown :deep(pre code) {
  background: transparent;
  padding: 0;
  color: inherit;
}

.openai-web-markdown :deep(ul),
.openai-web-markdown :deep(ol) {
  margin: 0 0 0.9rem 1.2rem;
}

.openai-web-markdown :deep(blockquote) {
  margin: 0 0 0.9rem;
  border-left: 3px solid rgba(14, 165, 233, 0.35);
  padding-left: 0.9rem;
  color: #475569;
}
</style>
