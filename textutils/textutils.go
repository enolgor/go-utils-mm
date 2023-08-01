package textutils

import (
	"regexp"
	"unicode"

	"github.com/icholy/replace"
	"golang.org/x/text/cases"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/unicode/rangetable"
)

var normalizerNoSpace transform.Transformer = transform.Chain(
	norm.NFD,
	runes.ReplaceIllFormed(),
	runes.Remove(runes.NotIn(rangetable.Merge(unicode.Digit, unicode.Letter))),
	runes.Map(unicode.ToLower),
	cases.Fold())

var normalizerSpace transform.Transformer = transform.Chain(
	norm.NFD,
	runes.ReplaceIllFormed(),
	runes.Remove(runes.NotIn(rangetable.Merge(unicode.Digit, unicode.Letter, unicode.Space))),
	runes.Map(unicode.ToLower),
	cases.Fold())

var normalizerSlug transform.Transformer = transform.Chain(
	norm.NFD,
	runes.ReplaceIllFormed(),
	runes.Remove(runes.NotIn(rangetable.Merge(unicode.Digit, unicode.Letter, unicode.Space))),
	replace.Regexp(regexp.MustCompile(`\s+`), []byte("-")),
	runes.Map(unicode.ToLower),
	cases.Fold())

type NormalizeOption byte

const (
	OptionSpace NormalizeOption = 1 << iota
	OptionSlugSpace
)

func Normalize(name string, options ...NormalizeOption) (s string, err error) {
	var option NormalizeOption
	var normalizer transform.Transformer = normalizerNoSpace
	for _, opt := range options {
		option = option | opt
	}
	if OptionSpace&option != 0 {
		normalizer = normalizerSpace
	}
	if OptionSlugSpace&option != 0 {
		normalizer = normalizerSlug
	}
	s, _, err = transform.String(normalizer, name)
	return
}

func MustNormalize(name string, options ...NormalizeOption) (s string) {
	var err error
	if s, err = Normalize(name, options...); err != nil {
		panic(err)
	}
	return
}
