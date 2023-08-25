package examples

import (
	"os"
	"time"

	"github.com/enolgor/go-utils/conf"
	"golang.org/x/text/language"
)

var PORT int
var HOST string
var TIMEZONE time.Location
var TEST bool
var LANGS []*conf.KeyValue[language.Tag, bool]

func init() {
	os.Setenv("HOST", "asdf")
	conf.Set(&PORT, "PORT", "p", 8080)
	conf.SetEnv(&HOST, "HOST", "localhost")
	conf.SetFlag(&TIMEZONE, "tz", *time.UTC)
	conf.Set(&TEST, "TEST", "t", false)
	conf.SetKVs(&LANGS, "LANGUAGE", "lang", []conf.KeyValue[language.Tag, bool]{{language.English, true}})
	conf.Read()
}
