package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// webserverCmd respresents the webserver command
var webserverCmd = &cobra.Command{
	Use:   "webserver",
	Short: "Runs a webserver that displays data from TodoLite DB on Sync Gateway",
	Long:  `This connects to the admin port (4985) on Sync Gateway and exposes a web UI to dump data which is useful for debugging.  `,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		fmt.Println("webserver called")

	},
}

func init() {
	RootCmd.AddCommand(webserverCmd)

	// Here you will define your flags and configuration settings

	// Cobra supports Persistent Flags which will work for this command and all subcommands
	// webserverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command is called directly
	// webserverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle" )

}
