package cliproxy

import (
	"testing"

	internalconfig "github.com/router-for-me/CLIProxyAPI/v6/internal/config"
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

func cliproxyModelsContainID(models []*ModelInfo, id string) bool {
	for _, model := range models {
		if model != nil && model.ID == id {
			return true
		}
	}
	return false
}
