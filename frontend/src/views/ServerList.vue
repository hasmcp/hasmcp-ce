<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useMcpserverStore } from '../stores/serverStore'

const router = useRouter()
const serverStore = useMcpserverStore()

const servers = computed(() => serverStore.servers)

const goToDetail = (id) => {
  router.push({ name: 'ServerDetail', params: { id } })
}

const goToCreate = () => {
  router.push({ name: 'ServerCreate' })
}
</script>

<template>
  <div class="p-4">
    <h1 class="text-3xl font-bold mb-6 text-gray-800 flex justify-between items-center">
      MCP Servers
      <button @click="goToCreate" class="p-2 rounded-full text-white bg-black hover:bg-gray-800 transition duration-150"
        title="Create New MCP Server">
        <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
      </button>
    </h1>

    <div v-if="servers.length === 0" class="text-center py-10 text-gray-500 bg-white rounded-lg shadow-md">
      No MCP Servers created yet.
    </div>

    <div v-else class="space-y-6">
      <div v-for="server in servers" :key="server.id"
        class="bg-white p-6 rounded-xl shadow-lg border-l-4 border-black group cursor-pointer"
        @click="goToDetail(server.id)">
        <div class="flex justify-between items-start">
          <div>
            <h2 class="text-xl font-bold text-gray-900 group-hover:text-black transition duration-150">
              {{ server.name }}
            </h2>
            <p class="text-sm text-gray-600 mt-1 mb-3">
              {{ server.instructions || 'Empty instructions.' }}
            </p>

            <div class="flex space-x-4 text-xs font-medium text-gray-500">
              <span class="inline-flex items-center">
                <svg class="w-4 h-4 mr-1 text-black" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M7 7h.01M7 3h5a2 2 0 012 2v5a2 2 0 01-2 2H7a2 2 0 01-2-2V7a2 2 0 012-2z" />
                </svg>
                Version: v{{ server.version }}
              </span>
              <span class="inline-flex items-center">
                <svg class="w-4 h-4 mr-1 text-black" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
                Providers: {{ server?.providers?.length || 0 }}
              </span>
              <span class="inline-flex items-center">
                <svg class="w-4 h-4 mr-1 text-black" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
                </svg>
                Tools:
                {{
                  server?.providers?.reduce((count, p) => count + p.enabledTools.length, 0)
                }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
