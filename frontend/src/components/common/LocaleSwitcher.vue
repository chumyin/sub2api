<template>
  <div class="relative" ref="dropdownRef">
    <button
      @click="toggleDropdown"
      class="flex items-center gap-1.5 px-2 py-1.5 text-sm font-bold text-brutal-black transition-colors border-2 border-transparent hover:border-brutal-black hover:bg-brutal-yellow dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:border-dark-400"
      :title="currentLocale?.name"
    >
      <span class="text-base">{{ currentLocale?.flag }}</span>
      <span class="hidden sm:inline">{{ currentLocale?.code.toUpperCase() }}</span>
      <Icon
        name="chevronDown"
        size="xs"
        class="text-gray-400 transition-transform duration-200"
        :class="{ 'rotate-180': isOpen }"
      />
    </button>

    <transition name="dropdown">
      <div
        v-if="isOpen"
        class="absolute right-0 z-50 mt-1 w-32 overflow-hidden border-2 border-brutal-black bg-white dark:border-dark-500 dark:bg-dark-800"
        style="box-shadow: 4px 4px 0px #1A1A1A;"
      >
        <button
          v-for="locale in availableLocales"
          :key="locale.code"
          @click="selectLocale(locale.code)"
          class="flex w-full items-center gap-2 px-3 py-2 text-sm font-medium text-brutal-black transition-colors hover:bg-brutal-orange hover:text-white dark:text-dark-200 dark:hover:bg-dark-600"
          :class="{
            'bg-brutal-yellow text-brutal-black dark:bg-brutal-orange/20 dark:text-brutal-orange':
              locale.code === currentLocaleCode
          }"
        >
          <span class="text-base">{{ locale.flag }}</span>
          <span>{{ locale.name }}</span>
          <Icon v-if="locale.code === currentLocaleCode" name="check" size="sm" class="ml-auto text-primary-500" />
        </button>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { setLocale, availableLocales } from '@/i18n'

const { locale } = useI18n()

const isOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

const currentLocaleCode = computed(() => locale.value)
const currentLocale = computed(() => availableLocales.find((l) => l.code === locale.value))

function toggleDropdown() {
  isOpen.value = !isOpen.value
}

function selectLocale(code: string) {
  setLocale(code)
  isOpen.value = false
}

function handleClickOutside(event: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.15s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-4px);
}
</style>
