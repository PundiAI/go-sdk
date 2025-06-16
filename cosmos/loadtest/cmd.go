package loadtest

import (
	"errors"
	"fmt"
	"strings"

	tmloadtest "github.com/informalsystems/tm-load-test/pkg/loadtest"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCmd() *cobra.Command {
	var cfg tmloadtest.Config
	rootCmd := &cobra.Command{
		Use:   "loadtest <genesis|node_url> <key_dir>",
		Short: "Load test a Cosmos node",
		Args:  cobra.ExactArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			viper.SetConfigType("json")
			viper.SetConfigName("config")
			viper.AddConfigPath(".")
			if err := viper.ReadInConfig(); err != nil {
				var configFileNotFoundError viper.ConfigFileNotFoundError
				if errors.As(err, &configFileNotFoundError) {
					// return viper.WriteConfig()
					return nil
				}
				return err
			}
			return viper.Unmarshal(&cfg, func(decoderConfig *mapstructure.DecoderConfig) {
				decoderConfig.TagName = "json"
				decoderConfig.MatchName = func(mapKey, fieldName string) bool {
					mapKey = strings.ReplaceAll(mapKey, "-", "_")
					fieldName = strings.ReplaceAll(fieldName, "-", "_")
					return strings.EqualFold(mapKey, fieldName)
				}
			})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			genesisOrNodeURL := args[0]
			keyDir := args[1]
			baseInfo, err := NewBaseInfo(genesisOrNodeURL, keyDir)
			if err != nil {
				return err
			}
			fmt.Println("init accounts success", baseInfo.Accounts.Len())

			msgSendClientFactory := NewMsgSendClientFactory(baseInfo, baseInfo.GetDenom())
			if err = tmloadtest.RegisterClientFactory(msgSendClientFactory.Name(), msgSendClientFactory); err != nil {
				return err
			}

			if err = cfg.Validate(); err != nil {
				return err
			}
			return tmloadtest.ExecuteStandalone(cfg)
		},
	}
	rootCmd.Flags().StringVar(&cfg.ClientFactory, "client-factory", "msg_send", "The identifier of the client factory to use for generating load testing transactions")
	rootCmd.Flags().IntVarP(&cfg.Connections, "connections", "c", 1, "The number of connections to open to each endpoint simultaneously")
	rootCmd.Flags().IntVarP(&cfg.Time, "time", "T", 60, "The duration (in seconds) for which to handle the load test")
	rootCmd.Flags().IntVarP(&cfg.SendPeriod, "send-period", "p", 1, "The period (in seconds) at which to send batches of transactions")
	rootCmd.Flags().IntVarP(&cfg.Rate, "rate", "r", 1000, "The number of transactions to generate each second on each connection, to each endpoint")
	rootCmd.Flags().IntVarP(&cfg.Size, "size", "s", 250, "The size of each transaction, in bytes - must be greater than 40")
	rootCmd.Flags().IntVarP(&cfg.Count, "count", "N", -1, "The maximum number of transactions to send - set to -1 to turn off this limit")
	rootCmd.Flags().StringVar(&cfg.BroadcastTxMethod, "broadcast-tx-method", "async", "The broadcast_tx method to use when submitting transactions - can be async, sync or commit")
	rootCmd.Flags().StringSliceVar(&cfg.Endpoints, "endpoints", []string{}, "A comma-separated list of URLs indicating Tendermint WebSockets RPC endpoints to which to connect")
	rootCmd.Flags().StringVar(&cfg.EndpointSelectMethod, "endpoint-select-method", tmloadtest.SelectSuppliedEndpoints, "The method by which to select endpoints")
	rootCmd.Flags().IntVar(&cfg.ExpectPeers, "expect-peers", 0, "The minimum number of peers to expect when crawling the P2P network from the specified endpoint(s) prior to waiting for workers to connect")
	rootCmd.Flags().IntVar(&cfg.MaxEndpoints, "max-endpoints", 0, "The maximum number of endpoints to use for testing, where 0 means unlimited")
	rootCmd.Flags().IntVar(&cfg.PeerConnectTimeout, "peer-connect-timeout", 600, "The number of seconds to wait for all required peers to connect if expect-peers > 0")
	rootCmd.Flags().IntVar(&cfg.MinConnectivity, "min-peer-connectivity", 0, "The minimum number of peers to which each peer must be connected before starting the load test")
	rootCmd.Flags().StringVar(&cfg.StatsOutputFile, "stats-output", "", "Where to store aggregate statistics (in CSV format) for the load test")
	return rootCmd
}
