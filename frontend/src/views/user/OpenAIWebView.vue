<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
        <div class="grid gap-6 border-b border-slate-200 bg-gradient-to-br from-slate-950 via-slate-900 to-cyan-950 px-6 py-6 text-white lg:grid-cols-[1.35fr_0.95fr]">
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
          <div class="grid gap-3 sm:grid-cols-3 lg:grid-cols-1">
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
              <p class="mt-2 text-3xl font-semibold">{{ activeThreadCount }}</p>
            </div>
            <div class="rounded-2xl border border-white/10 bg-white/8 p-4 backdrop-blur">
              <p class="text-xs uppercase tracking-[0.2em] text-slate-300">
                {{ t('openAIWeb.proReadyGroups') }}
              </p>
              <p class="mt-2 text-3xl font-semibold">{{ proReadyGroupCount }}</p>
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
                <Icon name="chatBubble" size="md" />
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

      <template v-else-if="!hasWorkspace">
        <section class="rounded-3xl border border-dashed border-slate-300 bg-white p-10 text-center shadow-sm">
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
      </template>

      <section v-else class="grid gap-6 xl:grid-cols-[0.92fr_1.08fr]">
        <div class="space-y-6">
          <section
            v-if="entitlements.length > 0"
            class="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm"
          >
            <div class="flex items-start justify-between gap-3">
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
                    {{ accessModeLabel(selectedEntitlement.access_mode) }}
                  </span>
                  <span class="rounded-full bg-slate-900 px-3 py-1 text-xs font-medium text-white">
                    {{ t('openAIWeb.defaultModelTag', { model: selectedEntitlement.default_model }) }}
                  </span>
                  <span
                    v-if="selectedEntitlement.has_pro_accounts"
                    class="rounded-full bg-amber-100 px-3 py-1 text-xs font-medium text-amber-700"
                  >
                    {{ t('openAIWeb.proAccountsTag') }}
                  </span>
                  <span
                    v-if="selectedEntitlement.subscription_end"
                    class="rounded-full bg-emerald-100 px-3 py-1 text-xs font-medium text-emerald-700"
                  >
                    {{ t('openAIWeb.subscriptionEndTag', { date: formatDateTime(selectedEntitlement.subscription_end) }) }}
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

              <div class="grid gap-4 lg:grid-cols-2">
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

                <div>
                  <label class="mb-2 block text-sm font-medium text-slate-700">
                    {{ t('openAIWeb.cachePolicyLabel') }}
                  </label>
                  <select
                    v-model="createForm.cache_policy"
                    class="w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-cyan-400 focus:bg-white focus:ring-4 focus:ring-cyan-100"
                  >
                    <option value="local_only">{{ cachePolicyLabel('local_only') }}</option>
                    <option value="local_encrypted">{{ cachePolicyLabel('local_encrypted') }}</option>
                  </select>
                </div>
              </div>

              <div class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-xs leading-6 text-slate-500">
                {{ t('openAIWeb.cachePolicyHint') }}
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
          </section>

          <section
            v-else
            class="rounded-3xl border border-amber-200 bg-amber-50/70 p-6 shadow-sm"
          >
            <div class="flex items-start gap-3">
              <div class="rounded-2xl bg-amber-100 p-2 text-amber-700">
                <Icon name="exclamationTriangle" size="md" />
              </div>
              <div class="space-y-1">
                <h2 class="text-lg font-semibold text-slate-900">
                  {{ t('openAIWeb.createUnavailableTitle') }}
                </h2>
                <p class="text-sm leading-6 text-slate-600">
                  {{ t('openAIWeb.createUnavailableDesc') }}
                </p>
              </div>
            </div>
          </section>

          <section class="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
            <div class="flex flex-col gap-4 border-b border-slate-100 pb-5">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <h2 class="text-lg font-semibold text-slate-900">
                    {{ t('openAIWeb.threadListTitle') }}
                  </h2>
                  <p class="mt-1 text-sm text-slate-500">
                    {{ t('openAIWeb.threadListDesc') }}
                  </p>
                </div>
                <div class="rounded-2xl bg-slate-100 px-3 py-2 text-sm font-medium text-slate-700">
                  {{ filteredThreads.length }} / {{ threads.length }}
                </div>
              </div>

              <div class="grid gap-3 lg:grid-cols-[1fr_auto]">
                <SearchInput
                  v-model="threadSearch"
                  :placeholder="t('openAIWeb.searchPlaceholder')"
                />
                <div class="inline-flex rounded-2xl border border-slate-200 bg-slate-50 p-1">
                  <button
                    v-for="option in filterOptions"
                    :key="option.value"
                    type="button"
                    class="rounded-xl px-3 py-2 text-sm font-medium transition"
                    :class="threadFilter === option.value ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-700'"
                    @click="threadFilter = option.value"
                  >
                    {{ option.label }}
                  </button>
                </div>
              </div>
            </div>

            <div v-if="filteredThreads.length === 0" class="mt-6 rounded-2xl border border-dashed border-slate-200 bg-slate-50 p-8 text-center">
              <div class="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-white text-slate-400 shadow-sm">
                <Icon :name="threads.length === 0 ? 'inbox' : 'search'" size="xl" />
              </div>
              <h3 class="mt-4 text-base font-semibold text-slate-900">
                {{ threads.length === 0 ? t('openAIWeb.emptyTitle') : t('openAIWeb.emptySearchTitle') }}
              </h3>
              <p class="mt-2 text-sm leading-6 text-slate-500">
                {{ threads.length === 0 ? t('openAIWeb.emptyDesc') : t('openAIWeb.emptySearchDesc') }}
              </p>
            </div>

            <div v-else class="mt-6 space-y-3">
              <button
                v-for="thread in filteredThreads"
                :key="thread.local_thread_id"
                type="button"
                class="w-full rounded-2xl border p-4 text-left transition"
                :class="selectedThreadId === thread.local_thread_id ? 'border-slate-900 bg-slate-950 text-white shadow-lg shadow-slate-900/10' : 'border-slate-200 bg-slate-50 hover:border-slate-300 hover:bg-white'"
                @click="selectedThreadId = thread.local_thread_id"
              >
                <div class="flex items-start justify-between gap-3">
                  <div class="min-w-0 space-y-3">
                    <div class="flex flex-wrap items-center gap-2">
                      <h3 class="truncate text-sm font-semibold">
                        {{ thread.title || t('openAIWeb.untitledThread') }}
                      </h3>
                      <span
                        class="rounded-full px-2.5 py-1 text-[11px] font-medium"
                        :class="selectedThreadId === thread.local_thread_id ? 'bg-white/10 text-slate-100' : 'bg-slate-200 text-slate-700'"
                      >
                        {{ thread.requested_model || t('openAIWeb.modelAutoResolved') }}
                      </span>
                    </div>

                    <div class="grid gap-2 text-xs sm:grid-cols-2" :class="selectedThreadId === thread.local_thread_id ? 'text-slate-300' : 'text-slate-500'">
                      <p>{{ thread.group?.name || `#${thread.group_id}` }}</p>
                      <p>{{ formatRelativeWithDateTime(thread.updated_at) }}</p>
                    </div>

                    <div class="flex flex-wrap gap-2 text-[11px]">
                      <span
                        class="rounded-full px-2.5 py-1 font-medium"
                        :class="selectedThreadId === thread.local_thread_id ? 'bg-emerald-500/20 text-emerald-100' : statusClass(thread.status)"
                      >
                        {{ statusLabel(thread.status) }}
                      </span>
                      <span
                        class="rounded-full px-2.5 py-1 font-medium"
                        :class="selectedThreadId === thread.local_thread_id ? 'bg-cyan-500/20 text-cyan-100' : syncStatusClass(thread.sync_status)"
                      >
                        {{ syncStatusLabel(thread.sync_status) }}
                      </span>
                      <span
                        v-if="threadNeedsAttention(thread)"
                        class="rounded-full bg-rose-500/15 px-2.5 py-1 font-medium text-rose-200"
                      >
                        {{ t('openAIWeb.threadNeedsAttention') }}
                      </span>
                    </div>
                  </div>

                  <button
                    type="button"
                    class="inline-flex items-center gap-2 rounded-xl border px-3 py-2 text-xs font-medium transition disabled:cursor-not-allowed disabled:opacity-60"
                    :class="selectedThreadId === thread.local_thread_id ? 'border-white/15 bg-white/10 text-white hover:bg-white/15' : 'border-rose-200 bg-white text-rose-600 hover:border-rose-300 hover:bg-rose-50'"
                    :disabled="archivingIds.has(thread.local_thread_id)"
                    @click.stop="handleArchiveThread(thread)"
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
              </button>
            </div>
          </section>
        </div>

        <div class="space-y-6">
          <section
            v-if="selectedThread"
            class="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm"
          >
            <div class="border-b border-slate-200 bg-gradient-to-br from-slate-950 via-slate-900 to-slate-800 px-6 py-6 text-white">
              <div class="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
                <div class="space-y-3">
                  <div class="flex flex-wrap items-center gap-2">
                    <span class="rounded-full bg-white/10 px-3 py-1 text-xs font-medium uppercase tracking-[0.2em] text-slate-200">
                      {{ t('openAIWeb.selectedThreadBadge') }}
                    </span>
                    <span class="rounded-full bg-cyan-500/15 px-3 py-1 text-xs font-medium text-cyan-100">
                      {{ selectedThread.requested_model || t('openAIWeb.modelAutoResolved') }}
                    </span>
                    <span class="rounded-full bg-emerald-500/15 px-3 py-1 text-xs font-medium text-emerald-100">
                      {{ capabilityLabel(selectedThread.capability_mode) }}
                    </span>
                  </div>

                  <div>
                    <h2 class="text-2xl font-semibold tracking-tight">
                      {{ selectedThread.title || t('openAIWeb.untitledThread') }}
                    </h2>
                    <p class="mt-2 text-sm leading-6 text-slate-300">
                      {{ t('openAIWeb.selectedThreadDesc') }}
                    </p>
                  </div>

                  <div class="flex flex-wrap gap-2">
                    <span class="rounded-full bg-white/10 px-3 py-1 text-xs font-medium text-slate-100">
                      {{ statusLabel(selectedThread.status) }}
                    </span>
                    <span class="rounded-full bg-white/10 px-3 py-1 text-xs font-medium text-slate-100">
                      {{ syncStatusLabel(selectedThread.sync_status) }}
                    </span>
                    <span
                      v-if="selectedThreadEntitlement"
                      class="rounded-full bg-white/10 px-3 py-1 text-xs font-medium text-slate-100"
                    >
                      {{ accessModeLabel(selectedThreadEntitlement.access_mode) }}
                    </span>
                    <span
                      v-if="selectedThread.account?.plan_type"
                      class="rounded-full bg-amber-500/15 px-3 py-1 text-xs font-medium text-amber-100"
                    >
                      {{ selectedThread.account.plan_type }}
                    </span>
                  </div>
                </div>

                <div class="flex flex-wrap gap-2">
                  <button
                    type="button"
                    class="inline-flex items-center gap-2 rounded-xl border border-white/15 bg-white/10 px-3 py-2 text-sm font-medium text-white transition hover:bg-white/15"
                    @click="copyThreadValue(selectedThread.local_thread_id)"
                  >
                    <Icon name="copy" size="sm" />
                    <span>{{ t('openAIWeb.copyLocalId') }}</span>
                  </button>
                  <button
                    v-if="selectedThread.page_session_id"
                    type="button"
                    class="inline-flex items-center gap-2 rounded-xl border border-white/15 bg-white/10 px-3 py-2 text-sm font-medium text-white transition hover:bg-white/15"
                    @click="copyThreadValue(selectedThread.page_session_id)"
                  >
                    <Icon name="copy" size="sm" />
                    <span>{{ t('openAIWeb.copySessionId') }}</span>
                  </button>
                  <button
                    v-if="selectedThread.upstream_conversation_id"
                    type="button"
                    class="inline-flex items-center gap-2 rounded-xl border border-white/15 bg-white/10 px-3 py-2 text-sm font-medium text-white transition hover:bg-white/15"
                    @click="copyThreadValue(selectedThread.upstream_conversation_id)"
                  >
                    <Icon name="copy" size="sm" />
                    <span>{{ t('openAIWeb.copyConversationId') }}</span>
                  </button>
                  <button
                    type="button"
                    class="inline-flex items-center gap-2 rounded-xl border border-rose-400/30 bg-rose-500/10 px-3 py-2 text-sm font-medium text-rose-50 transition hover:bg-rose-500/15 disabled:cursor-not-allowed disabled:opacity-60"
                    :disabled="archivingIds.has(selectedThread.local_thread_id)"
                    @click="handleArchiveThread(selectedThread)"
                  >
                    <Icon name="trash" size="sm" />
                    <span>
                      {{
                        archivingIds.has(selectedThread.local_thread_id)
                          ? t('openAIWeb.archiving')
                          : t('openAIWeb.archive')
                      }}
                    </span>
                  </button>
                </div>
              </div>
            </div>

            <div class="grid gap-6 px-6 py-6 xl:grid-cols-[0.95fr_1.05fr]">
              <div class="space-y-6">
                <section class="rounded-3xl border border-slate-200 bg-slate-50/70 p-5">
                  <div class="flex items-center gap-2">
                    <Icon name="database" size="sm" class="text-slate-500" />
                    <h3 class="text-base font-semibold text-slate-900">
                      {{ t('openAIWeb.threadMetaTitle') }}
                    </h3>
                  </div>

                  <div class="mt-5 grid gap-4 text-sm sm:grid-cols-2">
                    <div class="rounded-2xl bg-white p-4 shadow-sm">
                      <p class="text-xs uppercase tracking-[0.16em] text-slate-400">{{ t('openAIWeb.groupInfo') }}</p>
                      <p class="mt-2 font-medium text-slate-900">{{ selectedThread.group?.name || `#${selectedThread.group_id}` }}</p>
                    </div>
                    <div class="rounded-2xl bg-white p-4 shadow-sm">
                      <p class="text-xs uppercase tracking-[0.16em] text-slate-400">{{ t('openAIWeb.accountInfo') }}</p>
                      <p class="mt-2 font-medium text-slate-900">{{ selectedThread.account?.name || `#${selectedThread.account_id}` }}</p>
                    </div>
                    <div class="rounded-2xl bg-white p-4 shadow-sm">
                      <p class="text-xs uppercase tracking-[0.16em] text-slate-400">{{ t('openAIWeb.planInfo') }}</p>
                      <p class="mt-2 font-medium text-slate-900">{{ selectedThread.account?.plan_type || '-' }}</p>
                    </div>
                    <div class="rounded-2xl bg-white p-4 shadow-sm">
                      <p class="text-xs uppercase tracking-[0.16em] text-slate-400">{{ t('openAIWeb.providerInfo') }}</p>
                      <p class="mt-2 font-medium text-slate-900">{{ selectedThread.provider }}</p>
                    </div>
                    <div class="rounded-2xl bg-white p-4 shadow-sm">
                      <p class="text-xs uppercase tracking-[0.16em] text-slate-400">{{ t('openAIWeb.syncStatusInfo') }}</p>
                      <p class="mt-2 font-medium text-slate-900">{{ syncStatusLabel(selectedThread.sync_status) }}</p>
                    </div>
                    <div class="rounded-2xl bg-white p-4 shadow-sm">
                      <p class="text-xs uppercase tracking-[0.16em] text-slate-400">{{ t('openAIWeb.historyModeInfo') }}</p>
                      <p class="mt-2 font-medium text-slate-900">{{ historyModeLabel(selectedThread.history_mode) }}</p>
                    </div>
                    <div class="rounded-2xl bg-white p-4 shadow-sm">
                      <p class="text-xs uppercase tracking-[0.16em] text-slate-400">{{ t('openAIWeb.cachePolicyInfo') }}</p>
                      <p class="mt-2 font-medium text-slate-900">{{ cachePolicyLabel(selectedThread.cache_policy) }}</p>
                    </div>
                    <div class="rounded-2xl bg-white p-4 shadow-sm">
                      <p class="text-xs uppercase tracking-[0.16em] text-slate-400">{{ t('openAIWeb.updatedAt') }}</p>
                      <p class="mt-2 font-medium text-slate-900">{{ formatRelativeWithDateTime(selectedThread.updated_at) }}</p>
                    </div>
                  </div>
                </section>

                <section class="rounded-3xl border border-slate-200 bg-white p-5">
                  <div class="flex items-center gap-2">
                    <Icon name="link" size="sm" class="text-slate-500" />
                    <h3 class="text-base font-semibold text-slate-900">
                      {{ t('openAIWeb.sessionIdentifiersTitle') }}
                    </h3>
                  </div>

                  <div class="mt-5 space-y-3">
                    <div class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3">
                      <p class="text-xs uppercase tracking-[0.16em] text-slate-400">{{ t('openAIWeb.localThreadId') }}</p>
                      <p class="mt-2 break-all font-mono text-sm text-slate-900">{{ selectedThread.local_thread_id }}</p>
                    </div>
                    <div class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3">
                      <p class="text-xs uppercase tracking-[0.16em] text-slate-400">{{ t('openAIWeb.pageSessionId') }}</p>
                      <p class="mt-2 break-all font-mono text-sm text-slate-900">{{ selectedThread.page_session_id }}</p>
                    </div>
                    <div class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3">
                      <p class="text-xs uppercase tracking-[0.16em] text-slate-400">{{ t('openAIWeb.upstreamConversationId') }}</p>
                      <p class="mt-2 break-all font-mono text-sm text-slate-900">
                        {{ selectedThread.upstream_conversation_id || t('openAIWeb.pendingBinding') }}
                      </p>
                    </div>
                    <div class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3">
                      <p class="text-xs uppercase tracking-[0.16em] text-slate-400">{{ t('openAIWeb.upstreamSessionId') }}</p>
                      <p class="mt-2 break-all font-mono text-sm text-slate-900">
                        {{ selectedThread.upstream_session_id || t('openAIWeb.pendingBinding') }}
                      </p>
                    </div>
                  </div>
                </section>
              </div>

              <div class="space-y-6">
                <section class="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
                  <div class="flex items-center justify-between gap-3 border-b border-slate-200 bg-slate-50 px-5 py-4">
                    <div>
                      <div class="flex items-center gap-2">
                        <Icon name="chat" size="sm" class="text-slate-500" />
                        <h3 class="text-base font-semibold text-slate-900">
                          {{ t('openAIWeb.chatTitle') }}
                        </h3>
                      </div>
                      <p class="mt-1 text-sm text-slate-500">
                        {{ t('openAIWeb.chatDesc') }}
                      </p>
                    </div>
                    <div class="rounded-2xl bg-white px-3 py-2 text-xs font-medium text-slate-600 shadow-sm">
                      {{ t('openAIWeb.messageCount', { count: selectedMessages.length }) }}
                    </div>
                  </div>

                  <div
                    ref="messageViewport"
                    class="max-h-[620px] overflow-y-auto bg-[radial-gradient(circle_at_top,_rgba(14,165,233,0.08),_transparent_42%),linear-gradient(180deg,_#f8fafc_0%,_#ffffff_32%)] px-5 py-5"
                  >
                    <div v-if="selectedMessages.length === 0" class="rounded-3xl border border-dashed border-slate-200 bg-white/90 p-8 text-center shadow-sm">
                      <div class="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-slate-100 text-slate-500">
                        <Icon name="sparkles" size="xl" />
                      </div>
                      <h4 class="mt-4 text-base font-semibold text-slate-900">
                        {{ t('openAIWeb.emptyChatTitle') }}
                      </h4>
                      <p class="mx-auto mt-2 max-w-xl text-sm leading-6 text-slate-500">
                        {{ t('openAIWeb.emptyChatDesc', {
                          group: selectedThread.group?.name || `#${selectedThread.group_id}`,
                          model: selectedThread.requested_model || t('openAIWeb.modelAutoResolved'),
                        }) }}
                      </p>
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
                            <span class="font-semibold uppercase tracking-[0.18em]" :class="message.role === 'user' ? 'text-slate-300' : message.status === 'error' ? 'text-rose-500' : 'text-slate-400'">
                              {{ message.role === 'user' ? t('openAIWeb.userLabel') : t('openAIWeb.assistantLabel') }}
                            </span>
                            <span :class="message.role === 'user' ? 'text-slate-400' : 'text-slate-400'">
                              {{ formatDateTime(message.created_at) }}
                            </span>
                          </div>

                          <div v-if="message.role === 'user' && message.content?.trim()" class="whitespace-pre-wrap break-words text-sm leading-7">
                            {{ message.content }}
                          </div>
                          <div v-else-if="message.status === 'pending'" class="space-y-3 text-sm leading-7 text-slate-600">
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
                              <div
                                class="space-y-3 border-t border-slate-200 px-3 py-3 text-xs text-slate-500"
                              >
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

                          <div
                            v-if="message.request_id || message.response_id || message.usage"
                            class="mt-4 flex flex-wrap gap-2 text-[11px]"
                          >
                            <span
                              v-if="message.request_id"
                              class="rounded-full bg-slate-100 px-2.5 py-1 font-medium text-slate-500"
                            >
                              {{ t('openAIWeb.requestIdTag', { id: message.request_id }) }}
                            </span>
                            <span
                              v-if="message.response_id"
                              class="rounded-full bg-cyan-50 px-2.5 py-1 font-medium text-cyan-700"
                            >
                              {{ t('openAIWeb.responseIdTag', { id: message.response_id }) }}
                            </span>
                            <span
                              v-if="message.usage"
                              class="rounded-full bg-emerald-50 px-2.5 py-1 font-medium text-emerald-700"
                            >
                              {{ formatUsageSummary(message.usage) }}
                            </span>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>

                  <div class="border-t border-slate-200 bg-white px-5 py-5">
                    <input
                      ref="composerFileInput"
                      type="file"
                      accept="image/*"
                      multiple
                      class="hidden"
                      @change="handleComposerFilesSelected"
                    >
                    <div class="mb-4 grid gap-3 lg:grid-cols-2">
                      <div>
                        <label class="mb-2 block text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                          {{ t('openAIWeb.chatModelLabel') }}
                        </label>
                        <select
                          v-model="composerRequestedModel"
                          class="w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-cyan-400 focus:bg-white focus:ring-4 focus:ring-cyan-100"
                        >
                          <option
                            v-for="option in selectedThreadModelOptions"
                            :key="option"
                            :value="option"
                          >
                            {{ option }}
                          </option>
                        </select>
                      </div>
                      <div>
                        <label class="mb-2 block text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                          {{ t('openAIWeb.reasoningEffortLabel') }}
                        </label>
                        <select
                          v-model="composerReasoningEffort"
                          class="w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-cyan-400 focus:bg-white focus:ring-4 focus:ring-cyan-100"
                        >
                          <option
                            v-for="option in reasoningEffortOptions"
                            :key="option.value || 'auto'"
                            :value="option.value"
                          >
                            {{ option.label }}
                          </option>
                        </select>
                      </div>
                    </div>
                    <div class="mb-3 flex flex-wrap gap-2 text-xs">
                      <span class="rounded-full bg-slate-100 px-3 py-1 font-medium text-slate-600">
                        {{ cachePolicyLabel(selectedThread.cache_policy) }}
                      </span>
                      <span class="rounded-full bg-slate-900 px-3 py-1 font-medium text-white">
                        {{ composerRequestedModel || selectedThread.requested_model || t('openAIWeb.modelAutoResolved') }}
                      </span>
                      <span
                        v-if="composerReasoningEffort"
                        class="rounded-full bg-violet-50 px-3 py-1 font-medium text-violet-700"
                      >
                        {{ t(`openAIWeb.reasoningEffort.${composerReasoningEffort}`) }}
                      </span>
                      <span class="rounded-full bg-cyan-50 px-3 py-1 font-medium text-cyan-700">
                        {{ t('openAIWeb.chatBillingHint') }}
                      </span>
                      <button
                        type="button"
                        class="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-white px-3 py-1 font-medium text-slate-600 transition hover:border-slate-300 hover:text-slate-900"
                        @click="openComposerFilePicker"
                      >
                        <Icon name="upload" size="sm" />
                        <span>{{ t('openAIWeb.addImages') }}</span>
                      </button>
                      <span
                        v-if="composerAttachments.length > 0"
                        class="rounded-full bg-amber-50 px-3 py-1 font-medium text-amber-700"
                      >
                        {{ t('openAIWeb.imageCountTag', { count: composerAttachments.length }) }}
                      </span>
                      <button
                        v-if="composerAttachments.length > 0"
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
                    <textarea
                      v-model="composer"
                      rows="4"
                      :placeholder="t('openAIWeb.composerPlaceholder')"
                      class="w-full resize-none rounded-3xl border border-slate-200 bg-slate-50 px-4 py-4 text-sm text-slate-900 outline-none transition focus:border-cyan-400 focus:bg-white focus:ring-4 focus:ring-cyan-100"
                      @keydown.enter.exact.prevent="handleSendMessage"
                    />
                    <div class="mt-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                      <p class="text-xs leading-6 text-slate-500">
                        {{ t('openAIWeb.composerHint') }}
                      </p>
                      <button
                        type="button"
                        class="inline-flex items-center justify-center gap-2 rounded-2xl bg-slate-950 px-4 py-3 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
                        :disabled="sending || !selectedThread || (!composer.trim() && composerAttachments.length === 0)"
                        @click="handleSendMessage"
                      >
                        <Icon :name="sending ? 'refresh' : 'arrowUp'" size="sm" />
                        <span>{{ sending ? t('openAIWeb.sending') : t('openAIWeb.send') }}</span>
                      </button>
                    </div>
                  </div>
                </section>

                <section class="rounded-3xl border border-slate-200 bg-white p-5">
                  <div class="flex items-center gap-2">
                    <Icon name="clock" size="sm" class="text-slate-500" />
                    <h3 class="text-base font-semibold text-slate-900">
                      {{ t('openAIWeb.timelineTitle') }}
                    </h3>
                  </div>

                  <div class="mt-5 space-y-4">
                    <div class="flex items-start gap-3">
                      <div class="mt-1 h-2.5 w-2.5 rounded-full bg-slate-900" />
                      <div>
                        <p class="text-sm font-medium text-slate-900">{{ t('openAIWeb.timelineCreated') }}</p>
                        <p class="mt-1 text-sm text-slate-500">{{ formatRelativeWithDateTime(selectedThread.created_at) }}</p>
                      </div>
                    </div>
                    <div class="flex items-start gap-3">
                      <div class="mt-1 h-2.5 w-2.5 rounded-full bg-cyan-500" />
                      <div>
                        <p class="text-sm font-medium text-slate-900">{{ t('openAIWeb.timelineUpdated') }}</p>
                        <p class="mt-1 text-sm text-slate-500">{{ formatRelativeWithDateTime(selectedThread.updated_at) }}</p>
                      </div>
                    </div>
                    <div class="flex items-start gap-3">
                      <div class="mt-1 h-2.5 w-2.5 rounded-full bg-emerald-500" />
                      <div>
                        <p class="text-sm font-medium text-slate-900">{{ t('openAIWeb.timelineSynced') }}</p>
                        <p class="mt-1 text-sm text-slate-500">
                          {{
                            selectedThread.last_synced_at
                              ? formatRelativeWithDateTime(selectedThread.last_synced_at)
                              : t('openAIWeb.timelineNotSynced')
                          }}
                        </p>
                      </div>
                    </div>
                    <div class="flex items-start gap-3">
                      <div class="mt-1 h-2.5 w-2.5 rounded-full bg-cyan-500" />
                      <div>
                        <p class="text-sm font-medium text-slate-900">{{ t('openAIWeb.timelineCacheMode') }}</p>
                        <p class="mt-1 text-sm text-slate-500">{{ cachePolicyLabel(selectedThread.cache_policy) }}</p>
                      </div>
                    </div>
                  </div>
                </section>
              </div>
            </div>
          </section>

          <section
            v-else
            class="rounded-3xl border border-dashed border-slate-300 bg-white p-12 text-center shadow-sm"
          >
            <div class="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-slate-100 text-slate-500">
              <Icon name="chatBubble" size="xl" />
            </div>
            <h2 class="mt-5 text-xl font-semibold text-slate-900">
              {{ t('openAIWeb.selectThreadTitle') }}
            </h2>
            <p class="mx-auto mt-3 max-w-2xl text-sm leading-6 text-slate-500">
              {{ t('openAIWeb.selectThreadDesc') }}
            </p>
          </section>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import DOMPurify from 'dompurify'
import { marked } from 'marked'
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import SearchInput from '@/components/common/SearchInput.vue'
import Icon from '@/components/icons/Icon.vue'
import openAIWebAPI from '@/api/openaiWeb'
import { useClipboard } from '@/composables/useClipboard'
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

type ThreadFilter = 'all' | 'ready' | 'attention'
type LocalMessageStatus = 'done' | 'pending' | 'error'
type LocalMessageRole = 'user' | 'assistant'
type OpenAIWebComposerReasoningEffort = '' | 'low' | 'medium' | 'high' | 'xhigh'

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
const { copyToClipboard } = useClipboard()

const loading = ref(true)
const refreshing = ref(false)
const creating = ref(false)
const sending = ref(false)
const entitlements = ref<OpenAIWebEntitlement[]>([])
const threads = ref<OpenAIWebThread[]>([])
const threadMessages = ref<Record<string, OpenAIWebLocalMessage[]>>({})
const archivingIds = ref<Set<string>>(new Set())
const threadSearch = ref('')
const threadFilter = ref<ThreadFilter>('all')
const selectedThreadId = ref(typeof route.query.thread === 'string' ? route.query.thread : '')
const composer = ref('')
const composerRequestedModel = ref('')
const composerReasoningEffort = ref<OpenAIWebComposerReasoningEffort>('')
const composerFileInput = ref<HTMLInputElement | null>(null)
const composerAttachments = ref<ComposerAttachment[]>([])
const messageViewport = ref<HTMLElement | null>(null)
const localCacheWarningState = ref<'none' | 'trimmed' | 'disabled'>('none')

const createForm = reactive({
  group_id: 0,
  requested_model: '',
  title: '',
  cache_policy: 'local_only' as 'local_only' | 'local_encrypted',
})

const filterOptions = computed(() => ([
  { value: 'all' as ThreadFilter, label: t('openAIWeb.filterAll') },
  { value: 'ready' as ThreadFilter, label: t('openAIWeb.filterReady') },
  { value: 'attention' as ThreadFilter, label: t('openAIWeb.filterAttention') },
]))

const hasWorkspace = computed(() => entitlements.value.length > 0 || threads.value.length > 0)
const activeThreadCount = computed(() => threads.value.filter((item) => item.status === 'active').length)
const proReadyGroupCount = computed(() => entitlements.value.filter((item) => item.has_pro_accounts).length)

const selectedEntitlement = computed(() => {
  return entitlements.value.find((item) => item.group_id === createForm.group_id) ?? null
})

function openAIWebModelOptionsForEntitlement(
  entitlement?: OpenAIWebEntitlement | null,
  fallbackModel?: string | null
): string[] {
  const models = new Set<string>()

  if (entitlement?.default_model) {
    models.add(entitlement.default_model)
  }
  models.add('gpt-5.4-mini')
  models.add('gpt-5.4')

  const fallback = fallbackModel?.trim() || ''
  if (entitlement?.has_pro_accounts || fallback === 'gpt-5.4-pro') {
    models.add('gpt-5.4-pro')
  }
  if (fallback) {
    models.add(fallback)
  }

  return Array.from(models)
}

const selectedModelOptions = computed(() => {
  return openAIWebModelOptionsForEntitlement(selectedEntitlement.value)
})

const sortedThreads = computed(() => {
  return [...threads.value].sort((left, right) => {
    return new Date(right.updated_at).getTime() - new Date(left.updated_at).getTime()
  })
})

const filteredThreads = computed(() => {
  const keyword = threadSearch.value.trim().toLowerCase()

  return sortedThreads.value.filter((thread) => {
    if (threadFilter.value === 'ready' && threadNeedsAttention(thread)) {
      return false
    }
    if (threadFilter.value === 'attention' && !threadNeedsAttention(thread)) {
      return false
    }

    if (!keyword) {
      return true
    }

    const haystacks = [
      thread.title,
      thread.group?.name,
      thread.account?.name,
      thread.requested_model,
      thread.local_thread_id,
      thread.upstream_conversation_id,
    ]

    return haystacks.some((value) => value?.toLowerCase().includes(keyword))
  })
})

const selectedThread = computed(() => {
  return filteredThreads.value.find((item) => item.local_thread_id === selectedThreadId.value)
    ?? sortedThreads.value.find((item) => item.local_thread_id === selectedThreadId.value)
    ?? filteredThreads.value[0]
    ?? null
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

watch(
  () => filteredThreads.value,
  (items) => {
    if (items.length === 0) {
      selectedThreadId.value = ''
      return
    }
    if (!items.some((item) => item.local_thread_id === selectedThreadId.value)) {
      selectedThreadId.value = items[0].local_thread_id
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
    if (threadID) {
      ensureThreadMessagesLoaded(threadID)
      applyComposerSettingsForThread(selectedThread.value)
    } else {
      composerRequestedModel.value = ''
      composerReasoningEffort.value = ''
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
    if (!threadID) {
      return
    }
    persistComposerSettings(threadID, {
      requested_model: composerRequestedModel.value,
      reasoning_effort: composerReasoningEffort.value,
    })
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
      cache_policy: createForm.cache_policy,
    })

    threads.value = [created, ...threads.value.filter((item) => item.local_thread_id !== created.local_thread_id)]
    createForm.title = ''
    createForm.requested_model = ''
    selectedThreadId.value = created.local_thread_id
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

function normalizeComposerReasoningEffort(raw: string | null | undefined): OpenAIWebComposerReasoningEffort {
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

function resolveComposerRequestedModel(thread: OpenAIWebThread | null): string {
  const explicit = composerRequestedModel.value.trim()
  if (explicit) {
    return explicit
  }
  if (thread?.requested_model?.trim()) {
    return thread.requested_model.trim()
  }
  if (selectedThreadEntitlement.value?.default_model?.trim()) {
    return selectedThreadEntitlement.value.default_model.trim()
  }
  return selectedThreadModelOptions.value[0] || 'gpt-5.4-mini'
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
    || 'gpt-5.4-mini'
  const requestedModel = saved.requested_model?.trim() || defaultModel
  const availableModels = openAIWebModelOptionsForEntitlement(selectedThreadEntitlement.value, requestedModel)

  composerRequestedModel.value = availableModels.includes(requestedModel) ? requestedModel : defaultModel
  composerReasoningEffort.value = normalizeComposerReasoningEffort(saved.reasoning_effort)
}

async function handleSendMessage() {
  const thread = selectedThread.value
  const content = composer.value.trim()
  const requestedModel = resolveComposerRequestedModel(thread)
  const reasoningEffort = normalizeComposerReasoningEffort(composerReasoningEffort.value)
  const rawAttachments = [...composerAttachments.value]
  const attachments = rawAttachments.map(composerAttachmentToRequest)
  if (!thread || sending.value || (!content && attachments.length === 0)) {
    return
  }

  sending.value = true
  composer.value = ''
  composerAttachments.value = []

  const userMessage = buildLocalMessage('user', content, 'done', rawAttachments.map(composerAttachmentToMessageImage))
  const pendingAssistant = buildLocalMessage('assistant', '', 'pending')
  const draftMessages = [...selectedMessages.value, userMessage, pendingAssistant]
  persistThreadMessages(thread.local_thread_id, draftMessages)

  try {
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
    composerRequestedModel.value = response.thread.requested_model || requestedModel

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
    persistThreadMessages(
      thread.local_thread_id,
      draftMessages.map((message) => {
        if (message.id !== pendingAssistant.id) {
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

async function copyThreadValue(value: string | null | undefined) {
  if (!value) {
    return
  }
  await copyToClipboard(value)
}

function threadNeedsAttention(thread: OpenAIWebThread): boolean {
  return thread.status === 'broken' || thread.sync_status === 'error' || !!thread.last_error
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

function syncStatusClass(status: string): string {
  switch (status) {
    case 'ready':
      return 'bg-emerald-100 text-emerald-700'
    case 'pending':
      return 'bg-amber-100 text-amber-700'
    default:
      return 'bg-rose-100 text-rose-700'
  }
}

function statusLabel(status: string): string {
  return t(`openAIWeb.status.${status}`)
}

function syncStatusLabel(status: string): string {
  return t(`openAIWeb.syncStatus.${status}`)
}

function capabilityLabel(mode: string): string {
  return t(`openAIWeb.capability.${mode}`)
}

function cachePolicyLabel(policy: string): string {
  return t(`openAIWeb.cachePolicy.${policy}`)
}

function historyModeLabel(mode: string): string {
  return t(`openAIWeb.historyMode.${mode}`)
}

function accessModeLabel(mode: string): string {
  return t(`openAIWeb.accessMode.${mode || 'standard'}`)
}

function renderMessageHtml(content: string): string {
  const html = marked.parse(content || '') as string
  return DOMPurify.sanitize(html)
}

function formatUsageSummary(usage: OpenAIWebThreadMessageUsage): string {
  return t('openAIWeb.usageTag', {
    input: usage.input_tokens,
    output: usage.output_tokens,
    total: usage.total_tokens,
  })
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
  return `${(bytes / (1024 * 1024)).toFixed(bytes >= 10*1024*1024 ? 0 : 1)} MB`
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
