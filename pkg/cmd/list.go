package cmd

import (
	"context"
	"fmt"
	"github.com/camabeh/etcd-extractor/pkg"
	"github.com/spf13/cobra"
	"sync"
)

func NewListCommand(service *pkg.EtcdService) *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all keys in etcd database based on prefix (default is \"/\")",
		Long:    "Restore etcd database and start etcd server then list all keys.",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			err := service.Connect(ctx)
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

				if len(args) == 1 {
					errCh <- service.PrintKeysWithPrefix(ctx, args[0])
				} else {
					errCh <- service.PrintAllKeys(ctx)
				}
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
	return listCmd
}
