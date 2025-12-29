package crud

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type (
	// Enummerations for types
	VariableType    uint8
	ApiType         uint8
	VisibilityType  uint8
	ObjectType      uint8
	ObjectEventType uint8
	MethodType      uint8

	ResourceChange struct {
		ObjectType      ObjectType
		EventType       ObjectEventType
		ResoureID       int64
		ResourceOwnerID int64
	}

	// Variable hosts env and secret variable
	Variable struct {
		ID        int64
		CreatedAt time.Time
		UpdatedAt time.Time

		Type  VariableType // 0: INVALID, 1: ENV, 2: SECRET
		Value []byte
		Nonce []byte
		Name  string
	}

	CreateVariableRequest struct {
		Variable Variable
	}

	CreateVariableResponse struct {
		Variable Variable
	}

	UpdateVariableRequest struct {
		Variable Variable
	}

	UpdateVariableResponse struct {
		Variable Variable
	}

	SaveVariableRequest struct {
		Variable Variable
	}

	DeleteVariableRequest struct {
		ID int64
	}

	ListVariablesResponse struct {
		Variables []Variable
	}

	// Provider hosts API providers
	Provider struct {
		ID        int64
		CreatedAt time.Time
		UpdatedAt time.Time

		Version        int32
		ApiType        ApiType        // 0: INVALID, 1: REST
		VisibilityType VisibilityType // 0: INVALID, 1: INTERNAL, 2: PUBLIC
		BaseURL        string
		DocumentURL    string
		IconURL        string
		SecretPrefix   string
		Name           string
		Description    string

		Tools        []ProviderTool
		Oauth2Config ProviderOauth2Config
	}

	CreateProviderRequest struct {
		Provider Provider
	}

	CreateProviderResponse struct {
		Provider Provider
	}

	DeleteProviderRequest struct {
		ID int64
	}

	UpdateProviderRequest struct {
		Provider Provider
	}

	UpdateProviderResponse struct {
		Provider Provider
	}

	ListProvidersRequest struct {
		Filters    ProviderFilters
		Pagination Pagination
	}

	GetProviderRequest struct {
		ID int64
	}

	GetProviderResponse struct {
		Provider Provider
	}

	ProviderFilters struct {
		NameContains    string
		BaseURLContains string
		ApiType         ApiType
		VisibilityType  VisibilityType
	}

	Pagination struct {
		Limit int
		Token int64
	}

	Sorting struct {
		By        string
		Ascending bool
	}

	ListProvidersResponse struct {
		Providers []Provider
	}

	ProviderOauth2Config struct {
		ClientID                    string
		ClientSecret                string
		ClientSecretEncrypted       []byte
		ClientSecretEncryptionNonce []byte
		AuthURL                     string
		TokenURL                    string
	}

	// ProviderTool hosts the tools for the provider
	ProviderTool struct {
		ID int64

		ProviderID          int64
		Method              MethodType // 0: INVALID, 1: GET, 2: HEAD, 3: POST, 4: PUT, 5: PATCH, 6: DELETE, 7: CONNECT, 8: OPTIONS, 9: TRACE
		Path                string
		Name                string
		Title               string
		Description         string
		PathArgsJSONSchema  []byte
		QueryArgsJSONSchema []byte
		ReqBodyJSONSchema   []byte
		ResBodyJSONSchema   []byte
		Headers             []ToolHeader
		Oauth2Scopes        []string
	}

	CreateProviderToolRequest struct {
		Tool ProviderTool
	}

	CreateProviderToolResponse struct {
		Tool ProviderTool
	}

	UpdateProviderToolRequest struct {
		Tool ProviderTool
	}

	UpdateProviderToolResponse struct {
		Tool ProviderTool
	}

	DeleteProviderToolRequest struct {
		ProviderID int64
		ToolID     int64
	}

	ToolHeader struct {
		Key   string
		Value string
	}

	ListProviderToolsRequest struct {
		ProviderID int64
		ToolIDs    []int64
	}

	ListProviderToolsResponse struct {
		Tools []ProviderTool
	}

	GetProviderToolRequest struct {
		ProviderID int64
		ToolID     int64
	}

	GetProviderToolResponse struct {
		Tool ProviderTool
	}

	// Server hosts a server of a set of provider tools
	Server struct {
		ID        int64
		CreatedAt time.Time
		UpdatedAt time.Time

		// RequestHeadersProxyEnabled allows passing the request headers from MCP client to the actual tool when it is set to
		// true. The default value is false.
		RequestHeadersProxyEnabled bool

		Name           string
		Instructions   string
		Version        int32
		Providers      []Provider
		Resources      []Resource
		Prompts        []Prompt
		VisibilityType VisibilityType
	}

	CreateServerRequest struct {
		Server Server
	}

	CreateServerResponse struct {
		Server Server
	}

	UpdateServerRequest struct {
		Server Server
	}

	UpdateServerResponse struct {
		Server Server
	}

	ListServersResponse struct {
		Servers []Server
	}

	ListServersRequest struct {
	}

	DeleteServerRequest struct {
		ID int64
	}

	GetServerRequest struct {
		ID int64
	}

	GetServerResponse struct {
		Server Server
	}

	// ServerToken host hashed value for the token that interacts with the server
	ServerToken struct {
		ID        int64
		ServerID  int64
		CreatedAt time.Time
		ExpiresAt time.Time
		Scope     string

		ActualValue []byte
	}

	CreateServerTokenRequest struct {
		Token ServerToken
	}

	CreateServerTokenResponse struct {
		ServerToken ServerToken
	}

	ListServerTokensRequest struct {
		ServerID int64
	}

	ListServerTokensResponse struct {
		Tokens []ServerToken
	}

	DeleteServerTokenRequest struct {
		ID int64
	}

	CreateServerToolRequest struct {
		Tool ServerTool
	}

	ServerTool struct {
		ServerID   int64
		ProviderID int64
		ToolID     int64
	}

	CreateServerToolResponse struct {
		Tool ServerTool
	}

	DeleteServerToolRequest struct {
		Tool ServerTool
	}

	ListServerToolsRequest struct {
		ServerID int64
	}

	ListServerToolsResponse struct {
		Tools []ServerTool
	}

	// Resource hosts a known resource that the server is capable of reading.
	Resource struct {
		ID        int64
		CreatedAt time.Time
		UpdatedAt time.Time

		Name        string
		Description string
		URI         string
		MimeType    string
		Size        int64
		Annotations json.RawMessage

		VisibilityType VisibilityType
	}

	ServerResource struct {
		ServerID   int64
		ResourceID int64
	}

	ServerPrompt struct {
		ServerID int64
		PromptID int64
	}

	CreateResourceRequest struct {
		Resource Resource
	}

	CreateResourceResponse struct {
		Resource Resource
	}

	GetResourceRequest struct {
		ID int64
	}

	GetResourceResponse struct {
		Resource Resource
	}

	ListResourcesRequest struct {
	}

	ListResourcesResponse struct {
		Resources []Resource
	}

	UpdateResourceRequest struct {
		Resource Resource
	}

	UpdateResourceResponse struct {
		Resource Resource
	}

	DeleteResourceRequest struct {
		ID int64
	}

	// Prompt hosts a prompt or prompt template that the server offers.
	Prompt struct {
		ID        int64
		CreatedAt time.Time
		UpdatedAt time.Time

		Name        string
		Description string
		Arguments   json.RawMessage // []PromptArgument
		Messages    json.RawMessage // []PromptMessage

		VisibilityType VisibilityType
	}

	CreatePromptRequest struct {
		Prompt Prompt
	}

	CreatePromptResponse struct {
		Prompt Prompt
	}

	GetPromptRequest struct {
		ID int64
	}

	GetPromptResponse struct {
		Prompt Prompt
	}

	ListPromptsRequest struct {
	}

	ListPromptsResponse struct {
		Prompts []Prompt
	}

	UpdatePromptRequest struct {
		Prompt Prompt
	}

	UpdatePromptResponse struct {
		Prompt Prompt
	}

	DeletePromptRequest struct {
		ID int64
	}

	// ServerResource Junction
	CreateServerResourceRequest struct {
		Resource ServerResource // includes only the resource id
	}

	CreateServerResourceResponse struct {
		Resoure ServerResource // includes only the resource id
	}

	DeleteServerResourceRequest struct {
		ServerID   int64
		ResourceID int64
	}

	ListServerResourcesRequest struct {
		ServerID int64
	}

	ListServerResourcesResponse struct {
		Resources []ServerResource // includes only the resource ids
	}

	// ServerPrompt Junction
	CreateServerPromptRequest struct {
		Prompt ServerPrompt // includes only the prompt id
	}

	CreateServerPromptResponse struct {
		Prompt ServerPrompt // includes only the prompt id
	}

	DeleteServerPromptRequest struct {
		ServerID int64
		PromptID int64
	}

	ListServerPromptsRequest struct {
		ServerID int64
	}

	ListServerPromptsResponse struct {
		Prompts []ServerPrompt // includes only the prompt ids
	}
)

const (
	VariableTypeInvalid VariableType = iota
	VariableTypeEnv
	VariableTypeSecret
)

func (vt VariableType) String() string {
	switch vt {
	case VariableTypeEnv:
		return "ENV"
	case VariableTypeSecret:
		return "SECRET"
	default:
		return ""
	}
}

func StringToVariableType(s string) VariableType {
	s = strings.ToUpper(s)
	switch s {
	case "ENV":
		return VariableTypeEnv
	case "SECRET":
		return VariableTypeSecret
	default:
		return VariableTypeInvalid
	}
}

const (
	ApiTypeInvalid ApiType = iota
	ApiTypeRest
	ApiTypeInvalidMax
)

func (at ApiType) String() string {
	switch at {
	case ApiTypeRest:
		return "REST"
	default:
		return ""
	}
}

func StringToApiType(s string) ApiType {
	s = strings.ToUpper(s)
	switch s {
	case "REST":
		return ApiTypeRest
	default:
		return ApiTypeInvalid
	}
}

const (
	VisibilityTypeInvalid VisibilityType = iota
	VisibilityTypeInternal
	VisibilityTypePublic
	VisibilityTypeInvalidMax
)

func (vt VisibilityType) String() string {
	switch vt {
	case VisibilityTypeInternal:
		return "INTERNAL"
	case VisibilityTypePublic:
		return "PUBLIC"
	default:
		return ""
	}
}

func StringToVisibilityType(s string) VisibilityType {
	s = strings.ToUpper(s)
	switch s {
	case "INTERNAL":
		return VisibilityTypeInternal
	case "PUBLIC":
		return VisibilityTypePublic
	default:
		return VisibilityTypeInvalid
	}
}

const (
	ObjectTypeInvalid ObjectType = iota
	ObjectTypeVariable
	ObjectTypeProvider
	ObjectTypeProviderTool
	ObjectTypeServer
	ObjectTypeServerToken
	ObjectTypeServerTool
	ObjectTypeServerPrompt
	ObjectTypeServerResource
	ObjectTypeResource
	ObjectTypePrompt
)

const (
	ObjectEventTypeInvalid ObjectEventType = iota
	ObjectEventTypeCreate
	ObjectEventTypeUpdate
	ObjectEventTypeDelete
)

func (rt ObjectType) String() string {
	switch rt {
	case ObjectTypeVariable:
		return "VARIABLE"
	case ObjectTypeProvider:
		return "PROVIDER"
	case ObjectTypeProviderTool:
		return "PROVIDER_ENDPOINT"
	case ObjectTypeServer:
		return "MCPSERVER"
	case ObjectTypeServerToken:
		return "MCPSERVER_TOKEN"
	case ObjectTypeResource:
		return "RESOURCE"
	case ObjectTypePrompt:
		return "PROMPT"
	default:
		return ""
	}
}

func StringToObjectType(s string) ObjectType {
	s = strings.ToUpper(s)
	switch s {
	case "VARIABLE":
		return ObjectTypeVariable
	case "PROVIDER":
		return ObjectTypeProvider
	case "PROVIDER_ENDPOINT":
		return ObjectTypeProviderTool
	case "MCPSERVER":
		return ObjectTypeServer
	case "MCPSERVER_TOKEN":
		return ObjectTypeServerToken
	case "RESOURCE":
		return ObjectTypeResource
	case "PROMPT":
		return ObjectTypePrompt
	default:
		return ObjectTypeInvalid
	}
}

const (
	MethodTypeInvalid MethodType = iota
	MethodTypeGet
	MethodTypeHead
	MethodTypePost
	MethodTypePut
	MethodTypePatch
	MethodTypeDelete
	MethodTypeConnect
	MethodTypeOptions
	MethodTypeTrace
)

func (mt MethodType) String() string {
	switch mt {
	case MethodTypeGet:
		return "GET"
	case MethodTypeHead:
		return "HEAD"
	case MethodTypePost:
		return "POST"
	case MethodTypePut:
		return "PUT"
	case MethodTypePatch:
		return "PATCH"
	case MethodTypeDelete:
		return "DELETE"
	case MethodTypeConnect:
		return "CONNECT"
	case MethodTypeOptions:
		return "OPTIONS"
	case MethodTypeTrace:
		return "TRACE"
	default:
		return ""
	}
}

func StringToMethodType(method string) MethodType {
	method = strings.ToUpper(method)
	switch method {
	case "GET":
		return MethodTypeGet
	case "HEAD":
		return MethodTypeHead
	case "POST":
		return MethodTypePost
	case "PUT":
		return MethodTypePut
	case "PATCH":
		return MethodTypePatch
	case "DELETE":
		return MethodTypeDelete
	case "CONNECT":
		return MethodTypeConnect
	case "OPTIONS":
		return MethodTypeOptions
	case "TRACE":
		return MethodTypeTrace
	default:
		return MethodTypeInvalid
	}
}

func (v Variable) MarshalZerologObject(e *zerolog.Event) {
	e.
		Int64("id", v.ID).
		Str("name", v.Name).
		Str("value", string(v.Value)).
		Str("type", v.Type.String()).
		Uint16("type", uint16(v.Type))
}
