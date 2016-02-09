package cmd

import (
	"fmt"

	"github.com/couchbaselabs/logg"
	"github.com/spf13/cobra"
	"github.com/tleyden/todolite-appserver/libtodolite"
)

var (
	url        *string
	openOcrUrl *string
	since      *string
)

func init() {
	logg.LogKeys["CLI"] = true
	logg.LogKeys["TODOLITE"] = true
}

// changesfollowerCmd respresents the changesfollower command
var changesfollowerCmd = &cobra.Command{
	Use:   "changesfollower",
	Short: "Follows the todolite changes feed and performs actions",
	Long:  `This has the ability to run images through OCR.  In the future, it will be able to send push notifications, etc.`,
	Run: func(cmd *cobra.Command, args []string) {

		if *url == "" {
			logg.LogError(fmt.Errorf("URL is empty"))
			return
		}
		if *openOcrUrl == "" {
			logg.LogError(fmt.Errorf("OpenOcr URL is empty"))
			return
		}

		logg.LogTo("CLI", "url: %v openOcrUrl: %v", *url, *openOcrUrl)
		todoliteApp := libtodolite.NewTodoLiteApp(*url, *openOcrUrl)
		err := todoliteApp.InitApp()
		if err != nil {
			logg.LogPanic("Error initializing todo lite app: %v", err)
		}
		go todoliteApp.FollowChangesFeed(*since)
		select {}

	},
}

func init() {
	RootCmd.AddCommand(changesfollowerCmd)

	url = changesfollowerCmd.PersistentFlags().String("url", "", "Sync Gateway URL")

	openOcrUrl = changesfollowerCmd.PersistentFlags().String("openOcrUrl", "", "OpenOCR Server URL")

	since = changesfollowerCmd.PersistentFlags().String("since", "", "Since value to start changes feed at (last sequence)")

}
