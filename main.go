package main

import (
	"fmt"

	"github.com/enolgor/go-utils/examples"
)

func main() {
	fmt.Printf("%s:%d tz=%s t=%v\n", examples.HOST, examples.PORT, examples.TIMEZONE.String(), examples.TEST)
}
