package main

import "github.com/pundiai/go-sdk/cosmos/loadtest"

func main() {
	rootCmd := loadtest.NewCmd()
	if err := rootCmd.Execute(); err != nil {
		rootCmd.PrintErrln(err)
	}
}
