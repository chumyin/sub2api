<script setup lang="ts">
import { ref } from 'vue'

defineProps<{
  content?: string
}>()

const show = ref(false)
</script>

<template>
  <div
    class="group relative ml-1 inline-flex items-center align-middle"
    @mouseenter="show = true"
    @mouseleave="show = false"
  >
    <!-- Trigger Icon -->
    <slot name="trigger">
      <svg
        class="h-4 w-4 cursor-help text-gray-400 transition-colors hover:text-primary-600 dark:text-gray-500 dark:hover:text-primary-400"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        stroke-width="2"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
        />
      </svg>
    </slot>

    <!-- Popover Content -->
    <div
      v-show="show"
      class="absolute bottom-full left-1/2 z-50 mb-2 w-64 -translate-x-1/2 bg-brutal-black p-3 text-xs font-medium leading-relaxed text-white border-2 border-brutal-black opacity-0 transition-opacity duration-150 group-hover:opacity-100 dark:bg-dark-900 dark:border-dark-500"
      style="box-shadow: 3px 3px 0px #FF6B00;"
    >
      <slot>{{ content }}</slot>
      <div class="absolute -bottom-1 left-1/2 h-2 w-2 -translate-x-1/2 rotate-45 bg-brutal-black dark:bg-dark-900"></div>
    </div>
  </div>
</template>

