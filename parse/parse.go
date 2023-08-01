package parse

import (
	"strconv"
	"time"

	"golang.org/x/text/language"
)

type Parseable interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		float32 | float64 |
		bool | string |
		complex64 | complex128 |
		time.Duration | time.Time | time.Location | language.Tag
}

func Int(str string) (int, error) {
	return strconv.Atoi(str)
}

func Int8(str string) (int8, error) {
	v, err := strconv.ParseInt(str, 10, 8)
	return int8(v), err
}

func Int16(str string) (int16, error) {
	v, err := strconv.ParseInt(str, 10, 16)
	return int16(v), err
}

func Int32(str string) (int32, error) {
	v, err := strconv.ParseInt(str, 10, 32)
	return int32(v), err
}

func Int64(str string) (int64, error) {
	return strconv.ParseInt(str, 10, 64)
}

func Uint(str string) (uint, error) {
	v, err := strconv.ParseUint(str, 0, strconv.IntSize)
	return uint(v), err
}

func Uint8(str string) (uint8, error) {
	v, err := strconv.ParseUint(str, 0, 8)
	return uint8(v), err
}

func Uint16(str string) (uint16, error) {
	v, err := strconv.ParseUint(str, 0, 16)
	return uint16(v), err
}

func Uint32(str string) (uint32, error) {
	v, err := strconv.ParseUint(str, 0, 32)
	return uint32(v), err
}

func Uint64(str string) (uint64, error) {
	return strconv.ParseUint(str, 0, 64)
}

func Float32(str string) (float32, error) {
	v, err := strconv.ParseFloat(str, 32)
	return float32(v), err
}

func Float64(str string) (float64, error) {
	return strconv.ParseFloat(str, 64)
}

func Bool(str string) (bool, error) {
	return strconv.ParseBool(str)
}

func String(str string) (string, error) {
	return str, nil
}

func Complex64(str string) (complex64, error) {
	v, err := strconv.ParseComplex(str, 64)
	return complex64(v), err
}

func Complex128(str string) (complex128, error) {
	return strconv.ParseComplex(str, 128)
}

func Duration(str string) (time.Duration, error) {
	return time.ParseDuration(str)
}

func Time(str string) (time.Time, error) {
	return time.Parse(time.RFC3339, str)
}

func Location(str string) (time.Location, error) {
	loc, err := time.LoadLocation(str)
	if err != nil {
		loc = time.UTC
	}
	return *loc, err
}

func Language(str string) (language.Tag, error) {
	return language.Parse(str)
}

func Must[P Parseable](parser func(string) (P, error)) func(string) P {
	return func(str string) P {
		v, err := parser(str)
		if err != nil {
			panic(err)
		}
		return v
	}
}

func GetParser[P Parseable](take *P) func(string) (P, error) {
	var p any
	switch any(take).(type) {
	case *int:
		p = any(Int)
	case *int8:
		p = any(Int8)
	case *int16:
		p = any(Int16)
	case *int32:
		p = any(Int32)
	case *int64:
		p = any(Int64)
	case *uint:
		p = any(Uint)
	case *uint8:
		p = any(Uint8)
	case *uint16:
		p = any(Uint16)
	case *uint32:
		p = any(Uint32)
	case *uint64:
		p = any(Uint64)
	case *float32:
		p = any(Float32)
	case *float64:
		p = any(Float64)
	case *bool:
		p = any(Bool)
	case *string:
		p = any(String)
	case *complex64:
		p = any(Complex64)
	case *complex128:
		p = any(Complex128)
	case *time.Duration:
		p = any(Duration)
	case *time.Time:
		p = any(Time)
	case *time.Location:
		p = any(Location)
	case *language.Tag:
		p = any(Language)
	}
	return p.(func(string) (P, error))
}

func Parse[P Parseable](take *P, str string) error {
	parse := GetParser(take)
	var err error
	*take, err = parse(str)
	return err
}

func MustParse[P Parseable](take *P, str string) {
	if err := Parse(take, str); err != nil {
		panic(err)
	}
}
