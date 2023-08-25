package examples

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/enolgor/go-utils/ht"
)

type Body struct {
	Salute string `json:"salute"`
}

func Http() {
	resp, err := ht.Post("http://postman-echo.com/post").WithHeader("test-header", "asdf").WithBody(Body{Salute: "hello"}).Do()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Status)
	io.Copy(os.Stdout, resp.Body)
}
