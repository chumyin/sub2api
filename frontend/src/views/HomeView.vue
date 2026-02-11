<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="min-h-screen">
    <!-- iframe mode -->
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Default Home Page - Neo Brutalism -->
  <div
    v-else
    class="relative flex min-h-screen flex-col overflow-hidden bg-cream dark:bg-dark-950"
  >
    <!-- Background - Bold geometric shapes -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div
        class="absolute -right-20 -top-20 h-72 w-72 bg-brutal-orange border-3 border-brutal-black"
        style="transform: rotate(12deg);"
      ></div>
      <div
        class="absolute -bottom-24 -left-24 h-64 w-64 bg-brutal-yellow border-3 border-brutal-black"
        style="transform: rotate(-8deg);"
      ></div>
      <div
        class="absolute left-1/3 top-1/4 h-40 w-40 bg-brutal-teal border-3 border-brutal-black"
        style="transform: rotate(20deg);"
      ></div>
      <div
        class="absolute bottom-1/3 right-1/6 h-32 w-32 bg-brutal-pink border-3 border-brutal-black"
        style="transform: rotate(-15deg);"
      ></div>
      <!-- Grid Pattern -->
      <div
        class="absolute inset-0 bg-[linear-gradient(#1A1A1A_1px,transparent_1px),linear-gradient(90deg,#1A1A1A_1px,transparent_1px)] bg-[size:64px_64px] opacity-[0.03]"
      ></div>
    </div>

    <!-- Header -->
    <header class="relative z-20 px-6 py-4">
      <nav class="mx-auto flex max-w-6xl items-center justify-between">
        <!-- Logo -->
        <div class="flex items-center">
          <div class="h-10 w-10 overflow-hidden border-2 border-brutal-black bg-white" style="box-shadow: 2px 2px 0px #1A1A1A;">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
        </div>

        <!-- Nav Actions -->
        <div class="flex items-center gap-3">
          <!-- Language Switcher -->
          <LocaleSwitcher />

          <!-- Doc Link -->
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="p-2 text-brutal-black transition-colors border-2 border-transparent hover:border-brutal-black hover:bg-brutal-yellow dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:border-dark-400"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>

          <!-- Theme Toggle -->
          <button
            @click="toggleTheme"
            class="p-2 text-brutal-black transition-colors border-2 border-transparent hover:border-brutal-black hover:bg-brutal-yellow dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:border-dark-400"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>

          <!-- Login / Dashboard Button -->
          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="inline-flex items-center gap-1.5 bg-brutal-black py-1.5 pl-1.5 pr-3 text-white border-2 border-brutal-black transition-all hover:-translate-y-0.5 dark:bg-dark-700 dark:border-dark-400"
            style="box-shadow: 3px 3px 0px #FF6B00;"
          >
            <span
              class="flex h-6 w-6 items-center justify-center bg-brutal-orange text-[10px] font-bold text-white border border-white"
            >
              {{ userInitial }}
            </span>
            <span class="text-xs font-bold uppercase tracking-wide">{{ t('home.dashboard') }}</span>
            <svg
              class="h-3 w-3 text-gray-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M4.5 19.5l15-15m0 0H8.25m11.25 0v11.25"
              />
            </svg>
          </router-link>
          <router-link
            v-else
            to="/login"
            class="inline-flex items-center bg-brutal-black px-4 py-1.5 text-xs font-bold uppercase tracking-wide text-white border-2 border-brutal-black transition-all hover:-translate-y-0.5 dark:bg-dark-700 dark:border-dark-400"
            style="box-shadow: 3px 3px 0px #FF6B00;"
          >
            {{ t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <!-- Main Content -->
    <main class="relative z-10 flex-1 px-6 py-16">
      <div class="mx-auto max-w-6xl">
        <!-- Hero Section - Left/Right Layout -->
        <div class="mb-12 flex flex-col items-center justify-between gap-12 lg:flex-row lg:gap-16">
          <!-- Left: Text Content -->
          <div class="flex-1 text-center lg:text-left">
            <h1
              class="mb-4 text-4xl font-bold uppercase tracking-wide text-brutal-black dark:text-white md:text-5xl lg:text-6xl"
            >
              {{ siteName }}
            </h1>
            <p class="mb-8 text-lg font-medium text-gray-700 dark:text-dark-300 md:text-xl">
              {{ siteSubtitle }}
            </p>

            <!-- CTA Button -->
            <div>
              <router-link
                :to="isAuthenticated ? dashboardPath : '/login'"
                class="btn btn-primary px-8 py-3 text-base"
              >
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
                <Icon name="arrowRight" size="md" class="ml-2" :stroke-width="2" />
              </router-link>
            </div>
          </div>

          <!-- Right: Terminal Animation -->
          <div class="flex flex-1 justify-center lg:justify-end">
            <div class="terminal-container">
              <div class="terminal-window">
                <!-- Window header -->
                <div class="terminal-header">
                  <div class="terminal-buttons">
                    <span class="btn-close"></span>
                    <span class="btn-minimize"></span>
                    <span class="btn-maximize"></span>
                  </div>
                  <span class="terminal-title">terminal</span>
                </div>
                <!-- Terminal content -->
                <div class="terminal-body">
                  <div class="code-line line-1">
                    <span class="code-prompt">$</span>
                    <span class="code-cmd">curl</span>
                    <span class="code-flag">-X POST</span>
                    <span class="code-url">/v1/messages</span>
                  </div>
                  <div class="code-line line-2">
                    <span class="code-comment"># Routing to upstream...</span>
                  </div>
                  <div class="code-line line-3">
                    <span class="code-success">200 OK</span>
                    <span class="code-response">{ "content": "Hello!" }</span>
                  </div>
                  <div class="code-line line-4">
                    <span class="code-prompt">$</span>
                    <span class="cursor"></span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Feature Tags - Centered -->
        <div class="mb-12 flex flex-wrap items-center justify-center gap-4 md:gap-6">
          <div
            class="inline-flex items-center gap-2.5 bg-white px-5 py-2.5 border-2 border-brutal-black dark:bg-dark-800 dark:border-dark-500"
            style="box-shadow: 3px 3px 0px #1A1A1A;"
          >
            <Icon name="swap" size="sm" class="text-brutal-orange" />
            <span class="text-sm font-bold text-brutal-black dark:text-dark-200">{{
              t('home.tags.subscriptionToApi')
            }}</span>
          </div>
          <div
            class="inline-flex items-center gap-2.5 bg-white px-5 py-2.5 border-2 border-brutal-black dark:bg-dark-800 dark:border-dark-500"
            style="box-shadow: 3px 3px 0px #1A1A1A;"
          >
            <Icon name="shield" size="sm" class="text-brutal-orange" />
            <span class="text-sm font-bold text-brutal-black dark:text-dark-200">{{
              t('home.tags.stickySession')
            }}</span>
          </div>
          <div
            class="inline-flex items-center gap-2.5 bg-white px-5 py-2.5 border-2 border-brutal-black dark:bg-dark-800 dark:border-dark-500"
            style="box-shadow: 3px 3px 0px #1A1A1A;"
          >
            <Icon name="chart" size="sm" class="text-brutal-orange" />
            <span class="text-sm font-bold text-brutal-black dark:text-dark-200">{{
              t('home.tags.realtimeBilling')
            }}</span>
          </div>
        </div>

        <!-- Features Grid -->
        <div class="mb-12 grid gap-6 md:grid-cols-3">
          <!-- Feature 1: Unified Gateway -->
          <div
            class="group bg-white p-6 border-2 border-brutal-black transition-all duration-100 hover:-translate-x-0.5 hover:-translate-y-0.5 dark:bg-dark-800 dark:border-dark-500"
            style="box-shadow: 4px 4px 0px #1A1A1A;"
            @mouseenter="$event.currentTarget.style.boxShadow = '6px 6px 0px #1A1A1A'"
            @mouseleave="$event.currentTarget.style.boxShadow = '4px 4px 0px #1A1A1A'"
          >
            <div
              class="mb-4 flex h-12 w-12 items-center justify-center bg-brutal-blue border-2 border-brutal-black"
              style="box-shadow: 2px 2px 0px #1A1A1A;"
            >
              <Icon name="server" size="lg" class="text-white" />
            </div>
            <h3 class="mb-2 text-lg font-bold uppercase tracking-wide text-brutal-black dark:text-white">
              {{ t('home.features.unifiedGateway') }}
            </h3>
            <p class="text-sm font-medium leading-relaxed text-gray-700 dark:text-dark-400">
              {{ t('home.features.unifiedGatewayDesc') }}
            </p>
          </div>

          <!-- Feature 2: Account Pool -->
          <div
            class="group bg-white p-6 border-2 border-brutal-black transition-all duration-100 hover:-translate-x-0.5 hover:-translate-y-0.5 dark:bg-dark-800 dark:border-dark-500"
            style="box-shadow: 4px 4px 0px #1A1A1A;"
            @mouseenter="$event.currentTarget.style.boxShadow = '6px 6px 0px #1A1A1A'"
            @mouseleave="$event.currentTarget.style.boxShadow = '4px 4px 0px #1A1A1A'"
          >
            <div
              class="mb-4 flex h-12 w-12 items-center justify-center bg-brutal-teal border-2 border-brutal-black"
              style="box-shadow: 2px 2px 0px #1A1A1A;"
            >
              <svg
                class="h-6 w-6 text-brutal-black"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M18 18.72a9.094 9.094 0 003.741-.479 3 3 0 00-4.682-2.72m.94 3.198l.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0112 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 016 18.719m12 0a5.971 5.971 0 00-.941-3.197m0 0A5.995 5.995 0 0012 12.75a5.995 5.995 0 00-5.058 2.772m0 0a3 3 0 00-4.681 2.72 8.986 8.986 0 003.74.477m.94-3.197a5.971 5.971 0 00-.94 3.197M15 6.75a3 3 0 11-6 0 3 3 0 016 0zm6 3a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0zm-13.5 0a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0z"
                />
              </svg>
            </div>
            <h3 class="mb-2 text-lg font-bold uppercase tracking-wide text-brutal-black dark:text-white">
              {{ t('home.features.multiAccount') }}
            </h3>
            <p class="text-sm font-medium leading-relaxed text-gray-700 dark:text-dark-400">
              {{ t('home.features.multiAccountDesc') }}
            </p>
          </div>

          <!-- Feature 3: Billing & Quota -->
          <div
            class="group bg-white p-6 border-2 border-brutal-black transition-all duration-100 hover:-translate-x-0.5 hover:-translate-y-0.5 dark:bg-dark-800 dark:border-dark-500"
            style="box-shadow: 4px 4px 0px #1A1A1A;"
            @mouseenter="$event.currentTarget.style.boxShadow = '6px 6px 0px #1A1A1A'"
            @mouseleave="$event.currentTarget.style.boxShadow = '4px 4px 0px #1A1A1A'"
          >
            <div
              class="mb-4 flex h-12 w-12 items-center justify-center bg-brutal-yellow border-2 border-brutal-black"
              style="box-shadow: 2px 2px 0px #1A1A1A;"
            >
              <svg
                class="h-6 w-6 text-brutal-black"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z"
                />
              </svg>
            </div>
            <h3 class="mb-2 text-lg font-bold uppercase tracking-wide text-brutal-black dark:text-white">
              {{ t('home.features.balanceQuota') }}
            </h3>
            <p class="text-sm font-medium leading-relaxed text-gray-700 dark:text-dark-400">
              {{ t('home.features.balanceQuotaDesc') }}
            </p>
          </div>
        </div>

        <!-- Supported Providers -->
        <div class="mb-8 text-center">
          <h2 class="mb-3 text-2xl font-bold uppercase tracking-wide text-brutal-black dark:text-white">
            {{ t('home.providers.title') }}
          </h2>
          <p class="text-sm font-medium text-gray-700 dark:text-dark-400">
            {{ t('home.providers.description') }}
          </p>
        </div>

        <div class="mb-16 flex flex-wrap items-center justify-center gap-4">
          <!-- Claude - Supported -->
          <div
            class="flex items-center gap-2 bg-white px-5 py-3 border-2 border-brutal-black dark:bg-dark-800 dark:border-dark-500"
            style="box-shadow: 3px 3px 0px #FF6B00;"
          >
            <div
              class="flex h-8 w-8 items-center justify-center bg-brutal-orange border-2 border-brutal-black"
            >
              <span class="text-xs font-bold text-white">C</span>
            </div>
            <span class="text-sm font-bold text-brutal-black dark:text-dark-200">{{ t('home.providers.claude') }}</span>
            <span
              class="bg-brutal-teal px-1.5 py-0.5 text-[10px] font-bold uppercase text-brutal-black border border-brutal-black"
              >{{ t('home.providers.supported') }}</span
            >
          </div>
          <!-- GPT - Supported -->
          <div
            class="flex items-center gap-2 bg-white px-5 py-3 border-2 border-brutal-black dark:bg-dark-800 dark:border-dark-500"
            style="box-shadow: 3px 3px 0px #FF6B00;"
          >
            <div
              class="flex h-8 w-8 items-center justify-center bg-brutal-teal border-2 border-brutal-black"
            >
              <span class="text-xs font-bold text-brutal-black">G</span>
            </div>
            <span class="text-sm font-bold text-brutal-black dark:text-dark-200">GPT</span>
            <span
              class="bg-brutal-teal px-1.5 py-0.5 text-[10px] font-bold uppercase text-brutal-black border border-brutal-black"
              >{{ t('home.providers.supported') }}</span
            >
          </div>
          <!-- Gemini - Supported -->
          <div
            class="flex items-center gap-2 bg-white px-5 py-3 border-2 border-brutal-black dark:bg-dark-800 dark:border-dark-500"
            style="box-shadow: 3px 3px 0px #FF6B00;"
          >
            <div
              class="flex h-8 w-8 items-center justify-center bg-brutal-blue border-2 border-brutal-black"
            >
              <span class="text-xs font-bold text-white">G</span>
            </div>
            <span class="text-sm font-bold text-brutal-black dark:text-dark-200">{{ t('home.providers.gemini') }}</span>
            <span
              class="bg-brutal-teal px-1.5 py-0.5 text-[10px] font-bold uppercase text-brutal-black border border-brutal-black"
              >{{ t('home.providers.supported') }}</span
            >
          </div>
          <!-- Antigravity - Supported -->
          <div
            class="flex items-center gap-2 bg-white px-5 py-3 border-2 border-brutal-black dark:bg-dark-800 dark:border-dark-500"
            style="box-shadow: 3px 3px 0px #FF6B00;"
          >
            <div
              class="flex h-8 w-8 items-center justify-center bg-brutal-pink border-2 border-brutal-black"
            >
              <span class="text-xs font-bold text-white">A</span>
            </div>
            <span class="text-sm font-bold text-brutal-black dark:text-dark-200">{{ t('home.providers.antigravity') }}</span>
            <span
              class="bg-brutal-teal px-1.5 py-0.5 text-[10px] font-bold uppercase text-brutal-black border border-brutal-black"
              >{{ t('home.providers.supported') }}</span
            >
          </div>
          <!-- More - Coming Soon -->
          <div
            class="flex items-center gap-2 bg-gray-100 px-5 py-3 border-2 border-brutal-black/50 opacity-60 dark:bg-dark-800 dark:border-dark-600"
            style="box-shadow: 3px 3px 0px #1A1A1A;"
          >
            <div
              class="flex h-8 w-8 items-center justify-center bg-gray-400 border-2 border-brutal-black"
            >
              <span class="text-xs font-bold text-white">+</span>
            </div>
            <span class="text-sm font-bold text-brutal-black dark:text-dark-200">{{ t('home.providers.more') }}</span>
            <span
              class="bg-gray-200 px-1.5 py-0.5 text-[10px] font-bold uppercase text-gray-600 border border-brutal-black/50 dark:bg-dark-600 dark:text-dark-300"
              >{{ t('home.providers.soon') }}</span
            >
          </div>
        </div>
      </div>
    </main>

    <!-- Footer -->
    <footer class="relative z-10 border-t-3 border-brutal-black px-6 py-8 dark:border-dark-600">
      <div
        class="mx-auto flex max-w-6xl flex-col items-center justify-center gap-4 text-center sm:flex-row sm:text-left"
      >
        <p class="text-sm font-bold text-gray-600 dark:text-dark-400">
          &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
        </p>
        <div class="flex items-center gap-4">
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="text-sm font-bold text-brutal-black transition-colors hover:text-brutal-orange dark:text-dark-300 dark:hover:text-brutal-orange"
          >
            {{ t('home.docs') }}
          </a>
          <a
            :href="githubUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="text-sm font-bold text-brutal-black transition-colors hover:text-brutal-orange dark:text-dark-300 dark:hover:text-brutal-orange"
          >
            GitHub
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

// Check if homeContent is a URL (for iframe display)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

// Theme
const isDark = ref(document.documentElement.classList.contains('dark'))

// GitHub URL
const githubUrl = 'https://github.com/Wei-Shaw/sub2api'

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})

// Current year for footer
const currentYear = computed(() => new Date().getFullYear())

// Toggle theme
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// Initialize theme
function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()

  // Check auth state
  authStore.checkAuth()

  // Ensure public settings are loaded (will use cache if already loaded from injected config)
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
/* Terminal Container - Neo Brutalism */
.terminal-container {
  position: relative;
  display: inline-block;
}

/* Terminal Window */
.terminal-window {
  width: 420px;
  background: #1A1A1A;
  border: 3px solid #1A1A1A;
  box-shadow: 8px 8px 0px #FF6B00;
  overflow: hidden;
  transition: transform 0.1s ease;
}

.terminal-window:hover {
  transform: translate(-2px, -2px);
  box-shadow: 10px 10px 0px #FF6B00;
}

/* Terminal Header */
.terminal-header {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: #292524;
  border-bottom: 2px solid #44403c;
}

.terminal-buttons {
  display: flex;
  gap: 8px;
}

.terminal-buttons span {
  width: 12px;
  height: 12px;
  border: 1px solid #1A1A1A;
}

.btn-close {
  background: #ef4444;
}
.btn-minimize {
  background: #FFD600;
}
.btn-maximize {
  background: #00C9A7;
}

.terminal-title {
  flex: 1;
  text-align: center;
  font-size: 12px;
  font-family: 'Space Mono', ui-monospace, monospace;
  color: #78716c;
  margin-right: 52px;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  font-weight: 700;
}

/* Terminal Body */
.terminal-body {
  padding: 20px 24px;
  font-family: 'Space Mono', ui-monospace, monospace;
  font-size: 14px;
  line-height: 2;
}

.code-line {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  opacity: 0;
  animation: line-appear 0.5s ease forwards;
}

.line-1 {
  animation-delay: 0.3s;
}
.line-2 {
  animation-delay: 1s;
}
.line-3 {
  animation-delay: 1.8s;
}
.line-4 {
  animation-delay: 2.5s;
}

@keyframes line-appear {
  from {
    opacity: 0;
    transform: translateY(5px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.code-prompt {
  color: #00C9A7;
  font-weight: bold;
}
.code-cmd {
  color: #38bdf8;
}
.code-flag {
  color: #FF6B00;
  font-weight: 700;
}
.code-url {
  color: #00C9A7;
}
.code-comment {
  color: #78716c;
  font-style: italic;
}
.code-success {
  color: #00C9A7;
  background: rgba(0, 201, 167, 0.2);
  padding: 2px 8px;
  font-weight: 700;
  border: 1px solid #00C9A7;
}
.code-response {
  color: #FFD600;
  font-weight: 700;
}

/* Blinking Cursor */
.cursor {
  display: inline-block;
  width: 10px;
  height: 18px;
  background: #00C9A7;
  animation: blink 1s step-end infinite;
}

@keyframes blink {
  0%,
  50% {
    opacity: 1;
  }
  51%,
  100% {
    opacity: 0;
  }
}
</style>
