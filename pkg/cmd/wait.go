package cmd

import (
	"context"
	"github.com/camabeh/etcd-extractor/pkg"
	"github.com/spf13/cobra"
	"math"
	"sync"
	"time"
)

func NewWaitCommand(service *pkg.EtcdService) *cobra.Command {
	waitCmd := &cobra.Command{
		Use:   "wait",
		Short: "Restore etcd database and start etcd server running infinitely",
		RunE: func(cmd *cobra.Command, strings []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			err := service.RestoreSnapshot(ctx)
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
				sleepWithContext(ctx, math.MaxInt)
				errCh <- nil
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

	return waitCmd
}

func sleepWithContext(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	select {
	case <-ctx.Done():
		if !timer.Stop() {
			// If the timer has been stopped then read from the channel to empty the channel
			<-timer.C
		}
	case <-timer.C:
		// Nothing here...
	}
}
