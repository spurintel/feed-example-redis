package main

import (
	"context"
	"feedexampleredis/internal/app"
	"feedexampleredis/internal/commands"
	"feedexampleredis/internal/server"
	"feedexampleredis/internal/storage"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

var (
	// Binary info
	Version string
	Commit  string
	Date    string

	// Flags
	file string
	api  bool

	// Args
	command string
)

func init() {
	app.InitLogging(Version, Commit, Date)
}

// main - the main function, starts the process and listens for signals to stop. Use a errgroup to manage the goroutines.
func main() {
	cfg, err := app.ParseConfigFromEnvironment()
	if err != nil {
		fmt.Println("error parsing config:", err)
	}

	flag.StringVar(&file, "file", "", "path to the feed file or realtime file to process")
	flag.BoolVar(&api, "api", false, "start the API server")
	flag.Parse()

	// Get the command from the args
	args := flag.Args()
	if len(args) > 0 {
		command = args[0]
	} else {
		fmt.Fprintf(os.Stderr, "error: no command specified, it must be one of: daemon, insert, merge\n")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	// Start the signal handler
	g.Go(func() error {
		return signalHandler(ctx)
	})

	// Setup redis client, it is used in multiple commands
	ttl := time.Duration(cfg.TTL) * time.Hour
	redisClient := storage.NewRedis(cfg.RedisAddr, cfg.RedisPass, cfg.RedisDB, ttl, cfg.ConcurrentNum, cfg.ChunkSize)
	redisClient.Connect()
	slog.Info(
		"redis client created",
		slog.String("redis_addr", cfg.RedisAddr),
		slog.String("redis_db", strconv.Itoa(cfg.RedisDB)),
	)

	// Start the main process
	switch command {
	case "daemon":
		// Start the API server if the flag is set
		if api {
			g.Go(func() error {
				defer cancel()
				api := server.NewServer(cfg, redisClient)
				if cfg.CertFile != "" && cfg.KeyFile != "" {
					return api.StartTLS(ctx)
				}
				return api.Start(ctx)
			})
		}
		g.Go(func() error {
			defer cancel()
			return commands.Daemon(ctx, cfg, redisClient)
		})
	case "insert":
		// TODO
		if file == "" {
			fmt.Fprintf(os.Stderr, "error: no file specified\n")
			os.Exit(1)
		}
		g.Go(func() error {
			defer cancel()
			return commands.InsertFeedFile(ctx, file, redisClient)
		})
	case "merge":
		// TODO
		if file == "" {
			fmt.Fprintf(os.Stderr, "error: no file specified\n")
			os.Exit(1)
		}
		g.Go(func() error {
			defer cancel()
			return commands.MergeRealtimeFile(ctx, file, redisClient)
		})
	default:
		fmt.Fprintf(os.Stderr, "error: invalid command specified, it must be one of: daemon, insert, merge\n")
		os.Exit(1)
	}

	// Wait for the first error to occur
	if err := g.Wait(); err != nil {
		if err == ErrorStop {
			slog.Info("received signal to stop")
		} else if err == context.Canceled || err == context.DeadlineExceeded {
			slog.Info("done")
		} else {
			slog.Error("received error", "error", err.Error())
		}
	}
}

var ErrorStop = fmt.Errorf("received signal to stop")

// signalHandler - listens for signals to stop the process.
func signalHandler(ctx context.Context) error {
	slog.Info("starting signal handler")
	defer slog.Info("stopping signal handler")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-sigCh:
		return ErrorStop
	}
}
