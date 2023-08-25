package examples

import (
	"net/http"

	"github.com/enolgor/go-utils/server"
)

func Server() {
	auth := server.BasicAuthHandler("admin", "test")
	http.HandleFunc("/", server.Handle(server.Get, auth, server.Html("<html><body>hello</body></html>")))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
