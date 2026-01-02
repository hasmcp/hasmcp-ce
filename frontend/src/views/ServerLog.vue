<script setup>
import { ref, onMounted, onUnmounted, computed, reactive, nextTick } from 'vue'
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

// --- Realtime Analytics State ---
const analytics = reactive({
  totalCalls: 0,
  totalToolCalls: 0,
  sessions: new Set(),
  clients: new Set(),
  toolCalls: {},   // { toolName: count }
  methodCalls: {}  // { methodName: count }
})

const stats = computed(() => ({
  clientsCount: analytics.clients.size,
  sessionsCount: analytics.sessions.size,
  toolCallsCount: analytics.totalToolCalls,
  totalCallsCount: analytics.totalCalls,
  topTools: Object.entries(analytics.toolCalls)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 5),
  topMethods: Object.entries(analytics.methodCalls)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 5)
}))

const updateAnalytics = (entry) => {
  // Only focus on incoming requests indicated by "«"
  if (!entry.event || !entry.event.startsWith('«')) return

  analytics.totalCalls++

  // Regex captures: « [session].[client]/[version].[method]
  // This handles client names with parentheses and versions with dots by:
  // 1. Capturing session until first dot
  // 2. Capturing client info until the protocol slash
  // 3. Capturing version until the final dot
  // 4. Capturing the remainder as the method name
  const parts = entry.event.match(/^«\s+([^.]+)\.(.+)\/([^.]+)\.(.+)$/)

  if (!parts) return

  const [_, sessionId, clientInfo, _version, methodName] = parts

  // Update unique tracking sets
  if (sessionId) analytics.sessions.add(sessionId)
  if (clientInfo) analytics.clients.add(clientInfo)

  // Track the general method call (e.g., "resources/list", "notifications/initialized")
  analytics.methodCalls[methodName] = (analytics.methodCalls[methodName] || 0) + 1

  // Handle specialized tool calls specifically
  if (methodName === 'tools/call' && entry.data) {
    analytics.totalToolCalls++
    try {
      const dataJson = JSON.parse(entry.data)
      const toolName = dataJson.name.substr(13)
      if (toolName) {
        analytics.toolCalls[toolName] = (analytics.toolCalls[toolName] || 0) + 1
      }
    } catch (e) {
      // Ignore malformed JSON in the stream
    }
  }
}

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
  if (eventName.startsWith('req') || eventName.includes('«')) return 'text-blue-400'
  if (eventName.startsWith('res') || eventName.includes('»')) return 'text-green-400'
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
      const newEntry = {
        id: currentEventGroup.id,
        event: currentEventGroup.event,
        data: rawData,
        raw: currentEventGroup.raw.join('\n'),
      }

      updateAnalytics(newEntry)
      logEntries.value.unshift(newEntry)

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
  const server = serverStore.getMcpserverById(serverId.value)
  if (!server) {
    try {
      await serverStore.loadMcpserverById(serverId.value)
    } catch (e) {
      status.value = 'Error'
      error.value = 'Failed to load MCP server data.'
      return
    }
  }
  initialize()
})

onUnmounted(() => abortController.abort())
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

      <div class="mb-4 space-y-4 shrink-0">
        <div class="grid grid-cols-4 gap-4">
          <div class="bg-gray-50 p-3 rounded-lg border border-gray-100 shadow-sm text-center">
            <div class="text-[10px] uppercase font-bold text-gray-400">Clients</div>
            <div class="text-xl font-mono font-black text-blue-600">{{ stats.clientsCount }}</div>
          </div>
          <div class="bg-gray-50 p-3 rounded-lg border border-gray-100 shadow-sm text-center">
            <div class="text-[10px] uppercase font-bold text-gray-400">Sessions</div>
            <div class="text-xl font-mono font-black text-purple-600">{{ stats.sessionsCount }}</div>
          </div>
          <div class="bg-gray-50 p-3 rounded-lg border border-gray-100 shadow-sm text-center">
            <div class="text-[10px] uppercase font-bold text-gray-400">Tool Calls</div>
            <div class="text-xl font-mono font-black text-green-600">{{ stats.toolCallsCount }}</div>
          </div>
          <div class="bg-gray-50 p-3 rounded-lg border border-gray-100 shadow-sm text-center">
            <div class="text-[10px] uppercase font-bold text-gray-400">Method Calls</div>
            <div class="text-xl font-mono font-black text-gray-800">{{ stats.totalCallsCount }}</div>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div class="bg-gray-50 p-3 rounded-lg border border-gray-100 shadow-sm">
            <div class="text-[10px] uppercase font-bold text-gray-500 mb-2 border-b pb-1">Top Tool Calls</div>
            <div v-if="stats.topTools.length === 0"
              class="h-20 flex items-center justify-center text-[10px] text-gray-400 italic">Waiting for tool calls...
            </div>
            <div v-else class="space-y-2 h-20 overflow-y-auto">
              <div v-for="[name, count] in stats.topTools" :key="name" class="space-y-1">
                <div class="flex justify-between text-[10px] font-mono">
                  <span class="truncate pr-2">{{ name }}</span>
                  <span class="font-bold">{{ count }}</span>
                </div>
                <div class="w-full bg-gray-200 h-1 rounded-full overflow-hidden">
                  <div class="bg-green-500 h-full transition-all duration-500"
                    :style="{ width: (count / stats.toolCallsCount * 100) + '%' }"></div>
                </div>
              </div>
            </div>
          </div>

          <div class="bg-gray-50 p-3 rounded-lg border border-gray-100 shadow-sm">
            <div class="text-[10px] uppercase font-bold text-gray-500 mb-2 border-b pb-1">Top Method Calls</div>
            <div v-if="stats.topMethods.length === 0"
              class="h-20 flex items-center justify-center text-[10px] text-gray-400 italic">Waiting for requests...
            </div>
            <div v-else class="space-y-2 h-20 overflow-y-auto">
              <div v-for="[name, count] in stats.topMethods" :key="name" class="space-y-1">
                <div class="flex justify-between text-[10px] font-mono">
                  <span class="truncate pr-2">{{ name }}</span>
                  <span class="font-bold">{{ count }}</span>
                </div>
                <div class="w-full bg-gray-200 h-1 rounded-full overflow-hidden">
                  <div class="bg-blue-500 h-full transition-all duration-500"
                    :style="{ width: (count / stats.totalCallsCount * 100) + '%' }"></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div
        class="bg-black text-white font-mono rounded-lg shadow-inner grow overflow-y-auto p-4 flex flex-col-reverse min-h-0">
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
          <div class="shrink-0 py-3">
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