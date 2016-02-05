package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

var webhook_receiverCmd = &cobra.Command{

	Use:   "webhook_receiver",
	Short: "Receive webhooks from Sync Gateway and send push notifications",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Are you looking for /webhook_receiver?")
		})

		// http.HandleFunc("/webhook_receiver", libtodolite.WebhookReceiver)

		log.Printf("Listening on port 8080 for webhook events")
		log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))

	},
}

func init() {

	RootCmd.AddCommand(webhook_receiverCmd)

	// Here you will define your flags and configuration settings

	// Cobra supports Persistent Flags which will work for this command and all subcommands
	// webhook_receiverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command is called directly
	// webhook_receiverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle" )

}
