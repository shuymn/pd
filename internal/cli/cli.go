package cli

import (
	"errors"
	"log/slog"
)

// ErrDiagnostics requests exit code 1 after a command emitted diagnostics.
var ErrDiagnostics = errors.New("command emitted diagnostics")

// Root is the kong root struct for the pd CLI.
type Root struct {
	OutputLogger     *slog.Logger `kong:"-"`
	DiagnosticLogger *slog.Logger `kong:"-"`

	Root    string  `help:"Directory to scan, relative to the current directory." default:"." name:"root"`
	Verbose bool    `help:"Emit list diagnostics to stderr." name:"verbose"`
	List    ListCmd `cmd:"" help:"List discovery metadata from docs directory."`
	Show    ShowCmd `cmd:"" help:"Show discovery metadata for a single document."`
}
