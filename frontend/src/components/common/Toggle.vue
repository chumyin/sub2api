<template>
  <button
    type="button"
    @click="toggle"
    class="relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer border-2 border-brutal-black transition-colors duration-100 ease-in-out focus:outline-none dark:border-dark-400"
    :class="[modelValue ? 'bg-brutal-orange' : 'bg-gray-300 dark:bg-dark-600']"
    role="switch"
    :aria-checked="modelValue"
  >
    <span
      class="pointer-events-none inline-block h-[18px] w-[18px] transform bg-white border border-brutal-black transition duration-100 ease-in-out"
      :class="[modelValue ? 'translate-x-5' : 'translate-x-0']"
    />
  </button>
</template>

<script setup lang="ts">
const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

function toggle() {
  emit('update:modelValue', !props.modelValue)
}
</script>
