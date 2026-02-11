<template>
  <div class="relative flex min-h-screen items-center justify-center overflow-hidden p-4 bg-cream dark:bg-dark-950">
    <!-- Neo Brutalism Background Pattern -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <!-- Bold geometric shapes -->
      <div
        class="absolute -right-20 -top-20 h-64 w-64 bg-brutal-orange border-3 border-brutal-black"
        style="transform: rotate(15deg);"
      ></div>
      <div
        class="absolute -bottom-16 -left-16 h-48 w-48 bg-brutal-yellow border-3 border-brutal-black"
        style="transform: rotate(-10deg);"
      ></div>
      <div
        class="absolute right-1/4 bottom-1/4 h-32 w-32 bg-brutal-teal border-3 border-brutal-black"
        style="transform: rotate(25deg);"
      ></div>
      <!-- Grid Pattern -->
      <div
        class="absolute inset-0 bg-[linear-gradient(#1A1A1A_1px,transparent_1px),linear-gradient(90deg,#1A1A1A_1px,transparent_1px)] bg-[size:64px_64px] opacity-[0.03]"
      ></div>
    </div>

    <!-- Content Container -->
    <div class="relative z-10 w-full max-w-md">
      <!-- Logo/Brand -->
      <div class="mb-8 text-center">
        <div
          class="mb-4 inline-flex h-16 w-16 items-center justify-center overflow-hidden border-3 border-brutal-black bg-white"
          style="box-shadow: 4px 4px 0px #1A1A1A;"
        >
          <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
        </div>
        <h1 class="mb-2 text-3xl font-bold uppercase tracking-wide text-brutal-black dark:text-white">
          {{ siteName }}
        </h1>
        <p class="text-sm font-medium text-gray-600 dark:text-dark-400">
          {{ siteSubtitle }}
        </p>
      </div>

      <!-- Card Container -->
      <div class="bg-white dark:bg-dark-800 border-3 border-brutal-black dark:border-dark-500 p-8" style="box-shadow: 6px 6px 0px #1A1A1A;">
        <slot />
      </div>

      <!-- Footer Links -->
      <div class="mt-6 text-center text-sm font-bold">
        <slot name="footer" />
      </div>

      <!-- Copyright -->
      <div class="mt-8 text-center text-xs font-medium text-gray-500 dark:text-dark-400">
        &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'Subscription to API Conversion Platform')

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>
