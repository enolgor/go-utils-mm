package main

import (
	"github.com/enolgor/go-utils/examples"
)

// func main() {
// 	fmt.Printf("%s:%d tz=%s t=%v\n", examples.HOST, examples.PORT, examples.TIMEZONE.String(), examples.TEST)
// 	for _, lang := range examples.LANGS {
// 		fmt.Printf("%v\n", lang)
// 	}
// 	//examples.Http()
// }

func main() {
	examples.Server()
}
