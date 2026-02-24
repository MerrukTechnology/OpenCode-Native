package app

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/MerrukTechnology/OpenCode-Native/internal/config"
	"github.com/MerrukTechnology/OpenCode-Native/internal/logging"
	"github.com/MerrukTechnology/OpenCode-Native/internal/lsp"
	"github.com/MerrukTechnology/OpenCode-Native/internal/lsp/install"
	"github.com/MerrukTechnology/OpenCode-Native/internal/lsp/protocol"
	"github.com/MerrukTechnology/OpenCode-Native/internal/lsp/watcher"
)

type serverNameContextKey string

const ServerNameContextKey serverNameContextKey = "server_name"

type lspService struct {
	clients   map[string]*lsp.Client
	clientsCh chan *lsp.Client
	mu        sync.RWMutex

	watcherCancelFuncs []context.CancelFunc
	cancelMu           sync.Mutex
	watcherWG          sync.WaitGroup
}

func NewLspService() lsp.LspService {
	return &lspService{
		clients:   make(map[string]*lsp.Client),
		clientsCh: make(chan *lsp.Client, 50),
	}
}

func (s *lspService) Init(ctx context.Context) error {
	cfg := config.Get()
	wg := sync.WaitGroup{}
	for name, server := range install.ResolveServers(cfg) {
		wg.Add(1)
		go func(nm string, srv install.ResolvedServer) {
			lspName := "LSP-" + nm
			defer logging.RecoverPanic(lspName, func() {
				logging.ErrorPersist(fmt.Sprintf("Panic while starting %s", lspName))
			})
			defer wg.Done()
			s.startLSPServer(ctx, nm, srv)
		}(name, server)
	}
	go func() {
		wg.Wait()
		logging.Info("LSP clients initialization completed")
		close(s.clientsCh)
	}()
	logging.Info("LSP clients initialization started in background")

	// Note: We intentionally do not return an error if an LSP server fails to start during initialization.
	// This is because we want the app to be resilient to LSP startup failures and to allow the app to start
	// even if some LSP servers fail to start. The LSP clients will continue to run in the background and
	// will be restarted automatically if they crash, so we can afford to be lenient during initialization.
	// TODO: handle this return value in the app startup logic to show a warning if some LSP servers failed to start.
	// we need to return an error here if the initialization fails.
	return nil
}

func (s *lspService) Shutdown(ctx context.Context) error {
	s.cancelMu.Lock()
	for _, cancel := range s.watcherCancelFuncs {
		cancel()
	}
	s.cancelMu.Unlock()
	s.watcherWG.Wait()

	s.mu.RLock()
	clients := make(map[string]*lsp.Client, len(s.clients))
	maps.Copy(clients, s.clients)
	s.mu.RUnlock()

	for name, client := range clients {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := client.Shutdown(shutdownCtx); err != nil {
			logging.Error("Failed to shutdown LSP client", "name", name, "error", err)
		}
		cancel()
	}
	// TODO: Add timeout for shutdown and force close if it exceeds the timeout to prevent hanging on shutdown.
	// handle this return value in the app shutdown logic to force close if it exceeds the timeout.
	// we need to return an error here if the shutdown exceeds the timeout.
	return nil
}

func (s *lspService) ForceShutdown() {
	s.cancelMu.Lock()
	for _, cancel := range s.watcherCancelFuncs {
		cancel()
	}
	s.cancelMu.Unlock()

	s.mu.RLock()
	clients := make(map[string]*lsp.Client, len(s.clients))
	maps.Copy(clients, s.clients)
	s.mu.RUnlock()

	for name, client := range clients {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		if err := client.Shutdown(shutdownCtx); err != nil {
			logging.Debug("Failed to gracefully shutdown LSP client, forcing close", "name", name, "error", err)
			client.Close()
		}
		cancel()
	}
}

func (s *lspService) Clients() map[string]*lsp.Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snapshot := make(map[string]*lsp.Client, len(s.clients))
	maps.Copy(snapshot, s.clients)
	return snapshot
}

func (s *lspService) ClientsCh() <-chan *lsp.Client {
	return s.clientsCh
}

func (s *lspService) ClientsForFile(filePath string) []*lsp.Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ext := strings.ToLower(filepath.Ext(filePath))
	var matched []*lsp.Client
	for _, client := range s.clients {
		if slices.Contains(client.GetExtensions(), ext) {
			matched = append(matched, client)
		}
	}
	return matched
}

func (s *lspService) NotifyOpenFile(ctx context.Context, filePath string) {
	for _, client := range s.Clients() {
		_ = client.OpenFile(ctx, filePath)
	}
}

func (s *lspService) WaitForDiagnostics(ctx context.Context, filePath string) error {
	clients := s.Clients()
	if len(clients) == 0 {
		// TODO: This is a temporary workaround to avoid waiting for diagnostics when no LSP clients are available.
		// In the future, we should refactor this to only wait for diagnostics from relevant clients, and to not wait at all if there are no relevant clients.
		logging.Debug("No LSP clients available to wait for diagnostics")
		return nil
	}

	diagChan := make(chan struct{}, 1)

	for _, client := range clients {
		originalDiags := make(map[protocol.DocumentUri][]protocol.Diagnostic)
		newDiags := client.GetDiagnostics()

		for k, v := range newDiags {
			originalDiags[k] = v
		}

		handler := func(params json.RawMessage) {
			lsp.HandleDiagnostics(client, params)
			var diagParams protocol.PublishDiagnosticsParams
			if err := json.Unmarshal(params, &diagParams); err != nil {
				return
			}

			if diagParams.URI.Path() == filePath || lsp.HasDiagnosticsChanged(client.GetDiagnostics(), originalDiags) {
				select {
				case diagChan <- struct{}{}:
				default:
				}
			}
		}

		client.RegisterNotificationHandler("textDocument/publishDiagnostics", handler)

		if client.IsFileOpen(filePath) {
			_ = client.NotifyChange(ctx, filePath)
		} else {
			_ = client.OpenFile(ctx, filePath)
		}
	}

	select {
	case <-diagChan:
	case <-time.After(5 * time.Second):
	case <-ctx.Done():
	}
	return ctx.Err()
}

func (s *lspService) FormatDiagnostics(filePath string) string {
	clients := s.Clients()
	return lsp.FormatDiagnostics(filePath, clients)
}

// hasMatchingFiles checks whether the working directory contains any files
// with extensions handled by the given server. It does a shallow walk
// (max 3 levels deep) to keep startup fast.
func hasMatchingFiles(rootDir string, extensions []string) bool {
	if len(extensions) == 0 {
		return true
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
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "dist" || name == "build" || name == "target" {
				return filepath.SkipDir
			}
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

func (s *lspService) startLSPServer(ctx context.Context, name string, server install.ResolvedServer) {
	cfg := config.Get()

	if !hasMatchingFiles(config.WorkingDirectory(), server.Extensions) {
		logging.Debug("No matching files found, skipping LSP server", "name", name, "extensions", server.Extensions)
		return
	}

	command, args, err := install.ResolveCommand(ctx, server, cfg.DisableLSPDownload)
	if err != nil {
		logging.Debug("LSP server not available, skipping", "name", name, "reason", err)
		return
	}

	s.createAndStartLSPClient(ctx, name, server, command, args...)
}

func (s *lspService) createAndStartLSPClient(ctx context.Context, name string, server install.ResolvedServer, command string, args ...string) {
	logging.Info("Creating LSP client", "name", name, "command", command, "args", args)

	lspClient, err := lsp.NewClient(ctx, command, server.Env, args...)
	if err != nil {
		logging.Error("Failed to create LSP client for", name, err)
		return
	}

	lspClient.SetExtensions(server.Extensions)

	initCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

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

	watchCtx, cancelFunc := context.WithCancel(ctx)
	watchCtx = context.WithValue(watchCtx, ServerNameContextKey, name)

	workspaceWatcher := watcher.NewWorkspaceWatcher(lspClient)

	s.cancelMu.Lock()
	s.watcherCancelFuncs = append(s.watcherCancelFuncs, cancelFunc)
	s.cancelMu.Unlock()

	s.watcherWG.Add(1)

	s.mu.Lock()
	s.clients[name] = lspClient
	s.clientsCh <- lspClient
	s.mu.Unlock()

	go s.runWorkspaceWatcher(watchCtx, name, workspaceWatcher)
}

func (s *lspService) runWorkspaceWatcher(ctx context.Context, name string, workspaceWatcher *watcher.WorkspaceWatcher) {
	defer s.watcherWG.Done()
	defer logging.RecoverPanic("LSP-"+name, func() {
		s.restartLSPClient(ctx, name)
	})

	workspaceWatcher.WatchWorkspace(ctx, config.WorkingDirectory())
	logging.Info("Workspace watcher stopped", "client", name)
}

// Non-blocking, thread-safe, and designed to prevent zombie processes.
func (s *lspService) restartLSPClient(ctx context.Context, name string) {
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

		s.mu.Lock()
		oldClient, exists := s.clients[name]
		if exists {
			delete(s.clients, name)
		}
		s.mu.Unlock()

		// 3. Cleanup the old process safely
		if exists && oldClient != nil {
			// Go 1.21+ feature: context.WithoutCancel(ctx)
			// This keeps the context values (tracing/logs) but IGNORES cancellation.
			// This ensures the cleanup finishes even if the user switches tabs.
			detachedCtx := context.WithoutCancel(ctx)

			// Give the old server 2 seconds to die gracefully
			shutdownCtx, cancel := context.WithTimeout(detachedCtx, 2*time.Second)
			defer cancel()

			// Try graceful shutdown first
			if err := oldClient.Shutdown(shutdownCtx); err != nil {
				logging.Warn("Graceful shutdown failed, force closing to save RAM", "client", name)

				// 4. Fallback: force close the client to free resources
				if kErr := oldClient.Close(); kErr != nil {
					logging.Error("Failed to force close LSP client", "error", kErr)
				}
			}
		}

		// 4. Start the new server
		s.startLSPServer(ctx, name, server)
		logging.Info("Successfully restarted LSP client", "client", name)
	}()
}
