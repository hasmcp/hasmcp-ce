import { ref } from 'vue'
import { defineStore } from 'pinia'

export const useToastStore = defineStore('toast', () => {
  const message = ref('')
  const isVisible = ref(false)
  const timeoutId = ref(null)
  const level = ref('info') // 'info', 'warning', 'alert'

  /**
   * Displays a toast notification with a given message.
   * Automatically hides the toast after a specified duration.
   * @param {string} msg The message to display in the toast.
   * @param {string} levelType The type of toast ('info', 'warning', 'alert', 'error').
   * @param {number} duration The duration in milliseconds (default: 3000).
   */
  const showToast = (msg, levelType = 'info', duration = 3000) => {
    // Clear any existing timeout to reset the timer
    if (timeoutId.value) {
      clearTimeout(timeoutId.value)
      timeoutId.value = null
    }

    message.value = msg
    level.value = levelType
    isVisible.value = true

    timeoutId.value = setTimeout(() => {
      hideToast()
    }, duration)
  }

  const hideToast = () => {
    isVisible.value = false
    level.value = 'info'
    message.value = ''
    if (timeoutId.value) {
      clearTimeout(timeoutId.value)
      timeoutId.value = null
    }
  }

  return {
    message,
    isVisible,
    level, // <-- This line is critical
    showToast,
    hideToast,
  }
})
