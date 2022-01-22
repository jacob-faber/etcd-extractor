package cmd

import (
	"bufio"
	"context"
	"fmt"
	"github.com/camabeh/etcd-extractor/pkg"
	"github.com/spf13/cobra"
	"io"
	"os"
	"sync"
)

func NewGetCommand(opts *pkg.RunOptions, service *pkg.EtcdService) *cobra.Command {
	getCmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"extract"},
		Short:   "Get all values (or subset if specifying stdin, args or files as args as keys) in YAML format in etcd database",
		Long:    "Restore etcd database and start etcd server then get all values in YAML format. Keys can be also specified as a prefix, e.g. /openshift.io/users",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			keys, err := getEtcdKeys(os.Stdin, args)
			if err != nil {
				return err
			}
			if len(keys) == 0 {
				return fmt.Errorf("specify stdin or arguments")
			}

			err = service.Connect(ctx)
			if err != nil {
				return err
			}
			defer service.Close()

			err = service.RestoreSnapshot(ctx)
			if err != nil {
				return err
			}

			var errCh = make(chan error, 2)
			var wg sync.WaitGroup

			wg.Add(1)
			go func() {
				defer wg.Done()
				errCh <- service.StartServer(ctx)
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()

				waitCtx, cancel := context.WithTimeout(ctx, pkg.DefaultEtcdTimeout)
				defer cancel()
				err = service.WaitForReady(waitCtx)
				if err != nil {
					errCh <- fmt.Errorf("failed to start etcd server: %v", err)
					return
				}

				errCh <- service.PrintValues(ctx, keys, opts.SkipDecodeErrors)
			}()

			wg.Wait()
			close(errCh)
			for err := range errCh {
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	getCmd.PersistentFlags().BoolVarP(&opts.SkipDecodeErrors, "skip-decode-errors", "", false,
		"Skip decode errors")

	return getCmd
}

// Gets etcd keys from stdin, args ar from files specified in args
func getEtcdKeys(stdin *os.File, args []string) ([]string, error) {
	var items []string
	// Handle stdin - Every line entry is a key
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		items, err = appendScan(stdin, items)
		if err != nil {
			return nil, err
		}
	}

	// Handle args - Every argument can be a key or a file
	for _, potentialFile := range args {
		if pkg.IsFile(potentialFile) == nil {
			file, err := os.Open(potentialFile)
			if err != nil {
				return nil, err
			}
			// We have a file, every line entry is a key
			items, err = appendScan(file, items)
			if err != nil {
				return nil, err
			}
		} else {
			// We don't have a file, so it's a key
			items = append(items, potentialFile)
		}
	}

	return items, nil
}

func appendScan(input io.Reader, dest []string) ([]string, error) {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		text := scanner.Text()
		// We don't want empty lines
		if text == "" {
			continue
		}
		dest = append(dest, text)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return dest, nil
}
