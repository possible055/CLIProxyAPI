package registry

import "testing"

func TestCodexStaticModelsUseFixedCatalog(t *testing.T) {
	tierModels := map[string]struct {
		models []*ModelInfo
		want   []string
	}{
		"free": {
			models: GetCodexFreeModels(),
			want: []string{
				"gpt-5.2",
				"gpt-5.3-codex",
				"gpt-5.4",
				"gpt-5.4-mini",
				"gpt-image-2",
				"codex-auto-review",
			},
		},
		"team": {
			models: GetCodexTeamModels(),
			want: []string{
				"gpt-5.2",
				"gpt-5.3-codex",
				"gpt-5.4",
				"gpt-5.4-mini",
				"gpt-5.5",
				"gpt-image-2",
				"codex-auto-review",
			},
		},
		"plus": {
			models: GetCodexPlusModels(),
			want: []string{
				"gpt-5.2",
				"gpt-5.3-codex",
				"gpt-5.3-codex-spark",
				"gpt-5.4",
				"gpt-5.4-mini",
				"gpt-5.5",
				"gpt-image-2",
				"codex-auto-review",
			},
		},
		"pro": {
			models: GetCodexProModels(),
			want: []string{
				"gpt-5.2",
				"gpt-5.3-codex",
				"gpt-5.3-codex-spark",
				"gpt-5.4",
				"gpt-5.4-mini",
				"gpt-5.5",
				"gpt-image-2",
				"codex-auto-review",
			},
		},
	}

	for tier, tc := range tierModels {
		t.Run(tier, func(t *testing.T) {
			assertModelIDs(t, tier, tc.models, tc.want)
		})
	}

	if model := findModelInfo(tierModels["free"].models, "gpt-5.5"); model != nil {
		t.Fatalf("free tier unexpectedly includes gpt-5.5")
	}
}

func TestCodexStaticModelsIgnoreRemoteCatalogChanges(t *testing.T) {
	original := swapModelsCatalogForTest(t, func(data staticModelsJSON) staticModelsJSON {
		data.CodexFree = []*ModelInfo{{ID: "remote-only-free"}}
		data.CodexTeam = []*ModelInfo{{ID: "remote-only-team"}}
		data.CodexPlus = []*ModelInfo{{ID: "remote-only-plus"}}
		data.CodexPro = []*ModelInfo{{ID: "remote-only-pro"}}
		return data
	})
	defer restoreModelsCatalog(original)

	models := GetCodexProModels()
	if findModelInfo(models, "remote-only-pro") != nil {
		t.Fatal("fixed codex pro catalog included remote-only-pro")
	}
	if findModelInfo(models, "gpt-5.4-mini") == nil {
		t.Fatal("fixed codex pro catalog missing gpt-5.4-mini")
	}
	if findModelInfo(models, "codex-auto-review") == nil {
		t.Fatal("fixed codex pro catalog missing codex-auto-review")
	}
}

func TestDetectChangedProvidersIgnoresCodexCatalogChanges(t *testing.T) {
	oldData := getModels()
	newData := *oldData
	newData.CodexFree = []*ModelInfo{{ID: "remote-only-free"}}
	newData.CodexTeam = []*ModelInfo{{ID: "remote-only-team"}}
	newData.CodexPlus = []*ModelInfo{{ID: "remote-only-plus"}}
	newData.CodexPro = []*ModelInfo{{ID: "remote-only-pro"}}

	changed := detectChangedProviders(oldData, &newData)
	if len(changed) != 0 {
		t.Fatalf("expected codex-only catalog changes to be ignored, got %v", changed)
	}
}

func TestCodexStaticLookupIncludesFixedBuiltins(t *testing.T) {
	model := LookupStaticModelInfo("gpt-5.5")
	if model == nil {
		t.Fatal("expected LookupStaticModelInfo to find gpt-5.5")
	}
	assertGPT55ModelInfo(t, "lookup", model)

	autoReview := LookupStaticModelInfo("codex-auto-review")
	if autoReview == nil {
		t.Fatal("expected LookupStaticModelInfo to find codex-auto-review")
	}
	if autoReview.ID != "codex-auto-review" {
		t.Fatalf("auto-review id mismatch: got %q", autoReview.ID)
	}
}

func TestCodexStaticChannelReturnsFixedProSet(t *testing.T) {
	models := GetStaticModelDefinitionsByChannel("codex")
	if findModelInfo(models, "gpt-5.4-mini") == nil {
		t.Fatal("codex static channel missing gpt-5.4-mini")
	}
	if findModelInfo(models, "codex-auto-review") == nil {
		t.Fatal("codex static channel missing codex-auto-review")
	}
	if findModelInfo(models, "gpt-5.3-codex-spark") == nil {
		t.Fatal("codex static channel missing gpt-5.3-codex-spark")
	}
}

func swapModelsCatalogForTest(t *testing.T, mutate func(staticModelsJSON) staticModelsJSON) *staticModelsJSON {
	t.Helper()

	modelsCatalogStore.mu.Lock()
	defer modelsCatalogStore.mu.Unlock()

	original := modelsCatalogStore.data
	if original == nil {
		t.Fatal("models catalog is nil")
	}
	mutated := mutate(*original)
	modelsCatalogStore.data = &mutated
	return original
}

func restoreModelsCatalog(original *staticModelsJSON) {
	modelsCatalogStore.mu.Lock()
	defer modelsCatalogStore.mu.Unlock()
	modelsCatalogStore.data = original
}

func assertModelIDs(t *testing.T, source string, models []*ModelInfo, want []string) {
	t.Helper()

	if len(models) != len(want) {
		t.Fatalf("%s model count mismatch: got %d, want %d (%v)", source, len(models), len(want), modelIDs(models))
	}
	for i, wantID := range want {
		if models[i] == nil {
			t.Fatalf("%s model %d is nil", source, i)
		}
		if models[i].ID != wantID {
			t.Fatalf("%s model %d mismatch: got %q, want %q", source, i, models[i].ID, wantID)
		}
	}
}

func modelIDs(models []*ModelInfo) []string {
	ids := make([]string, 0, len(models))
	for _, model := range models {
		if model == nil {
			ids = append(ids, "")
			continue
		}
		ids = append(ids, model.ID)
	}
	return ids
}

func findModelInfo(models []*ModelInfo, id string) *ModelInfo {
	for _, model := range models {
		if model != nil && model.ID == id {
			return model
		}
	}
	return nil
}

func assertGPT55ModelInfo(t *testing.T, source string, model *ModelInfo) {
	t.Helper()

	if model.ID != "gpt-5.5" {
		t.Fatalf("%s id mismatch: got %q", source, model.ID)
	}
	if model.Object != "model" {
		t.Fatalf("%s object mismatch: got %q", source, model.Object)
	}
	if model.Created != 1776902400 {
		t.Fatalf("%s created timestamp mismatch: got %d", source, model.Created)
	}
	if model.OwnedBy != "openai" {
		t.Fatalf("%s owned_by mismatch: got %q", source, model.OwnedBy)
	}
	if model.Type != "openai" {
		t.Fatalf("%s type mismatch: got %q", source, model.Type)
	}
	if model.DisplayName != "GPT 5.5" {
		t.Fatalf("%s display name mismatch: got %q", source, model.DisplayName)
	}
	if model.Version != "gpt-5.5" {
		t.Fatalf("%s version mismatch: got %q", source, model.Version)
	}
	if model.Description != "Frontier model for complex coding, research, and real-world work." {
		t.Fatalf("%s description mismatch: got %q", source, model.Description)
	}
	if model.ContextLength != 272000 {
		t.Fatalf("%s context length mismatch: got %d", source, model.ContextLength)
	}
	if model.MaxCompletionTokens != 128000 {
		t.Fatalf("%s max completion tokens mismatch: got %d", source, model.MaxCompletionTokens)
	}
	if len(model.SupportedParameters) != 1 || model.SupportedParameters[0] != "tools" {
		t.Fatalf("%s supported parameters mismatch: got %v", source, model.SupportedParameters)
	}
	if model.Thinking == nil {
		t.Fatalf("%s missing thinking support", source)
	}

	want := []string{"low", "medium", "high", "xhigh"}
	if len(model.Thinking.Levels) != len(want) {
		t.Fatalf("%s thinking level count mismatch: got %d, want %d", source, len(model.Thinking.Levels), len(want))
	}
	for i, level := range want {
		if model.Thinking.Levels[i] != level {
			t.Fatalf("%s thinking level %d mismatch: got %q, want %q", source, i, model.Thinking.Levels[i], level)
		}
	}
}
