// from https://github.com/pillarjs/path-to-regexp/tree/c7ec332e87d8560673884d5629e1cb23cb03cb87
package path

import (
	"fmt"
	"regexp"
	"strings"
)

type lexTokentype string

const (
	_open         lexTokentype = "OPEN"
	_close        lexTokentype = "CLOSE"
	_pattern      lexTokentype = "PATTERN"
	_name         lexTokentype = "NAME"
	_char         lexTokentype = "CHAR"
	_escaped_char lexTokentype = "ESCAPED_CHAR"
	_modifier     lexTokentype = "MODIFIER"
	_end          lexTokentype = "END"
)

type lexToken struct {
	_type lexTokentype
	index int
	value string
}

func lexer(str string) ([]lexToken, error) {
	tokens := []lexToken{}
	var char byte
	i := 0
	for i < len(str) {
		char = str[i]
		if char == '*' || char == '+' || char == '?' {
			tokens = append(tokens, lexToken{_type: _modifier, index: i, value: string(str[i])})
			i = i + 1
			continue
		}
		if char == '\\' {
			tokens = append(tokens, lexToken{_type: _escaped_char, index: i, value: string(str[i])})
			i = i + 1
			continue
		}
		if char == '{' {
			tokens = append(tokens, lexToken{_type: _open, index: i, value: string(str[i])})
			i = i + 1
			continue
		}

		if char == '}' {
			tokens = append(tokens, lexToken{_type: _close, index: i, value: string(str[i])})
			i = i + 1
			continue
		}

		if char == ':' {
			var name string
			j := i + 1

			for j < len(str) {
				code := str[j]
				if
				// `0-9`
				(code >= 48 && code <= 57) ||
					// `A-Z`
					(code >= 65 && code <= 90) ||
					// `a-z`
					(code >= 97 && code <= 122) ||
					// `_`
					code == 95 {
					name = name + string(str[j])
					j = j + 1
					continue
				}
				break
			}

			if name == "" {
				return nil, fmt.Errorf("missing parameter name at %d", i)
			}
			tokens = append(tokens, lexToken{_type: _name, index: i, value: name})
			i = j
			continue
		}

		if char == '(' {
			count := 1
			pattern := ""
			j := i + 1
			if str[j] == '?' {
				return nil, fmt.Errorf("pattern cannot start with '?' at %d", j)
			}
			for j < len(str) {
				if str[j] == '\\' {
					pattern = pattern + string(str[j]) + string(str[j+1])
					j = j + 2
				}
				if str[j] == ')' {
					count = count - 1
					if count == 0 {
						j = j + 1
						break
					}
				} else if str[j] == '(' {
					count = count + 1
					if str[j+1] != '?' {
						return nil, fmt.Errorf("capturing groups are not allowed at %d", j)
					}
				}
				pattern = pattern + string(str[j])
				j = j + 1
			}
			if count != 0 {
				return nil, fmt.Errorf("unbalanced pattern at %d", i)
			}
			if pattern == "" {
				return nil, fmt.Errorf("missing pattern at %d", i)
			}
			tokens = append(tokens, lexToken{_type: _pattern, index: i, value: pattern})
			i = j
			continue
		}
		tokens = append(tokens, lexToken{_type: _char, index: i, value: string(str[i])})
		i = i + 1
	}
	tokens = append(tokens, lexToken{_type: _end, index: i, value: ""})
	return tokens, nil
}

type TokensToRegexpOptions struct {
	Prefixes  string
	Delimiter string
	Strict    bool
	Start     bool
	End       bool
	Encode    func(string) string
	EndsWith  string
}

func defaultOptions() *TokensToRegexpOptions {
	return &TokensToRegexpOptions{
		Strict:    false,
		Start:     true,
		End:       true,
		Encode:    func(s string) string { return s },
		Prefixes:  "./",
		Delimiter: "/#?",
		EndsWith:  "",
	}
}

type Key struct {
	Name     any
	Prefix   string
	Suffix   string
	Pattern  string
	Modifier string
}

type Token any

func parse(str string) ([]Token, error) {
	tokens, err := lexer(str)
	if err != nil {
		return nil, err
	}
	options := defaultOptions()
	defaultPattern := `[^${` + escapeString(options.Delimiter) + `}]+?`
	result := []Token{}
	key := 0
	i := 0
	path := ""

	tryConsume := func(_type lexTokentype) *string {
		if i < len(tokens) && tokens[i]._type == _type {
			i = i + 1
			return &tokens[i-1].value
		}
		return nil
	}

	mustConsume := func(_type lexTokentype) (string, error) {
		value := tryConsume(_type)
		if value != nil {
			return *value, nil
		}
		return "", fmt.Errorf("unexpected %s at %d, expected %s", tokens[i]._type, tokens[i].index, _type)
	}

	consumeText := func() string {
		result := ""
		var value *string
		for {
			value = tryConsume(_char)
			if value == nil {
				value = tryConsume(_escaped_char)
			}
			if value == nil {
				break
			}
			result = result + *value
		}
		return result
	}

	for i < len(tokens) {
		_char := tryConsume(_char)
		name := tryConsume(_name)
		pattern := tryConsume(_pattern)
		if name != nil || pattern != nil {
			prefix := ""
			if _char != nil {
				prefix = *_char
			}
			if !strings.Contains(options.Prefixes, prefix) {
				path = path + prefix
				prefix = ""
			}
			if path != "" {
				result = append(result, path)
				path = ""
			}
			_key := Key{}
			if name != nil {
				_key.Name = *name
			} else {
				_key.Name = key
				key = key + 1
			}
			_key.Prefix = prefix
			_key.Suffix = ""
			if pattern != nil {
				_key.Pattern = *pattern
			} else {
				_key.Pattern = defaultPattern
			}
			if modifier := tryConsume(_modifier); modifier != nil {
				_key.Modifier = *modifier
			} else {
				_key.Modifier = ""
			}
			result = append(result, _key)
			continue
		}
		var value *string
		if _char != nil {
			value = _char
		} else {
			value = tryConsume(_escaped_char)
		}
		if value != nil {
			path = path + (*value)
			continue
		}
		if path != "" {
			result = append(result, path)
			path = ""
		}
		open := tryConsume(_open)
		if open != nil {
			prefix := consumeText()
			var name, pattern string
			if _name := tryConsume(_name); _name != nil {
				name = *_name
			} else {
				name = ""
			}
			if _pattern := tryConsume(_pattern); _pattern != nil {
				pattern = *_pattern
			} else {
				pattern = ""
			}
			suffix := consumeText()
			if _, err := mustConsume(_close); err != nil {
				return nil, err
			}
			_key := Key{}
			if name != "" {
				_key.Name = name
			} else if pattern != "" {
				_key.Name = key
				key = key + 1
			} else {
				_key.Name = ""
			}
			if name != "" && pattern == "" {
				_key.Pattern = defaultPattern
			} else {
				_key.Pattern = pattern
			}
			_key.Prefix = prefix
			_key.Suffix = suffix
			if modifier := tryConsume(_modifier); modifier != nil {
				_key.Modifier = *modifier
			} else {
				_key.Modifier = ""
			}
			continue
		}
		if _, err := mustConsume(_end); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func tokensToRegexp(tokens []Token, keys *[]Key) (*regexp.Regexp, error) {
	options := defaultOptions()
	encode := options.Encode
	endsWithRe := fmt.Sprintf(`[%s]|$`, escapeString(options.EndsWith))
	delimiterRe := fmt.Sprintf(`[%s]`, escapeString(options.Delimiter))
	route := ""
	if options.Start {
		route = "^"
	}
	for _, _token := range tokens {
		if token, ok := _token.(string); ok {
			route = route + escapeString(encode(token))
			continue
		}
		token := _token.(Key)
		prefix := escapeString(encode(token.Prefix))
		suffix := escapeString(encode(token.Suffix))
		if token.Pattern != "" {
			if keys != nil {
				*keys = append(*keys, token)
			}
			if prefix != "" || suffix != "" {
				if token.Modifier == "+" || token.Modifier == "*" {
					mod := ""
					if token.Modifier == "*" {
						mod = "?"
					}
					route = route + fmt.Sprintf(`(?:%s((?:%s)(?:%s%s(?:%s))*)%s)%s`, prefix, token.Pattern, suffix, prefix, token.Pattern, suffix, mod)
				} else {
					route = route + fmt.Sprintf(`(?:%s(%s)%s)%s`, prefix, token.Pattern, suffix, token.Modifier)
				}
			} else {
				if token.Modifier == "+" || token.Modifier == "*" {
					route = route + fmt.Sprintf(`((?:%s)%s)`, token.Pattern, token.Modifier)
				} else {
					route = route + fmt.Sprintf(`(%s)%s`, token.Pattern, token.Modifier)
				}
			}
		} else {
			route = route + fmt.Sprintf(`(?:%s%s)%s`, prefix, suffix, token.Modifier)
		}
	}
	if options.End {
		if !options.Strict {
			route = route + fmt.Sprintf(`%s?`, delimiterRe)
		}
		if options.EndsWith == "" {
			route = route + "$"
		} else {
			route = route + fmt.Sprintf(`(?=%s)`, endsWithRe)
		}
	} else {
		endToken := tokens[len(tokens)-1]
		var isEndDelimited bool
		if v, ok := endToken.(string); ok {
			isEndDelimited = strings.Contains(delimiterRe, string(v[len(v)-1]))
		} else {
			isEndDelimited = endToken == nil
		}
		if !options.Strict {
			route = route + fmt.Sprintf(`(?:%s(?=%s))?`, delimiterRe, endsWithRe)
		}
		if !isEndDelimited {
			route = route + fmt.Sprintf(`(?=%s|%s)`, delimiterRe, endsWithRe)
		}
	}
	return regexp.Compile(route)
	//return new RegExp(route, flags(options));
	//function flags(options?: { sensitive?: boolean }) {
	//return options && options.sensitive ? "" : "i";
}

func PathToRegexp(path string, keys *[]Key) (*regexp.Regexp, error) {
	tokens, err := parse(path)
	if err != nil {
		return nil, err
	}
	return tokensToRegexp(tokens, keys)
}

var escape *regexp.Regexp = regexp.MustCompile(`([.+*?=^!:${}()[\]|/\\])`)

func escapeString(str string) string {
	return escape.ReplaceAllString(str, "\\$1")
}

func Matcher(expr string) (func(string, map[any]string) bool, error) {
	keys := []Key{}
	re, err := PathToRegexp(expr, &keys)
	if err != nil {
		return nil, err
	}
	return func(path string, valueMap map[any]string) bool {
		if values := re.FindAllStringSubmatch(path, -1); values != nil {
			for i := 1; i < len(values[0]); i++ {
				if values[0][i] == "" {
					continue
				}
				valueMap[keys[i-1].Name] = values[0][i]
			}
			return true
		}
		return false
	}, nil
}
