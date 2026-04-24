package registry

import "testing"

func TestGetAvailableModelsReturnsClonedSnapshots(t *testing.T) {
	r := newTestModelRegistry()
	r.RegisterClient("client-1", "OpenAI", []*ModelInfo{{ID: "m1", OwnedBy: "team-a", DisplayName: "Model One"}})

	first := r.GetAvailableModels("openai")
	if len(first) != 1 {
		t.Fatalf("expected 1 model, got %d", len(first))
	}
	first[0]["id"] = "mutated"
	first[0]["display_name"] = "Mutated"

	second := r.GetAvailableModels("openai")
	if got := second[0]["id"]; got != "m1" {
		t.Fatalf("expected cached snapshot to stay isolated, got id %v", got)
	}
	if got := second[0]["display_name"]; got != "Model One" {
		t.Fatalf("expected cached snapshot to stay isolated, got display_name %v", got)
	}
}

func TestGetAvailableModelsInvalidatesCacheOnRegistryChanges(t *testing.T) {
	r := newTestModelRegistry()
	r.RegisterClient("client-1", "OpenAI", []*ModelInfo{{ID: "m1", OwnedBy: "team-a", DisplayName: "Model One"}})

	models := r.GetAvailableModels("openai")
	if len(models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(models))
	}
	if got := models[0]["display_name"]; got != "Model One" {
		t.Fatalf("expected initial display_name Model One, got %v", got)
	}

	r.RegisterClient("client-1", "OpenAI", []*ModelInfo{{ID: "m1", OwnedBy: "team-a", DisplayName: "Model One Updated"}})
	models = r.GetAvailableModels("openai")
	if got := models[0]["display_name"]; got != "Model One Updated" {
		t.Fatalf("expected updated display_name after cache invalidation, got %v", got)
	}

	r.SuspendClientModel("client-1", "m1", "manual")
	models = r.GetAvailableModels("openai")
	if len(models) != 0 {
		t.Fatalf("expected no available models after suspension, got %d", len(models))
	}

	r.ResumeClientModel("client-1", "m1")
	models = r.GetAvailableModels("openai")
	if len(models) != 1 {
		t.Fatalf("expected model to reappear after resume, got %d", len(models))
	}
}

func TestGetAvailableModelsKeepsCodexSuspendedModelsListed(t *testing.T) {
	r := newTestModelRegistry()
	r.RegisterClient("codex-1", "codex", []*ModelInfo{{ID: "gpt-5.4-mini", OwnedBy: "openai", Type: "openai"}})

	r.SuspendClientModel("codex-1", "gpt-5.4-mini", "model_not_supported")

	models := r.GetAvailableModels("openai")
	if !modelMapContainsID(models, "gpt-5.4-mini") {
		t.Fatalf("expected suspended codex model to remain listed, got %+v", models)
	}

	providerModels := r.GetAvailableModelsByProvider("codex")
	if findModelInfo(providerModels, "gpt-5.4-mini") == nil {
		t.Fatalf("expected suspended codex provider model to remain listed, got %+v", providerModels)
	}
}

func modelMapContainsID(models []map[string]any, id string) bool {
	for _, model := range models {
		if model == nil {
			continue
		}
		if got, _ := model["id"].(string); got == id {
			return true
		}
	}
	return false
}
