package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	ErrRecordNotFound = fmt.Errorf("record not found")
	iface             string
	dnsServer         string
	interval          uint64 = 300
	debug             bool
)

var rootCmd = &cobra.Command{
	Use:   "wg-svcb",
	Short: "WireGuard SVCB Endpoint",
	RunE: func(cmd *cobra.Command, args []string) error {
		return run()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&iface, "interface", "i", "", "interface name")
	rootCmd.PersistentFlags().StringVarP(&dnsServer, "dns", "s", "223.5.5.5:53", "dns server")
	rootCmd.PersistentFlags().Uint64VarP(&interval, "interval", "t", 300, "sync interval")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "show debug log")
	rootCmd.MarkPersistentFlagRequired("interface")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
