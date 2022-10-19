package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/buglloc/cf-ddns/internal/watcher"
)

var watchArgs = struct {
	ZoneID  string
	DNSName string
	Proxied bool
	IPv4    bool
	IPv6    bool
	TTL     time.Duration
}{
	Proxied: false,
	IPv4:    true,
	IPv6:    false,
	TTL:     5 * time.Minute,
}

var watchCmd = &cobra.Command{
	Use:           "watch",
	SilenceUsage:  true,
	SilenceErrors: true,
	Short:         "Starts watcher",
	RunE: func(_ *cobra.Command, _ []string) error {
		cfToken := os.Getenv("CF_API_TOKEN")
		if cfToken == "" {
			return errors.New("no env[CF_API_TOKEN]")
		}

		cfc, err := cloudflare.NewWithAPIToken(cfToken)
		if err != nil {
			return fmt.Errorf("unable to create cloudflare client: %w", err)
		}

		instance, err := watcher.NewWatcher(
			watcher.WithCloudflareAPI(cfc),
			watcher.WithZoneID(watchArgs.ZoneID),
			watcher.WithDNSName(watchArgs.DNSName),
			watcher.WithProxied(watchArgs.Proxied),
			watcher.WithIPv4(watchArgs.IPv4),
			watcher.WithIPv6(watchArgs.IPv6),
			watcher.WithTTL(watchArgs.TTL),
		)
		if err != nil {
			return fmt.Errorf("unable to create watcher: %w", err)
		}

		errChan := make(chan error, 1)
		okChan := make(chan struct{})
		go func() {
			err := instance.Watch()
			if err != nil {
				errChan <- err
				return
			}

			close(okChan)
		}()

		stopChan := make(chan os.Signal, 1)
		signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-stopChan:
			log.Info().Msg("shutting down gracefully by signal")

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()

			instance.Shutdown(ctx)
		case err := <-errChan:
			log.Error().Err(err).Msg("start failed")
			return err
		case <-okChan:
		}
		return nil
	},
}

func init() {
	flags := watchCmd.PersistentFlags()
	flags.StringVar(&watchArgs.ZoneID, "zone-id", watchArgs.ZoneID, "ID of the zone that will get the records")
	flags.StringVar(&watchArgs.DNSName, "dns-name", watchArgs.DNSName, "DNS name to update the A/AAAA records")
	flags.BoolVar(&watchArgs.Proxied, "proxied", watchArgs.Proxied, "proxy through Cloudflare CDN")
	flags.BoolVar(&watchArgs.IPv4, "ipv4", watchArgs.IPv4, "enable IPv4 support")
	flags.BoolVar(&watchArgs.IPv6, "ipv6", watchArgs.IPv6, "enable IPv6 support")
	flags.DurationVar(&watchArgs.TTL, "ttl", watchArgs.TTL, "DNS record & check TTL")
}
