package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		Chart,
		Get,
	}
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
