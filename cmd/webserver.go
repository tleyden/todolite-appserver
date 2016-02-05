package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gocraft/web"
	"github.com/spf13/cobra"
	"github.com/tleyden/todolite-appserver/libtodolite"
)

var webserverCmd = &cobra.Command{
	Use:   "webserver",
	Short: "Runs a webserver that displays data from TodoLite DB on Sync Gateway",
	Long:  `This connects to the admin port (4985) on Sync Gateway and exposes a web UI to dump data which is useful for debugging.  `,
	Run: func(cmd *cobra.Command, args []string) {

		if *url == "" {
			log.Fatalf("Sync Gateway URL must be provided.  See --help")
			return
		}

		port := 3000

		router := web.New(libtodolite.Context{}).
			Middleware(web.LoggerMiddleware).     // Use some included middleware
			Middleware(web.ShowErrorsMiddleware). // ...
			Middleware(libtodolite.ConfigMiddleware(*url)).
			Middleware((*libtodolite.Context).ConnectToSyncGw).
			Get("/", (*libtodolite.Context).Root).
			Get("/changes", (*libtodolite.Context).ChangesFeed).
			Post("/webhook_receiver", (*libtodolite.Context).WebhookReceiver)

		log.Printf("Listening on port %v", port)
		http.ListenAndServe(fmt.Sprintf(":%v", port), router) // Start the server!

	},
}

func init() {
	RootCmd.AddCommand(webserverCmd)

	url = webserverCmd.PersistentFlags().String("url", "", "Sync Gateway URL")

}
