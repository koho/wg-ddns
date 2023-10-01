package main

import (
	"github.com/spf13/cobra"
)

var (
	iface    string
	interval uint64
	debug    bool
)

var rootCmd = &cobra.Command{
	Use:   "wg-ip",
	Short: "WireGuard DDNS Endpoint",
	RunE: func(cmd *cobra.Command, args []string) error {
		return run()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&iface, "interface", "i", "", "interface name")
	rootCmd.PersistentFlags().Uint64VarP(&interval, "interval", "t", 600, "sync interval")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "show debug log")
	rootCmd.MarkPersistentFlagRequired("interface")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
