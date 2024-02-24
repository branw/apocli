package main

import (
	"github.com/alecthomas/kong"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"go-apo/pkg/apocli"
	"log/slog"
	"os"
	"time"
)

type Context struct {
	Config *apocli.Config
}

var cli struct {
	Config struct {
		Path     ConfigPathCmd     `cmd:"" help:"Print path to config file"`
		List     ConfigListCmd     `cmd:"" help:"List all config items"`
		Get      ConfigGetCmd      `cmd:"" help:"Get the value of a config item"`
		Set      ConfigSetCmd      `cmd:"" help:"Set a config item value"`
		Clear    ConfigClearCmd    `cmd:"" help:"Reset a single config item to its default"`
		ClearAll ConfigClearAllCmd `cmd:"" help:"Reset all config values to the defaults"`
	} `cmd:"" help:"Manage configuration"`
}

func run() error {
	w := os.Stdout
	opts := &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.DateTime,
		NoColor:    !isatty.IsTerminal(w.Fd()),
	}
	logger := slog.New(tint.NewHandler(w, opts))
	slog.SetDefault(logger)

	config, err := apocli.LoadConfig()
	if err != nil {
		return err
	}

	ctx := kong.Parse(&cli)
	return ctx.Run(&Context{Config: config})
}

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", slog.Any("err", err))
	}
}
