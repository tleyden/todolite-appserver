package libtodolite

/*
type WebhookReceiver struct {
	DatabaseURL string
}

func WebhookReceiver(w http.ResponseWriter, r *http.Request) {

	fmt.Println("/webhook_receiver request received")
	// body := r.Body
	defer r.Body.Close()
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading body: %v", err)
	} else {
		fmt.Println("body %v", string(bytes))
	}

	/
		- get doc id
		- fetch doc
		- is it a task?  If no, return
		- fetch list
		- loop over all members except task creator
		- send push: new item added to foo list

	/

	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}
*/
