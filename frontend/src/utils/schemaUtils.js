/**
 * Checks if an object appears to be a JSON Schema based on common top-level keywords.
 * @param {object} obj
 * @returns {boolean}
 */
const isJsonSchema = (obj) => {
  if (typeof obj !== 'object' || obj === null || Array.isArray(obj)) return false

  // Check for common top-level schema keywords
  if ('$schema' in obj) return true
  if ('$ref' in obj) return true
  if ('type' in obj) {
    const type = obj.type
    if (type === 'object' && 'properties' in obj) return true
    if (type === 'array' && 'items' in obj) return true
    if (['string', 'number', 'integer', 'boolean', 'null'].includes(type)) return true
  }
  return false
}

/**
 * Infers a basic JSON Schema from a sample JSON object.
 * @param {any} sampleData
 * @returns {object} The inferred JSON Schema object.
 */
const inferBasicSchema = (sampleData) => {
  const result = {}
  if (Array.isArray(sampleData)) {
    result.type = 'array'
    if (sampleData.length > 0) {
      // Use the first element to define the items schema
      result.items = inferBasicSchema(sampleData[0])
    } else {
      result.items = {}
    }
  } else if (typeof sampleData === 'object' && sampleData !== null) {
    result.type = 'object'
    result.properties = {}
    result.required = Object.keys(sampleData)

    for (const key in sampleData) {
      // Recursively infer schema for the property value
      const propertySchema = inferBasicSchema(sampleData[key])

      // ADDITION: Inject empty description field for user convenience
      propertySchema.description = ''

      result.properties[key] = propertySchema
    }
  } else if (typeof sampleData === 'string') {
    result.type = 'string'
  } else if (typeof sampleData === 'number') {
    result.type = Number.isInteger(sampleData) ? 'integer' : 'number'
  } else if (typeof sampleData === 'boolean') {
    result.type = 'boolean'
  } else if (sampleData === null) {
    result.type = 'null'
  }
  return result
}

/**
 * Helper to strip parameters from the path and convert the remaining segments to PascalCase.
 * Applies a simple singularization heuristic to the main resource part if a path parameter is present.
 */
const toPascalCaseResourceName = (path) => {
  // 1. Check for parameters (e.g., :id or {id}) to trigger singularization heuristic
  const hasParam = path.includes(':') || path.includes('{')
  // Strip parameter placeholders
  let cleanPath = path.replace(/:\w+/g, '').replace(/{(\w+)}/g, '')

  // 2. Split by '/' and filter empty strings
  let segments = cleanPath.split('/').filter((s) => s.length > 0)

  if (segments.length === 0) return ''

  // 3. Apply naive singularization heuristic to the first segment if a parameter was present
  // E.g., 'posts' becomes 'post' if a parameter exists in the path.
  if (hasParam && segments[0].endsWith('s')) {
    segments[0] = segments[0].slice(0, -1)
  }

  // 4. PascalCase all segments and join (e.g., ['post', 'comment'] -> 'PostComment')
  return segments.map((s) => s.charAt(0).toUpperCase() + s.slice(1)).join('')
}

/**
 * Generates the full schema title based on method and path to match the user's examples/formula.
 */
const generateSchemaTitle = (method, path) => {
  if (!method || !path) return 'Generated Request Body Schema'

  const lowerMethod = method.toLowerCase()
  const pascalName = toPascalCaseResourceName(path)
  const segments = path
    .replace(/:\w+/g, '')
    .replace(/{(\w+)}/g, '')
    .split('/')
    .filter((s) => s.length > 0)

  // Custom action prefix mapping based on method
  let actionPrefix
  switch (lowerMethod) {
    case 'post':
      actionPrefix = 'create'
      break
    case 'put':
      actionPrefix = 'update'
      break
    case 'patch':
      actionPrefix = 'patch'
      break
    case 'delete':
      actionPrefix = 'delete'
      break
    default:
      actionPrefix = 'get' // Includes GET, HEAD, OPTIONS
  }

  // Case 1: Simple Plural Resource Tool (e.g., /posts) - no parameter in path
  if (
    segments.length === 1 &&
    segments[0].endsWith('s') &&
    !path.includes(':') &&
    !path.includes('{')
  ) {
    const resource = segments[0]

    if (lowerMethod === 'get') {
      // get /posts -> getPosts
      return 'get' + pascalName
    } else if (lowerMethod === 'post') {
      // post /posts -> createPost (singularize for creation)
      const singularResource = resource.slice(0, -1)
      return 'create' + singularResource.charAt(0).toUpperCase() + singularResource.slice(1)
    }
  }

  // Case 2: Everything else (e.g., /posts/:id, /posts/:id/comment)
  // This relies on `toPascalCaseResourceName` to have handled singularization.
  return actionPrefix + pascalName
}

/**
 * Checks if JSON content is a schema or a sample payload, and converts the latter to a schema.
 * @param {string | null | undefined} jsonString The raw JSON string from the user input.
 * @param {string | null} [method] The HTTP method of the tool (e.g., 'POST').
 * @param {string | null} [path] The path of the tool (e.g., '/users/:id').
 * @returns {string | null} The resulting JSON Schema string or null if input is invalid/empty.
 */
export const inferSchemaOrReturnOriginal = (jsonString, method = null, path = null) => {
  if (!jsonString) return null
  const trimmed = jsonString.trim()
  if (!trimmed) return null

  try {
    const parsed = JSON.parse(trimmed)

    // 1. If it looks like a schema, return the original string (prettified)
    if (isJsonSchema(parsed)) {
      return JSON.stringify(parsed, null, 2)
    }

    // 2. If it's sample data, infer the schema
    const inferredSchema = inferBasicSchema(parsed)

    // Generate the schema title using the new logic
    const schemaTitle = generateSchemaTitle(method, path)

    // Add required metadata for a root schema
    const finalSchema = {
      $schema: 'http://json-schema.org/draft-07/schema#',
      title: schemaTitle, // Use generated title
      description: '', // ADDITION: Add description to the root schema
      ...inferredSchema,
    }

    return JSON.stringify(finalSchema, null, 2)
  } catch {
    // If JSON parsing fails, return the original string.
    return jsonString
  }
}

/**
 * Creates a human-readable description from a camelCase or snake_case name.
 * e.g., 'organizationId' -> 'organization id'
 * e.g., 'domain_id' -> 'domain id'
 * @param {string} name
 * @returns {string}
 */
const createDescriptionFromName = (name) => {
  if (!name) return ''
  return name
    .replace(/_/g, ' ') // Convert snake_case to spaces
    .replace(/([A-Z])/g, ' $1') // Add space before capital letters
    .toLowerCase() // Convert to lowercase
    .trim() // Remove leading/trailing spaces
}

/**
 * Generates a JSON Schema for path parameters found in a URL path.
 * e.g., '/users/{userId}/posts/{postId}'
 * @param {string} path The tool path.
 * @returns {object | null} The JSON Schema object or null if no params are found.
 */
export const generatePathArgsSchema = (path) => {
  const pathParamRegex = /{([A-Za-z0-9_]+)}/g
  const properties = {}
  const required = []

  let match
  while ((match = pathParamRegex.exec(path)) !== null) {
    const paramName = match[1]
    if (paramName && !required.includes(paramName)) {
      required.push(paramName)
      properties[paramName] = {
        type: 'string',
        description: createDescriptionFromName(paramName),
      }
    }
  }

  if (required.length === 0) {
    return null
  }

  return {
    $schema: 'http://json-schema.org/draft-07/schema#',
    type: 'object',
    properties: properties,
    required: required,
  }
}
