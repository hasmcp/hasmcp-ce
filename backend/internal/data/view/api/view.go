package view

import (
	"encoding/json"
)

type (
	// Variable hosts env and secret variable
	Variable struct {
		ID        string `json:"id,omitempty"`
		CreatedAt string `json:"createdAt,omitempty"`
		UpdatedAt string `json:"updatedAt,omitempty"`

		Type  string `json:"type,omitempty"` // 0: INVALID, 1: ENV, 2: SECRET
		Value string `json:"value,omitempty"`
		Name  string `json:"name,omitempty"`
	}

	CreateVariableRequest struct {
		Variable Variable `json:"variable"`
	}

	CreateVariableResponse struct {
		Variable Variable `json:"variable"`
	}

	UpdateVariableRequest struct {
		Variable Variable `json:"variable"`
	}

	UpdateVariableResponse struct {
		Variable Variable `json:"variable"`
	}

	ListVariablesResponse struct {
		Variables []Variable `json:"variables"`
	}

	DeleteVariableRequest struct {
		ID string `json:"id"`
	}

	// Provider hosts API providers
	Provider struct {
		ID        string `json:"id,omitempty"`
		CreatedAt string `json:"createdAt,omitempty"`
		UpdatedAt string `json:"updatedAt,omitempty"`

		Version        int32  `json:"version,omitempty"`
		ApiType        string `json:"apiType,omitempty"`        // 0: INVALID, 1: REST
		VisibilityType string `json:"visibilityType,omitempty"` // 0: INVALID, 1: INTERNAL, 2: PUBLIC
		BaseURL        string `json:"baseURL,omitempty"`
		DocumentURL    string `json:"documentURL,omitempty"`
		IconURL        string `json:"iconURL,omitempty"`
		SecretPrefix   string `json:"secretPrefix,omitempty"`
		Name           string `json:"name,omitempty"`
		Description    string `json:"description,omitempty"`

		Tools        []ProviderTool        `json:"tools,omitempty"`
		Oauth2Config *ProviderOauth2Config `json:"oauth2Config,omitempty"`
	}

	CreateProviderRequest struct {
		Provider Provider `json:"provider,omitempty"`
	}

	CreateProviderResponse struct {
		Provider Provider `json:"provider,omitempty"`
	}

	UpdateProviderRequest struct {
		Provider Provider `json:"provider,omitempty"`
	}

	ListProviderRequest struct {
		Filters    ListProviderFilters
		Pagination Pagination
	}

	ListProvidersResponse struct {
		Providers  []Provider `json:"providers,omitempty"`
		MatchCount int        `json:"matchCount,omitempty"`
		NextToken  string     `json:"nextToken,omitempty"`
	}

	ListProviderFilters struct {
		NameContains    string
		BaseURLContains string
		ApiType         string
		Visibility      string
	}

	GetProviderResponse struct {
		Provider Provider `json:"provider,omitempty"`
	}

	UpdateProviderResponse struct {
		Provider Provider `json:"provider,omitempty"`
	}

	Pagination struct {
		Limit int    `json:"limit,omitempty"`
		Token string `json:"token,omitempty"`
	}

	Sorting struct {
		By        string
		Ascending bool
	}

	ListProviderResponse struct {
		Providers  []Provider `json:"providers,omitempty"`
		MatchCount int        `json:"matchCount,omitempty"`
		NextToken  string     `json:"nextToken,omitempty"`
	}

	// ProviderOauth2Config hosts oauth2 configuration for the provider (1:1)
	ProviderOauth2Config struct {
		ClientID     string `json:"clientID"`
		ClientSecret string `json:"clientSecret"`
		AuthURL      string `json:"authURL"`
		TokenURL     string `json:"tokenURL"`
	}

	// ProviderTool hosts the tools for the provider
	ProviderTool struct {
		ID string `json:"id"`

		ProviderID          string          `json:"providerID,omitempty"`
		Method              string          `json:"method,omitempty"`
		Path                string          `json:"path,omitempty"`
		Name                string          `json:"name,omitempty"`
		Title               string          `json:"title,omitempty"`
		Description         string          `json:"description,omitempty"`
		PathArgsJSONSchema  json.RawMessage `json:"pathArgsJSONSchema,omitempty"`
		QueryArgsJSONSchema json.RawMessage `json:"queryArgsJSONSchema,omitempty"`
		ReqBodyJSONSchema   json.RawMessage `json:"reqBodyJSONSchema,omitempty"`
		ResBodyJSONSchema   json.RawMessage `json:"resBodyJSONSchema,omitempty"`
		Headers             []ToolHeader    `json:"headers,omitempty"`
		Oauth2Scopes        []string        `json:"oauth2Scopes,omitempty"`
	}

	CreateProviderToolRequest struct {
		Tool ProviderTool `json:"tool,omitempty"`
	}

	CreateProviderToolResponse struct {
		Tool ProviderTool `json:"tool,omitempty"`
	}

	UpdateProviderToolRequest struct {
		Tool ProviderTool `json:"tool,omitempty"`
	}

	UpdateProviderToolResponse struct {
		Tool ProviderTool `json:"tool,omitempty"`
	}

	ToolHeader struct {
		Key   string `json:"key,omitempty"`
		Value string `json:"value,omitempty"`
	}

	ListProviderToolsResponse struct {
		Tools []ProviderTool `json:"tools,omitempty"`
	}

	GetProviderToolResponse struct {
		Tool ProviderTool `json:"tool,omitempty"`
	}

	// Server hosts a server of a set of provider tools
	Server struct {
		ID        string `json:"id,omitempty"`
		CreatedAt string `json:"createdAt,omitempty"`
		UpdatedAt string `json:"updatedAt,omitempty"`

		RequestHeadersProxyEnabled bool `json:"requestHeadersProxyEnabled"`

		Name         string     `json:"name,omitempty"`
		Instructions string     `json:"instructions,omitempty"`
		Version      int32      `json:"version,omitempty"`
		Providers    []Provider `json:"providers,omitempty"`
		Resources    []Resource `json:"resources,omitempty"`
		Prompts      []Prompt   `json:"prompts,omitempty"`

		VisibilityType string `json:"visibilityType,omitempty"` // 0: INVALID, 1: INTERNAL, 2: PUBLIC
	}

	ProviderToolPair struct {
		ProviderID string `json:"providerID,omitempty"`
		ToolID     string `json:"toolID,omitempty"`
	}

	CreateServerRequest struct {
		Server Server `json:"server,omitempty"`
	}

	CreateServerResponse struct {
		Server Server `json:"server,omitempty"`
	}

	ListServersResponse struct {
		Servers []Server `json:"servers,omitempty"`
	}

	GetServerResponse struct {
		Server Server `json:"server,omitempty"`
	}

	UpdateServerRequest struct {
		Server Server `json:"server,omitempty"`
	}

	UpdateServerResponse struct {
		Server Server `json:"server,omitempty"`
	}

	// ServerToken host hashed value for the token that interacts with the server
	ServerToken struct {
		ID        string `json:"id,omitempty"`
		ServerID  string `json:"serverID,omitempty"`
		CreatedAt string `json:"createdAt,omitempty"`
		ExpiresAt string `json:"expiresAt,omitempty"`
		Scope     string `json:"scope,omitempty"`

		Name  string `json:"name,omitempty"`
		Value string `json:"value,omitempty"`
	}

	CreateServerTokenRequest struct {
		Token ServerToken `json:"token,omitempty"`
	}

	CreateServerTokenResponse struct {
		Token ServerToken `json:"token,omitempty"`
	}

	ListServerTokensResponse struct {
		Tokens []ServerToken `json:"tokens,omitempty"`
	}

	CreateServerToolRequest struct {
		Tool ServerTool `json:"tool,omitempty"`
	}

	CreateServerToolResponse struct {
		Tool ServerTool `json:"tool,omitempty"`
	}

	ServerTool struct {
		ServerID   string `json:"serverID,omitempty"`
		ProviderID string `json:"providerID,omitempty"`
		ToolID     string `json:"toolID,omitempty"`
	}

	DeleteServerToolsRequest struct {
		ProviderID string `json:"providerID,omitempty"`
		ToolID     string `json:"toolID,omitempty"`
	}

	ListServerToolsResponse struct {
		Tools []ServerTool `json:"tools,omitempty"`
	}

	// Resource hosts a known resource that the server is capable of reading.
	Resource struct {
		ID        string `json:"id,omitempty"`
		CreatedAt string `json:"createdAt,omitempty"`
		UpdatedAt string `json:"updatedAt,omitempty"`

		Name        string          `json:"name,omitempty"`
		Description string          `json:"description,omitempty"`
		URI         string          `json:"uri,omitempty"`
		MimeType    string          `json:"mimeType,omitempty"`
		Size        int64           `json:"size,omitempty"`
		Annotations json.RawMessage `json:"annotations,omitempty"`

		VisibilityType string `json:"visibilityType,omitempty"` // 0: INVALID, 1: INTERNAL, 2: PUBLIC
	}

	CreateResourceRequest struct {
		Resource Resource `json:"resource,omitempty"`
	}

	CreateResourceResponse struct {
		Resource Resource `json:"resource,omitempty"`
	}

	GetResourceResponse struct {
		Resource Resource `json:"resource,omitempty"`
	}

	ListResourcesResponse struct {
		Resources []Resource `json:"resources,omitempty"`
	}

	UpdateResourceRequest struct {
		Resource Resource `json:"resource,omitempty"`
	}

	UpdateResourceResponse struct {
		Resource Resource `json:"resource,omitempty"`
	}

	// Prompt hosts a prompt or prompt template that the server offers.
	Prompt struct {
		ID        string `json:"id,omitempty"`
		CreatedAt string `json:"createdAt,omitempty"`
		UpdatedAt string `json:"updatedAt,omitempty"`

		Name        string          `json:"name,omitempty"`
		Description string          `json:"description,omitempty"`
		Arguments   json.RawMessage `json:"arguments,omitempty"` // []PromptArgument
		Messages    json.RawMessage `json:"messages,omitempty"`  // []PromptMessage

		VisibilityType string `json:"visibilityType,omitempty"` // 0: INVALID, 1: INTERNAL, 2: PUBLIC
	}

	CreatePromptRequest struct {
		Prompt Prompt `json:"prompt,omitempty"`
	}

	CreatePromptResponse struct {
		Prompt Prompt `json:"prompt,omitempty"`
	}

	GetPromptResponse struct {
		Prompt Prompt `json:"prompt,omitempty"`
	}

	ListPromptsResponse struct {
		Prompts []Prompt `json:"prompts,omitempty"`
	}

	UpdatePromptRequest struct {
		Prompt Prompt `json:"prompt,omitempty"`
	}

	ServerResource struct {
		ServerID   string `json:"serverID,omitempty"`
		ResourceID string `json:"resourceID,omitempty"`
	}

	// ServerResource Junction
	CreateServerResourceRequest struct {
		Resource ServerResource `json:"resource,omitempty"`
	}

	CreateServerResourceResponse struct {
		Resource ServerResource `json:"resource,omitempty"`
	}

	ListServerResourcesResponse struct {
		Resources []ServerResource `json:"resources,omitempty"`
	}

	ServerPrompt struct {
		ServerID string `json:"serverID,omitempty"`
		PromptID string `json:"promptID,omitempty"`
	}

	// ServerPrompt Junction
	CreateServerPromptRequest struct {
		Prompt ServerPrompt `json:"prompt,omitempty"`
	}

	CreateServerPromptResponse struct {
		Prompt ServerPrompt `json:"prompt,omitempty"`
	}

	ListServerPromptsResponse struct {
		Prompts []ServerPrompt `json:"prompt,omitempty"`
	}
)
