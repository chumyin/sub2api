<template>
  <div class="table-page-layout" :class="{ 'mobile-mode': isMobile }">
    <!-- Fixed area: Action buttons -->
    <div v-if="$slots.actions" class="layout-section-fixed">
      <slot name="actions" />
    </div>

    <!-- Fixed area: Search and filters -->
    <div v-if="$slots.filters" class="layout-section-fixed">
      <slot name="filters" />
    </div>

    <!-- Scrollable area: Table -->
    <div class="layout-section-scrollable">
      <div class="card table-scroll-container">
        <slot name="table" />
      </div>
    </div>

    <!-- Fixed area: Pagination -->
    <div v-if="$slots.pagination" class="layout-section-fixed">
      <slot name="pagination" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const isMobile = ref(false)

const checkMobile = () => {
  isMobile.value = window.innerWidth < 1024
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style scoped>
/* Desktop: Flexbox layout */
.table-page-layout {
  @apply flex flex-col gap-6;
  height: calc(100vh - 64px - 4rem);
}

.layout-section-fixed {
  @apply flex-shrink-0;
}

.layout-section-scrollable {
  @apply flex-1 min-h-0 flex flex-col;
}

/* Neo Brutalism table container */
.table-scroll-container {
  @apply flex flex-col overflow-hidden h-full bg-white dark:bg-dark-800 border-2 border-brutal-black dark:border-dark-500;
  box-shadow: 4px 4px 0px #1A1A1A;
}

.dark .table-scroll-container {
  box-shadow: 4px 4px 0px #44403c;
}

.table-scroll-container :deep(.table-wrapper) {
  @apply flex-1 overflow-x-auto overflow-y-auto;
  scrollbar-gutter: stable;
}

.table-scroll-container :deep(table) {
  @apply w-full;
  min-width: max-content;
  display: table;
}

.table-scroll-container :deep(thead) {
  @apply bg-brutal-yellow dark:bg-dark-700;
}

.table-scroll-container :deep(tbody) {
  /* Default table-row-group */
}

.table-scroll-container :deep(th) {
  @apply px-5 py-4 text-left text-sm font-bold text-brutal-black dark:text-white border-b-2 border-brutal-black dark:border-dark-500;
  @apply uppercase tracking-wider;
}

.table-scroll-container :deep(td) {
  @apply px-5 py-4 text-sm text-brutal-black dark:text-dark-200 border-b-2 border-brutal-black/20 dark:border-dark-600;
}

/* Mobile: Normal scroll */
.table-page-layout.mobile-mode .table-scroll-container {
  @apply h-auto overflow-visible border-2 border-brutal-black bg-white dark:bg-dark-800;
  box-shadow: 4px 4px 0px #1A1A1A;
}

.table-page-layout.mobile-mode .layout-section-scrollable {
  @apply flex-none min-h-fit;
}

.table-page-layout.mobile-mode .table-scroll-container :deep(.table-wrapper) {
  @apply overflow-visible;
}

.table-page-layout.mobile-mode .table-scroll-container :deep(table) {
  @apply flex-none;
  display: table;
  min-width: 100%;
}
</style>
