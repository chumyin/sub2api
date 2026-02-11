<template>
  <div
    class="flex items-center justify-between border-t-2 border-brutal-black bg-white px-4 py-3 dark:border-dark-600 dark:bg-dark-800 sm:px-6"
  >
    <div class="flex flex-1 items-center justify-between sm:hidden">
      <!-- Mobile pagination -->
      <button
        @click="goToPage(page - 1)"
        :disabled="page === 1"
        class="relative inline-flex items-center border-2 border-brutal-black bg-white px-4 py-2 text-sm font-bold text-brutal-black hover:bg-brutal-yellow disabled:cursor-not-allowed disabled:opacity-50 dark:border-dark-500 dark:bg-dark-700 dark:text-dark-200 dark:hover:bg-dark-600"
        style="box-shadow: 2px 2px 0px #1A1A1A;"
      >
        {{ t('pagination.previous') }}
      </button>
      <span class="text-sm font-bold text-brutal-black dark:text-dark-200">
        {{ t('pagination.pageOf', { page, total: totalPages }) }}
      </span>
      <button
        @click="goToPage(page + 1)"
        :disabled="page === totalPages"
        class="relative ml-3 inline-flex items-center border-2 border-brutal-black bg-white px-4 py-2 text-sm font-bold text-brutal-black hover:bg-brutal-yellow disabled:cursor-not-allowed disabled:opacity-50 dark:border-dark-500 dark:bg-dark-700 dark:text-dark-200 dark:hover:bg-dark-600"
        style="box-shadow: 2px 2px 0px #1A1A1A;"
      >
        {{ t('pagination.next') }}
      </button>
    </div>

    <div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
      <!-- Desktop pagination info -->
      <div class="flex items-center space-x-4">
        <p class="text-sm font-medium text-brutal-black dark:text-dark-200">
          {{ t('pagination.showing') }}
          <span class="font-bold font-mono">{{ fromItem }}</span>
          {{ t('pagination.to') }}
          <span class="font-bold font-mono">{{ toItem }}</span>
          {{ t('pagination.of') }}
          <span class="font-bold font-mono">{{ total }}</span>
          {{ t('pagination.results') }}
        </p>

        <!-- Page size selector -->
        <div v-if="showPageSizeSelector" class="flex items-center space-x-2">
          <span class="text-sm font-bold text-brutal-black dark:text-dark-200"
            >{{ t('pagination.perPage') }}:</span
          >
          <div class="page-size-select w-20">
            <Select
              :model-value="pageSize"
              :options="pageSizeSelectOptions"
              @update:model-value="handlePageSizeChange"
            />
          </div>
        </div>

        <div v-if="showJump" class="flex items-center space-x-2">
          <span class="text-sm font-bold text-brutal-black dark:text-dark-200">{{ t('pagination.jumpTo') }}</span>
          <input
            v-model="jumpPage"
            type="number"
            min="1"
            :max="totalPages"
            class="input w-20 text-sm"
            :placeholder="t('pagination.jumpPlaceholder')"
            @keyup.enter="submitJump"
          />
          <button type="button" class="btn btn-ghost btn-sm" @click="submitJump">
            {{ t('pagination.jumpAction') }}
          </button>
        </div>
      </div>

      <!-- Desktop pagination buttons -->
      <nav
        class="relative z-0 inline-flex"
        aria-label="Pagination"
      >
        <!-- Previous button -->
        <button
          @click="goToPage(page - 1)"
          :disabled="page === 1"
          class="relative inline-flex items-center border-2 border-brutal-black bg-white px-2 py-2 text-sm font-bold text-brutal-black hover:bg-brutal-yellow disabled:cursor-not-allowed disabled:opacity-50 dark:border-dark-500 dark:bg-dark-700 dark:text-dark-300 dark:hover:bg-dark-600 -mr-px"
          :aria-label="t('pagination.previous')"
        >
          <Icon name="chevronLeft" size="md" />
        </button>

        <!-- Page numbers -->
        <button
          v-for="pageNum in visiblePages"
          :key="pageNum"
          @click="typeof pageNum === 'number' && goToPage(pageNum)"
          :disabled="typeof pageNum !== 'number'"
          :class="[
            'relative inline-flex items-center border-2 border-brutal-black px-4 py-2 text-sm font-bold -mr-px',
            pageNum === page
              ? 'z-10 bg-brutal-orange text-white dark:bg-brutal-orange'
              : 'bg-white text-brutal-black hover:bg-brutal-yellow dark:border-dark-500 dark:bg-dark-700 dark:text-dark-200 dark:hover:bg-dark-600',
            typeof pageNum !== 'number' && 'cursor-default'
          ]"
          :aria-label="
            typeof pageNum === 'number' ? t('pagination.goToPage', { page: pageNum }) : undefined
          "
          :aria-current="pageNum === page ? 'page' : undefined"
        >
          {{ pageNum }}
        </button>

        <!-- Next button -->
        <button
          @click="goToPage(page + 1)"
          :disabled="page === totalPages"
          class="relative inline-flex items-center border-2 border-brutal-black bg-white px-2 py-2 text-sm font-bold text-brutal-black hover:bg-brutal-yellow disabled:cursor-not-allowed disabled:opacity-50 dark:border-dark-500 dark:bg-dark-700 dark:text-dark-300 dark:hover:bg-dark-600"
          :aria-label="t('pagination.next')"
        >
          <Icon name="chevronRight" size="md" />
        </button>
      </nav>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import Select from './Select.vue'

const { t } = useI18n()

interface Props {
  total: number
  page: number
  pageSize: number
  pageSizeOptions?: number[]
  showPageSizeSelector?: boolean
  showJump?: boolean
}

interface Emits {
  (e: 'update:page', page: number): void
  (e: 'update:pageSize', pageSize: number): void
}

const props = withDefaults(defineProps<Props>(), {
  pageSizeOptions: () => [10, 20, 50, 100],
  showPageSizeSelector: true,
  showJump: false
})

const emit = defineEmits<Emits>()

const totalPages = computed(() => Math.ceil(props.total / props.pageSize))

const fromItem = computed(() => {
  if (props.total === 0) return 0
  return (props.page - 1) * props.pageSize + 1
})

const toItem = computed(() => {
  const to = props.page * props.pageSize
  return to > props.total ? props.total : to
})

const pageSizeSelectOptions = computed(() => {
  return props.pageSizeOptions.map((size) => ({
    value: size,
    label: String(size)
  }))
})

const jumpPage = ref('')

const visiblePages = computed(() => {
  const pages: (number | string)[] = []
  const maxVisible = 7
  const total = totalPages.value

  if (total <= maxVisible) {
    // Show all pages if total is small
    for (let i = 1; i <= total; i++) {
      pages.push(i)
    }
  } else {
    // Always show first page
    pages.push(1)

    const start = Math.max(2, props.page - 2)
    const end = Math.min(total - 1, props.page + 2)

    // Add ellipsis before if needed
    if (start > 2) {
      pages.push('...')
    }

    // Add middle pages
    for (let i = start; i <= end; i++) {
      pages.push(i)
    }

    // Add ellipsis after if needed
    if (end < total - 1) {
      pages.push('...')
    }

    // Always show last page
    pages.push(total)
  }

  return pages
})

const goToPage = (newPage: number) => {
  if (newPage >= 1 && newPage <= totalPages.value && newPage !== props.page) {
    emit('update:page', newPage)
  }
}

const handlePageSizeChange = (value: string | number | boolean | null) => {
  if (value === null || typeof value === 'boolean') return
  const newPageSize = typeof value === 'string' ? parseInt(value) : value
  emit('update:pageSize', newPageSize)
}

const submitJump = () => {
  const value = jumpPage.value.trim()
  if (!value) return
  const pageNum = Number.parseInt(value, 10)
  if (Number.isNaN(pageNum)) return
  const nextPage = Math.min(Math.max(pageNum, 1), totalPages.value)
  jumpPage.value = ''
  goToPage(nextPage)
}
</script>

<style scoped>
.page-size-select :deep(.select-trigger) {
  @apply px-3 py-1.5 text-sm;
}
</style>
