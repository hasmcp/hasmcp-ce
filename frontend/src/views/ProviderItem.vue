<script setup>
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useProviderStore } from '../stores/providerStore.js'
import { inferSchemaOrReturnOriginal, generatePathArgsSchema } from '../utils/schemaUtils'
import { useMcpserverStore } from '../stores/serverStore.js'
import { useToastStore } from '../stores/toastStore.js'
import DeleteModal from '../components/DeleteModal.vue'
import OpenApiImportModal from '../components/OpenApiImportModal.vue'

const route = useRoute()
const router = useRouter()
const store = useProviderStore()
const toastStore = useToastStore()
const serverStore = useMcpserverStore()

// ----------------------------------------------------
// UTILITY & Store Access
// ----------------------------------------------------

const HTTP_METHODS = store.helpers.HTTP_METHODS
const methodsWithBody = ['POST', 'PUT', 'PATCH']
const queryArgTypes = ref(['string', 'number', 'bool', 'array of strings', 'array of numbers'])
const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms))

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

// ----------------------------------------------------
// State Management (Local)
// ----------------------------------------------------

const providerId = computed(() => route.params.id.trim())

const provider = ref(null)
const isLoading = ref(true)

const expandedState = ref({})
const isToolFormVisible = ref(false)
const editingTool = ref(null)
const toolSearchQuery = ref('')
const toolsPerPage = 10
const visibleToolCount = ref(toolsPerPage)
const isLoadingTools = ref(false)

const isProviderEditing = ref(false)
const editedProvider = ref(null)
const isOAuthEditEnabled = ref(false)

// OPENAPI IMPORT STATE
const isImportModalVisible = ref(false)
const isImporting = ref(false)
const importProgress = ref(0)
const importTotal = ref(0)
const importMessage = ref('')

const headerValidationErrors = ref([])
const nameValidationError = ref('')


const showDeleteModal = ref(false)
const itemToDelete = ref(null)
const deleteNameForModal = ref('')
const itemTypeForModal = ref('')

// ----------------------------------------------------
// Computed Properties
// ----------------------------------------------------
const currentToolList = computed(() => {
  return provider.value?.tools || []
})

const filteredTools = computed(() => {
  // Create a new array to avoid mutating the original
  let list = [...currentToolList.value]

  // 1. Filtering
  if (toolSearchQuery.value.trim()) {
    const query = toolSearchQuery.value.toLowerCase().trim()
    list = list.filter(
      (e) =>
        e.path.toLowerCase().includes(query) ||
        e.description.toLowerCase().includes(query) ||
        e.method.toLowerCase().includes(query),
    )
  }

  // 2. Sorting (by path, then method)
  list.sort((a, b) => {
    // First, compare by path
    const pathCompare = a.path.localeCompare(b.path)
    if (pathCompare !== 0) {
      return pathCompare
    }
    // If paths are identical, compare by method
    return a.method.localeCompare(b.method)
  })

  return list
})

const displayedTools = computed(() => {
  return filteredTools.value.slice(0, visibleToolCount.value)
})

const hasMoreTools = computed(() => {
  return displayedTools.value.length < filteredTools.value.length
})

const isBodyRequired = computed(() => {
  if (!editingTool.value) return false
  return methodsWithBody.includes(editingTool.value.method)
})

const availableMethods = computed(() => {
  return HTTP_METHODS
})

const canAddTool = computed(() => {
  if (!provider.value) return false
  return true
})

// ----------------------------------------------------
// Methods - Data Fetching
// ----------------------------------------------------
const fetchProviderData = async (id) => {
  if (!id) return
  isLoading.value = true
  provider.value = null

  try {
    const data = await store.getProviderById(id)

    if (data) {
      provider.value = data
    } else {
      toastStore.showToast('Provider not found.', 'warning', 3000)
      router.push({ name: 'Providers' })
    }
  } catch (error) {
    console.error('Failed to fetch provider:', error)
    router.push({ name: 'Providers' })
  } finally {
    isLoading.value = false
  }
}

// ----------------------------------------------------
// Methods - Provider CRUD
// ----------------------------------------------------

const formatProviderName = (event) => {
  let value = event.target.value
  value = value.replace(/[^a-zA-Z0-9]/g, '').slice(0, 16)
  editedProvider.value.name = value

  if (event.target.value !== value) {
    event.target.value = value
  }
}

const startProviderEdit = () => {
  editedProvider.value = JSON.parse(JSON.stringify(provider.value))
  if (editedProvider.value.oauth2Config) {
    isOAuthEditEnabled.value = true
  } else {
    isOAuthEditEnabled.value = false
    editedProvider.value.oauth2Config = {
      clientID: '',
      clientSecret: '',
      authURL: '',
      tokenURL: '',
    }
  }
  isProviderEditing.value = true
}

const cancelProviderEdit = () => {
  editedProvider.value = null
  isProviderEditing.value = false
}

const saveProviderEdit = async () => {
  if (!editedProvider.value.name || !editedProvider.value.description) {
    toastStore.showToast('Name and Description are required!', 'warning', 3000)
    return
  }

  if (editedProvider.value.documentURL && !isValidUrl(editedProvider.value.documentURL)) {
    toastStore.showToast('Documentation URL must be a valid HTTP/HTTPS URL.', 'warning', 3000)
    return
  }

  if (editedProvider.value.iconURL && !isValidUrl(editedProvider.value.iconURL)) {
    toastStore.showToast('Icon URL must be a valid HTTP/HTTPS URL.', 'warning', 3000)
    return
  }

  if (isOAuthEditEnabled.value) {
    const conf = editedProvider.value.oauth2Config
    if (!conf.clientID || !conf.clientSecret || !conf.authURL || !conf.tokenURL) {
      toastStore.showToast('All OAuth2 fields are required when enabled.', 'warning', 3000)
      return
    }
    if (!isValidUrl(conf.authURL)) {
      toastStore.showToast('OAuth Auth URL must be a valid HTTP/HTTPS URL.', 'warning', 3000)
      return
    }
    if (!isValidUrl(conf.tokenURL)) {
      toastStore.showToast('OAuth Token URL must be a valid HTTP/HTTPS URL.', 'warning', 3000)
      return
    }
  } else {
    editedProvider.value.oauth2Config = null
  }

  await store.updateProvider(editedProvider.value)
  toastStore.showToast(`Provider "${editedProvider.value.name}" details saved successfully!`)
  cancelProviderEdit()

  await fetchProviderData(providerId.value)
}

const convertProviderToServer = async () => {
  if (!provider.value) return

  // 1. Create the Server draft
  const mcpServerDraft = {
    id: null,
    name: `${provider.value.name}`,
    instructions: provider.value.description,
    version: 1, // Start at version 1
    requestHeadersProxyEnabled: true,
    providers: [
      {
        id: provider.value.id,
        version: provider.value.version,
        // Enable all tools from the provider
        enabledTools: JSON.parse(JSON.stringify(provider.value.tools || [])),
      },
    ],
    tokens: [],
  }

  // 2. Check that there are tools to add
  if (mcpServerDraft.providers[0].enabledTools.length === 0) {
    toastStore.showToast('This provider has no tools to convert.', 'warning', 3000)
    return
  }

  // 3. Check for name collision
  if (serverStore.hasMcpserverName(mcpServerDraft.name, null)) {
    toastStore.showToast(
      `An MCP Server named "${mcpServerDraft.name}" already exists.`,
      'warning',
      4000,
    )
    return
  }

  // 4. Save and redirect
  try {
    const newId = await serverStore.saveMcpserver(mcpServerDraft)
    toastStore.showToast(`MCP Server "${mcpServerDraft.name}" created from provider.`, 'info', 3000)
    router.push({ name: 'ServerDetail', params: { id: newId } })
  } catch (error) {
    console.error('Failed to convert provider to MCP server:', error)
    toastStore.showToast('Failed to create MCP Server.', 'alert', 3000)
  }
}

// ----------------------------------------------------
// Methods - Tool Management
// ----------------------------------------------------
const convertQueryArgsToSchema = (argsList) => {
  const validArgs = argsList.filter((arg) => arg.name.trim() !== '')
  if (validArgs.length === 0) {
    return null
  }
  const schema = {
    type: 'object',
    properties: {},
    required: [],
  }
  validArgs.forEach((arg) => {
    const propSchema = {}
    const argName = arg.name.trim()
    switch (arg.type) {
      case 'string':
        propSchema.type = 'string'
        break
      case 'number':
        propSchema.type = 'number'
        break
      case 'bool':
        propSchema.type = 'boolean'
        break
      case 'array of strings':
        propSchema.type = 'array'
        propSchema.items = { type: 'string' }
        break
      case 'array of numbers':
        propSchema.type = 'array'
        propSchema.items = { type: 'number' }
        break
    }
    propSchema.description = arg.description
    if (arg.required) {
      schema.required.push(argName)
    }
    schema.properties[argName] = propSchema
  })
  return schema
}

const resolveQueryArgsFromSchema = (schema) => {
  if (!schema || schema.type !== 'object' || !schema.properties) {
    return [{ name: '', type: 'string', required: false, description: '' }]
  }
  const argsList = []
  const requiredList = new Set(schema.required || [])
  for (const [name, propSchema] of Object.entries(schema.properties)) {
    const arg = {
      name: name,
      required: requiredList.has(name),
    }
    if (propSchema.type === 'array') {
      if (propSchema.items && propSchema.items.type === 'string') {
        arg.type = 'array of strings'
      } else if (propSchema.items && propSchema.items.type === 'number') {
        arg.type = 'array of numbers'
      }
    } else if (propSchema.type === 'string') {
      arg.type = 'string'
    } else if (propSchema.type === 'number') {
      arg.type = 'number'
    } else if (propSchema.type === 'boolean') {
      arg.type = 'bool'
    } else {
      arg.type = 'string'
    }
    arg.description = ''
    if (propSchema.description) {
      arg.description = propSchema.description
    }
    argsList.push(arg)
  }
  if (argsList.length === 0) {
    return [{ name: '', type: 'string', required: false, description: '' }]
  }
  return argsList
}

const loadMoreTools = () => {
  isLoadingTools.value = true
  setTimeout(() => {
    visibleToolCount.value += toolsPerPage
    isLoadingTools.value = false
  }, 400)
}

const toggleToolForm = (tool = null) => {
  isImportModalVisible.value = false

  if (tool) {
    // If clicking edit on an already-open edit form, close it.
    if (
      isToolFormVisible.value &&
      editingTool.value &&
      editingTool.value.id === tool.id
    ) {
      cancelToolEdit()
      return
    }

    const deepCopy = JSON.parse(JSON.stringify(tool))
    editingTool.value = deepCopy
    if (!editingTool.value.headers || editingTool.value.headers.length === 0) {
      editingTool.value.headers = [{ key: '', value: '' }]
    }
    editingTool.value.queryArgs = resolveQueryArgsFromSchema(deepCopy.queryArgsJSONSchema)
    editingTool.value.reqBodyJSONSchema =
      JSON.stringify(deepCopy.reqBodyJSONSchema, null, 2) || null

    if (!editingTool.value.oauth2Scopes) editingTool.value.oauth2Scopes = []

    // Hide read-only expanded view for this tool
    expandedState.value[tool.id] = false
  } else {
    // This is for "Add New"
    const defaultMethod = 'GET'
    const defaultPath = '/'
    const defaultBody = null
    const defaultHeaders = [{ key: '', value: '' }]
    editingTool.value = {
      id: null,
      method: defaultMethod,
      path: defaultPath,
      description: '',
      headers: defaultHeaders,
      reqBodyJSONSchema: defaultBody,
      queryArgsJSONSchema: null,
      pathArgsJSONSchema: null,
      queryArgs: [{ name: '', type: 'string', required: false, description: '' }],
      oauth2Scopes: [],
    }
  }
  isToolFormVisible.value = true
  headerValidationErrors.value = []
}

const cancelToolEdit = () => {
  editingTool.value = null
  isToolFormVisible.value = false
}

const saveTool = async () => {
  if (
    !editingTool.value.method ||
    !editingTool.value.path ||
    !editingTool.value.description
  ) {
    toastStore.showToast('Method, Path, and Description are required!', 'warning', 3000)
    return
  }
  if (!editingTool.value.path.startsWith('/')) {
    toastStore.showToast('Tool Path must start with a forward slash (/)', 'warning', 3000)
    return
  }

  const nameRegex = /^[a-z][a-zA-Z0-9]{0,19}$/
  nameValidationError.value = ''
  if (editingTool.value.name && editingTool.value.name.trim() !== '') {
    if (!nameRegex.test(editingTool.value.name)) {
      nameValidationError.value = "Invalid Name: Must start with lowercase, followed by letters or numbers (max 20 chars)."
    }
  }

  if (nameValidationError.value !== '') {
    toastStore.showToast(nameValidationError.value, 'warning', 3000)
    return
  }

  const requiredPrefix = provider.value.secretPrefix
  let hasHeaderError = false
  headerValidationErrors.value = []
  const envVarRegex = /\$\{([A-Z0-9_]+)\}/g

  for (let i = 0; i < editingTool.value.headers.length; i++) {
    const header = editingTool.value.headers[i]
    let localError = ''
    if (header.value.trim() !== '') {
      let match
      envVarRegex.lastIndex = 0
      while ((match = envVarRegex.exec(header.value)) !== null) {
        const varName = match[1]
        if (!varName.startsWith(requiredPrefix)) {
          localError = `Variable '${varName}' must start with prefix: '${requiredPrefix}'.`
          hasHeaderError = true
          break
        }
      }
    }
    headerValidationErrors.value[i] = localError
  }

  if (hasHeaderError) {
    return
  }

  const cleanedHeaders = editingTool.value.headers.filter(
    (h) => h.key.trim() !== '' || h.value.trim() !== '',
  )
  editingTool.value.headers = cleanedHeaders.length > 0 ? cleanedHeaders : []

  let finalReqBodySchema = null
  if (isBodyRequired.value && editingTool.value.reqBodyJSONSchema) {
    const method = editingTool.value.method
    const path = editingTool.value.path
    const convertedSchema = inferSchemaOrReturnOriginal(
      editingTool.value.reqBodyJSONSchema,
      method,
      path,
    )
    finalReqBodySchema = JSON.parse(convertedSchema)
  }

  const querySchema = convertQueryArgsToSchema(editingTool.value.queryArgs)
  const pathSchema = generatePathArgsSchema(editingTool.value.path)

  let payload = {
    id: editingTool.value.id,
    name: editingTool.value.name,
    title: editingTool.value.title,
    description: editingTool.value.description,
    pathArgsJSONSchema: pathSchema || null,
    queryArgsJSONSchema: querySchema,
    reqBodyJSONSchema: finalReqBodySchema,
    resBodyJSONSchema: editingTool.value.resBodyJSONSchema || null,
    headers: editingTool.value.headers,
    oauth2Scopes: (editingTool.value.oauth2Scopes || []).filter((s) => s && s.trim() !== ''),
  }

  if (!editingTool.value.id) {
    payload.method = editingTool.value.method
    payload.path = editingTool.value.path
    await store.createProviderTool(providerId.value, payload)
  } else {
    await store.updateProviderTool(providerId.value, payload)
  }

  toastStore.showToast(`Tool ${editingTool.value.id ? 'updated' : 'created'} successfully!`)
  cancelToolEdit()
  visibleToolCount.value = toolsPerPage
  toolSearchQuery.value = ''

  await fetchProviderData(providerId.value)
}

const toggleToolExpansion = (toolId) => {
  // If edit form is open for this tool, close it
  if (
    isToolFormVisible.value &&
    editingTool.value &&
    editingTool.value.id === toolId
  ) {
    cancelToolEdit()
  }

  expandedState.value = {
    ...expandedState.value,
    [toolId]: !expandedState.value[toolId],
  }
}

// ----------------------------------------------------
// 7. Methods - Header & Query Arg UI
// ----------------------------------------------------
const addHeader = (tool) => {
  tool.headers.push({ key: '', value: '' })
}

const removeHeader = (tool, index) => {
  tool.headers.splice(index, 1)
  headerValidationErrors.value.splice(index, 1)
  if (tool.headers.length === 0) {
    addHeader(tool)
  }
}

const addQueryArg = (tool) => {
  tool.queryArgs.push({ name: '', type: 'string', required: false, description: '' })
}

const removeQueryArg = (tool, index) => {
  tool.queryArgs.splice(index, 1)
  if (tool.queryArgs.length === 0) {
    addQueryArg(tool)
  }
}

const addScope = (tool) => {
  if (!tool.oauth2Scopes) tool.oauth2Scopes = []
  tool.oauth2Scopes.push('')
}

const removeScope = (tool, index) => {
  tool.oauth2Scopes.splice(index, 1)
}

const getMethodColor = (method) => {
  const colors = {
    GET: 'bg-green-600 border-green-600',
    POST: 'bg-blue-600 border-blue-600',
    PUT: 'bg-yellow-600 border-yellow-600',
    DELETE: 'bg-red-600 border-red-600',
    PATCH: 'bg-orange-600 border-orange-600',
    HEAD: 'bg-gray-500 border-gray-500',
    OPTIONS: 'bg-indigo-500 border-indigo-500',
  }
  return colors[method] || 'bg-gray-400 border-gray-400'
}

const getMethodBorderColor = (method) => {
  const colorClasses = getMethodColor(method)
    .split(' ')
    .find((c) => c.startsWith('border-'))
  return colorClasses || 'border-gray-400'
}

// ----------------------------------------------------
// 8. Modal Handlers
// ----------------------------------------------------
const initiateDelete = (item, type) => {
  itemToDelete.value = { id: item.id, type }
  deleteNameForModal.value = type === 'provider' ? item.name : `${item.method} ${item.path}`
  itemTypeForModal.value = type === 'provider' ? 'Provider' : 'Tool'
  showDeleteModal.value = true
}

const confirmDelete = async () => {
  if (!itemToDelete.value) return

  if (itemToDelete.value.type === 'provider') {
    await store.deleteProvider(itemToDelete.value.id)
    router.push({ name: 'Providers' })
  } else if (itemToDelete.value.type === 'tool') {
    await store.deleteProviderTool(providerId.value, itemToDelete.value.id)
    if (editingTool.value && editingTool.value.id === itemToDelete.value.id) {
      cancelToolEdit()
    }
    await fetchProviderData(providerId.value)
  }
  cancelDelete()
}

const cancelDelete = () => {
  showDeleteModal.value = false
  itemToDelete.value = null
  deleteNameForModal.value = ''
  itemTypeForModal.value = ''
}

// ----------------------------------------------------
// 9. OpenAPI Import Logic
// ----------------------------------------------------

/**
 * Recursively resolves $ref pointers within a schema object.
 * Only supports local #/ references.
 */
function dereferenceSchema(obj, fullSpec) {
  if (typeof obj !== 'object' || obj === null) {
    return obj
  }

  if (Array.isArray(obj)) {
    return obj.map((item) => dereferenceSchema(item, fullSpec))
  }

  if (obj.$ref) {
    const refPath = obj.$ref
    if (typeof refPath === 'string' && refPath.startsWith('#/')) {
      try {
        const pathParts = refPath.substring(2).split('/')
        let target = fullSpec
        for (const part of pathParts) {
          target = target[part]
          if (target === undefined) {
            throw new Error(`Path ${refPath} not found in spec.`)
          }
        }
        return dereferenceSchema(JSON.parse(JSON.stringify(target)), fullSpec)
      } catch (e) {
        console.warn(`Could not resolve $ref: ${refPath}`, e.message)
        return obj // Return original $ref object on failure
      }
    } else {
      return obj
    }
  }

  const newObj = {}
  for (const key in obj) {
    newObj[key] = dereferenceSchema(obj[key], fullSpec)
  }
  return newObj
}

/**
 * Creates a standard, prefixed variable name from a requirement name.
 */
function createVariableName(prefix, reqName) {
  const sanitizedName = reqName.toUpperCase().replace(/-/g, '_')
  return `${prefix}_${sanitizedName}`
}

/**
 * Translates an OpenAPI path item into the format required by our tool store.
 */
function translateOpenApiToTool(path, method, toolData, providerSecretPrefix, fullSpec) {
  const methodUpper = method.toUpperCase()
  const name = (toolData.operationId || toolData.summary || '').replace(/[^a-z0-9A-Z]/gi, '').slice(0, 20)
  const title = (toolData.summary || '').slice(0, 64)
  const description = (toolData.description || toolData.summary || 'No description provided.').slice(0, 512)
  const headers = []
  const headerKeys = new Set()

  // 1. Process 'parameters' in 'header'
  if (toolData.parameters) {
    toolData.parameters
      .filter((p) => p.in === 'header')
      .forEach((p) => {
        const key = p.name
        if (!headerKeys.has(key)) {
          // Assume all header parameters are variables
          const varName = createVariableName(providerSecretPrefix, p.name)
          headers.push({ key: key, value: `$\{${varName}}` })
          headerKeys.add(key)
        }
      })
  }

  // 2. Process 'security'
  const security = toolData.security || fullSpec.security || []
  const securitySchemes = fullSpec.components?.securitySchemes || {}
  const oauth2Scopes = new Set()

  security.forEach((securityReq) => {
    for (const reqName in securityReq) {
      const scheme = securitySchemes[reqName]
      if (!scheme) continue

      // Handle OAuth2 Scopes
      if (scheme.type === 'oauth2') {
        const scopes = securityReq[reqName] || []
        scopes.forEach((s) => oauth2Scopes.add(s))
      }

      if (scheme.type === 'apiKey' && scheme.in === 'header') {
        const key = scheme.name // e.g., 'X-API-Key'
        if (!headerKeys.has(key)) {
          const varName = createVariableName(providerSecretPrefix, reqName) // e.g., 'API_KEY_AUTH'
          headers.push({ key: key, value: `$\{${varName}}` })
          headerKeys.add(key)
        }
      } else if (scheme.type === 'http' && scheme.scheme === 'bearer') {
        const key = 'Authorization'
        if (!headerKeys.has(key)) {
          const varName = createVariableName(providerSecretPrefix, reqName) // e.g., 'BEARER_AUTH'
          headers.push({ key: key, value: `Bearer $\{${varName}}` })
          headerKeys.add(key)
        }
      }
    }
  })

  // 3. Path Args Schema
  const pathArgsJSONSchema = generatePathArgsSchema(path) || null

  // 4. Query Args Schema
  let queryArgsJSONSchema = null
  if (toolData.parameters) {
    const queryParams = toolData.parameters.filter((p) => p.in === 'query')
    if (queryParams.length > 0) {
      const querySchema = {
        type: 'object',
        properties: {},
        required: [],
      }
      queryParams.forEach((p) => {
        const paramSchema = dereferenceSchema(p.schema || {}, fullSpec)

        querySchema.properties[p.name] = {
          type: paramSchema.type || 'string',
          description: p.description || '',
          ...(paramSchema.type === 'array' && { items: paramSchema.items || { type: 'string' } }),
        }
        if (p.required) {
          querySchema.required.push(p.name)
        }
      })
      queryArgsJSONSchema = querySchema
    }
  }

  // 5. Request Body Schema
  let reqBodyJSONSchema = null
  if (toolData.requestBody) {
    const dereferencedBody = dereferenceSchema(toolData.requestBody, fullSpec)
    if (dereferencedBody.content) {
      const jsonSchema = dereferencedBody.content['application/json']?.schema
      if (jsonSchema) {
        // Pass the *resolved* schema to our utility
        reqBodyJSONSchema = JSON.parse(
          inferSchemaOrReturnOriginal(JSON.stringify(jsonSchema), methodUpper, path),
        )
      }
    }
  }

  // 6. Build the final payload
  const toolPayload = {
    method: methodUpper,
    path: path,
    name: name,
    title: title,
    description: description,
    headers: headers, // Already filtered for uniqueness
    pathArgsJSONSchema: pathArgsJSONSchema,
    queryArgsJSONSchema: queryArgsJSONSchema,
    reqBodyJSONSchema: reqBodyJSONSchema,
    resBodyJSONSchema: null,
    oauth2Scopes: Array.from(oauth2Scopes).filter((s) => s && s.trim() !== ''),
  }

  return toolPayload
}

/**
 * Main handler for starting the import process.
 */
async function handleImport(parsedSpec) {
  isImporting.value = true
  importProgress.value = 0
  importTotal.value = 0
  importMessage.value = 'Preparing import...'

  const toolsToProcess = []

  // --- SPEC VERSION DETECTION ---
  const isSwagger2 =
    parsedSpec.swagger && (parsedSpec.swagger === '2.0' || parsedSpec.swagger.startsWith('2.'))
  const isOAS3 = parsedSpec.openapi && parsedSpec.openapi.startsWith('3.')

  if (!isSwagger2 && !isOAS3) {
    toastStore.showToast('Invalid or unsupported spec version.', 'warning', 3000)
    isImporting.value = false
    return
  }
  const validMethods = new Set(HTTP_METHODS.map((m) => m.toLowerCase()))

  // 1. Flatten the spec into a list of tasks
  for (const path in parsedSpec.paths) {
    for (const method in parsedSpec.paths[path]) {
      if (validMethods.has(method)) {
        toolsToProcess.push({
          path: path,
          method: method,
          data: parsedSpec.paths[path][method],
        })
      }
    }
  }

  importTotal.value = toolsToProcess.length
  if (importTotal.value === 0) {
    isImporting.value = false
    toastStore.showToast('No valid tools found to import.', 'warning', 3000)
    return
  }

  // 2. Get existing tools for conflict checking
  await fetchProviderData(providerId.value)
  const existingTools = provider.value?.tools || []
  const toolMap = new Map(
    existingTools.map((e) => [`${e.method.toUpperCase()}|${e.path}`, e]),
  )

  const providerPrefix = provider.value.secretPrefix

  // 3. Process each tool sequentially
  for (const task of toolsToProcess) {
    importProgress.value++
    const methodUpper = task.method.toUpperCase()
    importMessage.value = `Importing ${importProgress.value} of ${importTotal.value}: ${methodUpper} ${task.path}`

    nextTick()

    try {
      let payload
      if (isOAS3) {
        // --- OAS3 LOGIC ---
        payload = translateOpenApiToTool(
          task.path,
          task.method,
          task.data,
          providerPrefix,
          parsedSpec,
        )
      } else {
        // --- SWAGGER 2 LOGIC ---
        payload = translateSwagger2ToTool(
          task.path,
          task.method,
          task.data,
          providerPrefix,
          parsedSpec, // Pass the full spec for resolving definitions
        )
      }

      const existingTool = toolMap.get(`${methodUpper}|${task.path}`)

      if (existingTool) {
        payload.id = existingTool.id
        await store.updateProviderTool(providerId.value, payload)
      } else {
        await store.createProviderTool(providerId.value, payload)
      }

      await sleep(200)
    } catch (error) {
      console.error(`Failed to import tool: ${methodUpper} ${task.path}`, error)
      toastStore.showToast(`Failed to import ${methodUpper} ${task.path}: ${error.message}`, 4000)
    }
  }

  // 4. Finalize
  importMessage.value = `Import complete! ${importTotal.value} tools processed.`
  toastStore.showToast('Import complete! Refreshing tool list.', 'info', 3000)

  setTimeout(() => {
    isImporting.value = false
    importMessage.value = ''
  }, 4000)

  await fetchProviderData(providerId.value)
}

/**
 * Translates a Swagger 2.0 path item into the format required by our tool store.
 */
function translateSwagger2ToTool(path, method, toolData, providerSecretPrefix, fullSpec) {
  const methodUpper = method.toUpperCase()
  const description = toolData.description || toolData.summary || 'No description provided.'
  const headers = []
  const headerKeys = new Set()

  // Combine parameters from path level and operation level
  const pathParams = fullSpec.paths[path].parameters || []
  const operationParams = toolData.parameters || []
  const allParameters = [...pathParams, ...operationParams]

  // 1. Process 'parameters' in 'header'
  allParameters
    .filter((p) => p.in === 'header')
    .forEach((p) => {
      const key = p.name
      if (!headerKeys.has(key)) {
        const varName = createVariableName(providerSecretPrefix, p.name)
        headers.push({ key: key, value: `$\{${varName}}` })
        headerKeys.add(key)
      }
    })

  // 2. Process 'security'
  const security = toolData.security || fullSpec.security || []
  const securityDefinitions = fullSpec.securityDefinitions || {}
  const oauth2Scopes = new Set()

  security.forEach((securityReq) => {
    for (const reqName in securityReq) {
      const scheme = securityDefinitions[reqName]
      if (!scheme) continue

      // Handle OAuth2 Scopes
      if (scheme.type === 'oauth2') {
        const scopes = securityReq[reqName] || []
        scopes.forEach((s) => oauth2Scopes.add(s))
      }

      if (scheme.type === 'apiKey' && scheme.in === 'header') {
        const key = scheme.name // e.g., 'X-API-Key'
        if (!headerKeys.has(key)) {
          const varName = createVariableName(providerSecretPrefix, reqName) // e.g., 'API_KEY_AUTH'
          headers.push({ key: key, value: `$\{${varName}}` })
          headerKeys.add(key)
        }
      }
      // Swagger 2.0 'basic' auth is also common, maps to 'Authorization' header
      else if (scheme.type === 'basic') {
        const key = 'Authorization'
        if (!headerKeys.has(key)) {
          const varName = createVariableName(providerSecretPrefix, reqName) // e.g., 'BASIC_AUTH'
          headers.push({ key: key, value: `Basic $\{${varName}}` }) // Assuming varName holds base64(user:pass)
          headerKeys.add(key)
        }
      }
    }
  })

  // 3. Path Args Schema (This utility is generic and should work)
  const pathArgsJSONSchema = generatePathArgsSchema(path) || null

  // 4. Query Args Schema
  let queryArgsJSONSchema = null
  const queryParams = allParameters.filter((p) => p.in === 'query')
  if (queryParams.length > 0) {
    const querySchema = {
      type: 'object',
      properties: {},
      required: [],
    }
    queryParams.forEach((p) => {
      // In Swagger 2, the schema is *not* nested under a 'schema' prop for simple types.
      // 'schema' is only used for `in: 'body'`.
      const paramSchema = {
        type: p.type,
        description: p.description || '',
        ...(p.type === 'array' && { items: p.items || { type: 'string' } }),
      }

      querySchema.properties[p.name] = paramSchema
      if (p.required) {
        querySchema.required.push(p.name)
      }
    })
    queryArgsJSONSchema = querySchema
  }

  // 5. Request Body Schema
  // In Swagger 2, this is a single parameter with `in: "body"`
  let reqBodyJSONSchema = null
  const bodyParam = allParameters.find((p) => p.in === 'body')
  if (bodyParam && bodyParam.schema) {
    // Dereference the schema
    const dereferencedBodySchema = dereferenceSchema(bodyParam.schema, fullSpec)
    // Pass the *resolved* schema to our utility
    reqBodyJSONSchema = JSON.parse(
      inferSchemaOrReturnOriginal(JSON.stringify(dereferencedBodySchema), methodUpper, path),
    )
  }

  // 6. Build the final payload
  const toolPayload = {
    method: methodUpper,
    path: path,
    description: description,
    headers: headers, // Already filtered for uniqueness
    pathArgsJSONSchema: pathArgsJSONSchema,
    queryArgsJSONSchema: queryArgsJSONSchema,
    reqBodyJSONSchema: reqBodyJSONSchema,
    resBodyJSONSchema: null, // Swagger 2 responses are complex, skipping for now like OAS3
    oauth2Scopes: Array.from(oauth2Scopes).filter((s) => s && s.trim() !== ''),
  }

  return toolPayload
}

// ----------------------------------------------------
// Lifecycle Hooks
// ----------------------------------------------------

onMounted(() => {
  fetchProviderData(providerId.value)
})

watch(providerId, (newId) => {
  if (newId) {
    cancelProviderEdit()
    cancelToolEdit()
    fetchProviderData(newId)
  }
})
</script>

<template>
  <div class="p-4">
    <div v-if="!provider" class="text-center py-10 text-gray-500">Loading provider data...</div>

    <div v-else-if="provider">
      <div class="bg-white p-6 rounded-xl shadow-2xl mb-8">
        <div class="flex justify-between items-start mb-4 border-b pb-4">
          <div class="flex items-center space-x-4">
            <img v-if="provider.iconURL" :src="provider.iconURL" :alt="`${provider.name} Icon`"
              class="w-12 h-12 rounded-full object-cover" />

            <svg v-else class="w-8 h-8 text-black" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
            </svg>
            <div>
              <h1 class="text-3xl font-bold text-gray-900">
                {{ provider.name }}
                <span class="text-xs font-semibold px-2 py-1 rounded-full bg-black text-white ml-2">ID: {{ provider.id
                  }}</span>
              </h1>
              <p class="text-sm text-gray-500 mt-1 font-mono break-all">{{ provider.baseURL }}</p>
            </div>
          </div>
          <div class="flex space-x-3">
            <button @click="startProviderEdit" v-if="!isProviderEditing && !isToolFormVisible && !isImporting"
              class="p-2 rounded-lg text-gray-700 hover:text-black transition duration-150 border hover:bg-gray-50"
              title="Manually Edit Provider Info">
              <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
              </svg>
            </button>

            <button @click="convertProviderToServer" v-if="
              !isProviderEditing &&
              !isToolFormVisible &&
              !isImporting &&
              provider.tools &&
              provider.tools.length > 0
            " class="p-2 rounded-lg text-green-600 hover:text-green-800 transition duration-150 border hover:bg-green-50"
              title="Convert to MCP Server">
              <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z">
                </path>
              </svg>
            </button>

            <button @click="initiateDelete(provider, 'provider')" :disabled="isImporting"
              class="p-2 rounded-lg text-red-600 hover:text-red-800 transition duration-150 border hover:bg-red-50 disabled:opacity-50"
              title="Delete Provider">
              <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </div>
        </div>

        <div v-if="!isProviderEditing && !isToolFormVisible" class="space-y-2">
          <div class="flex space-x-6">
            <p class="text-sm text-gray-500">
              Version: <span class="font-semibold text-black">{{ provider.version }}</span>
            </p>
            <p class="text-sm text-gray-500">
              Type: <span class="font-semibold text-black">{{ provider.apiType }}</span>
            </p>
            <p class="text-sm text-gray-500">
              Visibility:
              <span :class="[
                'font-semibold',
                provider.visibility === 'INTERNAL' ? 'text-indigo-700' : 'text-green-700',
              ]">{{ provider.visibilityType }}</span>
            </p>
            <p class="text-sm text-gray-500">
              Secret Prefix:
              <span class="font-semibold text-black font-mono">{{ provider.secretPrefix }}</span>
            </p>
          </div>
          <p class="text-base font-medium text-gray-700">{{ provider.description }}</p>

          <p v-if="provider.documentURL" class="text-sm text-gray-500 flex items-center">
            Documentation:
            <a :href="provider.documentURL" target="_blank"
              class="font-semibold text-blue-600 hover:text-blue-800 ml-2 flex items-center">
              View Document
              <svg class="w-4 h-4 ml-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
              </svg>
            </a>
          </p>

          <p class="text-sm text-gray-500">
            Tools Configured:
            <span class="font-semibold text-black">{{ currentToolList.length }}</span>
          </p>

          <div v-if="provider.oauth2Config" class="mt-4 p-4 bg-gray-50 rounded-lg border border-gray-200">
            <h3 class="text-sm font-bold text-gray-700 mb-2">OAuth2 Configuration</h3>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-x-4 gap-y-2 text-sm">
              <div class="flex items-center overflow-hidden">
                <span class="text-gray-500 mr-2 shrink-0">Client ID:</span>
                <span class="font-mono text-gray-800 truncate" :title="provider.oauth2Config.clientID">
                  {{ provider.oauth2Config.clientID }}
                </span>
              </div>
              <div class="flex items-center overflow-hidden">
                <span class="text-gray-500 mr-2 shrink-0">Auth URL:</span>
                <span class="font-mono text-gray-800 truncate" :title="provider.oauth2Config.authURL">
                  {{ provider.oauth2Config.authURL }}
                </span>
              </div>
              <div class="flex items-center overflow-hidden">
                <span class="text-gray-500 mr-2 shrink-0">Token URL:</span>
                <span class="font-mono text-gray-800 truncate" :title="provider.oauth2Config.tokenURL">
                  {{ provider.oauth2Config.tokenURL }}
                </span>
              </div>
              <div>
                <span class="text-gray-500 mr-2">Client Secret:</span>
                <span class="font-mono text-gray-800">********</span>
              </div>
            </div>
          </div>
        </div>

        <form v-if="isProviderEditing && editedProvider" @submit.prevent="saveProviderEdit" class="space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-600">Version (Immutable)</label>
              <input :value="editedProvider.version" type="text" disabled
                class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm bg-gray-100 cursor-not-allowed" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-600">Provider Type (Immutable)</label>
              <input :value="editedProvider.apiType" type="text" disabled
                class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm bg-gray-100 cursor-not-allowed" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-600">Visibility (Immutable)</label>
              <input :value="editedProvider.visibilityType" type="text" disabled
                class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm bg-gray-100 cursor-not-allowed" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-600">Secret Prefix (Immutable)</label>
              <input v-model="editedProvider.secretPrefix" type="text" disabled
                class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm font-mono bg-gray-100 cursor-not-allowed" />
              <p class="mt-1 text-xs text-gray-500">
                This prefix cannot be changed after creation.
              </p>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-600">Name (a-zA-Z0-9)</label>
              <input v-model="editedProvider.name" @input="formatProviderName" type="text" required
                class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm focus:ring-black focus:border-black" />
            </div>
            <div class="md:col-span-2">
              <label class="block text-sm font-medium text-gray-600">Base URL (Immutable)</label>
              <input v-model="editedProvider.baseURL" type="url" required :disabled="true"
                class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm bg-gray-100 cursor-not-allowed" />
            </div>
            <div class="md:col-span-3">
              <label class="block text-sm font-medium text-gray-600">Documentation URL (Optional)</label>
              <input v-model="editedProvider.documentURL" type="url" placeholder="e.g., https://docs.stripe.com"
                class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm focus:ring-black focus:border-black" />
            </div>
            <div class="md:col-span-3">
              <label class="block text-sm font-medium text-gray-600">Description</label>
              <textarea v-model="editedProvider.description" required rows="2"
                class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm focus:ring-black focus:border-black"></textarea>
            </div>
            <div class="md:col-span-3">
              <label class="block text-sm font-medium text-gray-600">Icon URL</label>
              <input v-model="editedProvider.iconURL" type="url"
                class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm focus:ring-black focus:border-black" />
            </div>
          </div>

          <div class="border-t pt-4">
            <div class="flex items-center mb-4">
              <label class="relative inline-flex items-center cursor-pointer">
                <input type="checkbox" v-model="isOAuthEditEnabled" class="sr-only peer" />
                <div
                  class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-gray-300 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:left-0.5 after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-black">
                </div>
                <span class="ml-3 text-sm font-medium text-gray-700">Enable OAuth2 Configuration</span>
              </label>
            </div>

            <div v-if="isOAuthEditEnabled" class="grid grid-cols-1 md:grid-cols-2 gap-4 bg-gray-50 p-4 rounded-lg">
              <div>
                <label class="block text-sm font-medium text-gray-600">Client ID</label>
                <input v-model="editedProvider.oauth2Config.clientID" type="text"
                  class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-2 text-sm focus:ring-black focus:border-black" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-600">Client Secret</label>
                <input v-model="editedProvider.oauth2Config.clientSecret" type="text"
                  class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-2 text-sm focus:ring-black focus:border-black" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-600">Auth URL</label>
                <input v-model="editedProvider.oauth2Config.authURL" type="url"
                  class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-2 text-sm focus:ring-black focus:border-black" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-600">Token URL</label>
                <input v-model="editedProvider.oauth2Config.tokenURL" type="url"
                  class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-2 text-sm focus:ring-black focus:border-black" />
              </div>
            </div>
          </div>

          <div class="flex justify-end space-x-3 pt-4 border-t">
            <button type="button" @click="cancelProviderEdit"
              class="px-4 py-2 text-gray-700 border border-gray-300 rounded-lg hover:bg-gray-100 transition duration-150">
              Cancel
            </button>
            <button type="submit"
              class="px-4 py-2 bg-black text-white font-semibold rounded-lg shadow-md hover:bg-gray-800 transition duration-150">
              Save Changes
            </button>
          </div>
        </form>
      </div>

      <h2 class="text-2xl font-bold mb-4 text-gray-800 flex justify-between items-center">
        Provider Tools
        <div class="flex space-x-2">
          <button @click="isImportModalVisible = true" :class="[
            'p-2 rounded-full text-white bg-black hover:bg-gray-800 transition duration-150',
            {
              'opacity-50 cursor-not-allowed':
                isProviderEditing || !canAddTool || isImporting,
            },
          ]" title="Import from OpenAPI Spec" :disabled="isProviderEditing || !canAddTool || isImporting">
            <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
            </svg>
          </button>
          <button @click="toggleToolForm()" :class="[
            'p-2 rounded-full text-white bg-black hover:bg-gray-800 transition duration-150',
            {
              'opacity-50 cursor-not-allowed':
                isProviderEditing || !canAddTool || isImporting,
            },
          ]" title="Add New Tool Manually" :disabled="isProviderEditing || !canAddTool || isImporting">
            <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
          </button>
        </div>
      </h2>

      <div v-if="isImporting" class="bg-white p-6 rounded-xl shadow-lg mb-8 border-2 border-black/10">
        <h3 class="text-lg font-semibold mb-3 text-gray-800">Importing Tools...</h3>
        <p class="text-sm text-gray-600 mb-2">{{ importMessage }}</p>
        <div class="w-full bg-gray-200 rounded-full h-4 overflow-hidden">
          <div class="bg-black h-4 rounded-full transition-all duration-300 ease-out"
            :style="{ width: (importProgress / (importTotal || 1)) * 100 + '%' }"></div>
        </div>
        <p class="text-xs text-gray-500 mt-2 text-right">
          {{ importProgress }} / {{ importTotal }}
        </p>
      </div>

      <div v-if="isToolFormVisible && editingTool && editingTool.id === null"
        class="bg-white p-6 rounded-xl shadow-2xl mb-8 border-2 border-black/10">
        <h3 class="text-xl font-semibold mb-4 text-gray-700 border-b pb-2">
          {{
            editingTool.id ? `Edit Tool ID: ${editingTool.id}` : 'Add New Tool'
          }}
        </h3>
        <form @submit.prevent="saveTool" class="space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div class="col-span-1">
              <label class="block text-sm font-medium text-gray-600">Method</label>
              <select v-model="editingTool.method" required :disabled="!!editingTool.id"
                class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm focus:ring-black focus:border-black"
                :class="{ 'bg-gray-100 cursor-not-allowed': !!editingTool.id }">
                <option v-for="method in availableMethods" :key="method" :value="method">
                  {{ method }}
                </option>
              </select>
            </div>
            <div class="col-span-3">
              <label class="block text-sm font-medium text-gray-600">Path (Starts with '/', example:
                '/users/{id}/profile' where 'id' value is
                dynamic)</label>
              <input v-model="editingTool.path" type="text" required placeholder="/users/{id}/profile"
                :disabled="!!editingTool.id"
                class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm font-mono focus:ring-black focus:border-black"
                :class="{ 'bg-gray-100 cursor-not-allowed': !!editingTool.id }" />
            </div>
          </div>
          <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div class="col-span-1">
              <label class="block text-sm font-medium text-gray-600">Name <span class="text-gray-400 font-normal">(MCP
                  tool
                  name)</span></label>
              <input v-model="editingTool.name" type="text" placeholder="e.g. getUser" maxlength="20"
                class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm font-mono focus:ring-black focus:border-black"
                :class="{ 'border-red-500 ring-1 ring-red-500': nameValidationError }" />
              <p v-if="nameValidationError" class="mt-1 text-xs text-red-600 font-bold italic">{{ nameValidationError }}
              </p>
            </div>
            <div class="col-span-3">
              <label class="block text-sm font-medium text-gray-600">Title <span class="text-gray-400 font-normal">(MCP
                  tool
                  title)</span></label>
              <input v-model="editingTool.title" type="text" placeholder="e.g. Get User Profile" maxlength="64"
                class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm font-mono focus:ring-black focus:border-black" />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-600">Description</label>
            <textarea v-model="editingTool.description" required rows="1"
              placeholder="A short summary of what this tool does."
              class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm focus:ring-black focus:border-black"></textarea>
          </div>

          <div class="border border-gray-200 p-4 rounded-lg bg-gray-50">
            <h4 class="text-sm font-semibold mb-2 text-gray-700">
              HTTP Headers (Can reference ENV vars like ${VAR_NAME})
            </h4>
            <p class="text-xs text-red-600 font-medium mb-3">
              <span class="font-bold">Important:</span> Environment variables in headers *must*
              start with the prefix:
              <code class="font-mono bg-red-100 px-1 py-0.5 rounded">{{
                provider.secretPrefix
              }}</code>
            </p>
            <template v-for="(header, index) in editingTool.headers" :key="index">
              <div class="grid grid-cols-12 gap-2 items-center mb-2">
                <input v-model="header.key" placeholder="Key (e.g., Authorization)" type="text"
                  class="col-span-5 p-2 border rounded-lg text-xs font-mono focus:ring-black focus:border-black"
                  :class="{ 'border-red-500': headerValidationErrors[index] }" />
                <input v-model="header.value" placeholder="Value (e.g., Bearer ${AUTH_TOKEN})" type="text"
                  class="col-span-6 p-2 border rounded-lg text-xs font-mono focus:ring-black focus:border-black"
                  :class="{ 'border-red-500 ring-red-500': headerValidationErrors[index] }"
                  @input="headerValidationErrors[index] = ''" />
                <button type="button" @click="removeHeader(editingTool, index)" :disabled="editingTool.headers !== undefined &&
                  editingTool.headers.length === 1 &&
                  header.key === '' &&
                  header.value === ''
                  " class="col-span-1 text-red-600 hover:text-red-800 disabled:opacity-50 transition duration-150">
                  <svg class="w-5 h-5 mx-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>

                <p v-if="headerValidationErrors[index]" class="col-span-12 text-xs text-red-600 italic -mt-1 pl-2">
                  {{ headerValidationErrors[index] }}
                </p>
              </div>
            </template>
            <button type="button" @click="addHeader(editingTool)"
              class="mt-2 text-xs text-black border border-black px-3 py-1 rounded-full hover:bg-gray-100 transition duration-150">
              + Add Header
            </button>
          </div>

          <div class="border border-gray-200 p-4 rounded-lg bg-gray-50">
            <h4 class="text-sm font-semibold mb-2 text-gray-700">Query String Arguments</h4>
            <div class="space-y-3">
              <template v-for="(arg, index) in editingTool.queryArgs" :key="index">
                <div class="grid grid-cols-12 gap-2 items-center">
                  <input v-model="arg.name" placeholder="Name (e.g., user_id)" type="text"
                    class="col-span-3 p-2 border rounded-lg text-xs font-mono focus:ring-black focus:border-black" />

                  <select v-model="arg.type"
                    class="col-span-3 p-2 border rounded-lg text-xs focus:ring-black focus:border-black">
                    <option v-for="t in queryArgTypes" :key="t" :value="t">{{ t }}</option>
                  </select>

                  <input v-model="arg.description" placeholder="e.g., val1, val2" type="text"
                    class="col-span-3 p-2 border rounded-lg text-xs font-mono focus:ring-black focus:border-black" />

                  <div class="col-span-2 flex items-center justify-center space-x-2">
                    <label :for="`req-toggle-${index}`" class="text-xs font-medium text-gray-700">Required</label>
                    <button :id="`req-toggle-${index}`" type="button" @click="arg.required = !arg.required" :class="[
                      arg.required ? 'bg-black' : 'bg-gray-200',
                      'relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
                    ]" role="switch" :aria-checked="arg.required">
                      <span aria-hidden="true" :class="[
                        arg.required ? 'translate-x-4' : 'translate-x-0',
                        'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                      ]"></span>
                    </button>
                  </div>

                  <button type="button" @click="removeQueryArg(editingTool, index)"
                    :disabled="editingTool.queryArgs.length === 1 && arg.name === ''"
                    class="col-span-1 text-red-600 hover:text-red-800 disabled:opacity-50 transition duration-150">
                    <svg class="w-5 h-5 mx-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
              </template>
            </div>
            <button type="button" @click="addQueryArg(editingTool)"
              class="mt-3 text-xs text-black border border-black px-3 py-1 rounded-full hover:bg-gray-100 transition duration-150">
              + Add Argument
            </button>
          </div>

          <div v-if="isBodyRequired">
            <label class="block text-sm font-medium text-gray-600">Sample Body (JSON/Text) or Schema (JSON)</label>
            <textarea v-model="editingTool.reqBodyJSONSchema" rows="6"
              placeholder="Enter sample request body here (e.g., JSON payload) or a full JSON Schema."
              class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm font-mono focus:ring-black focus:border-black"></textarea>
          </div>

          <div v-if="provider.oauth2Config" class="border border-gray-200 p-4 rounded-lg bg-gray-50">
            <h4 class="text-sm font-semibold mb-2 text-gray-700">OAuth2 Scopes</h4>
            <div class="space-y-2">
              <template v-for="(scope, index) in editingTool.oauth2Scopes" :key="index">
                <div class="flex items-center space-x-2">
                  <input v-model="editingTool.oauth2Scopes[index]" placeholder="e.g., read:user" type="text"
                    class="grow p-2 border rounded-lg text-xs font-mono focus:ring-black focus:border-black" />
                  <button type="button" @click="removeScope(editingTool, index)"
                    class="text-red-600 hover:text-red-800 transition duration-150 p-1">
                    <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
              </template>
            </div>
            <button type="button" @click="addScope(editingTool)"
              class="mt-3 text-xs text-black border border-black px-3 py-1 rounded-full hover:bg-gray-100 transition duration-150">
              + Add Scope
            </button>
          </div>

          <div class="flex justify-end space-x-3 pt-4 border-t">
            <button type="button" @click="cancelToolEdit"
              class="px-4 py-2 text-gray-700 border border-gray-300 rounded-lg hover:bg-gray-100 transition duration-150">
              Cancel
            </button>
            <button type="submit"
              class="px-4 py-2 bg-black text-white font-semibold rounded-lg shadow-md hover:bg-gray-800 transition duration-150">
              {{ editingTool.id ? 'Save Tool' : 'Create Tool' }}
            </button>
          </div>
        </form>
      </div>
      <div class="mb-6">
        <input v-model="toolSearchQuery" type="text" :placeholder="`Search ${currentToolList.length} tools...`"
          :disabled="isImporting"
          class="block w-full border border-gray-300 rounded-full shadow-inner p-3 text-sm focus:ring-black focus:border-black disabled:bg-gray-100" />
      </div>

      <div v-if="displayedTools.length > 0" class="space-y-3">
        <div v-for="tool in displayedTools" :key="tool.id" :class="[
          'bg-white rounded-xl shadow-lg transition duration-300 border-l-4',
          getMethodBorderColor(tool.method),
        ]">
          <div :class="[
            'p-4 flex justify-between items-center cursor-pointer',
            expandedState[tool.id] ||
              (isToolFormVisible && editingTool && editingTool.id === tool.id)
              ? 'border-b border-gray-200'
              : '',
          ]" @click="toggleToolExpansion(tool.id)">
            <div class="flex items-center space-x-4 grow min-w-0">
              <span :class="[
                'font-mono text-xs font-bold px-3 py-1 rounded-full text-white shrink-0',
                getMethodColor(tool.method).split(' ')[0],
              ]">
                {{ tool.method }}
              </span>
              <span class="font-mono text-sm text-gray-800 font-semibold shrink-0">{{
                tool.path
                }}</span>
              <span class="text-sm text-gray-500 truncate hidden sm:block">- {{ tool.description }}</span>
            </div>
            <div class="flex space-x-2 shrink-0">
              <button @click.stop="toggleToolForm(tool)" :disabled="isImporting"
                class="p-1 text-gray-600 hover:text-black transition duration-150 disabled:opacity-50"
                title="Edit Tool">
                <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
              </button>
              <button @click.stop="initiateDelete(tool, 'tool')" :disabled="isImporting"
                class="p-1 text-red-500 hover:text-red-700 transition duration-150 disabled:opacity-50"
                title="Delete Tool">
                <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
              <button @click.stop="toggleToolExpansion(tool.id)"
                class="p-1 text-gray-600 transition duration-150 transform"
                :class="{ 'rotate-180': expandedState[tool.id] }">
                <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                </svg>
              </button>
            </div>
          </div>

          <div v-if="isToolFormVisible && editingTool && editingTool.id === tool.id"
            class="bg-white p-6 rounded-b-xl border-t-2 border-dashed">
            <h3 class="text-xl font-semibold mb-4 text-gray-700 border-b pb-2">
              {{
                editingTool.id ? `Edit Tool ID: ${editingTool.id}` : 'Add New Tool'
              }}
            </h3>
            <form @submit.prevent="saveTool" class="space-y-4">
              <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
                <div class="col-span-1">
                  <label class="block text-sm font-medium text-gray-600">Method</label>
                  <select v-model="editingTool.method" required :disabled="!!editingTool.id"
                    class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm focus:ring-black focus:border-black"
                    :class="{ 'bg-gray-100 cursor-not-allowed': !!editingTool.id }">
                    <option v-for="method in availableMethods" :key="method" :value="method">
                      {{ method }}
                    </option>
                  </select>
                </div>
                <div class="col-span-3">
                  <label class="block text-sm font-medium text-gray-600">Path (Starts with '/', example:
                    '/users/{id}/profile'
                    where 'id' value is
                    dynamic)</label>
                  <input v-model="editingTool.path" type="text" required placeholder="/users/{id}/profile"
                    :disabled="!!editingTool.id"
                    class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm font-mono focus:ring-black focus:border-black"
                    :class="{ 'bg-gray-100 cursor-not-allowed': !!editingTool.id }" />
                </div>
              </div>

              <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
                <div class="col-span-1">
                  <label class="block text-sm font-medium text-gray-600">Name <span
                      class="text-gray-400 font-normal">(MCP
                      tool
                      name)</span></label>
                  <input v-model="editingTool.name" type="text" placeholder="e.g. getUser" maxlength="20"
                    class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm font-mono focus:ring-black focus:border-black"
                    :class="{ 'border-red-500 ring-1 ring-red-500': nameValidationError }" />
                  <p v-if="nameValidationError" class="mt-1 text-xs text-red-600 font-bold italic">{{
                    nameValidationError }}
                  </p>
                </div>
                <div class="col-span-3">
                  <label class="block text-sm font-medium text-gray-600">Title <span
                      class="text-gray-400 font-normal">(MCP
                      tool
                      title)</span></label>
                  <input v-model="editingTool.title" type="text" placeholder="e.g. Get User Profile" maxlength="64"
                    class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm font-mono focus:ring-black focus:border-black" />
                </div>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-600">Description</label>
                <textarea v-model="editingTool.description" required rows="1"
                  placeholder="A short summary of what this tool does."
                  class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm focus:ring-black focus:border-black"></textarea>
              </div>

              <div class="border border-gray-200 p-4 rounded-lg bg-gray-50">
                <h4 class="text-sm font-semibold mb-2 text-gray-700">
                  HTTP Headers (Can reference ENV vars like ${VAR_NAME})
                </h4>
                <p class="text-xs text-red-600 font-medium mb-3">
                  <span class="font-bold">Important:</span> Environment variables in headers *must*
                  start with the prefix:
                  <code class="font-mono bg-red-100 px-1 py-0.5 rounded">{{
                    provider.secretPrefix
                  }}</code>
                </p>
                <template v-for="(header, index) in editingTool.headers" :key="index">
                  <div class="grid grid-cols-12 gap-2 items-center mb-2">
                    <input v-model="header.key" placeholder="Key (e.g., Authorization)" type="text"
                      class="col-span-5 p-2 border rounded-lg text-xs font-mono focus:ring-black focus:border-black"
                      :class="{ 'border-red-500': headerValidationErrors[index] }" />
                    <input v-model="header.value" placeholder="Value (e.g., Bearer ${AUTH_TOKEN})" type="text"
                      class="col-span-6 p-2 border rounded-lg text-xs font-mono focus:ring-black focus:border-black"
                      :class="{ 'border-red-500 ring-red-500': headerValidationErrors[index] }"
                      @input="headerValidationErrors[index] = ''" />
                    <button type="button" @click="removeHeader(editingTool, index)" :disabled="editingTool.headers !== undefined &&
                      editingTool.headers.length === 1 &&
                      header.key === '' &&
                      header.value === ''
                      " class="col-span-1 text-red-600 hover:text-red-800 disabled:opacity-50 transition duration-150">
                      <svg class="w-5 h-5 mx-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                          d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>

                    <p v-if="headerValidationErrors[index]" class="col-span-12 text-xs text-red-600 italic -mt-1 pl-2">
                      {{ headerValidationErrors[index] }}
                    </p>
                  </div>
                </template>
                <button type="button" @click="addHeader(editingTool)"
                  class="mt-2 text-xs text-black border border-black px-3 py-1 rounded-full hover:bg-gray-100 transition duration-150">
                  + Add Header
                </button>
              </div>

              <div class="border border-gray-200 p-4 rounded-lg bg-gray-50">
                <h4 class="text-sm font-semibold mb-2 text-gray-700">Query String Arguments</h4>
                <div class="space-y-3">
                  <template v-for="(arg, index) in editingTool.queryArgs" :key="index">
                    <div class="grid grid-cols-12 gap-2 items-center">
                      <input v-model="arg.name" placeholder="Name (e.g., user_id)" type="text"
                        class="col-span-3 p-2 border rounded-lg text-xs font-mono focus:ring-black focus:border-black" />

                      <select v-model="arg.type"
                        class="col-span-3 p-2 border rounded-lg text-xs focus:ring-black focus:border-black">
                        <option v-for="t in queryArgTypes" :key="t" :value="t">{{ t }}</option>
                      </select>

                      <input v-model="arg.description" placeholder="e.g., val1, val2" type="text"
                        class="col-span-3 p-2 border rounded-lg text-xs font-mono focus:ring-black focus:border-black" />

                      <div class="col-span-2 flex items-center justify-center space-x-2">
                        <label :for="`req-toggle-${index}`" class="text-xs font-medium text-gray-700">Required</label>
                        <button :id="`req-toggle-${index}`" type="button" @click="arg.required = !arg.required" :class="[
                          arg.required ? 'bg-black' : 'bg-gray-200',
                          'relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
                        ]" role="switch" :aria-checked="arg.required">
                          <span aria-hidden="true" :class="[
                            arg.required ? 'translate-x-4' : 'translate-x-0',
                            'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                          ]"></span>
                        </button>
                      </div>

                      <button type="button" @click="removeQueryArg(editingTool, index)"
                        :disabled="editingTool.queryArgs.length === 1 && arg.name === ''"
                        class="col-span-1 text-red-600 hover:text-red-800 disabled:opacity-50 transition duration-150">
                        <svg class="w-5 h-5 mx-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      </button>
                    </div>
                  </template>
                </div>
                <button type="button" @click="addQueryArg(editingTool)"
                  class="mt-3 text-xs text-black border border-black px-3 py-1 rounded-full hover:bg-gray-100 transition duration-150">
                  + Add Argument
                </button>
              </div>
              <div v-if="isBodyRequired">
                <label class="block text-sm font-medium text-gray-600">Sample Body (JSON/Text) or Schema (JSON)</label>
                <textarea v-model="editingTool.reqBodyJSONSchema" rows="6"
                  placeholder="Enter sample request body here (e.g., JSON payload) or a full JSON Schema."
                  class="mt-1 block w-full border border-gray-300 rounded-lg shadow-sm p-3 text-sm font-mono focus:ring-black focus:border-black"></textarea>
              </div>

              <div v-if="provider.oauth2Config" class="border border-gray-200 p-4 rounded-lg bg-gray-50">
                <h4 class="text-sm font-semibold mb-2 text-gray-700">OAuth2 Scopes</h4>
                <div class="space-y-2">
                  <template v-for="(scope, index) in editingTool.oauth2Scopes" :key="index">
                    <div class="flex items-center space-x-2">
                      <input v-model="editingTool.oauth2Scopes[index]" placeholder="e.g., read:user" type="text"
                        class="grow p-2 border rounded-lg text-xs font-mono focus:ring-black focus:border-black" />
                      <button type="button" @click="removeScope(editingTool, index)"
                        class="text-red-600 hover:text-red-800 transition duration-150 p-1">
                        <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      </button>
                    </div>
                  </template>
                </div>
                <button type="button" @click="addScope(editingTool)"
                  class="mt-3 text-xs text-black border border-black px-3 py-1 rounded-full hover:bg-gray-100 transition duration-150">
                  + Add Scope
                </button>
              </div>

              <div class="flex justify-end space-x-3 pt-4 border-t">
                <button type="button" @click="cancelToolEdit"
                  class="px-4 py-2 text-gray-700 border border-gray-300 rounded-lg hover:bg-gray-100 transition duration-150">
                  Cancel
                </button>
                <button type="submit"
                  class="px-4 py-2 bg-black text-white font-semibold rounded-lg shadow-md hover:bg-gray-800 transition duration-150">
                  {{ editingTool.id ? 'Save Tool' : 'Create Tool' }}
                </button>
              </div>
            </form>
          </div>
          <div v-else-if="expandedState[tool.id]" class="p-4 pt-0 bg-gray-50/50 rounded-b-xl border-t-2 border-dashed">
            <h4 class="text-sm font-semibold mb-2 text-gray-700">
              Headers ({{ tool?.headers?.length }})
            </h4>
            <div class="space-y-1 mb-4">
              <div v-for="(header, index) in tool.headers" :key="index"
                class="flex text-xs font-mono bg-white p-2 rounded-lg shadow-sm border border-gray-100">
                <span class="font-bold text-gray-800 w-1/3 truncate">{{ header.key }}:</span>
                <span class="text-gray-600 w-2/3 truncate pl-2">{{ header.value }}</span>
              </div>
              <p v-if="tool.headers === undefined || tool.headers.length === 0" class="text-xs text-gray-500 italic">
                No headers configured.
              </p>
            </div>

            <h4 class="text-sm font-semibold mb-2 text-gray-700">Query Arguments</h4>
            <div class="space-y-1 mb-4">
              <template v-if="
                tool.queryArgsJSONSchema &&
                tool.queryArgsJSONSchema.properties &&
                Object.keys(tool.queryArgsJSONSchema.properties).length > 0
              ">
                <div v-for="arg in resolveQueryArgsFromSchema(tool.queryArgsJSONSchema)" :key="arg.name"
                  class="grid grid-cols-4 gap-2 text-xs font-mono bg-white p-2 rounded-lg shadow-sm border border-gray-100 items-center">
                  <span class="font-bold text-gray-800 truncate col-span-1">{{ arg.name }}</span>
                  <span class="text-gray-600 truncate col-span-1">{{ arg.type }}</span>
                  <span :class="[
                    'truncate',
                    'col-span-1',
                    arg.required ? 'font-semibold text-red-600' : 'text-gray-500',
                  ]">
                    {{ arg.required ? 'Required' : 'Optional' }}
                  </span>
                  <span class="text-gray-500 truncate col-span-1" v-if="arg.description">D: {{ arg.description }}</span>
                </div>
              </template>
              <p v-else class="text-xs text-gray-500 italic">No query arguments configured.</p>
            </div>

            <div v-if="tool.reqBodyJSONSchema">
              <h4 class="text-sm font-semibold mb-2 text-gray-700">Request Body Schema</h4>
              <pre class="bg-gray-800 text-white p-3 rounded-lg text-xs overflow-x-auto shadow-inner">{{
                tool.reqBodyJSONSchema }}</pre>
            </div>
            <p v-else-if="isBodyRequired" class="text-sm text-gray-500 italic">
              No payload schema provided.
            </p>

            <template v-if="provider.oauth2Config">
              <h4 class="text-sm font-semibold mb-2 text-gray-700">OAuth2 Scopes</h4>
              <div class="flex flex-wrap gap-2 mb-4">
                <template v-if="tool.oauth2Scopes && tool.oauth2Scopes.length > 0">
                  <span v-for="scope in tool.oauth2Scopes" :key="scope"
                    class="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-md font-mono border border-blue-200">
                    {{ scope }}
                  </span>
                </template>
                <p v-else class="text-xs text-gray-500 italic">No scopes configured.</p>
              </div>
            </template>
          </div>
        </div>
      </div>

      <div v-if="filteredTools.length === 0 && !isImporting"
        class="text-center py-10 text-gray-500 bg-white rounded-xl shadow-md">
        No tools configured or matching the search criteria.
      </div>

      <div v-if="hasMoreTools" class="mt-8 text-center">
        <button @click="loadMoreTools" :disabled="isLoadingTools || isImporting"
          class="px-6 py-3 text-sm font-medium rounded-lg text-black border border-black hover:bg-gray-100 disabled:opacity-50 transition duration-150 shadow-md">
          {{
            isLoadingTools
              ? 'Loading...'
              : `Load 10 More Tools (Total: ${filteredTools.length})`
          }}
        </button>
      </div>
    </div>

    <DeleteModal :show="showDeleteModal" :variableName="deleteNameForModal" :itemType="itemTypeForModal"
      @close="cancelDelete" @confirm="confirmDelete" />

    <OpenApiImportModal :show="isImportModalVisible" @close="isImportModalVisible = false"
      @start-import="handleImport" />
  </div>
</template>