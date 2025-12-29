import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { useAuthStore } from './authStore'
import { useToastStore } from './toastStore'

/**
 * Transforms a raw API variable object into the client-side store format.
 * @param {object} apiVar - The variable object from the API (with `type`)
 * @returns {object} The variable object for the store (with `isSecret`)
 */
const transformApiVar = (apiVar) => {
  if (!apiVar) return null
  return {
    id: apiVar.id,
    name: apiVar.name,
    value: apiVar.value,
    isSecret: apiVar.type === 'SECRET',
  }
}

export const useEnvVarStore = defineStore('envVars', () => {
  // Centralized State
  const envVariables = ref([])
  const { apiClient } = useAuthStore()
  const toastStore = useToastStore()

  // Getters
  const envVarNameSet = computed(() => {
    return new Set(envVariables.value.map((v) => v.name))
  })

  const getEnvVariableByName = (name) => {
    return envVariables.value.find((v) => v.name === name)
  }

  /**
   * Checks if a list of required variable names are present in the store.
   * @param {string[]} requiredVarNames - Array of variable names (e.g., ['VITE_KEY_1', 'VITE_KEY_2']).
   * @returns {string[]} An array of missing variable names.
   */
  const getMissingEnvVars = (requiredVarNames) => {
    const presentNames = envVarNameSet.value
    return requiredVarNames.filter((name) => !presentNames.has(name))
  }

  // Actions

  /**
   * Fetches all env vars from the API.
   */
  const fetchEnvVars = async () => {
    try {
      // Assumes API tool is '/api/v1/variables'
      const data = await apiClient('/variables')

      // Corrected to handle {"variables": []} and transform
      envVariables.value = data.variables.map(transformApiVar).filter(Boolean) || []

    } catch {
      console.error('Failed to fetch env variables.')
    }
  }

  /**
   * Adds a new environment variable via the API.
   * @param {object} newVar - { name: string, value: string, isSecret: boolean }
   * @returns {{ success: boolean, message: string }}
   */
  const addEnvVar = async (newVar) => {
    if (getEnvVariableByName(newVar.name)) {
      const message = `Variable name "${newVar.name}" already exists.`
      toastStore.showToast(message, 'alert', 3000)
      return { success: false, message }
    }

    try {
      // Map the `isSecret` boolean to the `type` string as required by the API.
      const payload = {
        variable: {
          name: newVar.name.trim(),
          value: newVar.value.trim(),
          type: newVar.isSecret ? 'SECRET' : 'ENV',
        },
      }

      const createdVar = await apiClient('/variables', {
        method: 'POST',
        body: JSON.stringify(payload),
      })

      // Transform the raw API response before pushing it to the store
      const clientSideVar = transformApiVar(createdVar.variable)

      if (clientSideVar) {
        envVariables.value.push(clientSideVar)
        const message = `Environment Variable "${clientSideVar.name}" created successfully!`
        toastStore.showToast(message, 'info', 3000)
        return { success: true, message }
      } else {
        throw new Error('Failed to process created variable.')
      }

    } catch (error) {
      // apiClient handles toast
      return { success: false, message: error.message }
    }
  }

  /**
   * Updates the value of an existing environment variable via the API.
   * @param {number} id - The ID of the variable to update.
   * @param {string} newValue - The new value.
   * @returns {{ success: boolean, name: string, message: string }}
   */
  const updateEnvVarValue = async (id, newValue) => {
    try {
      // This payload is correct as it only sends the value
      const payload = {
        variable: {
          value: newValue.trim(),
        },
      }

      await apiClient(`/variables/${id}`, {
        method: 'PATCH',
        body: JSON.stringify(payload),
      })

      const variable = envVariables.value.find((v) => v.id === id)
      if (variable) {
        variable.value = payload.variable.value
      }
      const message = `Environment Variable "${variable.name}" updated successfully!`
      toastStore.showToast(message, 'info', 3000)
      return { success: true, name: variable.name, message }
    } catch (error) {
      return { success: false, message: error.message }
    }
  }

  /**
   * Deletes an environment variable by ID via the API.
   * @param {number} id - The ID of the variable to delete.
   * @returns {boolean} True if deleted, false otherwise.
   */
  const deleteEnvVar = async (id) => {
    try {
      await apiClient(`/variables/${id}`, { method: 'DELETE' })
      const initialLength = envVariables.value.length
      envVariables.value = envVariables.value.filter((v) => v.id !== id)
      toastStore.showToast('Environment variable deleted.', 'info', 3000)
      return initialLength !== envVariables.value.length
    } catch {
      return false
    }
  }

  return {
    envVariables,
    fetchEnvVars,
    getEnvVariableByName,
    getMissingEnvVars,
    addEnvVar,
    updateEnvVarValue,
    deleteEnvVar,
  }
})