import { ref } from 'vue'
import { defineStore } from 'pinia'
import { useProviderStore } from './providerStore'
import { useEnvVarStore } from './envVarStore'
import { useToastStore } from './toastStore'
import { useAuthStore } from './authStore'

/**
 * Extracts all unique environment variable names referenced in tool headers
 */
const extractRequiredEnvVars = (providersWithHydratedTools) => {
  const requiredVars = new Set()
  const envVarRegex = /\$\{([A-Z0-9_]+)\}/g

  if (providersWithHydratedTools) {
    for (const provider of providersWithHydratedTools) {
      // provider.enabledTools is now an array of *full* tool objects
      if (provider.enabledTools) {
        for (const tool of provider.enabledTools) {
          if (tool && tool.headers) {
            // Check if tool itself is defined
            for (const header of tool.headers) {
              if (header.value) {
                let match
                envVarRegex.lastIndex = 0
                while ((match = envVarRegex.exec(header.value)) !== null) {
                  requiredVars.add(match[1])
                }
              }
            }
          }
        }
      }
    }
  }
  return requiredVars
}

/**
 * Transforms a raw server object from the API (which uses 'tools')
 * into the structure used by the Pinia store (which uses 'enabledTools').
 */
const transformApiServer = (server) => {
  if (!server) return null // Safety check

  // Ensure providers array exists and is an array
  if (!server.providers || !Array.isArray(server.providers)) {
    return { ...server, providers: [] }
  }

  // Transform providers to match app's internal expectation
  const transformedProviders = server.providers.map((provider) => {
    return {
      ...provider,
      enabledTools: provider.tools || [], // Rename key and ensure it's an array
      tools: undefined, // Remove old key (optional, but clean)
    }
  })
  // Ensure requestHeadersProxyEnabled exists, default to false
  return {
    ...server,
    providers: transformedProviders,
    requestHeadersProxyEnabled: server.requestHeadersProxyEnabled || false,
  }
}

export const useMcpserverStore = defineStore('servers', () => {
  // State: Start with empty array
  const servers = ref([])
  const { apiClient } = useAuthStore()
  const toastStore = useToastStore()

  // --- Fetch Action ---
  const fetchMcpservers = async () => {
    try {
      const data = await apiClient('/servers')
      const fetchedServers = data.servers || []
      const transformedServers = fetchedServers.map(transformApiServer)
      servers.value = transformedServers
    } catch {
      console.error('Failed to fetch MCP servers.')
    }
  }

  const loadMcpserverById = async (id) => {
    // 1. Check if it's already in the store (use loose equality for safety)
    let existingIndex = servers.value.findIndex((d) => d.id == id)
    if (existingIndex !== -1) {
      return servers.value[existingIndex] // Already loaded
    }

    // 2. If not, fetch it from the API
    try {
      const data = await apiClient(`/servers/${id}`)
      if (data && data.server) {
        const server = transformApiServer(data.server)

        // 3. --- UPSERT LOGIC ---
        // Check *again* in case fetchMcpservers() finished while we were fetching
        existingIndex = servers.value.findIndex((d) => d.id == server.id)

        if (existingIndex !== -1) {
          // It was loaded by the main fetch. Just update it.
          servers.value[existingIndex] = server
        } else {
          // It's genuinely not in the list. Add it.
          servers.value.push(server)
        }
        return server
      }
    } catch (error) {
      console.error(`Failed to load MCP server ${id}:`, error)
      throw error // Re-throw for the component to handle
    }
  }

  // --- Getters ---

  const getMcpserverById = (id) => {
    const found = servers.value.find((d) => d.id == id)
    return found ? JSON.parse(JSON.stringify(found)) : null
  }

  const hasMcpserverName = (name, excludeId = null) => {
    return servers.value.some(
      (d) => d.name.toLowerCase() === name.toLowerCase() && d.id != excludeId,
    )
  }

  // --- Actions (API-driven) ---

  const saveMcpserver = async (serverData) => {
    // --- Payload creation logic ---
    const enabledProviders = serverData.providers
      .map((provider) => {
        const toolIds = provider.enabledTools.map((tool) => {
          return { id: tool.id }
        })
        return {
          id: provider.id,
          tools: toolIds, // Send 'tools' to the API
        }
      })
      .filter((p) => p.tools.length > 0)

    const payload = {
      name: serverData.name,
      version: serverData.version,
      instructions: serverData.instructions,
      providers: enabledProviders,
      requestHeadersProxyEnabled: serverData.requestHeadersProxyEnabled, // Add new attribute
    }
    const namespacedPayload = { server: payload }
    // --- End of payload creation ---

    if (serverData.id) {
      // === UPDATE ===
      const response = await apiClient(`/servers/${serverData.id}`, {
        method: 'PATCH',
        body: JSON.stringify(namespacedPayload),
      })
      const updatedServer = transformApiServer(response.server) // De-namespace
      const index = servers.value.findIndex((d) => d.id == updatedServer.id)
      if (index !== -1) {
        servers.value[index] = updatedServer
      }
      toastStore.showToast(`MCP Server "${updatedServer.name}" updated.`, 'info', 3000)
      return updatedServer.id
    }

    // === CREATE ===
    const response = await apiClient('/servers', {
      method: 'POST',
      body: JSON.stringify(namespacedPayload),
    })
    const newServer = transformApiServer(response.server) // De-namespace
    servers.value.unshift(newServer)
    toastStore.showToast(`MCP Server "${newServer.name}" created.`, 'info', 3000)
    return newServer.id
  }

  const deleteMcpserver = async (id) => {
    try {
      await apiClient(`/servers/${id}`, { method: 'DELETE' })
      const initialLength = servers.value.length
      servers.value = servers.value.filter((d) => d.id != id)
      toastStore.showToast('MCP Server deleted.', 'info', 3000)
      return initialLength !== servers.value.length
    } catch {
      return false
    }
  }

  /**
   * Creates a new token via the API.
   */
  const createToken = async (serverId, tokenData) => {
    const server = servers.value.find((d) => d.id == serverId)
    if (!server) {
      return { success: false, missingVars: [], error: 'Mcpserver not found.' }
    }

    const providerStore = useProviderStore()

    // Create an array of promises to fetch full provider data
    const providerFetchPromises = server.providers.map((minProvider) => {
      // Use the async getter from the provider store
      return providerStore.getProviderById(minProvider.id)
    })

    // Wait for all async fetches to complete
    const fullProviders = await Promise.all(providerFetchPromises)

    // Now, build the temporary list for validation
    const providersForValidation = server.providers.map((minProvider, index) => {
      // `minProvider` is { id: "...", enabledTools: [{id: "..."}] }
      const fullProvider = fullProviders[index] // Get the corresponding full provider

      if (!fullProvider || !fullProvider.tools) {
        // This provider isn't in the cache or has no tools
        return { enabledTools: [] }
      }

      // For each *enabled* tool ID, find the *full* tool object
      const hydratedEnabledTools = minProvider.enabledTools
        .map((minTool) => {
          // `minTool` is { id: "..." }
          // Find the full tool (which has headers) from the cached full provider
          return fullProvider.tools.find((ep) => ep.id === minTool.id)
        })
        .filter(Boolean) // Filter out any 'undefined'

      // Return the shape that `extractRequiredEnvVars` needs
      return {
        enabledTools: hydratedEnabledTools,
      }
    })

    // Run the check on the *temporary, hydrated* data
    const requiredVarsSet = extractRequiredEnvVars(providersForValidation)
    const requiredVarsArray = Array.from(requiredVarsSet)
    const envVarStore = useEnvVarStore()
    const missingVars = envVarStore.getMissingEnvVars(requiredVarsArray)

    if (missingVars.length > 0) {
      return { success: false, missingVars }
    }

    // 2. Validation successful, call API
    try {
      const namespacedPayload = { token: tokenData }
      const response = await apiClient(`/servers/${serverId}/tokens`, {
        method: 'POST',
        body: JSON.stringify(namespacedPayload),
      })

      const new_token = response.token // De-namespace the token

      if (!server.tokens) {
        server.tokens = []
      }

      server.tokens.unshift(new_token)
      toastStore.showToast(`Token created.`, 'info', 3000)
      return { success: true, token: new_token }
    } catch (error) {
      return { success: false, error: error.message, missingVars: [] }
    }
  }

  return {
    servers,
    fetchMcpservers,
    loadMcpserverById,
    getMcpserverById,
    hasMcpserverName,
    saveMcpserver,
    deleteMcpserver,
    createToken,
  }
})
