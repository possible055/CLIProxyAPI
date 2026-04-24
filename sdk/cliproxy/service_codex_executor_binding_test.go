package cliproxy

import (
	"context"
	"testing"

	internalconfig "github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v6/sdk/config"
)

func TestEnsureExecutorsForAuth_CodexDoesNotReplaceInNormalMode(t *testing.T) {
	service := &Service{
		cfg:         &config.Config{},
		coreManager: coreauth.NewManager(nil, nil, nil),
	}
	auth := &coreauth.Auth{
		ID:       "codex-auth-1",
		Provider: "codex",
		Status:   coreauth.StatusActive,
	}

	service.ensureExecutorsForAuth(auth)
	firstExecutor, okFirst := service.coreManager.Executor("codex")
	if !okFirst || firstExecutor == nil {
		t.Fatal("expected codex executor after first bind")
	}

	service.ensureExecutorsForAuth(auth)
	secondExecutor, okSecond := service.coreManager.Executor("codex")
	if !okSecond || secondExecutor == nil {
		t.Fatal("expected codex executor after second bind")
	}

	if firstExecutor != secondExecutor {
		t.Fatal("expected codex executor to stay unchanged in normal mode")
	}
}

func TestEnsureExecutorsForAuthWithMode_CodexForceReplace(t *testing.T) {
	service := &Service{
		cfg:         &config.Config{},
		coreManager: coreauth.NewManager(nil, nil, nil),
	}
	auth := &coreauth.Auth{
		ID:       "codex-auth-2",
		Provider: "codex",
		Status:   coreauth.StatusActive,
	}

	service.ensureExecutorsForAuth(auth)
	firstExecutor, okFirst := service.coreManager.Executor("codex")
	if !okFirst || firstExecutor == nil {
		t.Fatal("expected codex executor after first bind")
	}

	service.ensureExecutorsForAuthWithMode(auth, true)
	secondExecutor, okSecond := service.coreManager.Executor("codex")
	if !okSecond || secondExecutor == nil {
		t.Fatal("expected codex executor after forced rebind")
	}

	if firstExecutor == secondExecutor {
		t.Fatal("expected codex executor replacement in force mode")
	}
}

func TestBuildCodexConfigModelsAppendsBuiltins(t *testing.T) {
	models := buildCodexConfigModels(&internalconfig.CodexKey{
		Models: []internalconfig.CodexModel{{Name: "custom-codex-model"}},
	})

	for _, id := range []string{"custom-codex-model", "gpt-image-2", "codex-auto-review"} {
		if !cliproxyModelsContainID(models, id) {
			t.Fatalf("expected codex config models to include %q, got %+v", id, models)
		}
	}
}

func TestRegisterLoadedAuthModelsRegistersCodexOAuthCatalog(t *testing.T) {
	authID := "loaded-codex-oauth-auth"
	reg := registry.GetGlobalRegistry()
	reg.UnregisterClient(authID)
	t.Cleanup(func() {
		reg.UnregisterClient(authID)
	})

	service := &Service{
		cfg:         &config.Config{},
		coreManager: coreauth.NewManager(nil, nil, nil),
	}
	if _, err := service.coreManager.Register(context.Background(), &coreauth.Auth{
		ID:       authID,
		Provider: "codex",
		Status:   coreauth.StatusActive,
		Attributes: map[string]string{
			"auth_kind": "oauth",
			"plan_type": "team",
		},
	}); err != nil {
		t.Fatalf("register auth: %v", err)
	}

	service.registerLoadedAuthModels(context.Background())

	models := reg.GetModelsForClient(authID)
	if !cliproxyModelsContainID(models, "gpt-5.4-mini") {
		t.Fatalf("expected loaded Codex OAuth auth to register gpt-5.4-mini, got %+v", models)
	}
	if !cliproxyModelsContainID(models, "codex-auto-review") {
		t.Fatalf("expected loaded Codex OAuth auth to register codex-auto-review, got %+v", models)
	}
	if !stringSliceContains(reg.GetModelProviders("gpt-5.4-mini"), "codex") {
		t.Fatalf("expected gpt-5.4-mini provider to include codex")
	}
}

func cliproxyModelsContainID(models []*ModelInfo, id string) bool {
	for _, model := range models {
		if model != nil && model.ID == id {
			return true
		}
	}
	return false
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
