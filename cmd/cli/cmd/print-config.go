/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"

	config "github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	"github.com/lynx-go/x/encoding/json"
	"github.com/spf13/cobra"
)

// printConfigCmd represents the printConfig command
var printConfigCmd = &cobra.Command{
	Use:   "print-config",
	Short: "print config",
	Long:  `Print Configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		buildCLI(cmd, args, func(ctx context.Context, cc *CLIContext, args *CLIArgs) error {
			var cfg config.AppConfig
			if err := config.UnmarshalConfig(cc.App.Config(), &cfg); err != nil {
				return err
			}
			cc.Println("configuration: ")
			cc.Println("")
			out, _ := json.MarshalIndent(&cfg, "", "    ")
			cc.Printf("%s", out)
			cc.Println("")
			return nil
		}, WithLogToFile()).Run()
	},
}

func init() {
	rootCmd.AddCommand(printConfigCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// printConfigCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// printConfigCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
