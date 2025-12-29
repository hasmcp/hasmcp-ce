<script setup>
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useProviderStore } from '../stores/providerStore.js'
import { useToastStore } from '../stores/toastStore.js'

const router = useRouter()
const store = useProviderStore()
const toastStore = useToastStore()

// Access provider constants from the store helpers
const PROVIDER_TYPES = store.helpers.PROVIDER_TYPES
const PROVIDER_VISIBILITY = store.helpers.PROVIDER_VISIBILITY
const generateSecretPrefix = store.helpers.generateSecretPrefix

// ----------------------------------------------------
// State Management (Local)
// ----------------------------------------------------

const isCreationFormVisible = ref(false)
const searchQuery = ref('')
const itemsPerPage = 20
const visibleItemCount = ref(itemsPerPage)

const visibilityFilter = ref('ALL')

const newProvider = ref({
  name: '',
  description: '',
  baseURL: '',
  iconURL: '',
  documentURL: '',
  apiType: PROVIDER_TYPES[0] || 'REST',
  visibilityType: PROVIDER_VISIBILITY[0] || 'INTERNAL',
  secretPrefix: '',
  // OAuth2 Config
  enableOAuth: false,
  oauthClientID: '',
  oauthClientSecret: '',
  oauthAuthURL: '',
  oauthTokenURL: '',
})

// ----------------------------------------------------
// Computed Properties (using store.providers)
// ----------------------------------------------------

const sortedAndFilteredProviders = computed(() => {
  // Always work with a fresh copy of the store's state
  let list = [...store.providers]

  // 1. Filter by Search Query
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase().trim()
    list = list.filter(
      (p) =>
        p.name.toLowerCase().includes(query) ||
        p.description.toLowerCase().includes(query) ||
        p.baseURL.toLowerCase().includes(query),
    )
  }

  // 2. Filter by Visibility (NEW LOGIC)
  if (visibilityFilter.value !== 'ALL') {
    list = list.filter((p) => p.visibilityType === visibilityFilter.value)
  }

  // 3. Sort by User Count (Descending)
  // Note: Sorting happens in memory.
  list.sort((a, b) => b.name - a.name)

  return list
})

const displayedProviders = computed(() => {
  // Apply pagination slice to the sorted list
  return sortedAndFilteredProviders.value.slice(0, visibleItemCount.value)
})

const hasMoreProviders = computed(() => {
  return displayedProviders.value.length < sortedAndFilteredProviders.value.length
})

// ----------------------------------------------------
// Methods - General
// ----------------------------------------------------

const isValidUrl = (string) => {
  try {
    const url = new URL(string)

    if (url.protocol !== 'http:' && url.protocol !== 'https:') {
      return false
    }

    if (url.hostname === 'localhost' || url.hostname === '127.0.0.1' || url.hostname === '[::1]') {
      return true
    }

    if (!url.hostname.includes('.')) {
      return false
    }

    return true
  } catch {
    return false
  }
}

const formatProviderName = (event) => {
  let value = event.target.value
  value = value.replace(/[^a-zA-Z0-9]/g, '').slice(0, 16)
  newProvider.value.name = value

  if (event.target.value !== value) {
    event.target.value = value
  }
}

const loadMore = () => {
  visibleItemCount.value += itemsPerPage
}

// Toggle the manual creation form
const toggleCreationForm = () => {
  isCreationFormVisible.value = !isCreationFormVisible.value
  if (isCreationFormVisible.value) {
    newProvider.value = {
      name: '',
      description: '',
      baseURL: '',
      iconURL: '',
      documentURL: '',
      apiType: PROVIDER_TYPES[0] || 'REST',
      visibilityType: PROVIDER_VISIBILITY[0] || 'INTERNAL',
      secretPrefix: '',
      enableOAuth: false,
      oauthClientID: '',
      oauthClientSecret: '',
      oauthAuthURL: '',
      oauthTokenURL: '',
    }
  }
}

const createProvider = async () => {
  if (
    !newProvider.value.name ||
    !newProvider.value.description ||
    !newProvider.value.baseURL ||
    !newProvider.value.apiType ||
    !newProvider.value.visibilityType ||
    !newProvider.value.secretPrefix
  ) {
    toastStore.showToast(
      'Name, Description, Base URL, Type, Visibility, and Secret Prefix are required!',
      'warning',
      3000,
    )
    return
  }

  if (newProvider.value.enableOAuth) {
    if (
      !newProvider.value.oauthClientID ||
      !newProvider.value.oauthClientSecret ||
      !newProvider.value.oauthAuthURL ||
      !newProvider.value.oauthTokenURL
    ) {
      toastStore.showToast(
        'Client ID, Client Secret, Auth URL, and Token URL are required when OAuth2 is enabled.',
        'warning',
        3000,
      )
      return
    }
    if (!isValidUrl(newProvider.value.oauthAuthURL)) {
      toastStore.showToast('OAuth Auth URL must be a valid HTTP/HTTPS URL.', 'warning', 3000)
      return
    }
    if (!isValidUrl(newProvider.value.oauthTokenURL)) {
      toastStore.showToast('OAuth Token URL must be a valid HTTP/HTTPS URL.', 'warning', 3000)
      return
    }
  }

  if (!newProvider.value.baseURL.startsWith('http')) {
    toastStore.showToast(
      'Base URL must be a valid URL starting with http:// or https://',
      'warning',
      3000,
    )
    return
  }

  const name = newProvider.value.name.trim()
  const newEntry = {
    name: name,
    apiType: newProvider.value.apiType,
    visibilityType: newProvider.value.visibilityType,
    baseURL: newProvider.value.baseURL.trim(),
    iconURL: newProvider.value.iconURL.trim(),
    documentURL: newProvider.value.documentURL.trim(),
    description: newProvider.value.description.trim(),
    secretPrefix: newProvider.value.secretPrefix.trim(),
  }

  if (newProvider.value.enableOAuth) {
    newEntry.oauth2Config = {
      clientID: newProvider.value.oauthClientID.trim(),
      clientSecret: newProvider.value.oauthClientSecret.trim(),
      authURL: newProvider.value.oauthAuthURL.trim(),
      tokenURL: newProvider.value.oauthTokenURL.trim(),
    }
  }

  const newId = await store.addProvider(newEntry)

  if (newId) {
    // Navigate to detail page for tool creation
    router.push({ name: 'ProviderDetail', params: { id: newId } })

    // Reset form and hide it
    newProvider.value = {
      name: '',
      description: '',
      baseURL: '',
      iconURL: '',
      documentURL: '',
      apiType: PROVIDER_TYPES[0] || 'REST',
      visibilityType: PROVIDER_VISIBILITY[0] || 'INTERNAL',
      secretPrefix: '',
      enableOAuth: false,
      oauthClientID: '',
      oauthClientSecret: '',
      oauthAuthURL: '',
      oauthTokenURL: '',
    }
    isCreationFormVisible.value = false
  } else {
    // Handle the case where the API call failed and addProvider returned null
    toastStore.showToast('Failed to create provider. Please check the console.', 'alert', 3000)
  }
}

// Navigation to detailed view
const viewProvider = (providerId) => {
  router.push({ name: 'ProviderDetail', params: { id: providerId } })
}

// ----------------------------------------------------
// Watchers
// ----------------------------------------------------

// Watch for changes on the baseURL within the newProvider object
watch(
  () => newProvider.value.baseURL,
  (newBaseURL) => {
    newProvider.value.secretPrefix = generateSecretPrefix(newBaseURL)
  },
)
</script>

<template>
  <div class="p-4">
    <h1 class="text-3xl font-bold mb-6 text-gray-800 flex justify-between items-center">
      API Providers
      <div class="flex space-x-3">
        <button @click="toggleCreationForm" :class="[
          'p-2 rounded-full text-white transition-transform duration-300',
          isCreationFormVisible
            ? 'bg-red-500 hover:bg-red-600 rotate-45'
            : 'bg-black hover:bg-gray-800',
        ]" title="Toggle Add Provider Form">
          <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
        </button>
      </div>
    </h1>

    <div v-if="isCreationFormVisible" class="bg-white p-6 rounded-xl shadow-2xl mb-8 transition-all duration-300">
      <h2 class="text-xl font-semibold mb-4 text-gray-700 border-b pb-2">Add New API Provider</h2>
      <form @submit.prevent="createProvider" class="space-y-4">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-600">Provider Type (Immutable)</label>
            <div class="mt-1 flex items-center justify-start h-full">
              <label v-for="apiType in PROVIDER_TYPES" :key="apiType" :class="[
                'relative flex items-center p-3 rounded-lg cursor-pointer transition duration-150 mr-2',
                newProvider.apiType === apiType
                  ? 'bg-gray-200 text-black'
                  : 'bg-white text-gray-600 hover:bg-gray-50',
              ]">
                <input type="radio" :value="apiType" v-model="newProvider.apiType" name="type-toggle"
                  class="peer relative h-5 w-5 cursor-pointer appearance-none rounded-full border border-gray-300 text-black transition-all checked:border-black checked:bg-black checked:before:opacity-0"
                  :id="`type-${apiType}`" required />
                <span class="ml-2 text-sm font-medium">{{ apiType }}</span>
              </label>
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-600">Visibility (Immutable)</label>
            <div class="mt-1 flex items-center justify-start h-full">
              <label v-for="visibilityType in PROVIDER_VISIBILITY" :key="visibilityType" :class="[
                'relative flex items-center p-3 rounded-lg cursor-pointer transition duration-150 mr-2',
                newProvider.visibilityType === visibilityType
                  ? 'bg-gray-200 text-black'
                  : 'bg-white text-gray-600 hover:bg-gray-50',
              ]">
                <input type="radio" :value="visibilityType" v-model="newProvider.visibilityType"
                  name="visibility-toggle"
                  class="peer relative h-5 w-5 cursor-pointer appearance-none rounded-full border border-gray-300 text-black transition-all checked:border-black checked:bg-black checked:before:opacity-0"
                  :id="`vis-${visibilityType}`" required />
                <span class="ml-2 text-sm font-medium">{{ visibilityType }}</span>
              </label>
            </div>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label for="name-create" class="block text-sm font-medium text-gray-600">Name (Required, a-zA-Z0-9)</label>
            <input id="name-create" v-model="newProvider.name" @input="formatProviderName" type="text" required
              placeholder="e.g., StripeApi"
              class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm focus:ring-black focus:border-black" />
          </div>

          <div>
            <label for="baseurl-create" class="block text-sm font-medium text-gray-600">Base URL (Required,
              Immutable)</label>
            <input id="baseurl-create" v-model="newProvider.baseURL" type="url" required
              placeholder="e.g., https://api.stripe.com/v1"
              class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm focus:ring-black focus:border-black" />
          </div>

          <div class="md:col-span-2">
            <label for="prefix-create" class="block text-sm font-medium text-gray-600">Secret Prefix (Auto-generated
              from Base URL)</label>
            <input id="prefix-create" v-model="newProvider.secretPrefix" type="text" readonly
              placeholder="e.g., API_STRIPE_COM"
              class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm font-mono focus:ring-black focus:border-black bg-gray-100 cursor-not-allowed" />
            <p class="mt-1 text-xs text-gray-500">
              All related ENV vars must start with this (e.g., `{{
                newProvider.secretPrefix || 'API_STRIPE_COM'
              }}_API_KEY`).
            </p>
          </div>
        </div>

        <div>
          <label for="docurl-create" class="block text-sm font-medium text-gray-600">Documentation URL
            (Optional)</label>
          <input id="docurl-create" v-model="newProvider.documentURL" type="url"
            placeholder="e.g., https://docs.stripe.com"
            class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm focus:ring-black focus:border-black" />
        </div>

        <div>
          <label for="description-create" class="block text-sm font-medium text-gray-600">Description (Required)</label>
          <textarea id="description-create" v-model="newProvider.description" required
            placeholder="A short summary of what this provider does." rows="2"
            class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm focus:ring-black focus:border-black"></textarea>
        </div>

        <div>
          <label for="iconurl-create" class="block text-sm font-medium text-gray-600">Icon 192x192 px (Optional)</label>
          <input id="iconurl-create" v-model="newProvider.iconURL" type="url"
            placeholder="e.g., https://example.com/icon.png"
            class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm focus:ring-black focus:border-black" />
        </div>

        <div class="border-t pt-4">
          <div class="flex items-center mb-4">
            <label class="relative inline-flex items-center cursor-pointer">
              <input type="checkbox" v-model="newProvider.enableOAuth" class="sr-only peer" />
              <div
                class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-gray-300 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:left-0.5 after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-black">
              </div>
              <span class="ml-3 text-sm font-medium text-gray-700">Enable OAuth2 Configuration (Optional)</span>
            </label>
          </div>

          <div v-if="newProvider.enableOAuth" class="grid grid-cols-1 md:grid-cols-2 gap-4 bg-gray-50 p-4 rounded-lg">
            <div>
              <label class="block text-sm font-medium text-gray-600">Client ID</label>
              <input v-model="newProvider.oauthClientID" type="text"
                class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-2 text-sm focus:ring-black focus:border-black" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-600">Client Secret</label>
              <input v-model="newProvider.oauthClientSecret" type="password"
                class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-2 text-sm focus:ring-black focus:border-black" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-600">Auth URL</label>
              <input v-model="newProvider.oauthAuthURL" type="url" placeholder="https://..."
                class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-2 text-sm focus:ring-black focus:border-black" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-600">Token URL</label>
              <input v-model="newProvider.oauthTokenURL" type="url" placeholder="https://..."
                class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-2 text-sm focus:ring-black focus:border-black" />
            </div>
          </div>
        </div>

        <div class="flex justify-end pt-4 border-t">
          <button type="submit"
            class="w-full md:w-auto px-6 py-2 bg-black text-white font-semibold rounded-lg shadow-xl hover:bg-gray-800 transition duration-150 transform hover:scale-[1.01]">
            Create Provider
          </button>
        </div>
      </form>
    </div>

    <div class="flex flex-col md:flex-row justify-between items-center bg-white p-4 rounded-xl shadow-md mb-6">
      <div class="w-full md:w-1/3 mb-4 md:mb-0">
        <input v-model="searchQuery" type="text" placeholder="Search providers by name, description, or base URL..."
          class="block w-full border border-gray-300 rounded-full shadow-sm p-3 text-sm focus:ring-black focus:border-black" />
      </div>
      <div class="flex space-x-2 text-sm">
        <button @click="visibilityFilter = 'ALL'" :class="[
          'px-3 py-1 rounded-full font-medium transition-colors duration-150',
          visibilityFilter === 'ALL'
            ? 'bg-black text-white'
            : 'bg-gray-200 text-gray-700 hover:bg-gray-300',
        ]">
          ALL
        </button>
        <button @click="visibilityFilter = 'INTERNAL'" :class="[
          'px-3 py-1 rounded-full font-medium transition-colors duration-150',
          visibilityFilter === 'INTERNAL'
            ? 'bg-indigo-600 text-white'
            : 'bg-indigo-100 text-indigo-700 hover:bg-indigo-200',
        ]">
          INTERNAL
        </button>
        <button @click="visibilityFilter = 'PUBLIC'" :class="[
          'px-3 py-1 rounded-full font-medium transition-colors duration-150',
          visibilityFilter === 'PUBLIC'
            ? 'bg-green-600 text-white'
            : 'bg-green-100 text-green-700 hover:bg-green-200',
        ]">
          PUBLIC
        </button>
      </div>

      <div class="text-sm font-medium text-gray-600">
        Showing {{ displayedProviders.length }} of {{ sortedAndFilteredProviders.length }} Providers
      </div>
    </div>

    <div v-if="displayedProviders.length > 0" class="space-y-4">
      <div v-for="provider in displayedProviders" :key="provider.id"
        class="bg-white p-6 rounded-xl shadow-lg hover:shadow-xl transition duration-300 cursor-pointer border-l-4 border-black group"
        @click="viewProvider(provider.id)">
        <div class="grid grid-cols-12 items-center gap-4">
          <div class="col-span-12 md:col-span-10 flex items-center space-x-4">
            <div
              class="w-12 h-12 flex items-center justify-center shrink-0 bg-gray-100 rounded-full border-gray-300 p-3">
              <img v-if="provider.iconURL" :src="provider.iconURL" :alt="`${provider.name} Icon`"
                class="w-6 h-6 rounded-full object-cover"
                onerror="this.onerror=null; this.src='https://placehold.co/48x48/CCCCCC/333333?text=API';" />

              <svg v-else class="w-6 h-6 text-black" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
              </svg>
            </div>
            <div>
              <p class="font-bold text-lg text-gray-900 group-hover:text-black transition duration-150">
                {{ provider.name }}
                <span class="text-xs font-semibold px-2 py-0.5 rounded-full bg-gray-200 text-gray-700 ml-1">{{
                  provider.apiType }}</span>
                <span :class="[
                  'text-xs font-semibold px-2 py-0.5 rounded-full ml-1',
                  provider.visibilityType === 'INTERNAL'
                    ? 'bg-indigo-100 text-indigo-700'
                    : 'bg-green-100 text-green-700',
                ]">
                  {{ provider.visibilityType }}
                </span>
              </p>
              <p class="text-xs text-gray-500 truncate">{{ provider.baseURL }}</p>
            </div>
          </div>

          <div class="col-span-6 md:col-span-1">
            <p class="text-sm text-gray-600 font-medium">Version</p>
            <p class="text-lg font-extrabold text-black">
              {{ provider.version }}
            </p>
          </div>

          <div class="col-span-12 md:col-span-1 flex justify-end space-x-2">
            <button @click.stop="viewProvider(provider.id)"
              class="text-black hover:text-gray-700 font-semibold p-1 transition duration-150">
              <svg class="w-6 h-6 inline-block mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
              </svg>
              View
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="text-center py-10 text-gray-500 bg-white rounded-lg shadow-md">
      No providers match your search criteria.
    </div>

    <div v-if="hasMoreProviders" class="mt-8 text-center">
      <button @click="loadMore"
        class="px-6 py-3 text-sm font-medium rounded-lg text-black border border-black hover:bg-gray-100 transition duration-150 shadow-md">
        Load
        {{
          Math.min(itemsPerPage, sortedAndFilteredProviders.length - displayedProviders.length)
        }}
        More Providers (Total: {{ sortedAndFilteredProviders.length }})
      </button>
    </div>
  </div>
</template>