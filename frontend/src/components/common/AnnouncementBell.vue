<template>
  <div>
    <!-- 铃铛按钮 -->
    <button
      @click="openModal"
      class="relative flex h-9 w-9 items-center justify-center text-brutal-black transition-all border-2 border-transparent hover:border-brutal-black hover:bg-brutal-yellow dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:border-dark-400"
      :class="{ 'text-blue-600 dark:text-blue-400': unreadCount > 0 }"
      :aria-label="t('announcements.title')"
    >
      <Icon name="bell" size="md" />
      <!-- 未读红点 -->
      <span
        v-if="unreadCount > 0"
        class="absolute right-1 top-1 flex h-2 w-2"
      >
        <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-red-500 opacity-75"></span>
        <span class="relative inline-flex h-2 w-2 rounded-full bg-red-500"></span>
      </span>
    </button>

    <!-- 公告列表 Modal -->
    <Teleport to="body">
      <Transition name="modal-fade">
        <div
          v-if="isModalOpen"
          class="fixed inset-0 z-[100] flex items-start justify-center overflow-y-auto bg-brutal-black/60 p-4 pt-[8vh]"
          @click="closeModal"
        >
          <div
            class="w-full max-w-[620px] overflow-hidden bg-white border-3 border-brutal-black dark:bg-dark-800 dark:border-dark-500"
            style="box-shadow: 8px 8px 0px #1A1A1A;"
            @click.stop
          >
            <!-- Header with Gradient -->
            <div class="relative overflow-hidden border-b-2 border-brutal-black bg-brutal-yellow px-6 py-5 dark:bg-dark-700 dark:border-dark-500">
              <div class="relative z-10 flex items-start justify-between">
                <div>
                  <div class="flex items-center gap-2">
                    <div class="flex h-8 w-8 items-center justify-center bg-brutal-orange text-white border-2 border-brutal-black" style="box-shadow: 2px 2px 0px #1A1A1A;">
                      <Icon name="bell" size="sm" />
                    </div>
                    <h2 class="text-lg font-bold uppercase tracking-wide text-brutal-black dark:text-white">
                      {{ t('announcements.title') }}
                    </h2>
                  </div>
                  <p v-if="unreadCount > 0" class="mt-2 text-sm text-gray-600 dark:text-gray-400">
                    <span class="font-medium text-blue-600 dark:text-blue-400">{{ unreadCount }}</span>
                    {{ t('announcements.unread') }}
                  </p>
                </div>
                <div class="flex items-center gap-2">
                  <button
                    v-if="unreadCount > 0"
                    @click="markAllAsRead"
                    :disabled="loading"
                    class="btn btn-sm bg-brutal-orange text-white border-2 border-brutal-black hover:bg-primary-600 disabled:opacity-50"
                  >
                    {{ t('announcements.markAllRead') }}
                  </button>
                  <button
                    @click="closeModal"
                    class="flex h-9 w-9 items-center justify-center text-brutal-black border-2 border-transparent hover:border-brutal-black hover:bg-brutal-yellow dark:text-dark-400 dark:hover:bg-dark-600 dark:hover:border-dark-400"
                    :aria-label="t('common.close')"
                  >
                    <Icon name="x" size="sm" />
                  </button>
                </div>
              </div>
              <!-- Decorative element removed for Neo Brutalism -->
            </div>

            <!-- Body -->
            <div class="max-h-[65vh] overflow-y-auto">
              <!-- Loading -->
              <div v-if="loading" class="flex items-center justify-center py-16">
                <div class="spinner h-8 w-8 border-2 border-brutal-black border-t-transparent"></div>
              </div>

              <!-- Announcements List -->
              <div v-else-if="announcements.length > 0">
                <div
                  v-for="item in announcements"
                  :key="item.id"
                  class="group relative flex items-center gap-4 border-b-2 border-brutal-black/20 px-6 py-4 transition-all hover:bg-brutal-yellow/20 dark:border-dark-600 dark:hover:bg-dark-700"
                  :class="{ 'bg-brutal-orange/5 dark:bg-brutal-orange/5': !item.read_at }"
                  style="min-height: 72px"
                  @click="openDetail(item)"
                >
                  <!-- Status Indicator -->
                  <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center">
                    <div
                      v-if="!item.read_at"
                      class="relative flex h-10 w-10 items-center justify-center bg-brutal-orange text-white border-2 border-brutal-black" style="box-shadow: 2px 2px 0px #1A1A1A;"
                    >
                      <!-- Pulse ring removed for Neo Brutalism -->
                      <!-- Icon -->
                      <svg class="relative z-10 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                    <div
                      v-else
                      class="flex h-10 w-10 items-center justify-center bg-gray-200 text-gray-500 border-2 border-brutal-black/30 dark:bg-dark-700 dark:text-gray-500 dark:border-dark-500"
                    >
                      <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                  </div>

                  <!-- Content -->
                  <div class="flex min-w-0 flex-1 items-center justify-between gap-4">
                    <div class="min-w-0 flex-1">
                      <h3 class="truncate text-sm font-bold text-brutal-black dark:text-white">
                        {{ item.title }}
                      </h3>
                      <div class="mt-1 flex items-center gap-2">
                        <time class="text-xs text-gray-500 dark:text-gray-400">
                          {{ formatRelativeTime(item.created_at) }}
                        </time>
                        <span
                          v-if="!item.read_at"
                          class="inline-flex items-center gap-1 bg-brutal-orange px-1.5 py-0.5 text-xs font-bold text-white border border-brutal-black"
                        >
                          <span class="relative flex h-1.5 w-1.5">
                            <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-blue-500 opacity-75"></span>
                            <span class="relative inline-flex h-1.5 w-1.5 rounded-full bg-blue-600"></span>
                          </span>
                          {{ t('announcements.unread') }}
                        </span>
                      </div>
                    </div>

                    <!-- Arrow -->
                    <div class="flex-shrink-0">
                      <svg
                        class="h-5 w-5 text-gray-400 transition-transform group-hover:translate-x-1 dark:text-gray-600"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        stroke-width="2"
                      >
                        <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
                      </svg>
                    </div>
                  </div>

                  <!-- Unread indicator bar -->
                  <div
                    v-if="!item.read_at"
                    class="absolute left-0 top-0 h-full w-1 bg-brutal-orange"
                  ></div>
                </div>
              </div>

              <!-- Empty State -->
              <div v-else class="empty-state py-16">
                <div class="mb-4 flex h-16 w-16 items-center justify-center bg-brutal-yellow border-2 border-brutal-black" style="box-shadow: 3px 3px 0px #1A1A1A;">
                  <Icon name="inbox" size="xl" class="text-gray-400 dark:text-gray-500" />
                </div>
                <p class="text-sm font-bold uppercase text-brutal-black dark:text-white">{{ t('announcements.empty') }}</p>
                <p class="mt-1 text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('announcements.emptyDescription') }}</p>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- 公告详情 Modal -->
    <Teleport to="body">
      <Transition name="modal-fade">
        <div
          v-if="detailModalOpen && selectedAnnouncement"
          class="fixed inset-0 z-[110] flex items-start justify-center overflow-y-auto bg-brutal-black/60 p-4 pt-[6vh]"
          @click="closeDetail"
        >
          <div
            class="w-full max-w-[780px] overflow-hidden bg-white border-3 border-brutal-black dark:bg-dark-800 dark:border-dark-500"
            style="box-shadow: 8px 8px 0px #1A1A1A;"
            @click.stop
          >
            <!-- Header -->
            <div class="relative overflow-hidden border-b-2 border-brutal-black bg-brutal-yellow px-8 py-6 dark:border-dark-600 dark:bg-dark-700">
              <div class="relative z-10 flex items-start justify-between gap-4">
                <div class="flex-1 min-w-0">
                  <!-- Icon and Category -->
                  <div class="mb-3 flex items-center gap-2">
                    <div class="flex h-10 w-10 items-center justify-center bg-brutal-orange text-white border-2 border-brutal-black" style="box-shadow: 2px 2px 0px #1A1A1A;">
                      <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                    <div class="flex items-center gap-2">
                      <span class="bg-white px-2.5 py-1 text-xs font-bold text-brutal-black border border-brutal-black dark:bg-dark-600 dark:text-dark-200 dark:border-dark-400">
                        {{ t('announcements.title') }}
                      </span>
                      <span
                        v-if="!selectedAnnouncement.read_at"
                        class="inline-flex items-center gap-1.5 bg-brutal-orange px-2.5 py-1 text-xs font-bold text-white border border-brutal-black"
                      >
                        <span class="relative flex h-2 w-2">
                          <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-white opacity-75"></span>
                          <span class="relative inline-flex h-2 w-2 rounded-full bg-white"></span>
                        </span>
                        {{ t('announcements.unread') }}
                      </span>
                    </div>
                  </div>

                  <!-- Title -->
                  <h2 class="mb-3 text-2xl font-bold uppercase tracking-wide leading-tight text-brutal-black dark:text-white">
                    {{ selectedAnnouncement.title }}
                  </h2>

                  <!-- Meta Info -->
                  <div class="flex items-center gap-4 text-sm text-gray-600 dark:text-gray-400">
                    <div class="flex items-center gap-1.5">
                      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                      <time>{{ formatRelativeWithDateTime(selectedAnnouncement.created_at) }}</time>
                    </div>
                    <div class="flex items-center gap-1.5">
                      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                      </svg>
                      <span>{{ selectedAnnouncement.read_at ? t('announcements.read') : t('announcements.unread') }}</span>
                    </div>
                  </div>
                </div>

                <!-- Close button -->
                <button
                  @click="closeDetail"
                  class="flex h-10 w-10 flex-shrink-0 items-center justify-center text-brutal-black border-2 border-transparent hover:border-brutal-black hover:bg-white dark:text-dark-400 dark:hover:bg-dark-600 dark:hover:border-dark-400"
                  :aria-label="t('common.close')"
                >
                  <Icon name="x" size="md" />
                </button>
              </div>
            </div>

            <!-- Body with Enhanced Markdown -->
            <div class="max-h-[60vh] overflow-y-auto bg-white px-8 py-8 dark:bg-dark-800">
              <div class="relative">
                <div class="absolute left-0 top-0 bottom-0 w-1 bg-brutal-orange"></div>
                <div class="pl-6">
                  <div
                    class="markdown-body prose prose-sm max-w-none dark:prose-invert"
                    v-html="renderMarkdown(selectedAnnouncement.content)"
                  ></div>
                </div>
              </div>
            </div>

            <!-- Footer with Actions -->
            <div class="border-t-2 border-brutal-black bg-gray-50 px-8 py-5 dark:border-dark-600 dark:bg-dark-900">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <span>{{ selectedAnnouncement.read_at ? t('announcements.readStatus') : t('announcements.markReadHint') }}</span>
                </div>
                <div class="flex items-center gap-3">
                  <button
                    @click="closeDetail"
                    class="btn btn-secondary"
                  >
                    {{ t('common.close') }}
                  </button>
                  <button
                    v-if="!selectedAnnouncement.read_at"
                    @click="markAsReadAndClose(selectedAnnouncement.id)"
                    class="btn btn-primary"
                  >
                    <span class="flex items-center gap-2">
                      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                      </svg>
                      {{ t('announcements.markRead') }}
                    </span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { announcementsAPI } from '@/api'
import { useAppStore } from '@/stores/app'
import { formatRelativeTime, formatRelativeWithDateTime } from '@/utils/format'
import type { UserAnnouncement } from '@/types'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

// Configure marked
marked.setOptions({
  breaks: true,
  gfm: true,
})

// State
const announcements = ref<UserAnnouncement[]>([])
const isModalOpen = ref(false)
const detailModalOpen = ref(false)
const selectedAnnouncement = ref<UserAnnouncement | null>(null)
const loading = ref(false)

// Computed
const unreadCount = computed(() =>
  announcements.value.filter((a) => !a.read_at).length
)

// Methods
function renderMarkdown(content: string): string {
  if (!content) return ''
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
}

async function loadAnnouncements() {
  try {
    loading.value = true
    const allAnnouncements = await announcementsAPI.list(false)
    announcements.value = allAnnouncements.slice(0, 20)
  } catch (err: any) {
    console.error('Failed to load announcements:', err)
    appStore.showError(err?.message || t('common.unknownError'))
  } finally {
    loading.value = false
  }
}

function openModal() {
  isModalOpen.value = true
  if (announcements.value.length === 0) {
    loadAnnouncements()
  }
}

function closeModal() {
  isModalOpen.value = false
}

function openDetail(announcement: UserAnnouncement) {
  selectedAnnouncement.value = announcement
  detailModalOpen.value = true
  if (!announcement.read_at) {
    markAsRead(announcement.id)
  }
}

function closeDetail() {
  detailModalOpen.value = false
  selectedAnnouncement.value = null
}

async function markAsRead(id: number) {
  try {
    await announcementsAPI.markRead(id)
    const announcement = announcements.value.find((a) => a.id === id)
    if (announcement) {
      announcement.read_at = new Date().toISOString()
    }
    if (selectedAnnouncement.value?.id === id) {
      selectedAnnouncement.value.read_at = new Date().toISOString()
    }
  } catch (err: any) {
    appStore.showError(err?.message || t('common.unknownError'))
  }
}

async function markAsReadAndClose(id: number) {
  await markAsRead(id)
  appStore.showSuccess(t('announcements.markedAsRead'))
  closeDetail()
}

async function markAllAsRead() {
  try {
    loading.value = true
    const unreadAnnouncements = announcements.value.filter((a) => !a.read_at)
    await Promise.all(unreadAnnouncements.map((a) => announcementsAPI.markRead(a.id)))
    announcements.value.forEach((a) => {
      if (!a.read_at) {
        a.read_at = new Date().toISOString()
      }
    })
    appStore.showSuccess(t('announcements.allMarkedAsRead'))
  } catch (err: any) {
    appStore.showError(err?.message || t('common.unknownError'))
  } finally {
    loading.value = false
  }
}

function handleEscape(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    if (detailModalOpen.value) {
      closeDetail()
    } else if (isModalOpen.value) {
      closeModal()
    }
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleEscape)
  loadAnnouncements()
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleEscape)
  // Restore body overflow in case component is unmounted while modals are open
  document.body.style.overflow = ''
})

watch([isModalOpen, detailModalOpen], ([modal, detail]) => {
  if (modal || detail) {
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
})
</script>

<style scoped>
/* Modal Animations */
.modal-fade-enter-active {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.modal-fade-leave-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 1, 1);
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

.modal-fade-enter-from > div {
  transform: scale(0.94) translateY(-12px);
  opacity: 0;
}

.modal-fade-leave-to > div {
  transform: scale(0.96) translateY(-8px);
  opacity: 0;
}

/* Scrollbar Styling */
.overflow-y-auto::-webkit-scrollbar {
  width: 8px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: transparent;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: linear-gradient(to bottom, #cbd5e1, #94a3b8);
  border-radius: 4px;
}

.dark .overflow-y-auto::-webkit-scrollbar-thumb {
  background: linear-gradient(to bottom, #4b5563, #374151);
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(to bottom, #94a3b8, #64748b);
}

.dark .overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(to bottom, #6b7280, #4b5563);
}
</style>

<style>
/* Enhanced Markdown Styles */
.markdown-body {
  @apply text-[15px] leading-[1.75];
  @apply text-gray-700 dark:text-gray-300;
}

.markdown-body h1 {
  @apply mb-6 mt-8 border-b border-gray-200 pb-3 text-3xl font-bold text-gray-900 dark:border-dark-600 dark:text-white;
}

.markdown-body h2 {
  @apply mb-4 mt-7 border-b border-gray-100 pb-2 text-2xl font-bold text-gray-900 dark:border-dark-700 dark:text-white;
}

.markdown-body h3 {
  @apply mb-3 mt-6 text-xl font-semibold text-gray-900 dark:text-white;
}

.markdown-body h4 {
  @apply mb-2 mt-5 text-lg font-semibold text-gray-900 dark:text-white;
}

.markdown-body p {
  @apply mb-4 leading-relaxed;
}

.markdown-body a {
  @apply font-medium text-blue-600 underline decoration-blue-600/30 decoration-2 underline-offset-2 transition-all hover:decoration-blue-600 dark:text-blue-400 dark:decoration-blue-400/30 dark:hover:decoration-blue-400;
}

.markdown-body ul,
.markdown-body ol {
  @apply mb-4 ml-6 space-y-2;
}

.markdown-body ul {
  @apply list-disc;
}

.markdown-body ol {
  @apply list-decimal;
}

.markdown-body li {
  @apply leading-relaxed;
  @apply pl-2;
}

.markdown-body li::marker {
  @apply text-blue-600 dark:text-blue-400;
}

.markdown-body blockquote {
  @apply relative my-5 border-l-4 border-blue-500 bg-blue-50/50 py-3 pl-5 pr-4 italic text-gray-700 dark:border-blue-400 dark:bg-blue-900/10 dark:text-gray-300;
}

.markdown-body blockquote::before {
  content: '"';
  @apply absolute -left-1 top-0 text-5xl font-serif text-blue-500/20 dark:text-blue-400/20;
}

.markdown-body code {
  @apply rounded-lg bg-gray-100 px-2 py-1 text-[13px] font-mono text-pink-600 dark:bg-dark-700 dark:text-pink-400;
}

.markdown-body pre {
  @apply my-5 overflow-x-auto rounded-xl border border-gray-200 bg-gray-50 p-5 dark:border-dark-600 dark:bg-dark-900/50;
}

.markdown-body pre code {
  @apply bg-transparent p-0 text-[13px] text-gray-800 dark:text-gray-200;
}

.markdown-body hr {
  @apply my-8 border-0 border-t-2 border-gray-200 dark:border-dark-700;
}

.markdown-body table {
  @apply mb-5 w-full overflow-hidden border-2 border-brutal-black dark:border-dark-600;
}

.markdown-body th,
.markdown-body td {
  @apply border-r border-b border-gray-200 px-4 py-3 text-left dark:border-dark-600;
}

.markdown-body th:last-child,
.markdown-body td:last-child {
  @apply border-r-0;
}

.markdown-body tr:last-child td {
  @apply border-b-0;
}

.markdown-body th {
  @apply bg-brutal-yellow font-bold uppercase tracking-wide text-brutal-black dark:bg-dark-700 dark:text-white;
}

.markdown-body tbody tr {
  @apply transition-colors hover:bg-gray-50 dark:hover:bg-dark-700/30;
}

.markdown-body img {
  @apply my-5 max-w-full border-2 border-brutal-black dark:border-dark-600;
}

.markdown-body strong {
  @apply font-bold text-brutal-black dark:text-white;
}

.markdown-body em {
  @apply italic text-gray-600 dark:text-gray-400;
}
</style>
