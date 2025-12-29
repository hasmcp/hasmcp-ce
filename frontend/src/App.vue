<template>
  <ApiTokenModal v-if="!authStore.isAuthenticated" />

  <div v-if="authStore.isAuthenticated" id="app" class="flex h-screen bg-gray-50">
    <nav
      class="p-6 w-64 rounded-br-3xl rounded-tr-3xl bg-slate-100 text-gray-800 shrink-0 flex flex-col justify-between overflow-y-auto">
      <div>
        <div class="p-6 mt-4 bg-white rounded-2xl text-xl font-bold flex items-center gap-3">
          <svg width="24px" height="24px" viewBox="0 0 256 256">
            <g stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
              <g>
                <rect stroke="#000000" stroke-width="16" x="38" y="38" width="180" height="180" rx="18"></rect>
                <rect fill="#000000" x="91" y="96" width="24" height="64" rx="12"></rect>
                <rect fill="#000000" x="139" y="96" width="24" height="64" rx="12"></rect>
                <rect fill="#000000" x="95" y="116" width="64" height="24" rx="12"></rect>
                <rect fill="#000000" x="79" y="0" width="24" height="64" rx="12"></rect>
                <rect fill="#000000" x="152" y="0" width="24" height="64" rx="12"></rect>
                <rect fill="#000000" x="80" y="192" width="24" height="64" rx="12"></rect>
                <rect fill="#000000" x="152" y="192" width="24" height="64" rx="12"></rect>
                <rect fill="#000000" x="0" y="80" width="64" height="24" rx="12"></rect>
                <rect fill="#000000" x="0" y="152" width="64" height="24" rx="12"></rect>
                <rect fill="#000000" x="192" y="80" width="64" height="24" rx="12"></rect>
                <rect fill="#000000" x="192" y="152" width="64" height="24" rx="12"></rect>
              </g>
            </g>
          </svg>
          <span>Has MCP</span>
        </div>

        <div class="p-4 space-y-2 bg-white mt-6 rounded-2xl">
          <router-link to="/"
            class="block px-4 py-2 rounded-md hover:bg-gray-700 transition duration-150 text-gray-800 hover:text-white">
            <svg class="w-5 h-5 inline-block mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
            </svg>
            Overview
          </router-link>

          <router-link to="/servers"
            class="block px-4 py-2 rounded-md hover:bg-gray-700 transition duration-150 text-gray-800 hover:text-white"
            active-class="router-link-active">
            <svg class="w-5 h-5 inline-block mr-2" viewBox="0 0 16 17" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path
                d="M1.62524 8.11636L7.6712 2.07042C8.50598 1.23564 9.85941 1.23564 10.6941 2.07042C11.5289 2.90518 11.5289 4.25861 10.6941 5.09339L6.12821 9.65934"
                stroke="currentColor" stroke-width="1.5" stroke-linecap="round"></path>
              <path
                d="M6.19116 9.59684L10.6941 5.09385C11.5289 4.25908 12.8823 4.25908 13.7171 5.09385L13.7486 5.12534C14.5834 5.96011 14.5834 7.31354 13.7486 8.14831L8.28059 13.6164C8.00233 13.8946 8.00233 14.3457 8.28059 14.6239L9.40336 15.7468"
                stroke="currentColor" stroke-width="1.5" stroke-linecap="round"></path>
              <path
                d="M9.18266 3.58203L4.71116 8.05351C3.87639 8.88826 3.87639 10.2417 4.71116 11.0765C5.54593 11.9112 6.89936 11.9112 7.73414 11.0765L12.2056 6.605"
                stroke="currentColor" stroke-width="1.5" stroke-linecap="round"></path>
            </svg>
            MCP Servers
          </router-link>
          <router-link to="/providers"
            class="block px-4 py-2 rounded-md hover:bg-gray-700 transition duration-150 text-gray-800 hover:text-white">
            <svg class="w-5 h-5 inline-block mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z">
              </path>
            </svg>
            Providers
          </router-link>
          <router-link to="/variables"
            class="block px-4 py-2 rounded-md hover:bg-gray-700 transition duration-150 text-gray-800 hover:text-white">
            <svg class="w-5 h-5 inline-block mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37.526.315 1.134.48 1.724.48z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            Variables
          </router-link>
        </div>
      </div>

      <div class="p-0 space-y-2 mt-6 rounded-2xl">
        <a target="_blank" tabindex="-1" href="https://github.com/hasmcp/hasmcp-docs"><button data-slot="button"
            data-shape="rounded" data-target="default"
            class="font-inter group/button whitespace-nowrap [&amp;_svg]:pointer-events-none shrink-0 [&amp;_svg]:shrink-0 data-[loading=true]:pointer-events-none data-[loading=true]:select-none data-[shape='pill']:rounded-full bg-transparent border border-transparent outline-none focus-visible:ring-2 focus-visible:ring-tertiary-hover focus-visible:border-secondary disabled:cursor-not-allowed disabled:text-quaternary disabled:hover:bg-transparent typography-caption-l-medium data-[shape='rounded']:rounded-[10px] [&amp;_svg:not([class*='size-'])]:size-[14px] tracking-[-0.1px] flex gap-3 items-center rounded-lg px-2 h-8 text-sm leading-5 justify-between w-full font-normal! text-primary opacity-70 hover:bg-zinc-950/5 dark:hover:bg-white/5 hover:opacity-100">
            <div class="flex items-center gap-2 -translate-x-[0.5px]">
              <svg class="h-5 w-5" viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5"
                stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path
                  d="M9 16.25C9 15.145 9.895 14.25 11 14.25H15.25C15.802 14.25 16.25 13.802 16.25 13.25V3.75C16.25 3.198 15.802 2.75 15.25 2.75H11C9.895 2.75 9 3.645 9 4.75"
                  stroke="currentColor" width="1.5" linecap="round" linejoin="round"></path>
                <path
                  d="M9 16.25V4.75C9 3.645 8.105 2.75 7 2.75H2.75C2.198 2.75 1.75 3.198 1.75 3.75V13.25C1.75 13.802 2.198 14.25 2.75 14.25H7C8.105 14.25 9 15.145 9 16.25Z"
                  stroke="currentColor" width="1.5" linecap="round" linejoin="round"></path>
              </svg>
              <p>Documentation</p>
            </div>
          </button>
        </a>
        <button @click="authStore.clearToken()"
          class="cursor:pointer text-left font-inter group/button whitespace-nowrap [&amp;_svg]:pointer-events-none shrink-0 [&amp;_svg]:shrink-0 data-[loading=true]:pointer-events-none data-[loading=true]:select-none data-[shape='pill']:rounded-full bg-transparent border border-transparent outline-none focus-visible:ring-2 focus-visible:ring-tertiary-hover focus-visible:border-secondary disabled:cursor-not-allowed disabled:text-quaternary disabled:hover:bg-transparent typography-caption-l-medium data-[shape='rounded']:rounded-[10px] [&amp;_svg:not([class*='size-'])]:size-[14px] tracking-[-0.1px] flex gap-3 items-center rounded-lg px-2 h-8 text-sm leading-5 justify-between w-full font-normal! text-primary opacity-70 hover:bg-zinc-950/5 dark:hover:bg-white/5 hover:opacity-100">
          <div class="flex items-center gap-2 -translate-x-[0.5px]">
            <svg class="w-5 h-5 inline-block text-gray-700 group-hover:text-white transition duration-150" fill="none"
              viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
            Logout
          </div>
        </button>
        <div class="lg:my-2.5 my-1 lg:px-2 w-full">
          <div class="h-px bg-zinc-900/5 dark:bg-white/5"></div>
        </div>
        <div class="p-4 bg-white rounded-md text-xs">
          <a href="https://github.com/hasmcp/hasmcp-ce" target="_blank" class="hover:underline">
            <svg class="w-5 h-5 inline-block mr-2 text-gray-700 group-hover:text-white transition duration-150" x="0px"
              y="0px" viewBox="0 0 36 36" enable-background="new 0 0 36 36" xml:space="preserve">
              <path fill-rule="evenodd" clip-rule="evenodd" fill="#191717" d="M18,1.4C9,1.4,1.7,8.7,1.7,17.7c0,7.2,4.7,13.3,11.1,15.5
	c0.8,0.1,1.1-0.4,1.1-0.8c0-0.4,0-1.4,0-2.8c-4.5,1-5.5-2.2-5.5-2.2c-0.7-1.9-1.8-2.4-1.8-2.4c-1.5-1,0.1-1,0.1-1
	c1.6,0.1,2.5,1.7,2.5,1.7c1.5,2.5,3.8,1.8,4.7,1.4c0.1-1.1,0.6-1.8,1-2.2c-3.6-0.4-7.4-1.8-7.4-8.1c0-1.8,0.6-3.2,1.7-4.4
	c-0.2-0.4-0.7-2.1,0.2-4.3c0,0,1.4-0.4,4.5,1.7c1.3-0.4,2.7-0.5,4.1-0.5c1.4,0,2.8,0.2,4.1,0.5c3.1-2.1,4.5-1.7,4.5-1.7
	c0.9,2.2,0.3,3.9,0.2,4.3c1,1.1,1.7,2.6,1.7,4.4c0,6.3-3.8,7.6-7.4,8c0.6,0.5,1.1,1.5,1.1,3c0,2.2,0,3.9,0,4.5
	c0,0.4,0.3,0.9,1.1,0.8c6.5-2.2,11.1-8.3,11.1-15.5C34.3,8.7,27,1.4,18,1.4z" />
            </svg>
            HasMCP-CE v0.1.0
          </a>
        </div>
      </div>
    </nav>

    <main class="grow p-6 mb-6 max-w-screen-2xl flex flex-col min-h-0 overflow-y-auto">
      <router-view class="grow flex flex-col min-h-0" />
    </main>

    <ToastNotification />
  </div>
</template>

<script setup>
import { watchEffect } from 'vue'
import ToastNotification from './components/ToastNotification.vue'
import ApiTokenModal from './components/ApiTokenModal.vue'
import { useAuthStore } from './stores/authStore'
import { useProviderStore } from './stores/providerStore'
import { useEnvVarStore } from './stores/envVarStore'
import { useMcpserverStore } from './stores/serverStore'

const authStore = useAuthStore()
const providerStore = useProviderStore()
const envVarStore = useEnvVarStore()
const serverStore = useMcpserverStore()

watchEffect(() => {
  if (authStore.isAuthenticated) {
    // Once authenticated, fetch all core application data
    providerStore.fetchProviders()
    envVarStore.fetchEnvVars()
    serverStore.fetchMcpservers()
  }
})
</script>

<style scoped>
.router-link-active {
  background-color: #374151;
  color: white;
}
</style>
