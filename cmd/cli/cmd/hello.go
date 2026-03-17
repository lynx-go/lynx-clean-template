package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
)

// helloCmd represents the hello command
var helloCmd = &cobra.Command{
	Use:   "hello",
	Short: "pub hello message to PubSub",
	Long:  "pub hello message to PubSub",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli := buildCLI(cmd, args, func(ctx context.Context, cc *CLIContext, args *CLIArgs) error {
			toUid := args.GetString("to")
			cc.Printf("Hello world to %s \n", toUid)
			//log.InfoContext(ctx, "hello lynx cli", "args", args, "to_uid", toUid)
			//time.Sleep(1 * time.Second)
			return nil
		}, WithPreWaitTime(500*time.Millisecond), WithPostWaitTime(1*time.Second))
		cli.Run()
	},
}

func init() {
	rootCmd.AddCommand(helloCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// helloCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// helloCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	helloCmd.Flags().StringP("to", "t", "", "hello to uid")
}
