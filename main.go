package main

import (
	"fmt"
	"log"

	"github.com/enolgor/go-utils/examples"
	"github.com/enolgor/go-utils/server/path"
)

type Key string

func main() {
	examples.Server(8080)
	match, err := path.Matcher("/test/(.*)/hola/:asdf/(.*)")
	if err != nil {
		log.Fatal(err)
	}
	valueMap := map[any]string{}
	fmt.Println(match("/test/eneko/foo/hola/end/again", valueMap))
	fmt.Println(valueMap)
}
