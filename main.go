package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"

	"github.com/shuymn/pd/internal/cli"
)

func main() {
	ctx := context.Background()

	parser := kong.Must(new(cli.Root),
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

	runErr := kctx.Run()
	parser.FatalIfErrorf(runErr)
}
