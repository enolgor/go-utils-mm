package examples

import (
	"os"
	"time"

	"github.com/enolgor/go-utils/conf"
)

var PORT int
var HOST string
var TIMEZONE time.Location
var TEST bool

func init() {
	os.Setenv("HOST", "asdf")
	conf.Set(&PORT, "PORT", "p", 8080)
	conf.SetEnv(&HOST, "HOST", "localhost")
	conf.SetFlag(&TIMEZONE, "tz", *time.UTC)
	conf.Set(&TEST, "TEST", "t", false)
	conf.Read()
}
