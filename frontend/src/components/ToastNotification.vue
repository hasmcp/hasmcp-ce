<script setup>
import { useToastStore } from '../stores/toastStore'
import { computed } from 'vue'

const toastStore = useToastStore()

// Dynamically compute classes based on the toast level
const toastClasses = computed(() => {
  switch (toastStore.level) {
    case 'alert':
    case 'error':
      return 'bg-red-600 text-white'
    case 'warning':
      return 'bg-orange-500 text-white'
    case 'info':
    default:
      return 'bg-black text-white'
  }
})
</script>

<template>
  <teleport to="body">
    <Transition name="toast">
      <div v-if="toastStore.isVisible" :class="[
        'fixed bottom-4 right-4 z-50 flex items-center p-4 w-full max-w-xs rounded-lg shadow-xl',
        toastClasses,
      ]" role="alert">
        <svg v-if="toastStore.level === 'info'" class="w-5 h-5 text-green-400 mr-2 shrink-0" fill="none"
          stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
        </svg>

        <svg v-if="toastStore.level === 'alert' || toastStore.level === 'error'"
          class="w-5 h-5 text-white mr-2 shrink-0" fill="currentColor" viewBox="0 0 20 20"
          xmlns="http://www.w3.org/2000/svg">
          <path fill-rule="evenodd"
            d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
            clip-rule="evenodd"></path>
        </svg>

        <svg v-if="toastStore.level === 'warning'" class="w-5 h-5 text-white mr-2 shrink-0" fill="currentColor"
          viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
          <path fill-rule="evenodd"
            d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.332-.216 3.001-1.742 3.001H4.42c-1.526 0-2.492-1.669-1.742-3.001l5.58-9.92zM10 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
            clip-rule="evenodd"></path>
        </svg>

        <div class="text-sm font-medium min-w-0 break-all">
          {{ toastStore.message }}
        </div>

        <button type="button" @click="toastStore.hideToast"
          class="ml-auto -mx-1.5 -my-1.5 rounded-lg focus:ring-2 focus:ring-gray-300 p-1.5 inline-flex h-8 w-8 transition duration-150"
          :class="toastStore.level === 'info'
            ? 'bg-black text-white hover:text-gray-200 hover:bg-gray-800'
            : 'bg-transparent text-white/70 hover:text-white hover:bg-white/20'
            " aria-label="Close">
          <span class="sr-only">Close toast</span>
          <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
            <path fill-rule="evenodd"
              d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
              clip-rule="evenodd"></path>
          </svg>
        </button>
      </div>
    </Transition>
  </teleport>
</template>

<style scoped>
/* Transition styles for smooth show/hide effect */
.toast-enter-active,
.toast-leave-active {
  transition:
    opacity 0.3s ease,
    transform 0.3s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(100%);
}
</style>
