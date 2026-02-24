package lsp

import "context"

type LspService interface {
	Init(ctx context.Context) error
	Shutdown(ctx context.Context) error
	ForceShutdown()

	Clients() map[string]*Client
	ClientsCh() <-chan *Client
	ClientsForFile(filePath string) []*Client

	NotifyOpenFile(ctx context.Context, filePath string)
	WaitForDiagnostics(ctx context.Context, filePath string) error
	FormatDiagnostics(filePath string) string
}
