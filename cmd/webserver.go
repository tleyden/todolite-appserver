package cmd

import (
	"net/http"

	"github.com/gocraft/web"
	"github.com/spf13/cobra"
	"github.com/tleyden/todolite-appserver/libtodolite"
)

// webserverCmd respresents the webserver command
var webserverCmd = &cobra.Command{
	Use:   "webserver",
	Short: "Runs a webserver that displays data from TodoLite DB on Sync Gateway",
	Long:  `This connects to the admin port (4985) on Sync Gateway and exposes a web UI to dump data which is useful for debugging.  `,
	Run: func(cmd *cobra.Command, args []string) {

		router := web.New(libtodolite.Context{}).
			Middleware(web.LoggerMiddleware).     // Use some included middleware
			Middleware(web.ShowErrorsMiddleware). // ...
			Middleware((*libtodolite.Context).ConnectToSyncGw).
			Get("/", (*libtodolite.Context).Root).
			Get("/changes", (*libtodolite.Context).ChangesFeed)
		http.ListenAndServe("localhost:3000", router) // Start the server!

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
