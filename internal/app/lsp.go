package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MerrukTechnology/OpenCode-Native/internal/config"
	"github.com/MerrukTechnology/OpenCode-Native/internal/logging"
	"github.com/MerrukTechnology/OpenCode-Native/internal/lsp"
	"github.com/MerrukTechnology/OpenCode-Native/internal/lsp/install"
	"github.com/MerrukTechnology/OpenCode-Native/internal/lsp/watcher"
)

type contextKey string

func (app *App) initLSPClients(ctx context.Context) {
	cfg := config.Get()

	// Resolve which servers to start: merge built-in registry with user config
	servers := install.ResolveServers(cfg)

	for name, server := range servers {
		go app.startLSPServer(ctx, name, server)
	}

	logging.Info("LSP clients initialization started in background")
}

// hasMatchingFiles checks whether the working directory contains any files
// with extensions handled by the given server. It does a shallow walk
// (max 3 levels deep) to keep startup fast.
func hasMatchingFiles(rootDir string, extensions []string) bool {
	if len(extensions) == 0 {
		return true // no extensions specified, assume relevant
	}

	extSet := make(map[string]struct{}, len(extensions))
	for _, ext := range extensions {
		extSet[ext] = struct{}{}
	}

	found := false
	_ = filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return filepath.SkipDir
		}

		if d.IsDir() {
			// Skip hidden dirs and common non-source dirs
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "dist" || name == "build" || name == "target" {
				return filepath.SkipDir
			}
			// Limit depth to 3 levels
			rel, _ := filepath.Rel(rootDir, path)
			if strings.Count(rel, string(filepath.Separator)) >= 3 {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if _, ok := extSet[ext]; ok {
			found = true
			return filepath.SkipAll
		}
		return nil
	})

	return found
}

// startLSPServer resolves the binary (auto-installing if needed), then creates and starts the LSP client
func (app *App) startLSPServer(ctx context.Context, name string, server install.ResolvedServer) {
	cfg := config.Get()

	// Skip servers whose file extensions aren't present in the project
	if !hasMatchingFiles(config.WorkingDirectory(), server.Extensions) {
		logging.Debug("No matching files found, skipping LSP server", "name", name, "extensions", server.Extensions)
		return
	}

	// Resolve the command — check PATH, bin dir, or auto-install
	command, args, err := install.ResolveCommand(ctx, server, cfg.DisableLSPDownload)
	if err != nil {
		logging.Debug("LSP server not available, skipping", "name", name, "reason", err)
		return
	}

	app.createAndStartLSPClient(ctx, name, server, command, args...)
}

// createAndStartLSPClient creates a new LSP client, initializes it, and starts its workspace watcher
func (app *App) createAndStartLSPClient(ctx context.Context, name string, server install.ResolvedServer, command string, args ...string) {
	logging.Info("Creating LSP client", "name", name, "command", command, "args", args)

	// Build environment for the server
	lspClient, err := lsp.NewClient(ctx, command, server.Env, args...)
	if err != nil {
		logging.Error("Failed to create LSP client for", name, err)
		return
	}

	// Store extensions on the client for routing
	lspClient.SetExtensions(server.Extensions)

	initCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Pass server-specific initialization options
	var initOpts map[string]any
	if m, ok := server.Initialization.(map[string]any); ok {
		initOpts = m
	}

	_, err = lspClient.InitializeLSPClient(initCtx, config.WorkingDirectory(), initOpts)
	if err != nil {
		logging.Error("Initialize failed", "name", name, "error", err)
		lspClient.Close()
		return
	}

	if err := lspClient.WaitForServerReady(initCtx); err != nil {
		logging.Error("Server failed to become ready", "name", name, "error", err)
		lspClient.SetServerState(lsp.StateError)
	} else {
		logging.Info("LSP server is ready", "name", name)
		lspClient.SetServerState(lsp.StateReady)
	}

	logging.Info("LSP client initialized", "name", name)

	watchCtx, cancelFunc := context.WithCancel(ctx)
	watchCtx = context.WithValue(watchCtx, contextKey("serverName"), name)

	workspaceWatcher := watcher.NewWorkspaceWatcher(lspClient)

	app.cancelFuncsMutex.Lock()
	app.watcherCancelFuncs = append(app.watcherCancelFuncs, cancelFunc)
	app.cancelFuncsMutex.Unlock()

	app.watcherWG.Add(1)

	app.clientsMutex.Lock()
	app.LSPClients[name] = lspClient
	app.clientsMutex.Unlock()

	go app.runWorkspaceWatcher(watchCtx, name, workspaceWatcher)
}

// runWorkspaceWatcher executes the workspace watcher for an LSP client
func (app *App) runWorkspaceWatcher(ctx context.Context, name string, workspaceWatcher *watcher.WorkspaceWatcher) {
	defer app.watcherWG.Done()
	defer logging.RecoverPanic("LSP-"+name, func() {
		app.restartLSPClient(ctx, name)
	})

	workspaceWatcher.WatchWorkspace(ctx, config.WorkingDirectory())
	logging.Info("Workspace watcher stopped", "client", name)
}

// restartLSPClient attempts to restart a crashed or failed LSP client
// Optimized for macOS Catalina: Non-blocking, thread-safe, and prevents Zombie processes.
func (app *App) restartLSPClient(ctx context.Context, name string) {
	// 1. Get config immediately (fast)
	cfg := config.Get()

	// 2. Run the restart logic in a background goroutine.
	// This prevents the UI from freezing while we wait for the old process to die.
	go func() {
		logging.Info("Initiating LSP restart sequence", "client", name)

		// Resolve expensive paths inside the background thread to avoid I/O lag
		servers := install.ResolveServers(cfg)
		server, exists := servers[name]
		if !exists {
			logging.Error("Cannot restart client, configuration not found", "client", name)
			return
		}

		// Lock only for the map modification
		app.clientsMutex.Lock()
		oldClient, exists := app.LSPClients[name]
		if exists {
			// Remove it immediately so the UI knows it's "restarting/down"
			delete(app.LSPClients, name)
		}
		app.clientsMutex.Unlock()

		// 3. Cleanup the old process safely
		if exists && oldClient != nil {
			// GO 1.24 MAGIC: context.WithoutCancel(ctx)
			// This keeps the context values (tracing/logs) but IGNORES cancellation.
			// This ensures the cleanup finishes even if the user switches tabs.
			detachedCtx := context.WithoutCancel(ctx)

			// Give the old server 2 seconds to die gracefully
			shutdownCtx, cancel := context.WithTimeout(detachedCtx, 2*time.Second)
			defer cancel()

			// Try graceful shutdown first
			if err := oldClient.Shutdown(shutdownCtx); err != nil {
				logging.Warn("Graceful shutdown failed, force closing to save RAM", "client", name)

				// 4. THE FIX: Call your new ForceClose method
				if kErr := oldClient.Close(); kErr != nil {
					logging.Error("Failed to force close LSP client", "error", kErr)
				}
			}
		}

		// 4. Start the new server
		app.startLSPServer(ctx, name, server)
		logging.Info("Successfully restarted LSP client", "client", name)
	}()
}
