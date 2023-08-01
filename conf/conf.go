package conf

import (
	"flag"
	"fmt"
	"os"

	"github.com/enolgor/go-utils/parse"
)

type Configurable interface {
	parse.Parseable
}

func flagFunc[T Configurable](take *T) func(string) error {
	return func(s string) error {
		return parse.Parse(take, s)
	}
}

func getEnv[T Configurable](take *T, key string) (string, error) {
	var err error
	var str string
	var ok bool
	if str, ok = os.LookupEnv(key); ok && str != "" {
		err = parse.Parse(take, str)
	}
	return str, err
}

var envErr error

func Set[T Configurable](take *T, envKey, flagKey string, def T) {
	*take = def
	if envKey != "" {
		if val, err := getEnv(take, envKey); err != nil {
			envErr = fmt.Errorf(`invalid value "%s" for env "%s": %s`, val, envKey, err.Error())
		}
	}
	if flagKey != "" {
		if v, ok := any(take).(*bool); ok {
			flag.BoolVar(v, flagKey, *v, "")
		} else {
			flag.Func(flagKey, "", flagFunc(take))
		}
	}
}

func SetEnv[T Configurable](take *T, envKey string, def T) {
	Set(take, envKey, "", def)
}

func SetFlag[T Configurable](take *T, flagKey string, def T) {
	Set(take, "", flagKey, def)
}

func Read() {
	if envErr != nil {
		panic(envErr)
	}
	flag.Parse()
}
