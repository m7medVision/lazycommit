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
