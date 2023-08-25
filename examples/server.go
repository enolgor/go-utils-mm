package examples

import (
	"net/http"

	"github.com/enolgor/go-utils/ht"
)

func Server() {
	auth := ht.BasicAuthHandler("admin", "test")
	http.HandleFunc("/", auth.Finally(ht.Handle(ht.Html("<html><body>hello</body></html>"))))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
