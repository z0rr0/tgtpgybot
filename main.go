package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/z0rr0/tgtpgybot/bot"
	"github.com/z0rr0/tgtpgybot/config"
)

// Name is a bot name.
const Name = "TgTGPYBot"

var (
	// Version is git version
	Version = ""
	// Revision is revision number
	Revision = ""
	// BuildDate is build date
	BuildDate = ""
	// GoVersion is runtime Go language version
	GoVersion = runtime.Version()
)

func main() {
	var configFile = "config.json"

	defer func() {
		if r := recover(); r != nil {
			slog.Error("abnormal termination", "version", Version, "error", r, "stack", string(debug.Stack()))
		}
	}()

	version := flag.Bool("version", false, "show version")
	flag.StringVar(&configFile, "config", configFile, "configuration file")
	flag.Parse()

	versionInfo := fmt.Sprintf("%v: %v %v %v %v", Name, Version, Revision, GoVersion, BuildDate)
	if *version {
		fmt.Println(versionInfo)
		flag.PrintDefaults()
		return
	}

	cfg, err := config.New(configFile)
	if err != nil {
		panic(err)
	}

	slog.Info(
		"main", "logging", cfg.DebugLevel,
		"version", Version, "revision", Revision, "build_date", BuildDate, "go_version", GoVersion,
	)
	slog.Info("read config", "config", cfg)

	b, err := bot.New(cfg)
	if err != nil {
		panic(err)
	}

	sigChan := make(chan os.Signal, 1)
	b.Start(sigChan)
	b.Stop()

	slog.Info("stopped")
}
