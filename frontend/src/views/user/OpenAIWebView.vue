<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
        <div class="grid gap-6 border-b border-slate-200 bg-gradient-to-br from-slate-950 via-slate-900 to-cyan-950 px-6 py-6 text-white lg:grid-cols-[1.4fr_0.8fr]">
          <div class="space-y-3">
            <div class="inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/10 px-3 py-1 text-xs font-medium uppercase tracking-[0.24em] text-cyan-100">
              <Icon name="globe" size="sm" />
              <span>{{ t('openAIWeb.summaryBadge') }}</span>
            </div>
            <div class="space-y-2">
              <h1 class="text-2xl font-semibold tracking-tight">
                {{ t('openAIWeb.title') }}
              </h1>
              <p class="max-w-2xl text-sm leading-6 text-slate-200">
                {{ t('openAIWeb.description') }}
              </p>
            </div>
          </div>
          <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-1">
            <div class="rounded-2xl border border-white/10 bg-white/8 p-4 backdrop-blur">
              <p class="text-xs uppercase tracking-[0.2em] text-slate-300">
                {{ t('openAIWeb.eligibleGroups') }}
              </p>
              <p class="mt-2 text-3xl font-semibold">{{ entitlements.length }}</p>
            </div>
            <div class="rounded-2xl border border-white/10 bg-white/8 p-4 backdrop-blur">
              <p class="text-xs uppercase tracking-[0.2em] text-slate-300">
                {{ t('openAIWeb.activeThreads') }}
              </p>
              <p class="mt-2 text-3xl font-semibold">{{ threads.length }}</p>
            </div>
          </div>
        </div>

        <div class="grid gap-4 px-6 py-5 lg:grid-cols-2">
          <div class="rounded-2xl border border-cyan-100 bg-cyan-50/80 p-4">
            <div class="flex items-start gap-3">
              <div class="rounded-xl bg-cyan-100 p-2 text-cyan-700">
                <Icon name="shield" size="md" />
              </div>
              <div>
                <p class="font-medium text-slate-900">{{ t('openAIWeb.privacyTitle') }}</p>
                <p class="mt-1 text-sm leading-6 text-slate-600">
                  {{ t('openAIWeb.privacyDesc') }}
                </p>
              </div>
            </div>
          </div>

          <div class="rounded-2xl border border-amber-100 bg-amber-50/80 p-4">
            <div class="flex items-start gap-3">
              <div class="rounded-xl bg-amber-100 p-2 text-amber-700">
                <Icon name="lightbulb" size="md" />
              </div>
              <div>
                <p class="font-medium text-slate-900">{{ t('openAIWeb.scopeTitle') }}</p>
                <p class="mt-1 text-sm leading-6 text-slate-600">
                  {{ t('openAIWeb.scopeDesc') }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <div v-if="loading" class="flex justify-center py-16">
        <div class="h-10 w-10 animate-spin rounded-full border-2 border-cyan-500 border-t-transparent" />
      </div>

      <template v-else>
        <section v-if="entitlements.length === 0" class="rounded-3xl border border-dashed border-slate-300 bg-white p-10 text-center shadow-sm">
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

        <template v-else>
          <section class="grid gap-6 xl:grid-cols-[0.95fr_1.05fr]">
            <div class="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
              <div class="flex items-center justify-between">
                <div>
                  <h2 class="text-lg font-semibold text-slate-900">
                    {{ t('openAIWeb.createTitle') }}
                  </h2>
                  <p class="mt-1 text-sm text-slate-500">
                    {{ t('openAIWeb.createDesc') }}
                  </p>
                </div>
                <button
                  type="button"
                  class="inline-flex items-center gap-2 rounded-xl border border-slate-200 px-3 py-2 text-sm font-medium text-slate-600 transition hover:border-slate-300 hover:text-slate-900"
                  :disabled="refreshing"
                  @click="loadAll"
                >
                  <Icon name="refresh" size="sm" />
                  <span>{{ t('openAIWeb.refresh') }}</span>
                </button>
              </div>

              <form class="mt-6 space-y-4" @submit.prevent="handleCreateThread">
                <div>
                  <label class="mb-2 block text-sm font-medium text-slate-700">
                    {{ t('openAIWeb.groupLabel') }}
                  </label>
                  <select
                    v-model.number="createForm.group_id"
                    class="w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-cyan-400 focus:bg-white focus:ring-4 focus:ring-cyan-100"
                  >
                    <option
                      v-for="entitlement in entitlements"
                      :key="entitlement.group_id"
                      :value="entitlement.group_id"
                    >
                      {{ entitlement.group_name }}
                    </option>
                  </select>
                  <div v-if="selectedEntitlement" class="mt-3 flex flex-wrap gap-2">
                    <span class="rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-600">
                      {{ selectedEntitlement.group_desc || t('openAIWeb.noGroupDesc') }}
                    </span>
                    <span class="rounded-full bg-cyan-100 px-3 py-1 text-xs font-medium text-cyan-700">
                      {{ t('openAIWeb.defaultModelTag', { model: selectedEntitlement.default_model }) }}
                    </span>
                    <span
                      v-if="selectedEntitlement.has_pro_accounts"
                      class="rounded-full bg-amber-100 px-3 py-1 text-xs font-medium text-amber-700"
                    >
                      {{ t('openAIWeb.proAccountsTag') }}
                    </span>
                  </div>
                </div>

                <div>
                  <label class="mb-2 block text-sm font-medium text-slate-700">
                    {{ t('openAIWeb.titleLabel') }}
                  </label>
                  <input
                    v-model.trim="createForm.title"
                    type="text"
                    :placeholder="t('openAIWeb.titlePlaceholder')"
                    class="w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-cyan-400 focus:bg-white focus:ring-4 focus:ring-cyan-100"
                  >
                </div>

                <div>
                  <label class="mb-2 block text-sm font-medium text-slate-700">
                    {{ t('openAIWeb.modelLabel') }}
                  </label>
                  <select
                    v-model="createForm.requested_model"
                    class="w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-cyan-400 focus:bg-white focus:ring-4 focus:ring-cyan-100"
                  >
                    <option value="">{{ t('openAIWeb.modelAuto') }}</option>
                    <option
                      v-for="option in selectedModelOptions"
                      :key="option"
                      :value="option"
                    >
                      {{ option }}
                    </option>
                  </select>
                </div>

                <button
                  type="submit"
                  class="inline-flex items-center gap-2 rounded-2xl bg-slate-950 px-4 py-3 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
                  :disabled="creating"
                >
                  <Icon :name="creating ? 'refresh' : 'plus'" size="sm" />
                  <span>
                    {{ creating ? t('openAIWeb.creating') : t('openAIWeb.createButton') }}
                  </span>
                </button>
              </form>
            </div>

            <div class="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
              <div class="flex items-center justify-between">
                <div>
                  <h2 class="text-lg font-semibold text-slate-900">
                    {{ t('openAIWeb.threadListTitle') }}
                  </h2>
                  <p class="mt-1 text-sm text-slate-500">
                    {{ t('openAIWeb.threadListDesc') }}
                  </p>
                </div>
              </div>

              <div v-if="threads.length === 0" class="mt-6 rounded-2xl border border-dashed border-slate-200 bg-slate-50 p-8 text-center">
                <div class="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-white text-slate-400 shadow-sm">
                  <Icon name="inbox" size="xl" />
                </div>
                <h3 class="mt-4 text-base font-semibold text-slate-900">
                  {{ t('openAIWeb.emptyTitle') }}
                </h3>
                <p class="mt-2 text-sm leading-6 text-slate-500">
                  {{ t('openAIWeb.emptyDesc') }}
                </p>
              </div>

              <div v-else class="mt-6 space-y-4">
                <article
                  v-for="thread in threads"
                  :key="thread.local_thread_id"
                  class="rounded-2xl border border-slate-200 bg-slate-50 p-4 transition hover:border-slate-300 hover:bg-white"
                >
                  <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                    <div class="min-w-0 space-y-3">
                      <div class="flex flex-wrap items-center gap-2">
                        <h3 class="text-base font-semibold text-slate-900">
                          {{ thread.title || t('openAIWeb.untitledThread') }}
                        </h3>
                        <span class="rounded-full bg-slate-200 px-2.5 py-1 text-xs font-medium text-slate-700">
                          {{ thread.requested_model || t('openAIWeb.modelAutoResolved') }}
                        </span>
                        <span
                          class="rounded-full px-2.5 py-1 text-xs font-medium"
                          :class="statusClass(thread.status)"
                        >
                          {{ statusLabel(thread.status) }}
                        </span>
                      </div>

                      <div class="grid gap-2 text-sm text-slate-500 sm:grid-cols-2">
                        <p>
                          <span class="font-medium text-slate-700">{{ t('openAIWeb.groupInfo') }}:</span>
                          {{ thread.group?.name || `#${thread.group_id}` }}
                        </p>
                        <p>
                          <span class="font-medium text-slate-700">{{ t('openAIWeb.accountInfo') }}:</span>
                          {{ thread.account?.name || `#${thread.account_id}` }}
                        </p>
                        <p>
                          <span class="font-medium text-slate-700">{{ t('openAIWeb.capabilityInfo') }}:</span>
                          {{ capabilityLabel(thread.capability_mode) }}
                        </p>
                        <p>
                          <span class="font-medium text-slate-700">{{ t('openAIWeb.updatedAt') }}:</span>
                          {{ formatDateTime(thread.updated_at) }}
                        </p>
                      </div>
                    </div>

                    <div class="flex shrink-0 items-center gap-2">
                      <button
                        type="button"
                        class="inline-flex items-center gap-2 rounded-xl border border-rose-200 bg-white px-3 py-2 text-sm font-medium text-rose-600 transition hover:border-rose-300 hover:bg-rose-50 disabled:cursor-not-allowed disabled:opacity-60"
                        :disabled="archivingIds.has(thread.local_thread_id)"
                        @click="handleArchiveThread(thread)"
                      >
                        <Icon name="trash" size="sm" />
                        <span>
                          {{
                            archivingIds.has(thread.local_thread_id)
                              ? t('openAIWeb.archiving')
                              : t('openAIWeb.archive')
                          }}
                        </span>
                      </button>
                    </div>
                  </div>
                </article>
              </div>
            </div>
          </section>
        </template>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import openAIWebAPI from '@/api/openaiWeb'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDateTime } from '@/utils/format'
import type { OpenAIWebEntitlement, OpenAIWebThread } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const refreshing = ref(false)
const creating = ref(false)
const entitlements = ref<OpenAIWebEntitlement[]>([])
const threads = ref<OpenAIWebThread[]>([])
const archivingIds = ref<Set<string>>(new Set())

const createForm = reactive({
  group_id: 0,
  requested_model: '',
  title: '',
})

const selectedEntitlement = computed(() => {
  return entitlements.value.find((item) => item.group_id === createForm.group_id) ?? null
})

const selectedModelOptions = computed(() => {
  const entitlement = selectedEntitlement.value
  const models = new Set<string>()

  if (entitlement?.default_model) {
    models.add(entitlement.default_model)
  }
  models.add('gpt-5.4-mini')
  models.add('gpt-5.4')
  if (entitlement?.has_pro_accounts) {
    models.add('gpt-5.4-pro')
  }

  return Array.from(models)
})

watch(
  () => entitlements.value,
  (items) => {
    if (items.length > 0 && !items.some((item) => item.group_id === createForm.group_id)) {
      createForm.group_id = items[0].group_id
      createForm.requested_model = ''
    }
  },
  { immediate: true }
)

watch(
  () => createForm.group_id,
  () => {
    createForm.requested_model = ''
  }
)

async function loadAll() {
  refreshing.value = true
  try {
    const [loadedEntitlements, loadedThreads] = await Promise.all([
      openAIWebAPI.getOpenAIWebEntitlements(),
      openAIWebAPI.getOpenAIWebThreads(),
    ])

    entitlements.value = loadedEntitlements
    threads.value = loadedThreads
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('openAIWeb.failedToLoad')))
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

async function handleCreateThread() {
  if (!createForm.group_id) {
    return
  }

  creating.value = true
  try {
    const created = await openAIWebAPI.createOpenAIWebThread({
      group_id: createForm.group_id,
      requested_model: createForm.requested_model || undefined,
      title: createForm.title || undefined,
    })

    threads.value = [created, ...threads.value.filter((item) => item.local_thread_id !== created.local_thread_id)]
    createForm.title = ''
    createForm.requested_model = ''
    appStore.showSuccess(t('openAIWeb.createSuccess'))
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('openAIWeb.createFailed')))
  } finally {
    creating.value = false
  }
}

async function handleArchiveThread(thread: OpenAIWebThread) {
  const next = new Set(archivingIds.value)
  next.add(thread.local_thread_id)
  archivingIds.value = next

  try {
    await openAIWebAPI.archiveOpenAIWebThread(thread.local_thread_id)
    threads.value = threads.value.filter((item) => item.local_thread_id !== thread.local_thread_id)
    appStore.showSuccess(t('openAIWeb.archiveSuccess'))
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('openAIWeb.archiveFailed')))
  } finally {
    const updated = new Set(archivingIds.value)
    updated.delete(thread.local_thread_id)
    archivingIds.value = updated
  }
}

function statusClass(status: string): string {
  switch (status) {
    case 'active':
      return 'bg-emerald-100 text-emerald-700'
    case 'archived':
      return 'bg-slate-200 text-slate-700'
    default:
      return 'bg-rose-100 text-rose-700'
  }
}

function statusLabel(status: string): string {
  return t(`openAIWeb.status.${status}`)
}

function capabilityLabel(mode: string): string {
  return t(`openAIWeb.capability.${mode}`)
}

onMounted(() => {
  void loadAll()
})
</script>
