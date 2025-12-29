<script setup>
import { ref, computed } from 'vue'
import { useEnvVarStore } from '../stores/envVarStore'
import { useToastStore } from '../stores/toastStore.js'
import DeleteModal from '../components/DeleteModal.vue'

const envVarStore = useEnvVarStore()
const toastStore = useToastStore()

// ----------------------------------------------------
// Data and State
// ----------------------------------------------------

// State for the new variable form
const newEnv = ref({
  name: '',
  value: '',
  isSecret: false,
})

// State for the creation form visibility
const isCreationFormVisible = ref(false)

// State for editing
const editingId = ref(null)
const editingValue = ref('') // Only tracks the new VALUE being edited

// State for modal
const showDeleteModal = ref(false)
const variableToDelete = ref(null)
const variableNameForModal = ref('')
const itemTypeForModal = ref('Environment Variable')

// ----------------------------------------------------
// Computed Properties & Helpers
// ----------------------------------------------------

// Get variables from store
const envVariables = computed(() => envVarStore.envVariables)

// Enforces capital snake case for the name field on input
const enforceCapitalSnakeCase = (event) => {
  let value = event.target.value.toUpperCase()
  value = value.replace(/[^A-Z0-9_]/g, '').replace(/-|\s/g, '_')
  newEnv.value.name = value
}

// Helper to determine the masked value for secrets
const maskedValue = (value) => {
  if (value.length > 8) {
    return '*'.repeat(value.length - 4) + value.slice(-4)
  }
  return '*'.repeat(value.length)
}

// Helper for row class based on secret status
const rowClass = (isSecret) => {
  return isSecret ? 'bg-gray-100 text-gray-800' : 'bg-white'
}

// ----------------------------------------------------
// CRUD Methods
// ----------------------------------------------------

// Create a new environment variable
const createEnv = async () => {
  if (!newEnv.value.name || !newEnv.value.value) {
    toastStore.showToast('Name and Value are required!', 'warning', 3000)
    return
  }

  try {
    // Pass the local state object (with isSecret) to the store
    const result = await envVarStore.addEnvVar(newEnv.value)

    if (!result.success) {
      toastStore.showToast(result.message, 'alert', 4000)
      return
    }

    toastStore.showToast(result.message, 'info', 3000)

    // Reset form and hide it
    newEnv.value.name = ''
    newEnv.value.value = ''
    newEnv.value.isSecret = false
    isCreationFormVisible.value = false
  } catch (error) {
    console.error('Failed to create env var:', error)
  }
}

// Initiate deletion process (open modal)
const initiateDelete = (variable) => {
  variableToDelete.value = variable.id
  variableNameForModal.value = variable.name
  itemTypeForModal.value = 'Environment Variable'
  showDeleteModal.value = true
}

// Handle confirmation from modal
const confirmDelete = async () => {
  if (variableToDelete.value !== null) {
    const isDeleted = await envVarStore.deleteEnvVar(variableToDelete.value)

    if (isDeleted) {
      toastStore.showToast(
        `Environment Variable "${variableNameForModal.value}" deleted successfully!`,
      )
    }

    showDeleteModal.value = false
    variableToDelete.value = null
    variableNameForModal.value = ''
    itemTypeForModal.value = ''
  }
}

// Close modal without deleting
const cancelDelete = () => {
  showDeleteModal.value = false
  variableToDelete.value = null
  variableNameForModal.value = ''
  itemTypeForModal.value = ''
}

// Start editing a variable
const startEdit = (variable) => {
  editingId.value = variable.id
  editingValue.value = variable.value
}

// Cancel editing
const cancelEdit = () => {
  editingId.value = null
  editingValue.value = ''
}

// Save edited variable (Value Only)
const saveEdit = async () => {
  if (editingId.value === null) return

  const result = await envVarStore.updateEnvVarValue(editingId.value, editingValue.value)

  if (result.success) {
    toastStore.showToast(result.message)
    cancelEdit()
  }
}
</script>

<template>
  <div class="p-4">
    <h1 class="text-3xl font-bold mb-6 text-gray-800 flex justify-between items-center">
      Environment Variables
      <button @click="isCreationFormVisible = !isCreationFormVisible" :class="[
        'p-2 rounded-full text-white transition-transform duration-300',
        isCreationFormVisible
          ? 'bg-black hover:bg-gray-800 rotate-45'
          : 'bg-black hover:bg-gray-800',
      ]" title="Toggle Add Variable Form">
        <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
      </button>
    </h1>

    <div v-if="isCreationFormVisible" class="bg-white p-6 rounded-lg shadow-md mb-8 transition-all duration-300">
      <h2 class="text-xl font-semibold mb-4 text-gray-700">Add New Variable</h2>

      <form @submit.prevent="createEnv" class="grid grid-cols-1 md:grid-cols-6 gap-4 items-end">

        <div class="md:col-span-2">
          <label for="name" class="block text-sm font-medium text-gray-600">NAME</label>
          <input id="name" v-model="newEnv.name" @input="enforceCapitalSnakeCase" type="text" required
            placeholder="E.g., APP_SECRET_KEY"
            class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2 text-sm focus:ring-black focus:border-black" />
        </div>

        <div class="md:col-span-2">
          <label for="value" class="block text-sm font-medium text-gray-600">VALUE</label>
          <input id="value" v-model="newEnv.value" type="text" required placeholder="Value for the variable"
            class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2 text-sm focus:ring-black focus:border-black" />
        </div>

        <div class="md:col-span-1">
          <label class="block text-sm font-medium text-gray-600">&nbsp;</label>
          <div class="mt-1 flex items-center h-[38px]"> <label class="relative inline-flex items-center cursor-pointer">
              <input type="checkbox" v-model="newEnv.isSecret" class="sr-only peer" />
              <div
                class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-gray-300 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:left-0.5] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-black">
              </div>
              <span class="ml-3 text-sm font-medium text-gray-700">{{
                newEnv.isSecret ? 'Secret' : 'Env'
                }}</span>
            </label>
          </div>
        </div>

        <div class="md:col-span-1">
          <button type="submit"
            class="w-full px-4 py-2 bg-black text-white font-semibold rounded-md shadow-sm hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-black transition duration-150 text-sm">
            Save
          </button>
        </div>

      </form>
    </div>

    <div class="bg-white p-6 rounded-lg shadow-md">
      <h2 class="text-xl font-semibold mb-4 text-gray-700">
        Existing Variables ({{ envVariables.length }})
      </h2>

      <div v-if="envVariables.length === 0" class="text-center py-10 text-gray-500">
        No environment variables added yet.
      </div>

      <table v-else class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-1/4">
              Name
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-1/3">
              Value
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-1/6">
              Type
            </th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider w-1/4">
              Actions
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="env in envVariables" :key="env.id" :class="rowClass(env.isSecret)">
            <td class="px-6 py-4 whitespace-nowrap font-mono text-sm font-semibold">
              {{ env.name }}
            </td>

            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
              <template v-if="editingId === env.id">
                <input v-model="editingValue" type="text"
                  class="border border-gray-300 rounded-md shadow-sm p-1 w-full text-sm focus:border-black" />
              </template>
              <template v-else>
                {{ env.isSecret ? maskedValue(env.value) : env.value }}
              </template>
            </td>

            <td class="px-6 py-4 whitespace-nowrap text-sm">
              <span :class="env.isSecret
                ? 'px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-gray-800 text-white'
                : 'px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-gray-200 text-gray-800'
                ">
                {{ env.isSecret ? 'SECRET 🔒' : 'ENV' }}
              </span>
            </td>

            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-2">
              <template v-if="editingId === env.id">
                <button @click="saveEdit" class="text-black hover:text-gray-800 font-semibold">
                  Save
                </button>
                <button @click="cancelEdit" class="text-gray-500 hover:text-black ml-2">
                  Cancel
                </button>
              </template>
              <template v-else>
                <button @click="startEdit(env)" class="text-gray-700 hover:text-black p-1" title="Edit">
                  <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                  </svg>
                </button>
                <button @click="initiateDelete(env)" class="text-gray-700 hover:text-black p-1 ml-2" title="Delete">
                  <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </template>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <DeleteModal :show="showDeleteModal" :variableName="variableNameForModal" :itemType="itemTypeForModal"
      @close="cancelDelete" @confirm="confirmDelete" />
  </div>
</template>