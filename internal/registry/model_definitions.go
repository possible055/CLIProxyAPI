// Package registry provides model definitions and lookup helpers for various AI providers.
// Static model metadata is loaded from the embedded models.json file and can be refreshed from network.
package registry

import (
	"strings"
)

const (
	codexBuiltinImageModelID      = "gpt-image-2"
	codexBuiltinAutoReviewModelID = "codex-auto-review"
)

var fixedCodexModelInfos = map[string]*ModelInfo{
	"gpt-5.2": {
		ID:                  "gpt-5.2",
		Object:              "model",
		Created:             1765440000,
		OwnedBy:             "openai",
		Type:                "openai",
		DisplayName:         "GPT 5.2",
		Version:             "gpt-5.2",
		Description:         "Stable version of GPT 5.2",
		ContextLength:       400000,
		MaxCompletionTokens: 128000,
		SupportedParameters: []string{"tools"},
		Thinking:            &ThinkingSupport{Levels: []string{"none", "low", "medium", "high", "xhigh"}},
	},
	"gpt-5.3-codex": {
		ID:                  "gpt-5.3-codex",
		Object:              "model",
		Created:             1770307200,
		OwnedBy:             "openai",
		Type:                "openai",
		DisplayName:         "GPT 5.3 Codex",
		Version:             "gpt-5.3",
		Description:         "Stable version of GPT 5.3 Codex, The best model for coding and agentic tasks across domains.",
		ContextLength:       400000,
		MaxCompletionTokens: 128000,
		SupportedParameters: []string{"tools"},
		Thinking:            &ThinkingSupport{Levels: []string{"low", "medium", "high", "xhigh"}},
	},
	"gpt-5.3-codex-spark": {
		ID:                  "gpt-5.3-codex-spark",
		Object:              "model",
		Created:             1770912000,
		OwnedBy:             "openai",
		Type:                "openai",
		DisplayName:         "GPT 5.3 Codex Spark",
		Version:             "gpt-5.3",
		Description:         "Ultra-fast coding model.",
		ContextLength:       128000,
		MaxCompletionTokens: 128000,
		SupportedParameters: []string{"tools"},
		Thinking:            &ThinkingSupport{Levels: []string{"low", "medium", "high", "xhigh"}},
	},
	"gpt-5.4": {
		ID:                  "gpt-5.4",
		Object:              "model",
		Created:             1772668800,
		OwnedBy:             "openai",
		Type:                "openai",
		DisplayName:         "GPT 5.4",
		Version:             "gpt-5.4",
		Description:         "Stable version of GPT 5.4",
		ContextLength:       1050000,
		MaxCompletionTokens: 128000,
		SupportedParameters: []string{"tools"},
		Thinking:            &ThinkingSupport{Levels: []string{"low", "medium", "high", "xhigh"}},
	},
	"gpt-5.4-mini": {
		ID:                  "gpt-5.4-mini",
		Object:              "model",
		Created:             1773705600,
		OwnedBy:             "openai",
		Type:                "openai",
		DisplayName:         "GPT 5.4 Mini",
		Version:             "gpt-5.4-mini",
		Description:         "GPT-5.4 mini brings the strengths of GPT-5.4 to a faster, more efficient model designed for high-volume workloads.",
		ContextLength:       400000,
		MaxCompletionTokens: 128000,
		SupportedParameters: []string{"tools"},
		Thinking:            &ThinkingSupport{Levels: []string{"low", "medium", "high", "xhigh"}},
	},
	"gpt-5.5": {
		ID:                  "gpt-5.5",
		Object:              "model",
		Created:             1776902400,
		OwnedBy:             "openai",
		Type:                "openai",
		DisplayName:         "GPT 5.5",
		Version:             "gpt-5.5",
		Description:         "Frontier model for complex coding, research, and real-world work.",
		ContextLength:       272000,
		MaxCompletionTokens: 128000,
		SupportedParameters: []string{"tools"},
		Thinking:            &ThinkingSupport{Levels: []string{"low", "medium", "high", "xhigh"}},
	},
}

// staticModelsJSON mirrors the top-level structure of models.json.
type staticModelsJSON struct {
	Claude      []*ModelInfo `json:"claude"`
	Gemini      []*ModelInfo `json:"gemini"`
	Vertex      []*ModelInfo `json:"vertex"`
	GeminiCLI   []*ModelInfo `json:"gemini-cli"`
	AIStudio    []*ModelInfo `json:"aistudio"`
	CodexFree   []*ModelInfo `json:"codex-free"`
	CodexTeam   []*ModelInfo `json:"codex-team"`
	CodexPlus   []*ModelInfo `json:"codex-plus"`
	CodexPro    []*ModelInfo `json:"codex-pro"`
	Kimi        []*ModelInfo `json:"kimi"`
	Antigravity []*ModelInfo `json:"antigravity"`
}

// GetClaudeModels returns the standard Claude model definitions.
func GetClaudeModels() []*ModelInfo {
	return cloneModelInfos(getModels().Claude)
}

// GetGeminiModels returns the standard Gemini model definitions.
func GetGeminiModels() []*ModelInfo {
	return cloneModelInfos(getModels().Gemini)
}

// GetGeminiVertexModels returns Gemini model definitions for Vertex AI.
func GetGeminiVertexModels() []*ModelInfo {
	return cloneModelInfos(getModels().Vertex)
}

// GetGeminiCLIModels returns Gemini model definitions for the Gemini CLI.
func GetGeminiCLIModels() []*ModelInfo {
	return cloneModelInfos(getModels().GeminiCLI)
}

// GetAIStudioModels returns model definitions for AI Studio.
func GetAIStudioModels() []*ModelInfo {
	return cloneModelInfos(getModels().AIStudio)
}

// GetCodexFreeModels returns model definitions for the Codex free plan tier.
func GetCodexFreeModels() []*ModelInfo {
	return fixedCodexModels("gpt-5.2", "gpt-5.3-codex", "gpt-5.4", "gpt-5.4-mini")
}

// GetCodexTeamModels returns model definitions for the Codex team plan tier.
func GetCodexTeamModels() []*ModelInfo {
	return fixedCodexModels("gpt-5.2", "gpt-5.3-codex", "gpt-5.4", "gpt-5.4-mini", "gpt-5.5")
}

// GetCodexPlusModels returns model definitions for the Codex plus plan tier.
func GetCodexPlusModels() []*ModelInfo {
	return fixedCodexModels("gpt-5.2", "gpt-5.3-codex", "gpt-5.3-codex-spark", "gpt-5.4", "gpt-5.4-mini", "gpt-5.5")
}

// GetCodexProModels returns model definitions for the Codex pro plan tier.
func GetCodexProModels() []*ModelInfo {
	return fixedCodexModels("gpt-5.2", "gpt-5.3-codex", "gpt-5.3-codex-spark", "gpt-5.4", "gpt-5.4-mini", "gpt-5.5")
}

// GetKimiModels returns the standard Kimi (Moonshot AI) model definitions.
func GetKimiModels() []*ModelInfo {
	return cloneModelInfos(getModels().Kimi)
}

// GetAntigravityModels returns the standard Antigravity model definitions.
func GetAntigravityModels() []*ModelInfo {
	return cloneModelInfos(getModels().Antigravity)
}

// WithCodexBuiltins injects hard-coded Codex-only model definitions that should
// not depend on remote models.json updates. Built-ins replace any matching IDs
// already present in the provided slice.
func WithCodexBuiltins(models []*ModelInfo) []*ModelInfo {
	return upsertModelInfos(models, codexBuiltinImageModelInfo(), codexBuiltinAutoReviewModelInfo())
}

func fixedCodexModels(ids ...string) []*ModelInfo {
	models := make([]*ModelInfo, 0, len(ids)+2)
	for _, id := range ids {
		model := fixedCodexModelInfo(id)
		if model == nil {
			continue
		}
		models = append(models, model)
	}
	return WithCodexBuiltins(models)
}

func fixedCodexModelInfo(id string) *ModelInfo {
	return cloneModelInfo(fixedCodexModelInfos[id])
}

func codexBuiltinImageModelInfo() *ModelInfo {
	return &ModelInfo{
		ID:          codexBuiltinImageModelID,
		Object:      "model",
		Created:     1704067200, // 2024-01-01
		OwnedBy:     "openai",
		Type:        "openai",
		DisplayName: "GPT Image 2",
		Version:     codexBuiltinImageModelID,
	}
}

func codexBuiltinAutoReviewModelInfo() *ModelInfo {
	return &ModelInfo{
		ID:                  codexBuiltinAutoReviewModelID,
		Object:              "model",
		Created:             1776902400,
		OwnedBy:             "openai",
		Type:                "openai",
		DisplayName:         "Codex Auto Review",
		Version:             codexBuiltinAutoReviewModelID,
		Description:         "Codex auto-review routing model.",
		ContextLength:       272000,
		MaxCompletionTokens: 128000,
		SupportedParameters: []string{"tools"},
		Thinking:            &ThinkingSupport{Levels: []string{"low", "medium", "high", "xhigh"}},
	}
}

func upsertModelInfos(models []*ModelInfo, extras ...*ModelInfo) []*ModelInfo {
	if len(extras) == 0 {
		return models
	}

	extraIDs := make(map[string]struct{}, len(extras))
	extraList := make([]*ModelInfo, 0, len(extras))
	for _, extra := range extras {
		if extra == nil {
			continue
		}
		id := strings.TrimSpace(extra.ID)
		if id == "" {
			continue
		}
		key := strings.ToLower(id)
		if _, exists := extraIDs[key]; exists {
			continue
		}
		extraIDs[key] = struct{}{}
		extraList = append(extraList, cloneModelInfo(extra))
	}

	if len(extraList) == 0 {
		return models
	}

	filtered := make([]*ModelInfo, 0, len(models)+len(extraList))
	for _, model := range models {
		if model == nil {
			continue
		}
		id := strings.TrimSpace(model.ID)
		if id == "" {
			continue
		}
		if _, exists := extraIDs[strings.ToLower(id)]; exists {
			continue
		}
		filtered = append(filtered, model)
	}

	filtered = append(filtered, extraList...)
	return filtered
}

// cloneModelInfos returns a shallow copy of the slice with each element deep-cloned.
func cloneModelInfos(models []*ModelInfo) []*ModelInfo {
	if len(models) == 0 {
		return nil
	}
	out := make([]*ModelInfo, len(models))
	for i, m := range models {
		out[i] = cloneModelInfo(m)
	}
	return out
}

// GetStaticModelDefinitionsByChannel returns static model definitions for a given channel/provider.
// It returns nil when the channel is unknown.
//
// Supported channels:
//   - claude
//   - gemini
//   - vertex
//   - gemini-cli
//   - aistudio
//   - codex
//   - kimi
//   - antigravity
func GetStaticModelDefinitionsByChannel(channel string) []*ModelInfo {
	key := strings.ToLower(strings.TrimSpace(channel))
	switch key {
	case "claude":
		return GetClaudeModels()
	case "gemini":
		return GetGeminiModels()
	case "vertex":
		return GetGeminiVertexModels()
	case "gemini-cli":
		return GetGeminiCLIModels()
	case "aistudio":
		return GetAIStudioModels()
	case "codex":
		return GetCodexProModels()
	case "kimi":
		return GetKimiModels()
	case "antigravity":
		return GetAntigravityModels()
	default:
		return nil
	}
}

// LookupStaticModelInfo searches all static model definitions for a model by ID.
// Returns nil if no matching model is found.
func LookupStaticModelInfo(modelID string) *ModelInfo {
	if modelID == "" {
		return nil
	}

	data := getModels()
	allModels := [][]*ModelInfo{
		data.Claude,
		data.Gemini,
		data.Vertex,
		data.GeminiCLI,
		data.AIStudio,
		GetCodexProModels(),
		data.Kimi,
		data.Antigravity,
	}
	for _, models := range allModels {
		for _, m := range models {
			if m != nil && m.ID == modelID {
				return cloneModelInfo(m)
			}
		}
	}

	return nil
}
