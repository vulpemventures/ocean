package db_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
	dbbadger "github.com/vulpemventures/ocean/internal/infrastructure/storage/db/badger"
	"github.com/vulpemventures/ocean/internal/infrastructure/storage/db/inmemory"
)

func TestExternalScriptRepository(t *testing.T) {
	repositories, err := newExternalScriptRepositories(
		func(repoType string) ports.ScriptEventHandler {
			return func(event domain.ExternalScriptEvent) {
				fmt.Printf("received event from %s repo: %+v\n", repoType, event)
			}
		},
	)
	require.NoError(t, err)

	for name, repo := range repositories {
		t.Run(name, func(t *testing.T) {
			testExternalScriptRepository(t, repo)
		})
	}
}

func testExternalScriptRepository(t *testing.T, repo domain.ExternalScriptRepository) {
	accountName := "test1"
	newScript := domain.AddressInfo{
		Account:     accountName,
		Script:      hex.EncodeToString(randomScript()),
		BlindingKey: randomBytes(32),
	}

	t.Run("add_script", func(t *testing.T) {
		done, err := repo.AddScript(ctx, newScript)
		require.NoError(t, err)
		require.True(t, done)

		done, err = repo.AddScript(ctx, newScript)
		require.NoError(t, err)
		require.False(t, done)
	})

	t.Run("get_all_scripts", func(t *testing.T) {
		scripts, err := repo.GetAllScripts(ctx)
		require.NoError(t, err)
		require.Len(t, scripts, 1)
	})

	t.Run("delete_script", func(t *testing.T) {
		done, err := repo.DeleteScript(ctx, newScript.Account)
		require.NoError(t, err)
		require.True(t, done)

		done, err = repo.DeleteScript(ctx, newScript.Account)
		require.NoError(t, err)
		require.False(t, done)

		scripts, err := repo.GetAllScripts(ctx)
		require.NoError(t, err)
		require.Empty(t, scripts)
	})
}

func newExternalScriptRepositories(
	handlerFactory func(repoType string) ports.ScriptEventHandler,
) (map[string]domain.ExternalScriptRepository, error) {
	inmemoryRepoManager := inmemory.NewRepoManager()
	badgerRepoManager, err := dbbadger.NewRepoManager("", nil)
	if err != nil {
		return nil, err
	}

	handlers := []ports.ScriptEventHandler{
		handlerFactory("badger"), handlerFactory("inmemory"),
	}

	repoManagers := []ports.RepoManager{badgerRepoManager, inmemoryRepoManager, pgRepoManager}

	for i, handler := range handlers {
		repoManager := repoManagers[i]
		repoManager.RegisterHandlerForExternalScriptEvent(domain.ExternalScriptAdded, handler)
		repoManager.RegisterHandlerForExternalScriptEvent(domain.ExternalScriptDeleted, handler)
	}

	return map[string]domain.ExternalScriptRepository{
		"inmemory": inmemoryRepoManager.ExternalScriptRepository(),
		"badger":   badgerRepoManager.ExternalScriptRepository(),
		"postgres": pgRepoManager.ExternalScriptRepository(),
	}, nil
}
