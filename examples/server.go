package examples

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/enolgor/go-utils/server"
)

func Server() {
	auth := server.BasicAuthHandler("admin", "test")
	http.HandleFunc("/", server.Handle(server.Post, auth, handle))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

type MultiPartForm struct {
}

func handle(req *http.Request) (int, string, func(io.Writer)) {
	r, _ := http.NewRequest(req.Method, req.URL.String(), req.Body)
	r.Header = req.Header
	if err := r.ParseMultipartForm(1 << 30); err != nil {
		log.Fatal(err)
	}
	return 200, "text/plain", func(w io.Writer) {
		for k, v := range r.Form {
			fmt.Fprintf(w, "%s: %v\n", k, v[0])
		}
		fmt.Fprintln(w)
		f, _, _ := r.FormFile("file")
		if f != nil {
			io.Copy(w, f)
		}

	}
}

func handle2(req *http.Request) (int, string, func(io.Writer)) {

	return 200, "text/plain", func(w io.Writer) {
		io.Copy(w, req.Body)
	}
}
