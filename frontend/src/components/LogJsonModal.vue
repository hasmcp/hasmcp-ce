<script setup>
import { computed, ref } from 'vue'
import { useToastStore } from '../stores/toastStore'

const props = defineProps({
  show: Boolean,
  jsonString: String,
})

const emit = defineEmits(['close'])
const toastStore = useToastStore()

const copyButtonText = ref('Copy JSON')

/**
 * Pretty-prints the JSON string prop.
 */
const prettyJson = computed(() => {
  if (!props.jsonString) return ''
  try {
    const parsed = JSON.parse(props.jsonString)
    return JSON.stringify(parsed, null, 2)
  } catch (e) {
    console.error(e)
    // If it's not valid JSON, just return the raw text
    return props.jsonString
  }
})

const closeModal = () => {
  emit('close')
  copyButtonText.value = 'Copy JSON' // Reset button text on close
}

// 4. New function to handle copying
const copyJson = () => {
  if (navigator.clipboard) {
    navigator.clipboard
      .writeText(prettyJson.value)
      .then(() => {
        toastStore.showToast('JSON copied to clipboard!', 'info', 2000)
        copyButtonText.value = 'Copied!'
      })
      .catch((err) => {
        toastStore.showToast(`Failed to copy: ${err}`, 'warning')
        copyButtonText.value = 'Failed!'
      })
  } else {
    toastStore.showToast('Clipboard API not available.', 'warning')
  }
}
</script>

<template>
  <div v-if="show" class="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog"
    aria-modal="true">
    <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
      <div class="fixed inset-0 bg-gray-900 bg-opacity-75 transition-opacity" aria-hidden="true" @click="closeModal">
      </div>

      <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>
      <div
        class="inline-block relative z-50 align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-2xl sm:w-full">
        <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
          <div class="flex justify-between items-center mb-4">
            <h3 class="text-lg leading-6 font-medium text-gray-900" id="modal-title">
              Log Data (JSON)
            </h3>
            <button @click="closeModal" class="text-gray-400 hover:text-gray-600">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
              </svg>
            </button>
          </div>
          <div class="mt-2">
            <pre
              class="bg-gray-900 text-white font-mono text-xs p-4 rounded-md overflow-x-auto max-h-[60vh]">{{ prettyJson }}</pre>
          </div>
        </div>

        <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
          <button type="button" @click="copyJson"
            class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-black text-base font-medium text-white hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-black sm:ml-3 sm:w-auto sm:text-sm transition duration-150">
            {{ copyButtonText }}
          </button>
          <button type="button" @click="closeModal"
            class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-black sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm transition duration-150">
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
