package main

import (
	"context"
	"errors"
	"os"

	"github.com/alecthomas/kong"

	"github.com/shuymn/pd/internal/cli"
	"github.com/shuymn/pd/internal/log"
)

func main() {
	ctx := context.Background()

	stdout := os.Stdout
	stderr := os.Stderr

	root := &cli.Root{
		OutputLogger:     log.NewOutputLogger(stdout),
		DiagnosticLogger: log.NewDiagnosticLogger(stderr),
	}

	parser := kong.Must(root,
		kong.Description("Progressive discovery of docs."),
		kong.BindTo(ctx, (*context.Context)(nil)),
	)

	// kong has no built-in "show help when no args" option; inject --help explicitly.
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"--help"}
	}

	kctx, parseErr := parser.Parse(args)
	parser.FatalIfErrorf(parseErr)

	err := kctx.Run()
	if errors.Is(err, cli.ErrDiagnostics) {
		os.Exit(1)
	}

	parser.FatalIfErrorf(err)
}
