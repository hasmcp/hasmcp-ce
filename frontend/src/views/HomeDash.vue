<script setup>
import { computed } from 'vue'
import { useMcpserverStore } from '../stores/serverStore'
import { useProviderStore } from '../stores/providerStore'
import { useEnvVarStore } from '../stores/envVarStore'
import { useRouter } from 'vue-router'

const router = useRouter()
const serverStore = useMcpserverStore()
const providerStore = useProviderStore()
const envVarStore = useEnvVarStore()

// --- Computed Metrics ---
const totalServers = computed(() => serverStore.servers.length)
const totalProviders = computed(() => providerStore.providers.length)
const totalEnvVars = computed(() => envVarStore.envVariables.length)

const totalEnabledTools = computed(() => {
  return serverStore.servers.reduce((sum, server) => {
    return (
      sum +
      server.providers.reduce((pSum, provider) => pSum + provider.enabledTools.length, 0)
    )
  }, 0)
})

// --- Quick Links ---
const goTo = (name) => {
  router.push({ name })
}

// --- Dashboard Cards Data (Monochrome Style) ---
const dashboardCards = computed(() => [
  {
    title: 'MCP Servers',
    value: totalServers.value,
    icon: `<svg class="w-8 h-8 inline-block mr-2" viewBox="0 0 16 17" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M1.62524 8.11636L7.6712 2.07042C8.50598 1.23564 9.85941 1.23564 10.6941 2.07042C11.5289 2.90518 11.5289 4.25861 10.6941 5.09339L6.12821 9.65934" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"></path>
<path d="M6.19116 9.59684L10.6941 5.09385C11.5289 4.25908 12.8823 4.25908 13.7171 5.09385L13.7486 5.12534C14.5834 5.96011 14.5834 7.31354 13.7486 8.14831L8.28059 13.6164C8.00233 13.8946 8.00233 14.3457 8.28059 14.6239L9.40336 15.7468" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"></path>
<path d="M9.18266 3.58203L4.71116 8.05351C3.87639 8.88826 3.87639 10.2417 4.71116 11.0765C5.54593 11.9112 6.89936 11.9112 7.73414 11.0765L12.2056 6.605" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"></path></svg>`,
    route: 'Servers',
    // Monochrome: Black background, white text
    color: 'bg-black',
    hover: 'hover:bg-gray-800',
  },
  {
    title: 'API Providers',
    value: totalProviders.value,
    icon: `<svg class="w-8 h-8" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>`,
    route: 'Providers',
    // Monochrome: Black background, white text
    color: 'bg-black',
    hover: 'hover:bg-gray-800',
  },
  {
    title: 'Enabled Tools',
    value: totalEnabledTools.value,
    icon: `<svg class="w-8 h-8" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10l-4-4m4 4l4-4m4-4H9m8 0l-4-4m4 4l4 4"></path></svg>`,
    route: 'Servers',
    // Monochrome: Black background, white text
    color: 'bg-black',
    hover: 'hover:bg-gray-800',
  },
  {
    title: 'Variables',
    value: totalEnvVars.value,
    icon: `<svg class="w-8 h-8" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37.526.315 1.134.48 1.724.48z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /></svg>`,
    route: 'Variables',
    // Monochrome: Black background, white text
    color: 'bg-black',
    hover: 'hover:bg-gray-800',
  },
])
</script>

<template>
  <div class="p-4">
    <h1 class="text-3xl font-bold mb-6 text-gray-800 flex justify-between items-center">Has MCP</h1>
    <p class="text-md text-gray-600 mb-8">
      No-code <span class="font-bold">Model Context Protocol (MCP)</span> Servers builder. Have a
      HTTP tool? convert it to MCP server tool in seconds.
    </p>

    <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-12">
      <div v-for="card in dashboardCards" :key="card.title"
        class="p-6 rounded-xl shadow-xl border border-gray-200 bg-white text-gray-900 transition duration-300 transform hover:scale-[1.02] cursor-pointer hover:shadow-2xl"
        @click="goTo(card.route)">
        <div class="flex justify-between items-start">
          <h2 class="text-xl font-semibold">{{ card.title }}</h2>
          <div v-html="card.icon" class="text-black"></div>
        </div>
        <p class="text-5xl font-bold mt-4">{{ card.value }}</p>
      </div>
    </div>

    <div class="bg-white p-8 rounded-xl shadow-2xl">
      <h2 class="text-3xl font-bold text-gray-800 mb-6 border-b pb-3">Quick Actions</h2>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <button @click="goTo('ServerCreate')"
          class="flex flex-col items-center justify-center p-6 bg-black text-white rounded-lg shadow-md hover:bg-gray-800 transition duration-150 transform hover:scale-[1.03] focus:outline-none focus:ring-4 focus:ring-black/50">
          <svg class="w-8 h-8 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          <span class="text-lg font-semibold">Create New MCP Server</span>
          <p class="text-sm text-gray-300 mt-1">Bundle providers and generate a configuration.</p>
        </button>

        <button @click="goTo('Providers')"
          class="flex flex-col items-center justify-center p-6 bg-white text-gray-800 border border-gray-300 rounded-lg shadow-md hover:bg-gray-50 transition duration-150 transform hover:scale-[1.03] focus:outline-none focus:ring-4 focus:ring-black/50">
          <svg class="w-8 h-8 mb-2 text-black" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
          </svg>
          <span class="text-lg font-semibold">Manage API Providers</span>
          <p class="text-sm text-gray-500 mt-1">View, edit, or add external API tools.</p>
        </button>

        <button @click="goTo('Variables')"
          class="flex flex-col items-center justify-center p-6 bg-white text-gray-800 border border-gray-300 rounded-lg shadow-md hover:bg-gray-50 transition duration-150 transform hover:scale-[1.03] focus:outline-none focus:ring-4 focus:ring-black/50">
          <svg class="w-8 h-8 mb-2 text-black" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37.526.315 1.134.48 1.724.48z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          <span class="text-lg font-semibold">Configure Env Variables</span>
          <p class="text-sm text-gray-500 mt-1">
            Set up secrets and configuration for your providers.
          </p>
        </button>
      </div>
    </div>
  </div>
</template>
