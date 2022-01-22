package main

import (
	"context"
	"github.com/camabeh/etcd-extractor/pkg"
	cmd "github.com/camabeh/etcd-extractor/pkg/cmd"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"os/signal"
	"syscall"
)

// TODO - Get rid of globals
var logger *zap.Logger
var atom = zap.NewAtomicLevel()

func setupLogger() {
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	logger = zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	), zap.AddStacktrace(zapcore.ErrorLevel))
	atom.SetLevel(zap.InfoLevel)
}

func main() {
	setupLogger()
	defer func() {
		_ = logger.Sync()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Propagate app termination via context
	go func() {
		select {
		case <-stop:
			cancel()
		}
	}()

	opts := &pkg.RunOptions{}
	var rootCmd = &cobra.Command{
		Use: "etcd-extractor",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return atom.UnmarshalText([]byte(opts.LogLevel))
		},
	}

	rootCmd.PersistentFlags().StringVarP(&opts.Endpoint, "endpoint", "", "http://127.0.0.1:2379",
		"URL pointing to running etcd instance")
	rootCmd.PersistentFlags().StringVarP(&opts.Snapshot, "snapshot", "", "/tmp/snapshot.db",
		"File path to the snapshot location")
	rootCmd.PersistentFlags().StringVarP(&opts.LogLevel, "loglevel", "", "info",
		"Specify log level (debug, info, warn, error, dpanic, panic, fatal)")
	rootCmd.PersistentFlags().BoolVarP(&opts.SkipEtcdRestore, "skip-etcd-restore", "", false,
		"Skip restoring etcd")
	rootCmd.PersistentFlags().BoolVarP(&opts.SkipEtcdStart, "skip-etcd-start", "", false,
		"Skip starting etcd")

	etcdService := pkg.NewEtcdService(logger, opts)
	rootCmd.AddCommand(cmd.NewListCommand(etcdService))
	rootCmd.AddCommand(cmd.NewGetCommand(opts, etcdService))
	rootCmd.AddCommand(cmd.NewWaitCommand(etcdService))
	rootCmd.AddCommand(cmd.NewVersionCommand())

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		logger.Fatal(err.Error())
	}
}
