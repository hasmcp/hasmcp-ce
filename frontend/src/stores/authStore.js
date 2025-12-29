import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { useToastStore } from './toastStore'

const getStoredToken = () => {
  if (typeof window !== 'undefined' && window.localStorage) {
    return window.localStorage.getItem('apiToken')
  }
  return ''
}

const setStoredToken = (token) => {
  if (typeof window !== 'undefined' && window.localStorage) {
    window.localStorage.setItem('apiToken', token)
  }
}

const removeStoredToken = () => {
  if (typeof window !== 'undefined' && window.localStorage) {
    window.localStorage.removeItem('apiToken')
  }
}

export const useAuthStore = defineStore('auth', () => {
  // Use the safe getter
  const token = ref(getStoredToken() || '')

  const port = process.env.NODE_ENV === 'development' ? ':8887' : (location.port === '' ? '' : ':' + location.port);
  const apiRootUrl = location.protocol + '//' + location.hostname + port
  const apiBaseUrl = apiRootUrl + '/api/v1'

  const toastStore = useToastStore()

  const isAuthenticated = computed(() => token.value !== '')

  /**
   * Saves the API token to state and localStorage.
   * @param {string} newToken
   */
  function setToken(newToken) {
    token.value = newToken
    setStoredToken(newToken) // Use the safe setter
  }

  /**
   * Clears the API token from state and localStorage.
   */
  function clearToken() {
    token.value = ''
    removeStoredToken() // Use the safe remover
    toastStore.showToast('Authentication token cleared.', 'info', 3000)
  }

  /**
   * A replacer function for JSON.stringify to omit keys with null values.
   * JSON.stringify already omits keys with undefined values.
   * @param {string} key The key being serialized.
   * @param {any} value The value of the key.
   * @returns {any} The original value, or undefined if the value was null.
   */
  const omitNullReplacer = (key, value) => {
    if (value === null) {
      return undefined // JSON.stringify will omit keys with undefined values
    }
    return value
  }

  /**
   * A centralized API client for making authenticated requests.
   * @param {string} tool - The API tool (e.g., '/providers')
   * @param {object} options - Standard fetch options (method, body, etc.)
   * @returns {Promise<any>} - The JSON response
   */
  const apiClient = async (tool, options = {}) => {
    const headers = {
      'Content-Type': 'application/json',
      ...options.headers,
    }

    // Add the auth token if it exists
    if (token.value) {
      headers['Authorization'] = `Bearer ${token.value}`
    }

    if (options.body && typeof options.body === 'string') {
      try {
        const parsedBody = JSON.parse(options.body)
        options.body = JSON.stringify(parsedBody, omitNullReplacer)
      } catch (e) {
        console.warn('apiClient: Request body was not valid JSON, sending as-is.', e)
      }
    }

    try {
      const response = await fetch(`${apiBaseUrl}${tool}`, {
        ...options,
        headers,
      })

      if (!response.ok) {
        if (response.status === 401) {
          clearToken()
        }
        const errorData = await response.json().catch(() => ({ message: response.statusText }))
        throw new Error(errorData.message || `API request failed with status ${response.status}`)
      }

      if (response.status === 204) {
        // No Content
        return null
      }
      return response.json()
    } catch (error) {
      console.error('API Client Error:', error)
      toastStore.showToast(error.message, 'alert', 4000)
      throw error // Re-throw for the caller to handle
    }
  }

  return { token, isAuthenticated, setToken, clearToken, apiClient, apiBaseUrl, apiRootUrl }
})
