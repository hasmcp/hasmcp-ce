import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { useAuthStore } from './authStore'
import { useToastStore } from './toastStore'

// ----------------------------------------------------
// 1. Core Utilities (Remain the same)
// ----------------------------------------------------

const HTTP_METHODS = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS']
const methodsWithBody = ['POST', 'PUT', 'PATCH']
const PROVIDER_TYPES = ['REST']
const PROVIDER_VISIBILITY = ['INTERNAL', 'PUBLIC']

/**
 * Generates a secret prefix from a URL hostname.
 * e.g., "https://api.stripe.com/v1" -> "API_STRIPE_COM"
 * e.g., "https://www.example.com" -> "EXAMPLE_COM"
 */
const generateSecretPrefix = (url) => {
  if (!url) {
    return ''
  }
  try {
    // new URL() constructor needs a full URL, add https:// if missing protocol
    let fullUrl = url
    if (!fullUrl.match(/^https?:\/\//)) {
      fullUrl = 'https://' + url
    }
    const parsedUrl = new URL(fullUrl)
    let hostname = parsedUrl.hostname
    // Remove 'www.' if it exists
    if (hostname.startsWith('www.')) {
      hostname = hostname.substring(4)
    }
    // Replace dots with underscores and convert to uppercase
    return hostname.replace(/\./g, '_').toUpperCase()
  } catch {
    // Don't log errors for partially typed/invalid URLs
    return ''
  }
}

// ----------------------------------------------------
// 2. Pinia Store Definition
// ----------------------------------------------------

export const useProviderStore = defineStore('providers', () => {
  // State: Start with an empty array, to be filled by API
  const providers = ref([])
  const { apiClient } = useAuthStore()
  const toastStore = useToastStore()

  const getProviderById = async (id) => {
    try {
      const data = await apiClient('/providers/' + id)
      // Handle API response shape {"providers": []}
      return data.provider
    } catch {
      // Error is already handled by apiClient, but we can add specific logic here
      console.error('Failed to fetch providers.')
    }
  }
  /**
   * Fetches all providers from the API and populates the store.
   */
  const fetchProviders = async () => {
    try {
      const data = await apiClient('/providers')
      // Handle API response shape {"providers": []}
      providers.value = data.providers || []
    } catch {
      // Error is already handled by apiClient, but we can add specific logic here
      console.error('Failed to fetch providers.')
    }
  }

  /**
   * Adds a new provider via the API.
   */
  const addProvider = async (providerData) => {
    // ... (Keep local validation)
    if (!providerData.apiType) {
      throw new Error('Provider type is required and immutable.')
    }
    if (!providerData.visibilityType) {
      throw new Error('Visibility type is required and immutable.')
    }

    try {
      // Namespace the payload
      const namespacedPayload = { provider: providerData }

      const newProvider = await apiClient('/providers', {
        method: 'POST',
        body: JSON.stringify(namespacedPayload),
      })
      providers.value.push(newProvider.provider)
      toastStore.showToast(`Provider "${newProvider.provider.name}" created.`, 'info', 3000)
      return newProvider.provider.id
    } catch {
      // Error is handled by apiClient
      return null
    }
  }

  /**
   * Updates core provider details via the API.
   */
  const updateProvider = async (updatedProvider) => {
    try {
      // IMMUTABILITY ENFORCEMENT: Exclude fields
      const updatableData = { ...updatedProvider }
      delete updatableData.apiType
      delete updatableData.visibilityType
      delete updatableData.secretPrefix
      delete updatableData.baseURL
      delete updatableData.tools

      // Namespace the payload
      const namespacedPayload = { provider: updatableData }

      const updatedData = await apiClient(`/providers/${updatedProvider.id}`, {
        method: 'PATCH',
        body: JSON.stringify(namespacedPayload),
      })

      const index = providers.value.findIndex((p) => p.id === updatedProvider.id)
      if (index !== -1) {
        providers.value[index] = updatedData.provider
      }
      toastStore.showToast(`Provider "${updatedData.provider.name}" updated.`, 'info', 3000)
    } catch {
      // Error is handled by apiClient
    }
  }

  /**
   * Deletes a provider via the API.
   */
  const deleteProvider = async (id) => {
    try {
      await apiClient(`/providers/${id}`, { method: 'DELETE' })
      const index = providers.value.findIndex((p) => p.id === id)
      if (index !== -1) {
        providers.value.splice(index, 1)
      }
      toastStore.showToast(`Provider deleted.`, 'info', 3000)
    } catch {
      // Error is handled by apiClient
    }
  }

  /**
   * Updates or adds a single tool within a provider.
   * NOTE: This assumes an tool like '/providers/:id/tools'
   */
  const updateProviderTool = async (providerId, toolData) => {
    const provider = providers.value.find((p) => p.id === providerId)
    if (!provider) return

    // Namespace the payload
    const id = toolData.id
    delete toolData.id
    const namespacedPayload = { tool: toolData }

    try {
      let updatedTool
      // Update existing tool
      updatedTool = await apiClient(`/providers/${providerId}/tools/${id}`, {
        method: 'PATCH',
        body: JSON.stringify(namespacedPayload),
      })
      const index = provider.tools.findIndex((e) => e.id === id)
      if (index !== -1) {
        provider.tools[index] = updatedTool
      }
      // Update provider version from server
    } catch {
      // Error is handled by apiClient
    }
  }

  const createProviderTool = async (providerId, toolData) => {
    const provider = providers.value.find((p) => p.id === providerId)
    if (!provider) return

    // Namespace the payload
    const namespacedPayload = { tool: toolData }

    try {
      let tool

      // Add new tool
      tool = await apiClient(`/providers/${providerId}/tools`, {
        method: 'POST',
        body: JSON.stringify(namespacedPayload),
      })
      provider.tools.unshift(tool)
      // Update provider version from server
      provider.version = tool.providerVersion
      toastStore.showToast(`Tool saved for ${provider.name}.`, 'info', 3000)
    } catch {
      // Error is handled by apiClient
    }
  }

  /**
   * Deletes a provider tool via the API.
   */
  const deleteProviderTool = async (providerId, toolId) => {
    const provider = providers.value.find((p) => p.id === providerId)
    if (!provider) return

    try {
      await apiClient(`/providers/${providerId}/tools/${toolId}`, {
        method: 'DELETE',
      })

      const index = provider.tools.findIndex((e) => e.id === toolId)
      if (index !== -1) {
        provider.tools.splice(index, 1)
      }
      toastStore.showToast(`Tool deleted from ${provider.name}.`, 'info', 3000)
    } catch {
      // Error is handled by apiClient
    }
  }

  // Expose helpers for reuse in component logic
  const helpers = {
    HTTP_METHODS,
    methodsWithBody,
    PROVIDER_TYPES,
    PROVIDER_VISIBILITY,
    generateSecretPrefix,
  }

  return {
    providers,
    fetchProviders,
    getProviderById,
    addProvider,
    updateProvider,
    deleteProvider,
    createProviderTool,
    updateProviderTool,
    deleteProviderTool,
    helpers,
  }
})
