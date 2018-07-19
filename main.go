package main

import (
	"fmt"
	"os"

	"github.com/mkideal/cli"
)

func main() {
	if err := cli.Root(root,
		cli.Tree(help),
		cli.Tree(downloader),
		cli.Tree(list),
		cli.Tree(uploader),
	).Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var help = cli.HelpCommand("show help")

// main command
var root = &cli.Command{
	Fn: func(ctx *cli.Context) error {
		ctx.String(ctx.Command().Usage(ctx))
		return nil
	},
}
