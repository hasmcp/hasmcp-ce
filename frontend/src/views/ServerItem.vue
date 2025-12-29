<script setup>
import { ref, computed, onMounted, watchEffect, watch, toRaw } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMcpserverStore } from '../stores/serverStore'
import { useProviderStore } from '../stores/providerStore'
import { useToastStore } from '../stores/toastStore'
import { useAuthStore } from '../stores/authStore'

import DeleteModal from '../components/DeleteModal.vue'

const route = useRoute()
const router = useRouter()
const serverStore = useMcpserverStore()
const providerStore = useProviderStore()
const toastStore = useToastStore()
const authStore = useAuthStore()

// --- State ---
const isCreateMode = computed(() => route.name === 'ServerCreate')
const serverId = computed(() => route.params.id)

// Local state for creation/editing, holds the draft of the server
const serverDraft = ref(null)

// Computed: Fetches live data from the store (used for initialization and reading)
const currentMcpserver = computed(() => {
  if (isCreateMode.value) return null
  return serverStore.getMcpserverById(serverId.value)
})

// Computed: The server data used by the UI (live if not editing, draft if editing)
const server = computed(() => {
  if (isEditingDetails.value || isCreateMode.value) {
    return serverDraft.value
  }
  return currentMcpserver.value
})

// UI State
const toolSelectionOpen = ref({})
const isEditingDetails = ref(false)
const providerSearchQuery = ref('')
const activeConfigTab = ref('common');

const fullProviderDataCache = ref({})
const providerLoadingState = ref({})

// --- Helper: Formats a Date object for datetime-local input ---
const formatAsDateTimeLocal = (date) => {
  const pad = (num) => String(num).padStart(2, '0')
  const year = date.getFullYear()
  const month = pad(date.getMonth() + 1) // Months are 0-indexed
  const day = pad(date.getDate())
  const hours = pad(date.getHours())
  const minutes = pad(date.getMinutes())
  return `${year}-${month}-${day}T${hours}:${minutes}`
}

const isCreatingToken = ref(false)
const oneYearFromNow = new Date()
oneYearFromNow.setFullYear(oneYearFromNow.getFullYear() + 1)
const newTokenExpiresAt = ref(formatAsDateTimeLocal(oneYearFromNow))
const newlyCreatedToken = ref(null)
const missingEnvVars = ref([])

const activeExpiryPreset = ref(defaultExpiry)
const expiryPresets = [
  { label: '1h', value: '1h' },
  { label: '3h', value: '3h' },
  { label: '6h', value: '6h' },
  { label: '24h', value: '24h' },
  { label: '3d', value: '3d' },
  { label: '7d', value: '7d' },
  { label: '30d', value: '30d' },
  { label: '90d', value: '90d' },
  { label: '1y', value: '1y' },
  { label: 'Custom', value: 'custom' },
]
// -----------------------

// Modal State
const showDeleteModal = ref(false)
const itemToDelete = ref(null)
const deleteNameForModal = ref('')
const itemTypeForModal = ref('')

// --- Computed Properties ---

const providers = computed(() => providerStore.providers)

const searchedProviders = computed(() => {
  let list = []
  const query = providerSearchQuery.value.trim().toLowerCase()
  const selectedProviderIds = new Set(server.value?.providers?.map((p) => p.id) || [])

  if (query) {
    const matchingProviders = providers.value.filter(
      (p) =>
        (p.name && p.name.toLowerCase().includes(query)) ||
        (p.description && p.description.toLowerCase().includes(query)),
    )
    list.push(...matchingProviders)
  }

  for (const selectedId of selectedProviderIds) {
    if (!list.some((p) => p.id === selectedId)) {
      const selectedProvider = providers.value.find((p) => p.id === selectedId)
      if (selectedProvider) {
        list.push(selectedProvider)
      }
    }
  }

  const uniqueProviders = Array.from(new Set(list.map((p) => p.id))).map((id) =>
    list.find((p) => p.id === id),
  )

  return uniqueProviders.sort((a, b) => a.id - b.id)
})

const enabledToolsMap = computed(() => {
  return (
    server.value?.providers?.reduce((acc, depProvider) => {
      acc[depProvider.id] = depProvider.enabledTools.reduce((eAcc, tool) => {
        eAcc[tool.id] = true
        return eAcc
      }, {})
      return acc
    }, {}) || {}
  )
})

const mcpBaseURL = window.location.protocol + '//' + window.location.hostname;
const portAddr = process.env.NODE_ENV === 'development' ? ':8887' : (location.port === '' ? '' : ':' + location.port);
const mcpConfigJson = computed(() => {
  if (!server.value) return ''
  const mcpUrl = `${mcpBaseURL}${portAddr}/mcp/${server.value.id}`

  const tokenValue = newlyCreatedToken.value
    ? newlyCreatedToken.value.value
    : '<YOUR_TOKEN_VALUE_FROM_ABOVE_TOKEN>'

  const config = {
    mcpServers: {
      [serverKeyName.value]: {
        url: mcpUrl,
        headers: {
          'x-hasmcp-key': `Bearer ${tokenValue}`, // Use new token
        },
      },
    },
  }
  return JSON.stringify(config, null, 2)
})

const mcpConfigJsonUsingRemoteMcp = computed(() => {
  if (!server.value) return ''
  const mcpURL = `${mcpBaseURL}${portAddr}/mcp/${server.value.id}`

  const tokenValue = newlyCreatedToken.value
    ? newlyCreatedToken.value.value
    : 'YOUR_TOKEN_VALUE_FROM_ABOVE_TOKEN'

  const config = {
    mcpServers: {
      [serverKeyName.value]: {
        command: 'npx',
        args: ['mcp-remote', mcpURL, '--header', 'x-hasmcp-key: Bearer ${HASMCP_MCP_ACCESS_TOKEN}'],
        env: {
          HASMCP_MCP_ACCESS_TOKEN: tokenValue,
        },
      },
    },
  }

  return JSON.stringify(config, null, 2)
})

const mcpConfigJsonGeminiCli = computed(() => {
  if (!server.value) return ''
  const mcpUrl = `${mcpBaseURL}${portAddr}/mcp/${server.value.id}`

  const tokenValue = newlyCreatedToken.value
    ? newlyCreatedToken.value.value
    : '<YOUR_TOKEN_VALUE_FROM_ABOVE_TOKEN>'

  const config = {
    mcpServers: {
      [serverKeyName.value]: {
        httpUrl: mcpUrl,
        headers: {
          'x-hasmcp-key': `Bearer ${tokenValue}`, // Use new token
        },
      },
    },
  }
  return JSON.stringify(config, null, 2)
})

const selectedProvider = computed(() => {
  if (!server.value?.providers?.length) return null
  const providerId = server.value.providers[0].id
  return fullProviderDataCache.value[providerId] || providers.value.find((p) => p.id === providerId)
})

const canAuthorize = computed(() => {
  console.log(selectedProvider.value)
  return !!selectedProvider.value?.oauth2Config?.clientID
})

const authorizeUrl = computed(() => {
  if (!server.value) return '#'
  const apiRootUrl = new URL(authStore.apiRootUrl).origin
  return `${apiRootUrl}/oauth2/authorize?server_id=${server.value.id}`
})

const tokenDisplayValue = computed(() => {
  if (!newlyCreatedToken.value) return null
  return `${newlyCreatedToken.value.value}`
})

// --- Methods: Token Management ---

const defaultExpiry = '24h'

// Set expiry from preset button
const setExpiry = (presetValue) => {
  activeExpiryPreset.value = presetValue
  if (presetValue === 'custom') {
    return
  }

  const now = new Date()
  switch (presetValue) {
    case '1h':
      now.setHours(now.getHours() + 1)
      break
    case '3h':
      now.setHours(now.getHours() + 3)
      break
    case '6h':
      now.setHours(now.getHours() + 6)
      break
    case '24h':
      now.setHours(now.getHours() + 24)
      break
    case '3d':
      now.setDate(now.getDate() + 3)
      break
    case '7d':
      now.setDate(now.getDate() + 7)
      break
    case '30d':
      now.setDate(now.getDate() + 30)
      break
    case '90d':
      now.setDate(now.getDate() + 90)
      break
    case '1y':
      now.setFullYear(now.getFullYear() + 1)
      break
  }
  newTokenExpiresAt.value = formatAsDateTimeLocal(now)
}

// Set preset to 'custom' if user manually changes the date
const onCustomDateChange = () => {
  activeExpiryPreset.value = 'custom'
}

const resetTokenForm = () => {
  isCreatingToken.value = false
  // Reset expiry to a year from now for UX
  const oneYearFromNow = new Date()
  oneYearFromNow.setFullYear(oneYearFromNow.getFullYear() + 1)
  newTokenExpiresAt.value = formatAsDateTimeLocal(oneYearFromNow)
  activeExpiryPreset.value = defaultExpiry // Reset active preset
  missingEnvVars.value = []
}

const startTokenCreation = () => {
  if (isCreatingToken.value) {
    resetTokenForm()
    newlyCreatedToken.value = null // Hide token display on cancel
  } else {
    if (isCreateMode.value || isEditingDetails.value) {
      toastStore.showToast(
        'Please save the MCP Server and exit edit mode before generating a token.',
        'warning',
        3000,
      )
      return
    }
    resetTokenForm()
    isCreatingToken.value = true
    newlyCreatedToken.value = null // Clear previous token
  }
}

const createToken = async () => {
  if (!newTokenExpiresAt.value) {
    toastStore.showToast('Expiration date is required!', 'warning', 3000)
    return
  }

  const targetId = server.value?.id
  if (!targetId) {
    toastStore.showToast('MCP Server ID not found. Cannot create token.', 'warning', 3000)
    return
  }

  try {
    const result = await serverStore.createToken(targetId, {
      expiresAt: new Date(newTokenExpiresAt.value).toISOString(),
    })

    if (!result.success) {
      missingEnvVars.value = result.missingVars || []
      if (missingEnvVars.value.length > 0) {
        toastStore.showToast(
          'Token creation failed. Missing required environment variables.',
          'alert',
          5000,
        )
      } else {
        toastStore.showToast(result.error || 'Failed to generate token.', 'alert', 3000)
      }
      isCreatingToken.value = true
      return
    }

    newlyCreatedToken.value = JSON.parse(JSON.stringify(result.token))
    missingEnvVars.value = []
    isCreatingToken.value = false
  } catch (error) {
    toastStore.showToast(error.message || 'An unexpected error occurred.', 'alert', 3000)
  }
}

const copyFullToken = () => {
  if (!newlyCreatedToken.value) return
  const fullToken = tokenDisplayValue.value

  if (navigator.clipboard) {
    navigator.clipboard
      .writeText(fullToken)
      .then(() => {
        toastStore.showToast(`Full Token copied to clipboard. It is now hidden.`, 'info', 4000)
        newlyCreatedToken.value = false
      })
      .catch((e) => {
        toastStore.showToast(
          `Failed to automatically copy: ${e.message}. Please copy the token text manually.`,
          'warning',
          6000,
        )
      })
  } else {
    toastStore.showToast(
      'Clipboard API not available. Please copy the token value manually.',
      'warning',
      4000,
    )
  }
}

const deleteTokenHandler = async (token) => {
  const isDeleted = await serverStore.deleteToken(server.value.id, token.id)
  if (isDeleted) {
    if (newlyCreatedToken.value && newlyCreatedToken.value.id === token.id) {
      newlyCreatedToken.value = false
    }
  }
}

watch(
  () => server.value?.providers,
  async (serverProviders) => {
    if (!serverProviders) return
    for (const p of serverProviders) {
      // If we don't have the full data (tools, oauthConfig) in cache, fetch it
      if (!fullProviderDataCache.value[p.id]) {
        try {
          const fullData = await providerStore.getProviderById(p.id)
          if (fullData) {
            fullProviderDataCache.value[p.id] = fullData
          }
        } catch (e) {
          console.error(`Failed to fetch full details for provider ${p.id}`, e)
        }
      }
    }
  },
  { immediate: true, deep: true },
)

// --- Lifecycle & Initialization ---

const initForm = () => {
  // This function now safely relies on currentMcpserver (which is a computed)
  // watchEffect will ensure this runs *after* currentMcpserver is populated
  const newMcpserverData = currentMcpserver.value

  if (isCreateMode.value) {
    isEditingDetails.value = true
    // Only set draft if it's not already set (to avoid wiping form on re-renders)
    if (!serverDraft.value) {
      serverDraft.value = {
        id: null,
        name: '',
        instructions: '',
        providers: [],
        tokens: [],
        version: 1,
        requestHeadersProxyEnabled: false, // Add new default
      }
    }
  } else if (newMcpserverData) {
    // Only update draft if not in edit mode (prevents overwriting user's edits)
    if (!isEditingDetails.value) {
      serverDraft.value = JSON.parse(JSON.stringify(newMcpserverData))
      isEditingDetails.value = false
    }
  } else if (serverId.value) {
    // Data is still loading or not found, draft will be null
    serverDraft.value = null
  }

  if (!isCreateMode.value) {
    resetTokenForm()
  }
}

// Use watchEffect to reactively initialize or reset the form
watchEffect(() => {
  initForm()
})

// Use onMounted to ensure data is loaded on hard refresh
onMounted(async () => {
  const server = serverStore.getMcpserverById(serverId.value)
  if (!server && !isCreateMode.value && serverId.value) {
    try {
      await serverStore.loadMcpserverById(serverId.value)
      // After this, currentMcpserver will update, and the watchEffect will run initForm()
    } catch (e) {
      console.error('Failed to load MCP server data', e)
      router.push({ name: 'Servers' })
      toastStore.showToast('MCP Server not found.', 'warning', 2000)
    }
  }
})

// --- Methods: Mcpserver CRUD ---

const saveMcpserver = async () => {
  if (!serverDraft.value?.name?.trim()) {
    toastStore.showToast('MCP Server Name is required!', 'warning', 3000)
    return
  }
  const dataToSave = JSON.parse(JSON.stringify(toRaw(serverDraft.value)))
  dataToSave.providers = dataToSave.providers.filter((p) => p.enabledTools.length > 0)

  if (dataToSave.providers.length === 0) {
    toastStore.showToast(
      'Please select and enable at least one tool from one provider.',
      'warning',
      3000,
    )
    return
  }

  if (dataToSave.providers.length > 1) {
    toastStore.showToast('Only 1 provider per MCP server is allowed.', 'warning', 3000)
    return
  }

  if (serverStore.hasMcpserverName(dataToSave.name, dataToSave.id)) {
    toastStore.showToast(`A MCP Server named "${dataToSave.name}" already exists.`, 'warning', 3000)
    return
  }

  try {
    const newId = await serverStore.saveMcpserver(dataToSave)
    if (isCreateMode.value) {
      router.replace({ name: 'ServerDetail', params: { id: newId } })
    } else {
      isEditingDetails.value = false
    }
  } catch (error) {
    console.error('Failed to save MCP server:', error)
  }
}

const cancelEdit = () => {
  if (isCreateMode.value) {
    router.push({ name: 'Servers' })
  } else {
    isEditingDetails.value = false
    serverDraft.value = JSON.parse(JSON.stringify(currentMcpserver.value)) // Reset draft
  }
}

// --- Methods: Tool Selection ---

const getProviderLiveTools = (providerId) => {
  return fullProviderDataCache.value[providerId]?.tools || []
}

const toggleProviderAccordion = async (provider) => {
  const providerId = provider.id
  const isOpening = !toolSelectionOpen.value[providerId]

  if (isOpening && !fullProviderDataCache.value[providerId]) {
    providerLoadingState.value[providerId] = true
    try {
      const fullProvider = await providerStore.getProviderById(providerId)
      if (fullProvider) {
        fullProviderDataCache.value[providerId] = fullProvider
      } else {
        toastStore.showToast(
          `Could not find provider details for ${provider.name}`,
          'warning',
          3000,
        )
      }
    } catch (error) {
      toastStore.showToast(`Failed to load tools for ${provider.name}`, 'warning', 3000)
      console.error(error)
    } finally {
      providerLoadingState.value[providerId] = false
    }
  }
  toolSelectionOpen.value[providerId] = !toolSelectionOpen.value[providerId]
}

const toggleTool = (providerId, tool) => {
  if (!tool || !tool.id) return
  let depProvider = serverDraft.value.providers.find((p) => p.id === providerId)

  if (!depProvider) {
    // Allow only 1 provider per server
    if (serverDraft.value.providers.length > 1) {
      toastStore.showToast('Only 1 provider per MCP server is allowed.', 'warning', 3000)
      return
    }

    const liveProvider = fullProviderDataCache.value[providerId]
    depProvider = {
      id: providerId,
      version: liveProvider?.version || 1,
      enabledTools: [],
    }
    serverDraft.value.providers.push(depProvider)
  }

  const enabledTools = depProvider.enabledTools
  const index = enabledTools.findIndex((e) => e.id === tool.id)

  if (index !== -1) {
    enabledTools.splice(index, 1)
  } else {
    enabledTools.push(tool)
  }

  if (enabledTools.length === 0) {
    const providerIndex = serverDraft.value.providers.findIndex((p) => p.id === providerId)
    if (providerIndex !== -1) {
      serverDraft.value.providers.splice(providerIndex, 1)
    }
  }
}

const isToolEnabled = (providerId, tool) => {
  return enabledToolsMap.value[providerId] && enabledToolsMap.value[providerId][tool.id]
}

// --- Helper for "Enable/Disable All" button ---
const areAllToolsEnabled = (providerId) => {
  const liveTools = getProviderLiveTools(providerId)
  if (liveTools.length === 0) return false // Can't disable "all" if there are none
  return liveTools.every((ep) => isToolEnabled(providerId, ep))
}

// --- "Enable/Disable All" logic ---
const toggleAllTools = (provider) => {
  const providerId = provider.id
  const liveTools = getProviderLiveTools(providerId)
  if (liveTools.length === 0) return // Nothing to do

  const allEnabled = areAllToolsEnabled(providerId)
  let depProvider = serverDraft.value.providers.find((p) => p.id === providerId)
  const providerIndex = serverDraft.value.providers.findIndex((p) => p.id === providerId)

  if (allEnabled) {
    // --- DISABLE ALL ---
    // If the provider is in the draft, remove it
    if (providerIndex !== -1) {
      serverDraft.value.providers.splice(providerIndex, 1)
    }
  } else {
    // --- ENABLE ALL ---
    if (!depProvider) {
      // Allow only 1 provider per server
      if (serverDraft.value.providers.length > 1) {
        toastStore.showToast('Only 1 provider per MCP server is allowed.', 'warning', 3000)
        return
      }

      // Provider isn't in the draft, so add it
      const liveProvider = fullProviderDataCache.value[providerId]
      depProvider = {
        id: providerId,
        version: liveProvider?.version || 1,
        enabledTools: [], // Will be filled next
      }
      serverDraft.value.providers.push(depProvider)
    }
    // Set enabledTools to a deep copy of all live tools
    depProvider.enabledTools = JSON.parse(JSON.stringify(liveTools))
  }
}

const getMethodTextColor = (method) => {
  const colors = {
    GET: 'text-green-600',
    POST: 'text-blue-600',
    PUT: 'text-yellow-600',
    DELETE: 'text-red-600',
    PATCH: 'text-orange-600',
    HEAD: 'text-gray-500',
    OPTIONS: 'text-indigo-500',
  }
  return colors[method] || 'text-gray-400'
}

const copyMcpConfig = () => {
  const config = mcpConfigJson.value
  if (navigator.clipboard) {
    navigator.clipboard
      .writeText(config)
      .then(() => {
        toastStore.showToast(`MCP Server config copied to clipboard.`, 'info', 3000)
      })
      .catch((e) => {
        toastStore.showToast(
          `Failed to automatically copy: ${e.message}. Please copy the config text manually.`,
          'warning',
          6000,
        )
      })
  } else {
    toastStore.showToast(
      'Clipboard API not available. Please copy the config value manually.',
      'warning',
      4000,
    )
  }
}

const copyMcpConfigRemoteMcp = () => {
  const config = mcpConfigJsonUsingRemoteMcp.value
  if (navigator.clipboard) {
    navigator.clipboard
      .writeText(config)
      .then(() => {
        toastStore.showToast(
          `MCP Server config using remote mcp copied to clipboard.`,
          'info',
          3000,
        )
      })
      .catch((e) => {
        toastStore.showToast(
          `Failed to automatically copy: ${e.message}. Please copy the config text manually.`,
          'warning',
          6000,
        )
      })
  } else {
    toastStore.showToast(
      'Clipboard API not available. Please copy the config value manually.',
      'warning',
      4000,
    )
  }
}

const copyMcpConfigGeminiCli = () => {
  const config = mcpConfigJsonGeminiCli.value
  if (navigator.clipboard) {
    navigator.clipboard
      .writeText(config)
      .then(() => {
        toastStore.showToast(`MCP Server config for gemini-cli copied to clipboard.`, 'info', 3000)
      })
      .catch((e) => {
        toastStore.showToast(
          `Failed to automatically copy: ${e.message}. Please copy the config text manually.`,
          'warning',
          6000,
        )
      })
  } else {
    toastStore.showToast(
      'Clipboard API not available. Please copy the config value manually.',
      'warning',
      4000,
    )
  }
}

// --- Methods: General Modal Handlers ---

const initiateDelete = (item, type) => {
  itemToDelete.value = { id: item.id, type, name: item.name }
  deleteNameForModal.value = item.name || item.id
  itemTypeForModal.value = type === 'server' ? 'MCP Server' : 'Token'
  showDeleteModal.value = true
}

const confirmDeleteModal = async () => {
  if (!itemToDelete.value) return

  if (itemToDelete.value.type === 'server') {
    await serverStore.deleteMcpserver(itemToDelete.value.id)
    toastStore.showToast(`MCP Server "${deleteNameForModal.value}" deleted.`, 'info', 2000)
    router.push({ name: 'Servers' })
  } else if (itemToDelete.value.type === 'token') {
    const tokenStub = { id: itemToDelete.value.id, name: itemToDelete.value.name }
    await deleteTokenHandler(tokenStub)
  }
  cancelDeleteModal()
}

const formatServerName = (event) => {
  let value = event.target.value
  // Enforce alphanumeric and max length 32
  value = value.replace(/[^a-zA-Z0-9]/g, '').slice(0, 16)

  serverDraft.value.name = value

  // Ensure input display matches sanitized value
  if (event.target.value !== value) {
    event.target.value = value
  }
}

const cancelDeleteModal = () => {
  showDeleteModal.value = false
  itemToDelete.value = null
  deleteNameForModal.value = ''
  itemTypeForModal.value = ''
}

const serverKeyName = computed(() => {
  if (!server.value) return 'mcpServer'

  // Get the first provider info
  const providerInfo = server.value.providers?.[0]
  if (!providerInfo) return `hasmcp_${server.value.id}`

  // Determine provider name (check store or draft/cache)
  let name = providerInfo.name
  if (!name) {
    const fromStore = providerStore.providers.find((p) => p.id === providerInfo.id)
    if (fromStore) name = fromStore.name
  }
  if (!name && fullProviderDataCache.value[providerInfo.id]) {
    name = fullProviderDataCache.value[providerInfo.id].name
  }

  if (!name) return `hasmcp_${server.value.id}`

  const processed = name.toLowerCase().substring(0, 16)
  return processed || `hasmcp_${server.value.id}`
})
</script>

<template>
  <div class="p-4">
    <div v-if="!server" class="text-center py-10 text-gray-500">Loading MCP Server data...</div>

    <div v-else>
      <div class="bg-white p-6 rounded-xl shadow-2xl mb-8">
        <div class="flex justify-between items-start mb-4 border-b pb-4">
          <h1 class="text-3xl font-bold text-gray-900">
            {{ isCreateMode ? 'Create New MCP Server' : server.name }}
            <span v-if="!isCreateMode" class="text-sm font-semibold px-2 py-1 rounded-full bg-black text-white ml-2">v{{
              server.version }}</span>
          </h1>
          <div v-if="!isCreateMode" class="flex space-x-3">
            <a v-if="canAuthorize" :href="authorizeUrl" :class="[
              'px-4 py-2 text-sm font-semibold rounded-lg shadow-md transition duration-150  bg-black hover:bg-gray-800 text-white flex items-center',
              isEditingDetails ? 'opacity-50 pointer-events-none' : ''
            ]">
              Authorize
            </a>
            <button @click="startTokenCreation" :disabled="isEditingDetails" :class="[
              'px-4 py-2 text-sm font-semibold rounded-lg shadow-md transition duration-150 disabled:opacity-50',
              isCreatingToken
                ? 'bg-red-500 hover:bg-red-600 text-white'
                : 'bg-black hover:bg-gray-800 text-white',
            ]">
              {{ isCreatingToken ? 'Cancel' : 'Generate Token' }}
            </button>
            <router-link :to="{ name: 'ServerLogs', params: { id: server.id } }"
              class="p-2 rounded-lg text-gray-700 hover:text-black transition duration-150 border hover:bg-gray-50"
              title="View MCP Server Logs">
              <svg class="w-6 h-6" viewBox="0 0 116.5 122.88">
                <path class="cls-1"
                  d="M17.88,22.75a2.19,2.19,0,0,1,3.05.6L22,24.66l3.84-4.87a2.2,2.2,0,1,1,3.4,2.78L23.6,29.66a2.74,2.74,0,0,1-.52.5A2.21,2.21,0,0,1,20,29.55L17.28,25.8a2.21,2.21,0,0,1,.6-3.05ZM81.13,59a27.86,27.86,0,0,1,23.31,43.1l12.06,13.14-8.31,7.6L96.56,110.09A27.86,27.86,0,1,1,81.13,59ZM38.47,71.54a3.07,3.07,0,0,1-2.9-3.17,3,3,0,0,1,2.9-3.17h9a3.07,3.07,0,0,1,2.9,3.17,3,3,0,0,1-2.9,3.17ZM93,44.89c-.56,2.11-5.31,2.43-6.38,0V7.43a1.06,1.06,0,0,0-.3-.76,1.08,1.08,0,0,0-.75-.3H7.39a1,1,0,0,0-1,1.06V95.74a1,1,0,0,0,1,1.05H37.72c3.21.34,3.3,5.88,0,6.38H7.43A7.48,7.48,0,0,1,0,95.74V7.43A7.3,7.3,0,0,1,2.19,2.19,7.35,7.35,0,0,1,7.43,0H85.6a7.32,7.32,0,0,1,5.24,2.19A7.39,7.39,0,0,1,93,7.43c0,36.56,0-18,0,37.46ZM38.44,27.47a3.07,3.07,0,0,1-2.91-3.17,3,3,0,0,1,2.91-3.17H68.21a3.07,3.07,0,0,1,2.91,3.17,3,3,0,0,1-2.91,3.17Zm0,22a3.06,3.06,0,0,1-2.91-3.16,3,3,0,0,1,2.91-3.17H68.21a3.07,3.07,0,0,1,2.91,3.17,3,3,0,0,1-2.91,3.16Zm32.19,40a3.4,3.4,0,0,1-.38-.49,3.71,3.71,0,0,1-.29-.56A3.54,3.54,0,0,1,75.05,84a2.78,2.78,0,0,1,.56.41l0,0c1,.93,1.28,1.12,2.36,2.08l.92.83,7.58-8.13c3.21-3.3,8.32,1.53,5.12,4.9L82.15,94.26l-.47.5a3.56,3.56,0,0,1-5,.22l0,0L76,94.28c-.58-.52-1.18-1-1.79-1.57-1.4-1.22-2.22-1.89-3.54-3.21ZM81.15,64.85A22.17,22.17,0,1,1,59,87,22.17,22.17,0,0,1,81.15,64.85ZM23.54,63.59a5.1,5.1,0,1,1-5.09,5.09,5.09,5.09,0,0,1,5.09-5.09ZM25.66,42a2.09,2.09,0,0,1,3,0,2.12,2.12,0,0,1,0,3l-2.07,2.13,2.07,2.13a2.1,2.1,0,0,1-3,3l-2.05-2.1-2.07,2.11a2.07,2.07,0,0,1-3,0,2.13,2.13,0,0,1,0-3l2.08-2.13L18.57,45a2.1,2.1,0,0,1,0-3,2.07,2.07,0,0,1,2.94,0l2.06,2.11L25.66,42Z" />
              </svg>
            </router-link>
            <button @click="isEditingDetails = !isEditingDetails"
              class="p-2 rounded-lg text-gray-700 hover:text-black transition duration-150 border hover:bg-gray-50"
              title="Toggle Edit Mode">
              <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
              </svg>
            </button>

            <button @click="initiateDelete(server, 'server')"
              class="p-2 rounded-lg text-red-600 hover:text-red-800 transition duration-150 border hover:bg-red-50"
              title="Delete MCP Server">
              <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </div>
        </div>

        <div v-if="!isCreateMode && server" class="mb-8">
          <div v-if="missingEnvVars.length > 0"
            class="bg-red-50 border-l-4 border-red-500 p-6 rounded-lg shadow-xl mb-6">
            <h3 class="text-xl font-bold text-red-800 mb-2">Missing Environment Variables!</h3>
            <p class="text-sm text-red-700 mb-4 font-semibold">
              The following environment variables are required by the enabled tools and must be
              added via the
              <router-link to="/variables"
                class="text-blue-900 underline hover:text-blue-700 font-bold transition duration-150">
                "Env Variables"
              </router-link>
              page before a token can be created:
            </p>
            <ul class="list-disc list-inside bg-red-100 p-3 rounded-md text-sm font-mono text-red-800">
              <li v-for="varName in missingEnvVars" :key="varName">{{ varName }}</li>
            </ul>
          </div>

          <div v-if="newlyCreatedToken" class="bg-yellow-50 border-l-4 border-yellow-500 p-6 rounded-lg shadow-xl mb-6">
            <h3 class="text-xl font-bold text-yellow-800 mb-2">Token Generated Successfully!</h3>
            <p class="text-sm text-yellow-700 mb-4 font-semibold">
              Copy this value now. It will be hidden permanently once copied or dismissed.
            </p>
            <div class="flex items-center bg-gray-900 rounded-lg p-3">
              <code class="text-white text-sm font-mono break-all grow pr-4">{{
                tokenDisplayValue
              }}</code>
              <button @click="copyFullToken"
                class="shrink-0 px-4 py-2 bg-green-500 text-white font-semibold rounded-md hover:bg-green-600 transition duration-150">
                Copy Token
              </button>
              <button @click="newlyCreatedToken = false"
                class="ml-2 shrink-0 px-4 py-2 bg-gray-500 text-white font-semibold rounded-md hover:bg-gray-600 transition duration-150">
                Dismiss
              </button>
            </div>
          </div>

          <div v-if="isCreatingToken" class="bg-gray-100 p-6 rounded-xl shadow-inner mb-6">
            <div class="mb-4">
              <label class="block text-sm font-medium text-gray-700 mb-2">Set Expiry</label>
              <div class="flex flex-wrap gap-2">
                <button v-for="preset in expiryPresets" :key="preset.value" type="button"
                  @click="setExpiry(preset.value)" :class="[
                    'px-3 py-1 text-xs font-semibold rounded-full border transition-all duration-150',
                    activeExpiryPreset === preset.value
                      ? 'bg-black text-white border-black'
                      : 'bg-white border-gray-300 text-gray-700 hover:bg-gray-50',
                  ]">
                  {{ preset.label }}
                </button>
              </div>
            </div>

            <form @submit.prevent="createToken" class="grid grid-cols-1 md:grid-cols-4 gap-4 items-end">
              <div class="md:col-span-3">
                <label for="new-token-expires" class="block text-sm font-medium text-gray-700">
                  {{ activeExpiryPreset === 'custom' ? 'Custom Expiration' : 'Expires At' }}
                </label>
                <input id="new-token-expires" v-model="newTokenExpiresAt" @input="onCustomDateChange"
                  type="datetime-local" required
                  class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2 text-sm focus:ring-black focus:border-black" />
              </div>

              <div class="md:col-span-1">
                <button type="submit"
                  class="w-full px-4 py-2 bg-black text-white font-semibold rounded-md shadow-sm hover:bg-gray-800 transition duration-150">
                  Create Token
                </button>
              </div>
            </form>
          </div>
        </div>

        <div v-if="!isEditingDetails && !isCreateMode" class="space-y-2">
          <p class="text-base font-medium text-gray-700">
            {{ server.instructions || 'Empty instructions.' }}
          </p>

          <div class="flex items-center space-x-2 pt-2">
            <span class="text-sm font-medium text-gray-500">Proxy Incoming Headers:</span>
            <span :class="[
              'px-2 inline-flex text-xs leading-5 font-semibold rounded-full',
              server.requestHeadersProxyEnabled
                ? 'bg-green-100 text-green-800'
                : 'bg-gray-100 text-gray-800',
            ]">
              {{ server.requestHeadersProxyEnabled ? 'ON' : 'OFF' }}
            </span>
            <span class="text-xs text-gray-400" v-if="server.requestHeadersProxyEnabled">
              (e.g., 'Authorization' headers will be passed to providers)
            </span>
          </div>
          <div class="pt-4 border-t border-gray-200">
            <h3 class="text-lg font-semibold text-gray-800 mb-2">MCP Server Address (for LLMs)</h3>
            <p class="text-sm text-gray-600 mb-4">
              Use this configuration snippet to register the MCP Server as a Remote Schema in your
              LLM:
            </p>

            <div class="p-1 bg-gray-100 rounded-lg flex space-x-1">
              <button @click="activeConfigTab = 'common'" :class="[
                'w-full py-2 px-3 rounded-md font-medium text-sm transition-colors duration-150',
                activeConfigTab === 'common'
                  ? 'bg-white text-black shadow'
                  : 'text-gray-600 hover:bg-gray-200'
              ]">
                common format
              </button>
              <button @click="activeConfigTab = 'gemini-cli'" :class="[
                'w-full py-2 px-3 rounded-md font-medium text-sm transition-colors duration-150',
                activeConfigTab === 'gemini-cli'
                  ? 'bg-white text-black shadow'
                  : 'text-gray-600 hover:bg-gray-200'
              ]">
                gemini-cli
              </button>
              <button @click="activeConfigTab = 'remote-mcp'" :class="[
                'w-full py-2 px-3 rounded-md font-medium text-sm transition-colors duration-150',
                activeConfigTab === 'remote-mcp'
                  ? 'bg-white text-black shadow'
                  : 'text-gray-600 hover:bg-gray-200'
              ]">
                using remote-mcp
              </button>
            </div>

            <div class="mt-4">
              <div v-if="activeConfigTab === 'common'">
                <div class="relative bg-gray-800 rounded-lg shadow-inner">
                  <pre class="text-white p-3 text-xs font-mono overflow-x-auto">{{
                    mcpConfigJson
                  }}</pre>
                  <button @click="copyMcpConfig"
                    class="absolute top-2 right-2 p-1 rounded-md text-gray-400 hover:text-white hover:bg-gray-700 transition duration-150"
                    title="Copy Config JSON">
                    <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m-3 7H9M7 11h6" />
                    </svg>
                  </button>
                </div>
              </div>

              <div v-if="activeConfigTab === 'gemini-cli'">
                <div class="relative bg-gray-800 rounded-lg shadow-inner">
                  <pre class="text-white p-3 text-xs font-mono overflow-x-auto">{{
                    mcpConfigJsonGeminiCli
                  }}</pre>
                  <button @click="copyMcpConfigGeminiCli"
                    class="absolute top-2 right-2 p-1 rounded-md text-gray-400 hover:text-white hover:bg-gray-700 transition duration-150"
                    title="Copy Config JSON">
                    <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m-3 7H9M7 11h6" />
                    </svg>
                  </button>
                </div>
              </div>

              <div v-if="activeConfigTab === 'remote-mcp'">
                <p class="text-sm text-gray-600 mb-2">
                  If the MCP client does not support remote MCP then you can fallback to mcp-remote
                  config:
                </p>
                <div class="relative bg-gray-800 rounded-lg shadow-inner">
                  <pre class="text-white p-3 text-xs font-mono overflow-x-auto">{{
                    mcpConfigJsonUsingRemoteMcp
                  }}</pre>
                  <button @click="copyMcpConfigRemoteMcp"
                    class="absolute top-2 right-2 p-1 rounded-md text-gray-400 hover:text-white hover:bg-gray-700 transition duration-150"
                    title="Copy Config JSON">
                    <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m-3 7H9M7 11h6" />
                    </svg>
                  </button>
                </div>
              </div>
            </div>

            <p class="text-xs text-red-500 mt-2 font-medium">
              <span class="font-bold">Important:</span> You must replace
              `&lt;YOUR_TOKEN_VALUE_FROM_ABOVE_TOKEN&gt;` or `YOUR_TOKEN_VALUE_FROM_ABOVE_TOKEN`
              with an actual token generated above.
            </p>
          </div>
          <p class="text-sm text-gray-500 font-medium">
            Providers:
            <span class="font-semibold text-black">{{ server.providers.length }}</span>
          </p>
          <p class="text-sm text-gray-500 font-medium">
            Enabled MCP Tools:
            <span class="font-semibold text-black">{{
              server.providers.reduce((count, p) => count + p.enabledTools.length, 0)
              }}</span>
          </p>
          <button @click="isEditingDetails = true"
            class="mt-2 text-sm text-blue-600 hover:text-blue-800 transition duration-150">
            Edit Details & Tools
          </button>
        </div>

        <form v-if="isEditingDetails || isCreateMode" @submit.prevent="saveMcpserver" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-600">MCP Server Name</label>
            <input v-model="serverDraft.name" @input="formatServerName" type="text" required placeholder="e.g., stripe"
              class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm focus:ring-black focus:border-black" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-600">Instructions (Optional)</label>
            <textarea v-model="serverDraft.instructions" rows="2"
              class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm focus:ring-black focus:border-black"></textarea>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-600">Proxy Incoming Headers</label>
            <p class="text-xs text-gray-500 mb-2">
              If enabled, incoming request headers (like 'Authorization') will be proxied to the
              provider's gateway.
            </p>
            <label class="relative inline-flex items-center cursor-pointer">
              <input type="checkbox" v-model="serverDraft.requestHeadersProxyEnabled" class="sr-only peer" />
              <div
                class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-gray-300 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:left-0.5] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-black">
              </div>
              <span class="ml-3 text-sm font-medium text-gray-700">{{
                serverDraft.requestHeadersProxyEnabled ? 'Enabled' : 'Disabled'
                }}</span>
            </label>
          </div>
          <div class="pt-4 border-t">
            <h2 class="text-xl font-bold mb-4 text-gray-800">Select Providers & MCP Tools</h2>
            <div class="mb-4">
              <input v-model="providerSearchQuery" type="text" placeholder="Search providers to add..."
                class="block w-full border border-gray-300 rounded-full shadow-inner p-3 text-sm focus:ring-black focus:border-black" />
            </div>

            <div v-if="searchedProviders.length > 0" class="space-y-4">
              <div v-for="provider in searchedProviders" :key="provider.id" :class="[
                'bg-gray-50 p-4 rounded-xl shadow-md border-l-4 transition duration-300',
                server.providers.some((p) => p.id === provider.id)
                  ? 'border-black'
                  : 'border-gray-300',
              ]">
                <div class="flex justify-between items-center cursor-pointer"
                  @click="toggleProviderAccordion(provider)">
                  <div class="flex items-center space-x-3">
                    <img v-if="provider.iconURL" :src="provider.iconURL" :alt="`${provider.name} Icon`"
                      class="w-12 h-12 rounded-full object-cover" />
                    <svg v-else class="w-8 h-8 text-black" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                    </svg>
                    <span class="font-bold text-lg text-gray-900">{{ provider.name }}</span>
                    <span class="text-xs font-semibold px-2 py-0.5 rounded-full bg-gray-200 text-gray-700">{{
                      provider.type }}</span>
                    <span class="text-xs font-semibold px-2 py-0.5 rounded-full bg-gray-300 text-gray-700">v{{
                      provider.version }}</span>
                    <span v-if="!server.providers.some((p) => p.id === provider.id)" class="text-sm text-red-500">Not
                      Deployed</span>
                  </div>
                  <button type="button" class="p-1 text-gray-600 transition duration-150 transform"
                    :class="{ 'rotate-180': toolSelectionOpen[provider.id] }">
                    <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                    </svg>
                  </button>
                </div>

                <div v-if="toolSelectionOpen[provider.id]" class="mt-4 pt-4 border-t border-gray-200">
                  <div v-if="providerLoadingState[provider.id]">
                    <p class="text-sm text-gray-500 italic">Loading tools...</p>
                  </div>
                  <div v-else-if="
                    !fullProviderDataCache[provider.id] ||
                    !fullProviderDataCache[provider.id].tools
                  ">
                    <p class="text-sm text-red-500">Could not load tools for this provider.</p>
                  </div>
                  <div v-else>
                    <div class="flex justify-between items-center mb-2">
                      <h3 class="text-md font-semibold text-gray-700">
                        Select MCP Tools ({{
                          getProviderLiveTools(provider.id).length
                        }}
                        available)
                      </h3>

                      <button v-if="getProviderLiveTools(provider.id).length > 0" type="button"
                        @click="toggleAllTools(provider)"
                        class="text-xs font-semibold px-3 py-1 rounded-full transition duration-150" :class="areAllToolsEnabled(provider.id)
                          ? 'bg-gray-200 text-gray-800 hover:bg-gray-300'
                          : 'bg-black text-white hover:bg-gray-800'
                          ">
                        {{ areAllToolsEnabled(provider.id) ? 'Disable All' : 'Enable All' }}
                      </button>
                    </div>
                    <div v-if="getProviderLiveTools(provider.id).length === 0" class="text-sm text-gray-500 italic">
                      This provider has no tools configured.
                    </div>
                    <div v-else class="space-y-2 max-h-60 overflow-y-auto pr-2">
                      <div v-for="tool in getProviderLiveTools(provider.id)" :key="tool.id" :class="[
                        'flex items-center justify-between p-3 rounded-md border transition duration-150',
                        isToolEnabled(provider.id, tool)
                          ? 'border-black bg-white'
                          : 'border-gray-200 bg-white hover:bg-gray-100',
                      ]">
                        <div class="flex flex-col grow truncate">
                          <div class="flex items-center space-x-2">
                            <span :class="[
                              'font-mono text-xs font-bold px-2 py-1 rounded text-white shrink-0 bg-gray-600',
                              getMethodTextColor(tool.method),
                            ]">
                              {{ tool.method }}
                            </span>
                            <span class="text-sm text-gray-800 font-mono truncate">{{
                              tool.path
                              }}</span>
                          </div>
                          <p class="text-xs text-gray-500 mt-1 truncate">
                            {{ tool.description }}
                          </p>
                        </div>
                        <label class="relative inline-flex items-center cursor-pointer ml-4">
                          <input type="checkbox" :checked="isToolEnabled(provider.id, tool)"
                            @change="toggleTool(provider.id, tool)" class="sr-only peer" />
                          <div :class="[
                            'w-11 h-6 bg-gray-200 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[\'\'] after:absolute after:top-0.5 after:left-0.5] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-black',
                          ]"></div>
                        </label>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div v-else class="text-center py-6 text-gray-500">
              No providers selected. Please use the search bar to find and add API providers.
            </div>
          </div>

          <div class="flex justify-end space-x-3 pt-4 border-t">
            <button type="button" @click="cancelEdit"
              class="px-4 py-2 text-gray-700 border border-gray-300 rounded-lg hover:bg-gray-100 transition duration-150">
              Cancel
            </button>
            <button type="submit"
              class="px-4 py-2 bg-black text-white font-semibold rounded-lg shadow-md hover:bg-gray-800 transition duration-150">
              {{ isCreateMode ? 'Create MCP Server' : 'Save Changes' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <DeleteModal :show="showDeleteModal" :variableName="deleteNameForModal" :itemType="itemTypeForModal"
      @close="cancelDeleteModal" @confirm="confirmDeleteModal" />
  </div>
</template>