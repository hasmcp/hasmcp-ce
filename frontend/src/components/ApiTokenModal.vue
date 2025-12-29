<template>
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-gray-900 bg-opacity-75"
    aria-modal="true"
    role="dialog"
  >
    <div class="w-full max-w-md p-6 bg-white rounded-lg shadow-xl">
      <h2 class="text-2xl font-bold text-gray-800 mb-4">API Access Required</h2>
      <p class="text-gray-600 mb-6">
        Please enter your API access token to connect to the Has MCP backend.
      </p>
      <form @submit.prevent="saveToken">
        <div>
          <label for="apiToken" class="block text-sm font-medium text-gray-700">Access Token</label>
          <input
            v-model="localToken"
            type="password"
            id="apiToken"
            class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
            placeholder="mcp_tk_..."
            required
          />
        </div>
        <div class="mt-6">
          <button
            type="submit"
            class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
          >
            Save and Connect
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '@/stores/authStore'

const authStore = useAuthStore()
const localToken = ref('')

const saveToken = () => {
  if (localToken.value.trim()) {
    authStore.setToken(localToken.value.trim())
  }
}
</script>
