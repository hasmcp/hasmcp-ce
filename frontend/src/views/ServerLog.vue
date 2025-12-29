<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useMcpserverStore } from '../stores/serverStore'
import { useAuthStore } from '../stores/authStore'
import LogJsonModal from '../components/LogJsonModal.vue'

const route = useRoute()
const serverStore = useMcpserverStore()
const authStore = useAuthStore()

const serverId = computed(() => route.params.id)
const server = computed(() => serverStore.getMcpserverById(serverId.value))

const logEntries = ref([])
const status = ref('Initializing...')
const error = ref(null)

let abortController = new AbortController()

const isJsonModalVisible = ref(false)
const modalJsonContent = ref('')

const showJsonModal = (jsonString) => {
  modalJsonContent.value = jsonString
  isJsonModalVisible.value = true
}

const closeJsonModal = () => {
  isJsonModalVisible.value = false
  modalJsonContent.value = ''
}

const getEventClass = (eventName) => {
  if (!eventName) return 'text-gray-400'
  if (eventName.startsWith('req')) return 'text-blue-400'
  if (eventName.startsWith('res')) return 'text-green-400'
  if (eventName.includes('error')) return 'text-red-500'
  return 'text-yellow-400'
}

const startLogStream = async (tokenValue) => {
  status.value = 'Connecting to log stream...'
  let currentEventGroup = { id: null, event: null, data: [], raw: [] }

  const pushEventGroup = () => {
    if (
      currentEventGroup.id ||
      currentEventGroup.event ||
      currentEventGroup.data.length > 0 ||
      currentEventGroup.raw.length > 0
    ) {
      const rawData = currentEventGroup.data.join('')
      logEntries.value.unshift({
        id: currentEventGroup.id,
        event: currentEventGroup.event,
        data: rawData,
        raw: currentEventGroup.raw.join('\n'),
      })
      if (logEntries.value.length > 1000) {
        logEntries.value.splice(1000)
      }
    }
    currentEventGroup = { id: null, event: null, data: [], raw: [] }
  }

  try {
    const baseUrl = new URL(authStore.apiBaseUrl).origin
    const streamUrl = `${baseUrl}/mcp/${serverId.value}/logs`

    const response = await fetch(streamUrl, {
      headers: { 'x-hasmcp-key': `Bearer ${tokenValue}` },
      signal: abortController.signal,
    })

    if (!response.ok) {
      throw new Error(`Failed to connect: ${response.status} ${response.statusText}`)
    }

    status.value = 'Streaming...'
    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) {
        status.value = 'Stream ended.'
        if (buffer) buffer.split('\n').forEach(processLine)
        pushEventGroup()
        break
      }
      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop()
      lines.forEach(processLine)
    }
  } catch (err) {
    if (err.name === 'AbortError') {
      status.value = 'Disconnected.'
    } else {
      console.error('Log stream error:', err)
      error.value = err.message
      status.value = 'Error'
    }
    pushEventGroup()
  }

  function processLine(line) {
    if (line.trim() === '') {
      pushEventGroup()
    } else if (line.startsWith(':')) {
      // Comment tag
    } else if (line.startsWith('id:')) {
      currentEventGroup.id = line.substring(3).trim()
    } else if (line.startsWith('event:')) {
      currentEventGroup.event = line.substring(6).trim()
    } else if (line.startsWith('data:')) {
      currentEventGroup.data.push(line.substring(5).trim())
    } else {
      currentEventGroup.raw.push(line)
    }
  }
}

const initialize = async () => {
  if (!serverId.value) {
    status.value = 'Error'
    error.value = 'No MCP Server ID found in URL.'
    return
  }

  status.value = 'Generating temporary token...'
  const expires = new Date()
  expires.setHours(expires.getHours() + 24)

  const tokenResult = await serverStore.createToken(serverId.value, {
    expiresAt: expires.toISOString(),
    scope: 'server:tail',
  })

  if (!tokenResult.success) {
    status.value = 'Error'
    error.value = tokenResult.missingVars?.length
      ? `Token creation failed. Missing env vars: ${tokenResult.missingVars.join(', ')}`
      : tokenResult.error || 'Failed to generate temporary token.'
    return
  }

  const token = tokenResult.token.value
  await startLogStream(token)
}

onMounted(async () => {
  // Check if the server data is already in the store
  const server = serverStore.getMcpserverById(serverId.value)

  // If not (e.g., on a hard refresh), fetch it first
  if (!server) {
    try {
      await serverStore.loadMcpserverById(serverId.value)
    } catch (e) {
      console.error('Failed to load MCP server data', e)
      status.value = 'Error'
      error.value = 'Failed to load MCP server data.'
      return
    }
  }

  // Now that the store is populated, run the initialization
  initialize()
})

onUnmounted(() => {
  abortController.abort()
})
</script>

<template>
  <div class="p-4 flex flex-col h-full">
    <h1 class="text-3xl font-bold mb-6 text-gray-800 shrink-0">
      Logs: {{ server?.name || 'MCP Server' }}
      <span class="text-lg font-mono text-gray-600 ml-2">(ID: {{ serverId }})</span>
    </h1>

    <div class="bg-white p-6 rounded-xl shadow-2xl grow flex flex-col min-h-0">
      <div class="mb-4 p-3 bg-gray-800 text-white rounded-lg shadow-md font-mono text-sm shrink-0">
        <span class="font-semibold">Status: </span>
        <span :class="{
          'text-green-400': status === 'Streaming...',
          'text-red-400': status === 'Error',
          'text-yellow-400':
            status === 'Connecting...' || status === 'Generating temporary token...',
        }">
          {{ status }}
        </span>
        <p v-if="error" class="text-red-400 mt-2">{{ error }}</p>
      </div>

      <div
        class="bg-black text-white font-mono rounded-lg shadow-inner grow overflow-y-auto p-4 flex flex-col-reverse min-h-[400px]">
        <div>
          <div v-for="(entry, index) in logEntries" :key="logEntries.length - index"
            class="text-xs leading-normal border-t border-gray-800 py-3 flex flex-row gap-x-4">
            <span class="select-none text-gray-600">{{
              String(logEntries.length - index).padStart(4, ' ')
              }}</span>

            <div class="flex flex-col min-w-0">
              <div v-if="entry.raw && !entry.id && !entry.event && !entry.data" class="text-gray-400">
                <span>{{ entry.raw }}</span>
              </div>

              <div v-else>
                <div v-if="entry.id" class="grid grid-cols-[60px,1fr]">
                  <span class="select-none text-gray-500 font-bold">ID:</span>
                  <span class="text-gray-200">{{ entry.id }}</span>
                </div>

                <div v-if="entry.event" class="grid grid-cols-[60px,1fr]">
                  <span class="select-none text-gray-500 font-bold">Event:</span>
                  <span :class="getEventClass(entry.event)">{{ entry.event }}</span>
                </div>

                <div v-if="entry.data" class="grid grid-cols-[60px,1fr] min-w-0">
                  <span class="select-none text-gray-500 font-bold self-start">Data:</span>
                  <span @click="showJsonModal(entry.data)"
                    class="text-gray-300 truncate cursor-pointer hover:text-white" title="Click to view full JSON">
                    {{ entry.data.replace(/\s+/g, ' ') }}
                  </span>
                </div>

                <div v-if="entry.raw" class="grid grid-cols-[60px,1fr] mt-1">
                  <span class="select-none text-red-500 font-bold">Raw:</span>
                  <span class="text-red-400">{{ entry.raw }}</span>
                </div>
              </div>
            </div>
          </div>
          <div classs="shrink-0 py-3">
            <div v-if="status === 'Connecting...' || status === 'Generating temporary token...'"
              class="text-yellow-400 animate-pulse text-xs">
              Connecting...
            </div>
            <div v-if="status === 'Stream ended.'" class="text-green-500 text-xs">
              --- LOG STREAM ENDED ---
            </div>
            <div v-if="status === 'Disconnected.'" class="text-yellow-500 text-xs">
              --- DISCONNECTED ---
            </div>
          </div>
        </div>
      </div>
    </div>

    <LogJsonModal :show="isJsonModalVisible" :jsonString="modalJsonContent" @close="closeJsonModal" />
  </div>
</template>
