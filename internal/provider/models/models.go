package models

// ModelProvider represents the provider of the model
type ModelProvider string

// ModelID represents the unique identifier for a model
type ModelID string

// Model represents the details of a model
type Model struct {
	ID                  ModelID
	Name                string
	Provider            ModelProvider
	APIModel            string
	CostPer1MIn         float64
	CostPer1MInCached   float64
	CostPer1MOut        float64
	CostPer1MOutCached  float64
	ContextWindow       int
	DefaultMaxTokens    int
	CanReason           bool
	SupportsAttachments bool
}

const (
	ProviderOpenAI ModelProvider = "openai"

	GPT41        ModelID = "gpt-4.1"
	GPT41Mini    ModelID = "gpt-4.1-mini"
	GPT41Nano    ModelID = "gpt-4.1-nano"
	GPT45Preview ModelID = "gpt-4.5-preview"
	GPT4o        ModelID = "gpt-4o"
	GPT4oMini    ModelID = "gpt-4o-mini"
	O1           ModelID = "o1"
	O1Pro        ModelID = "o1-pro"
	O1Mini       ModelID = "o1-mini"
	O3           ModelID = "o3"
	O3Mini       ModelID = "o3-mini"
	O4Mini       ModelID = "o4-mini"
)

var OpenAIModels = map[ModelID]Model{
	GPT41: {
		ID:                  GPT41,
		Name:                "GPT 4.1",
		Provider:            ProviderOpenAI,
		APIModel:            "gpt-4.1",
		CostPer1MIn:         2.00,
		CostPer1MInCached:   0.50,
		CostPer1MOutCached:  0.0,
		CostPer1MOut:        8.00,
		ContextWindow:       1_047_576,
		DefaultMaxTokens:    20000,
		SupportsAttachments: true,
	},
	GPT41Mini: {
		ID:                  GPT41Mini,
		Name:                "GPT 4.1 mini",
		Provider:            ProviderOpenAI,
		APIModel:            "gpt-4.1",
		CostPer1MIn:         0.40,
		CostPer1MInCached:   0.10,
		CostPer1MOutCached:  0.0,
		CostPer1MOut:        1.60,
		ContextWindow:       200_000,
		DefaultMaxTokens:    20000,
		SupportsAttachments: true,
	},
	GPT41Nano: {
		ID:                  GPT41Nano,
		Name:                "GPT 4.1 nano",
		Provider:            ProviderOpenAI,
		APIModel:            "gpt-4.1-nano",
		CostPer1MIn:         0.10,
		CostPer1MInCached:   0.025,
		CostPer1MOutCached:  0.0,
		CostPer1MOut:        0.40,
		ContextWindow:       1_047_576,
		DefaultMaxTokens:    20000,
		SupportsAttachments: true,
	},
	GPT45Preview: {
		ID:                  GPT45Preview,
		Name:                "GPT 4.5 preview",
		Provider:            ProviderOpenAI,
		APIModel:            "gpt-4.5-preview",
		CostPer1MIn:         75.00,
		CostPer1MInCached:   37.50,
		CostPer1MOutCached:  0.0,
		CostPer1MOut:        150.00,
		ContextWindow:       128_000,
		DefaultMaxTokens:    15000,
		SupportsAttachments: true,
	},
	GPT4o: {
		ID:                  GPT4o,
		Name:                "GPT 4o",
		Provider:            ProviderOpenAI,
		APIModel:            "gpt-4o",
		CostPer1MIn:         2.50,
		CostPer1MInCached:   1.25,
		CostPer1MOutCached:  0.0,
		CostPer1MOut:        10.00,
		ContextWindow:       128_000,
		DefaultMaxTokens:    4096,
		SupportsAttachments: true,
	},
	GPT4oMini: {
		ID:                  GPT4oMini,
		Name:                "GPT 4o mini",
		Provider:            ProviderOpenAI,
		APIModel:            "gpt-4o-mini",
		CostPer1MIn:         0.15,
		CostPer1MInCached:   0.075,
		CostPer1MOutCached:  0.0,
		CostPer1MOut:        0.60,
		ContextWindow:       128_000,
		SupportsAttachments: true,
	},
	O1: {
		ID:                  O1,
		Name:                "O1",
		Provider:            ProviderOpenAI,
		APIModel:            "o1",
		CostPer1MIn:         15.00,
		CostPer1MInCached:   7.50,
		CostPer1MOutCached:  0.0,
		CostPer1MOut:        60.00,
		ContextWindow:       200_000,
		DefaultMaxTokens:    50000,
		CanReason:           true,
		SupportsAttachments: true,
	},
	O1Pro: {
		ID:                  O1Pro,
		Name:                "o1 pro",
		Provider:            ProviderOpenAI,
		APIModel:            "o1-pro",
		CostPer1MIn:         150.00,
		CostPer1MInCached:   0.0,
		CostPer1MOutCached:  0.0,
		CostPer1MOut:        600.00,
		ContextWindow:       200_000,
		DefaultMaxTokens:    50000,
		CanReason:           true,
		SupportsAttachments: true,
	},
	O1Mini: {
		ID:                  O1Mini,
		Name:                "o1 mini",
		Provider:            ProviderOpenAI,
		APIModel:            "o1-mini",
		CostPer1MIn:         1.10,
		CostPer1MInCached:   0.55,
		CostPer1MOutCached:  0.0,
		CostPer1MOut:        4.40,
		ContextWindow:       128_000,
		DefaultMaxTokens:    50000,
		CanReason:           true,
		SupportsAttachments: true,
	},
	O3: {
		ID:                  O3,
		Name:                "o3",
		Provider:            ProviderOpenAI,
		APIModel:            "o3",
		CostPer1MIn:         10.00,
		CostPer1MInCached:   2.50,
		CostPer1MOutCached:  0.0,
		CostPer1MOut:        40.00,
		ContextWindow:       200_000,
		CanReason:           true,
		SupportsAttachments: true,
	},
	O3Mini: {
		ID:                  O3Mini,
		Name:                "o3 mini",
		Provider:            ProviderOpenAI,
		APIModel:            "o3-mini",
		CostPer1MIn:         1.10,
		CostPer1MInCached:   0.55,
		CostPer1MOutCached:  0.0,
		CostPer1MOut:        4.40,
		ContextWindow:       200_000,
		DefaultMaxTokens:    50000,
		CanReason:           true,
		SupportsAttachments: false,
	},
	O4Mini: {
		ID:                  O4Mini,
		Name:                "o4 mini",
		Provider:            ProviderOpenAI,
		APIModel:            "o4-mini",
		CostPer1MIn:         1.10,
		CostPer1MInCached:   0.275,
		CostPer1MOutCached:  0.0,
		CostPer1MOut:        4.40,
		ContextWindow:       128_000,
		DefaultMaxTokens:    50000,
		CanReason:           true,
		SupportsAttachments: true,
	},
}

// OpenRouter model provider and model IDs
const (
	ProviderOpenRouter ModelProvider = "openrouter"

	OpenRouterGPT41          ModelID = "openrouter.gpt-4.1"
	OpenRouterGPT41Mini      ModelID = "openrouter.gpt-4.1-mini"
	OpenRouterGPT41Nano      ModelID = "openrouter.gpt-4.1-nano"
	OpenRouterGPT45Preview   ModelID = "openrouter.gpt-4.5-preview"
	OpenRouterGPT4o          ModelID = "openrouter.gpt-4o"
	OpenRouterGPT4oMini      ModelID = "openrouter.gpt-4o-mini"
	OpenRouterO1             ModelID = "openrouter.o1"
	OpenRouterO1Pro          ModelID = "openrouter.o1-pro"
	OpenRouterO1Mini         ModelID = "openrouter.o1-mini"
	OpenRouterO3             ModelID = "openrouter.o3"
	OpenRouterO3Mini         ModelID = "openrouter.o3-mini"
	OpenRouterO4Mini         ModelID = "openrouter.o4-mini"
	OpenRouterGemini25Flash  ModelID = "openrouter.gemini-2.5-flash"
	OpenRouterGemini25       ModelID = "openrouter.gemini-2.5"
	OpenRouterClaude35Sonnet ModelID = "openrouter.claude-3.5-sonnet"
	OpenRouterClaude3Haiku   ModelID = "openrouter.claude-3-haiku"
	OpenRouterClaude37Sonnet ModelID = "openrouter.claude-3.7-sonnet"
	OpenRouterClaude35Haiku  ModelID = "openrouter.claude-3.5-haiku"
	OpenRouterClaude3Opus    ModelID = "openrouter.claude-3-opus"
	OpenRouterDeepSeekR1Free ModelID = "openrouter.deepseek-r1-free"
)

var OpenRouterModels = map[ModelID]Model{
	OpenRouterGPT41: {
		ID:                 OpenRouterGPT41,
		Name:               "OpenRouter – GPT 4.1",
		Provider:           ProviderOpenRouter,
		APIModel:           "openai/gpt-4.1",
		CostPer1MIn:        OpenAIModels[GPT41].CostPer1MIn,
		CostPer1MInCached:  OpenAIModels[GPT41].CostPer1MInCached,
		CostPer1MOut:       OpenAIModels[GPT41].CostPer1MOut,
		CostPer1MOutCached: OpenAIModels[GPT41].CostPer1MOutCached,
		ContextWindow:      OpenAIModels[GPT41].ContextWindow,
		DefaultMaxTokens:   OpenAIModels[GPT41].DefaultMaxTokens,
	},
	OpenRouterGPT41Mini: {
		ID:                 OpenRouterGPT41Mini,
		Name:               "OpenRouter – GPT 4.1 mini",
		Provider:           ProviderOpenRouter,
		APIModel:           "openai/gpt-4.1-mini",
		CostPer1MIn:        OpenAIModels[GPT41Mini].CostPer1MIn,
		CostPer1MInCached:  OpenAIModels[GPT41Mini].CostPer1MInCached,
		CostPer1MOut:       OpenAIModels[GPT41Mini].CostPer1MOut,
		CostPer1MOutCached: OpenAIModels[GPT41Mini].CostPer1MOutCached,
		ContextWindow:      OpenAIModels[GPT41Mini].ContextWindow,
		DefaultMaxTokens:   OpenAIModels[GPT41Mini].DefaultMaxTokens,
	},
	OpenRouterGPT41Nano: {
		ID:                 OpenRouterGPT41Nano,
		Name:               "OpenRouter – GPT 4.1 nano",
		Provider:           ProviderOpenRouter,
		APIModel:           "openai/gpt-4.1-nano",
		CostPer1MIn:        OpenAIModels[GPT41Nano].CostPer1MIn,
		CostPer1MInCached:  OpenAIModels[GPT41Nano].CostPer1MInCached,
		CostPer1MOut:       OpenAIModels[GPT41Nano].CostPer1MOut,
		CostPer1MOutCached: OpenAIModels[GPT41Nano].CostPer1MOutCached,
		ContextWindow:      OpenAIModels[GPT41Nano].ContextWindow,
		DefaultMaxTokens:   OpenAIModels[GPT41Nano].DefaultMaxTokens,
	},
	OpenRouterGPT45Preview: {
		ID:                 OpenRouterGPT45Preview,
		Name:               "OpenRouter – GPT 4.5 preview",
		Provider:           ProviderOpenRouter,
		APIModel:           "openai/gpt-4.5-preview",
		CostPer1MIn:        OpenAIModels[GPT45Preview].CostPer1MIn,
		CostPer1MInCached:  OpenAIModels[GPT45Preview].CostPer1MInCached,
		CostPer1MOut:       OpenAIModels[GPT45Preview].CostPer1MOut,
		CostPer1MOutCached: OpenAIModels[GPT45Preview].CostPer1MOutCached,
		ContextWindow:      OpenAIModels[GPT45Preview].ContextWindow,
		DefaultMaxTokens:   OpenAIModels[GPT45Preview].DefaultMaxTokens,
	},
	OpenRouterGPT4o: {
		ID:                 OpenRouterGPT4o,
		Name:               "OpenRouter – GPT 4o",
		Provider:           ProviderOpenRouter,
		APIModel:           "openai/gpt-4o",
		CostPer1MIn:        OpenAIModels[GPT4o].CostPer1MIn,
		CostPer1MInCached:  OpenAIModels[GPT4o].CostPer1MInCached,
		CostPer1MOut:       OpenAIModels[GPT4o].CostPer1MOut,
		CostPer1MOutCached: OpenAIModels[GPT4o].CostPer1MOutCached,
		ContextWindow:      OpenAIModels[GPT4o].ContextWindow,
		DefaultMaxTokens:   OpenAIModels[GPT4o].DefaultMaxTokens,
	},
	OpenRouterGPT4oMini: {
		ID:                 OpenRouterGPT4oMini,
		Name:               "OpenRouter – GPT 4o mini",
		Provider:           ProviderOpenRouter,
		APIModel:           "openai/gpt-4o-mini",
		CostPer1MIn:        OpenAIModels[GPT4oMini].CostPer1MIn,
		CostPer1MInCached:  OpenAIModels[GPT4oMini].CostPer1MInCached,
		CostPer1MOut:       OpenAIModels[GPT4oMini].CostPer1MOut,
		CostPer1MOutCached: OpenAIModels[GPT4oMini].CostPer1MOutCached,
		ContextWindow:      OpenAIModels[GPT4oMini].ContextWindow,
	},
	OpenRouterO1: {
		ID:                 OpenRouterO1,
		Name:               "OpenRouter – O1",
		Provider:           ProviderOpenRouter,
		APIModel:           "openai/o1",
		CostPer1MIn:        OpenAIModels[O1].CostPer1MIn,
		CostPer1MInCached:  OpenAIModels[O1].CostPer1MInCached,
		CostPer1MOut:       OpenAIModels[O1].CostPer1MOut,
		CostPer1MOutCached: OpenAIModels[O1].CostPer1MOutCached,
		ContextWindow:      OpenAIModels[O1].ContextWindow,
		DefaultMaxTokens:   OpenAIModels[O1].DefaultMaxTokens,
		CanReason:          OpenAIModels[O1].CanReason,
	},
	OpenRouterO1Pro: {
		ID:                 OpenRouterO1Pro,
		Name:               "OpenRouter – o1 pro",
		Provider:           ProviderOpenRouter,
		APIModel:           "openai/o1-pro",
		CostPer1MIn:        OpenAIModels[O1Pro].CostPer1MIn,
		CostPer1MInCached:  OpenAIModels[O1Pro].CostPer1MInCached,
		CostPer1MOut:       OpenAIModels[O1Pro].CostPer1MOut,
		CostPer1MOutCached: OpenAIModels[O1Pro].CostPer1MOutCached,
		ContextWindow:      OpenAIModels[O1Pro].ContextWindow,
		DefaultMaxTokens:   OpenAIModels[O1Pro].DefaultMaxTokens,
		CanReason:          OpenAIModels[O1Pro].CanReason,
	},
	OpenRouterO1Mini: {
		ID:                 OpenRouterO1Mini,
		Name:               "OpenRouter – o1 mini",
		Provider:           ProviderOpenRouter,
		APIModel:           "openai/o1-mini",
		CostPer1MIn:        OpenAIModels[O1Mini].CostPer1MIn,
		CostPer1MInCached:  OpenAIModels[O1Mini].CostPer1MInCached,
		CostPer1MOut:       OpenAIModels[O1Mini].CostPer1MOut,
		CostPer1MOutCached: OpenAIModels[O1Mini].CostPer1MOutCached,
		ContextWindow:      OpenAIModels[O1Mini].ContextWindow,
		DefaultMaxTokens:   OpenAIModels[O1Mini].DefaultMaxTokens,
		CanReason:          OpenAIModels[O1Mini].CanReason,
	},
	OpenRouterO3: {
		ID:                 OpenRouterO3,
		Name:               "OpenRouter – o3",
		Provider:           ProviderOpenRouter,
		APIModel:           "openai/o3",
		CostPer1MIn:        OpenAIModels[O3].CostPer1MIn,
		CostPer1MInCached:  OpenAIModels[O3].CostPer1MInCached,
		CostPer1MOut:       OpenAIModels[O3].CostPer1MOut,
		CostPer1MOutCached: OpenAIModels[O3].CostPer1MOutCached,
		ContextWindow:      OpenAIModels[O3].ContextWindow,
		DefaultMaxTokens:   OpenAIModels[O3].DefaultMaxTokens,
		CanReason:          OpenAIModels[O3].CanReason,
	},
	OpenRouterO3Mini: {
		ID:                 OpenRouterO3Mini,
		Name:               "OpenRouter – o3 mini",
		Provider:           ProviderOpenRouter,
		APIModel:           "openai/o3-mini-high",
		CostPer1MIn:        OpenAIModels[O3Mini].CostPer1MIn,
		CostPer1MInCached:  OpenAIModels[O3Mini].CostPer1MInCached,
		CostPer1MOut:       OpenAIModels[O3Mini].CostPer1MOut,
		CostPer1MOutCached: OpenAIModels[O3Mini].CostPer1MOutCached,
		ContextWindow:      OpenAIModels[O3Mini].ContextWindow,
		DefaultMaxTokens:   OpenAIModels[O3Mini].DefaultMaxTokens,
		CanReason:          OpenAIModels[O3Mini].CanReason,
	},
	OpenRouterO4Mini: {
		ID:                 OpenRouterO4Mini,
		Name:               "OpenRouter – o4 mini",
		Provider:           ProviderOpenRouter,
		APIModel:           "openai/o4-mini-high",
		CostPer1MIn:        OpenAIModels[O4Mini].CostPer1MIn,
		CostPer1MInCached:  OpenAIModels[O4Mini].CostPer1MInCached,
		CostPer1MOut:       OpenAIModels[O4Mini].CostPer1MOut,
		CostPer1MOutCached: OpenAIModels[O4Mini].CostPer1MOutCached,
		ContextWindow:      OpenAIModels[O4Mini].ContextWindow,
		DefaultMaxTokens:   OpenAIModels[O4Mini].DefaultMaxTokens,
		CanReason:          OpenAIModels[O4Mini].CanReason,
	},
	// The following models reference GeminiModels and AnthropicModels, which are not defined in this codebase yet.
	// TODO: Implement GeminiModels and AnthropicModels, then update these entries accordingly.
	OpenRouterGemini25Flash: {
		ID:                 OpenRouterGemini25Flash,
		Name:               "OpenRouter – Gemini 2.5 Flash",
		Provider:           ProviderOpenRouter,
		APIModel:           "google/gemini-2.5-flash-preview:thinking",
		CostPer1MIn:        0, // TODO: GeminiModels[Gemini25Flash].CostPer1MIn
		CostPer1MInCached:  0, // TODO: GeminiModels[Gemini25Flash].CostPer1MInCached
		CostPer1MOut:       0, // TODO: GeminiModels[Gemini25Flash].CostPer1MOut
		CostPer1MOutCached: 0, // TODO: GeminiModels[Gemini25Flash].CostPer1MOutCached
		ContextWindow:      0, // TODO: GeminiModels[Gemini25Flash].ContextWindow
		DefaultMaxTokens:   0, // TODO: GeminiModels[Gemini25Flash].DefaultMaxTokens
	},
	OpenRouterGemini25: {
		ID:                 OpenRouterGemini25,
		Name:               "OpenRouter – Gemini 2.5 Pro",
		Provider:           ProviderOpenRouter,
		APIModel:           "google/gemini-2.5-pro-preview-03-25",
		CostPer1MIn:        0, // TODO: GeminiModels[Gemini25].CostPer1MIn
		CostPer1MInCached:  0, // TODO: GeminiModels[Gemini25].CostPer1MInCached
		CostPer1MOut:       0, // TODO: GeminiModels[Gemini25].CostPer1MOut
		CostPer1MOutCached: 0, // TODO: GeminiModels[Gemini25].CostPer1MOutCached
		ContextWindow:      0, // TODO: GeminiModels[Gemini25].ContextWindow
		DefaultMaxTokens:   0, // TODO: GeminiModels[Gemini25].DefaultMaxTokens
	},
	OpenRouterClaude35Sonnet: {
		ID:                 OpenRouterClaude35Sonnet,
		Name:               "OpenRouter – Claude 3.5 Sonnet",
		Provider:           ProviderOpenRouter,
		APIModel:           "anthropic/claude-3.5-sonnet",
		CostPer1MIn:        0, // TODO: AnthropicModels[Claude35Sonnet].CostPer1MIn
		CostPer1MInCached:  0, // TODO: AnthropicModels[Claude35Sonnet].CostPer1MInCached
		CostPer1MOut:       0, // TODO: AnthropicModels[Claude35Sonnet].CostPer1MOut
		CostPer1MOutCached: 0, // TODO: AnthropicModels[Claude35Sonnet].CostPer1MOutCached
		ContextWindow:      0, // TODO: AnthropicModels[Claude35Sonnet].ContextWindow
		DefaultMaxTokens:   0, // TODO: AnthropicModels[Claude35Sonnet].DefaultMaxTokens
	},
	OpenRouterClaude3Haiku: {
		ID:                 OpenRouterClaude3Haiku,
		Name:               "OpenRouter – Claude 3 Haiku",
		Provider:           ProviderOpenRouter,
		APIModel:           "anthropic/claude-3-haiku",
		CostPer1MIn:        0, // TODO: AnthropicModels[Claude3Haiku].CostPer1MIn
		CostPer1MInCached:  0, // TODO: AnthropicModels[Claude3Haiku].CostPer1MInCached
		CostPer1MOut:       0, // TODO: AnthropicModels[Claude3Haiku].CostPer1MOut
		CostPer1MOutCached: 0, // TODO: AnthropicModels[Claude3Haiku].CostPer1MOutCached
		ContextWindow:      0, // TODO: AnthropicModels[Claude3Haiku].ContextWindow
		DefaultMaxTokens:   0, // TODO: AnthropicModels[Claude3Haiku].DefaultMaxTokens
	},
	OpenRouterClaude37Sonnet: {
		ID:                 OpenRouterClaude37Sonnet,
		Name:               "OpenRouter – Claude 3.7 Sonnet",
		Provider:           ProviderOpenRouter,
		APIModel:           "anthropic/claude-3.7-sonnet",
		CostPer1MIn:        0,     // TODO: AnthropicModels[Claude37Sonnet].CostPer1MIn
		CostPer1MInCached:  0,     // TODO: AnthropicModels[Claude37Sonnet].CostPer1MInCached
		CostPer1MOut:       0,     // TODO: AnthropicModels[Claude37Sonnet].CostPer1MOut
		CostPer1MOutCached: 0,     // TODO: AnthropicModels[Claude37Sonnet].CostPer1MOutCached
		ContextWindow:      0,     // TODO: AnthropicModels[Claude37Sonnet].ContextWindow
		DefaultMaxTokens:   0,     // TODO: AnthropicModels[Claude37Sonnet].DefaultMaxTokens
		CanReason:          false, // TODO: AnthropicModels[Claude37Sonnet].CanReason
	},
	OpenRouterClaude35Haiku: {
		ID:                 OpenRouterClaude35Haiku,
		Name:               "OpenRouter – Claude 3.5 Haiku",
		Provider:           ProviderOpenRouter,
		APIModel:           "anthropic/claude-3.5-haiku",
		CostPer1MIn:        0, // TODO: AnthropicModels[Claude35Haiku].CostPer1MIn
		CostPer1MInCached:  0, // TODO: AnthropicModels[Claude35Haiku].CostPer1MInCached
		CostPer1MOut:       0, // TODO: AnthropicModels[Claude35Haiku].CostPer1MOut
		CostPer1MOutCached: 0, // TODO: AnthropicModels[Claude35Haiku].CostPer1MOutCached
		ContextWindow:      0, // TODO: AnthropicModels[Claude35Haiku].ContextWindow
		DefaultMaxTokens:   0, // TODO: AnthropicModels[Claude35Haiku].DefaultMaxTokens
	},
	OpenRouterClaude3Opus: {
		ID:                 OpenRouterClaude3Opus,
		Name:               "OpenRouter – Claude 3 Opus",
		Provider:           ProviderOpenRouter,
		APIModel:           "anthropic/claude-3-opus",
		CostPer1MIn:        0, // TODO: AnthropicModels[Claude3Opus].CostPer1MIn
		CostPer1MInCached:  0, // TODO: AnthropicModels[Claude3Opus].CostPer1MInCached
		CostPer1MOut:       0, // TODO: AnthropicModels[Claude3Opus].CostPer1MOut
		CostPer1MOutCached: 0, // TODO: AnthropicModels[Claude3Opus].CostPer1MOutCached
		ContextWindow:      0, // TODO: AnthropicModels[Claude3Opus].ContextWindow
		DefaultMaxTokens:   0, // TODO: AnthropicModels[Claude3Opus].DefaultMaxTokens
	},
	OpenRouterDeepSeekR1Free: {
		ID:                 OpenRouterDeepSeekR1Free,
		Name:               "OpenRouter – DeepSeek R1 Free",
		Provider:           ProviderOpenRouter,
		APIModel:           "deepseek/deepseek-r1-0528:free",
		CostPer1MIn:        0,
		CostPer1MInCached:  0,
		CostPer1MOut:       0,
		CostPer1MOutCached: 0,
		ContextWindow:      163_840,
		DefaultMaxTokens:   10000,
	},
}
