<script setup>
import { ref } from 'vue'
import yaml from 'js-yaml'

const props = defineProps({
  show: Boolean,
})

const emit = defineEmits(['close', 'start-import'])

const specText = ref('')
const errorMessage = ref('')

const closeModal = () => {
  specText.value = ''
  errorMessage.value = ''
  emit('close')
}

const handleImport = () => {
  errorMessage.value = ''
  if (!specText.value.trim()) {
    errorMessage.value = 'OpenAPI spec cannot be empty.'
    return
  }

  let parsedSpec
  try {
    parsedSpec = JSON.parse(specText.value)
  } catch (jsonError) {
    console.error(jsonError)
    try {
      parsedSpec = yaml.load(specText.value)
    } catch (yamlError) {
      errorMessage.value = `Failed to parse spec. Invalid JSON or YAML. (Error: ${yamlError.message})`
      return
    }
  }

  if (!parsedSpec || (!parsedSpec.openapi && !parsedSpec.swagger)) {
    errorMessage.value = 'Invalid or unsupported spec. Must be an OpenAPI 3.x or Swagger 2.x spec.'
    return
  }

  const isOAS3 = parsedSpec.openapi && parsedSpec.openapi.startsWith('3.')
  const isSwagger2 = parsedSpec.swagger && (parsedSpec.swagger === '2.0' || parsedSpec.swagger.startsWith('2.'))

  if (!isOAS3 && !isSwagger2) {
    errorMessage.value = 'Invalid or unsupported spec version. Only OpenAPI 3.x and Swagger 2.x are supported.'
    return
  }

  if (!parsedSpec.paths || Object.keys(parsedSpec.paths).length === 0) {
    errorMessage.value = 'Spec contains no paths to import.'
    return
  }

  emit('start-import', parsedSpec)
  closeModal()
}
</script>

<template>
  <div v-if="props.show" class="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog"
    aria-modal="true">
    <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
      <div class="fixed inset-0 bg-gray-900 bg-opacity-75 transition-opacity" aria-hidden="true" @click="closeModal">
      </div>

      <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>
      <div
        class="inline-block relative z-50 align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-2xl sm:w-full">
        <form @submit.prevent="handleImport">
          <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
            <div class="sm:flex sm:items-start">
              <div
                class="mx-auto shrink-0 flex items-center justify-center h-12 w-12 rounded-full bg-black sm:mx-0 sm:h-10 sm:w-10">
                <svg class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
                </svg>
              </div>
              <div class="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left grow">
                <h3 class="text-lg leading-6 font-medium text-gray-900" id="modal-title">
                  Import Tools from OpenAPI 3 / Swagger
                </h3>
                <div class="mt-4">
                  <p class="text-sm text-gray-500 mb-2">
                    Paste your OpenAPI 3.x.x or Swagger 2.x spec (JSON or YAML) below. Tools with a matching
                    <strong class="text-black">METHOD + Path</strong>
                    will be updated. New tools will be created.
                  </p>
                  <textarea v-model="specText" rows="15"
                    class="w-full p-2 border border-gray-300 rounded-md font-mono text-xs focus:ring-black focus:border-black"
                    placeholder="Paste your OpenAPI 3 spec here..."></textarea>
                  <p v-if="errorMessage" class="text-sm text-red-600 mt-2">{{ errorMessage }}</p>
                </div>
              </div>
            </div>
          </div>

          <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
            <button type="submit"
              class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-black text-base font-medium text-white hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-black sm:ml-3 sm:w-auto sm:text-sm transition duration-150">
              Start Import
            </button>
            <button type="button" @click="closeModal"
              class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-black sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm transition duration-150">
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>