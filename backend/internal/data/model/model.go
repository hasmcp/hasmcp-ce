package model

import (
	"encoding/json"
	"time"
)

type (
	// Variable hosts env and secret variable
	Variable struct {
		ID        int64 `gorm:"primaryKey;autoIncrement:false"`
		CreatedAt time.Time
		UpdatedAt time.Time

		Type  uint8  // 0: INVALID, 1: ENV, 2: SECRET
		Value string `gorm:"type:varchar(255)"`
		Nonce string `gorm:"type:varchar(255)"`
		Name  string `gorm:"type:varchar(128)"`
	}

	VariableAttribute string

	// Provider hosts API providers
	Provider struct {
		ID        int64 `gorm:"primaryKey;autoIncrement:false"`
		CreatedAt time.Time
		UpdatedAt time.Time

		Version        int32
		ApiType        uint8  `gorm:"default:1"` // 0: INVALID, 1: REST
		VisibilityType uint8  `gorm:"default:1"` // 0: INVALID, 1: INTERNAL, 2: PUBLIC
		BaseURL        string `gorm:"varchar(255)"`
		DocumentURL    string `gorm:"varchar(255)"`
		IconURL        string `gorm:"varchar(255)"`
		SecretPrefix   string `gorm:"varchar(64)"`
		Name           string `gorm:"varchar(64)"`
		Description    string `gorm:"type:text"`

		Tools        []ProviderTool       `gorm:"foreignKey:provider_id"`
		Oauth2Config ProviderOauth2Config `gorm:"foreignKey:provider_id"`
	}

	ProviderAttribute string

	// ProviderOauth2Config hosts oauth2 configuration for the provider (1:1)
	ProviderOauth2Config struct {
		ID                          int64 `gorm:"primaryKey;autoIncrement:false"`
		ProviderID                  int64
		ClientID                    string
		ClientSecretEncrypted       string
		ClientSecretEncryptionNonce string
		AuthURL                     string
		TokenURL                    string
	}

	// ProviderTool hosts the tools for the provider
	ProviderTool struct {
		ID        int64 `gorm:"primaryKey;autoIncrement:false"`
		CreatedAt time.Time
		UpdatedAt time.Time

		ProviderID          int64           `gorm:"index:uidx_tool_provider_id_method_path,unique"` // Foreign key to Provider
		Method              uint8           `gorm:"index:uidx_tool_provider_id_method_path,unique"` // 0: INVALID, 1: GET, 2: HEAD, 3: POST, 4: PUT, 5: PATCH, 6: DELETE, 7: CONNECT, 8: OPTIONS, 9: TRACE
		Path                string          `gorm:"index:uidx_tool_provider_id_method_path,unique;type:varchar(255)"`
		Name                string          `gorm:"type:varchar(32)"`
		Title               string          `gorm:"type:varchar(128)"`
		Description         string          `gorm:"type:text"`
		PathArgsJSONSchema  json.RawMessage `gorm:"type:bytea"`
		QueryArgsJSONSchema json.RawMessage `gorm:"type:bytea"`
		ReqBodyJSONSchema   json.RawMessage `gorm:"type:bytea"`
		ResBodyJSONSchema   json.RawMessage `gorm:"type:bytea"`
		Headers             json.RawMessage `gorm:"type:bytea"`
		Oauth2Scopes        string
	}

	ProviderToolAttribute string

	ServerAttribute string

	// Server hosts a server of a set of provider tools
	Server struct {
		ID        int64 `gorm:"primaryKey;autoIncrement:false"`
		CreatedAt time.Time
		UpdatedAt time.Time

		RequestHeadersProxyEnabled bool

		Name         string `gorm:"type:varchar(128)"`
		Instructions string `gorm:"type:text"`
		Version      int32

		Tools     []ServerTool     `gorm:"foreignKey:server_id"`
		Resources []ServerResource `gorm:"foreignKey:server_id"`
		Prompts   []ServerPrompt   `gorm:"foreignKey:server_id"`

		VisibilityType uint8 `gorm:"default:1"` // 0: INVALID, 1: INTERNAL, 2: PUBLIC
	}

	// ServerTools hosts the tools that are used in the server
	ServerTool struct {
		ServerID   int64 `gorm:"primaryKey;autoIncrement:false"`
		ProviderID int64 `gorm:"primaryKey;autoIncrement:false"`
		ToolID     int64 `gorm:"primaryKey;autoIncrement:false"`
	}

	// ServerResource hosts the resources that are used in the server
	ServerResource struct {
		ServerID   int64 `gorm:"primaryKey;autoIncrement:false"`
		ResourceID int64 `gorm:"primaryKey;autoIncrement:false"`
	}

	// ServerPrompt hosts the prompts that are used in the server
	ServerPrompt struct {
		ServerID int64 `gorm:"primaryKey;autoIncrement:false"`
		PromptID int64 `gorm:"primaryKey;autoIncrement:false"`
	}

	// Resource hosts a known resource that the server is capable of reading.
	Resource struct {
		ID        int64 `gorm:"primaryKey;autoIncrement:false"`
		CreatedAt time.Time
		UpdatedAt time.Time

		Name        string `gorm:"type:varchar(128)"`
		Description string `gorm:"type:varchar(1024)"`
		URI         string `gorm:"type:varchar(255)"`
		MimeType    string `gorm:"type:varchar(64)"`
		Size        int64
		Annotations json.RawMessage `gorm:"type:bytea"`

		VisibilityType uint8 `gorm:"default:1"` // 0: INVALID, 1: INTERNAL, 2: PUBLIC
	}

	ResourceAttribute string

	// Prompt hosts a prompt or prompt template that the server offers.
	Prompt struct {
		ID        int64 `gorm:"primaryKey;autoIncrement:false"`
		CreatedAt time.Time
		UpdatedAt time.Time

		Name        string          `gorm:"type:varchar(128)"`
		Description string          `gorm:"type:text"`
		Arguments   json.RawMessage `gorm:"type:bytea"` // Stores []PromptArgument
		Messages    json.RawMessage `gorm:"type:bytea"` // Stores []PromptMessage

		VisibilityType uint8 `gorm:"default:1"` // 0: INVALID, 1: INTERNAL, 2: PUBLIC
	}

	PromptAttribute string
)

// Provider mutable attributes
const (
	ProviderAttributeName         ProviderAttribute = "name"
	ProviderAttributeDocumentURL  ProviderAttribute = "document_url"
	ProviderAttributeIconURL      ProviderAttribute = "icon_url"
	ProviderAttributeDescription  ProviderAttribute = "description"
	ProviderAttributeUpdatedAt    ProviderAttribute = "updated_at"
	ProviderAttributeVersion      ProviderAttribute = "version"
	ProviderAttributeOauth2Config ProviderAttribute = "oauth2_config"
)

func (a ProviderAttribute) String() string {
	return string(a)
}

// Provider mutable attributes
const (
	ProviderToolAttributeName                ProviderToolAttribute = "name"
	ProviderToolAttributeTitle               ProviderToolAttribute = "title"
	ProviderToolAttributeDescription         ProviderToolAttribute = "description"
	ProviderToolAttributePathArgsJSONSchema  ProviderToolAttribute = "path_args_json_schema"
	ProviderToolAttributeQueryArgsJSONSchema ProviderToolAttribute = "query_args_json_schema"
	ProviderToolAttributeReqBodyJSONSchema   ProviderToolAttribute = "req_body_json_schema"
	ProviderToolAttributeResBodyJSONSchema   ProviderToolAttribute = "res_body_json_schema"
	ProviderToolAttributeHeaders             ProviderToolAttribute = "headers"
	ProviderToolAttributeOauth2Scopes        ProviderToolAttribute = "oauth2_scopes"
	ProviderToolAttributeUpdatedAt           ProviderToolAttribute = "updated_at"
)

func (a ProviderToolAttribute) String() string {
	return string(a)
}

const (
	ServerAttributeName                       ServerAttribute = "name"
	ServerAttributeInstructions               ServerAttribute = "instructions"
	ServerAttributeUpdatedAt                  ServerAttribute = "updated_at"
	ServerAttributeVersion                    ServerAttribute = "version"
	ServerAttributeTools                      ServerAttribute = "tools"
	ServerAttributeResources                  ServerAttribute = "resources"
	ServerAttributePrompts                    ServerAttribute = "prompts"
	ServerAttributeRequestHeadersProxyEnabled ServerAttribute = "request_headers_proxy_enabled"
)

func (a ServerAttribute) String() string {
	return string(a)
}

const (
	VariableAttributeValue     VariableAttribute = "value"
	VariableAttributeNonce     VariableAttribute = "nonce"
	VariableAttributeUpdatedAt VariableAttribute = "updated_at"
)

func (a VariableAttribute) String() string {
	return string(a)
}

// Resource mutable attributes
const (
	ResourceAttributeName        ResourceAttribute = "name"
	ResourceAttributeDescription ResourceAttribute = "description"
	ResourceAttributeURI         ResourceAttribute = "uri"
	ResourceAttributeMimeType    ResourceAttribute = "mime_type"
	ResourceAttributeSize        ResourceAttribute = "size"
	ResourceAttributeAnnotations ResourceAttribute = "annotations"
	ResourceAttributeUpdatedAt   ResourceAttribute = "updated_at"
)

func (a ResourceAttribute) String() string {
	return string(a)
}

// Prompt mutable attributes
const (
	PromptAttributeName        PromptAttribute = "name"
	PromptAttributeDescription PromptAttribute = "description"
	PromptAttributeArguments   PromptAttribute = "arguments"
	PromptAttributeMessages    PromptAttribute = "messages"
	PromptAttributeUpdatedAt   PromptAttribute = "updated_at"
)

func (a PromptAttribute) String() string {
	return string(a)
}
