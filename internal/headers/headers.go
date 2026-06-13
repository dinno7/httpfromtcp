package headers

import (
	"bytes"
	"errors"
	"strings"
)

const CRLF = "\r\n"

var (
	ErrInvalidHeaderLine               = errors.New("invalid header line")
	ErrInvalidWhitespaceAfterHeaderKey = errors.New("whitespace after header key is not valid")
	ErrInvalidCharacter                = errors.New("invalid charactor")
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false
	for {
		index := bytes.Index(data[read:], []byte(CRLF))
		if index == -1 {
			break
		}

		if index == 0 {
			done = true
			read += len(CRLF)
			break
		}

		key, value, err := parseHeaderLine(data[read : read+index])
		if err != nil {
			return 0, false, err
		}

		h.Append(key, value)

		read += index + len(CRLF)
	}

	return read, done, nil
}

func (h Headers) Get(key string) string {
	key = strings.ToLower(key)
	return h[key]
}

func (h Headers) Append(key, value string) {
	key = strings.ToLower(key)
	if _, ok := h[key]; ok {
		h[key] = h[key] + ", " + value
	} else {
		h[key] = value
	}
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	h[key] = value
}

func (h Headers) Keys() []string {
	keys := []string{}
	for key := range h {
		keys = append(keys, key)
	}
	return keys
}

func (h Headers) Values() []string {
	values := []string{}
	for key := range h {
		values = append(values, h[key])
	}
	return values
}

func (h Headers) ForEach(cb func(key, value string)) {
	for headerKey, headerValue := range h {
		cb(headerKey, headerValue)
	}
}

func parseHeaderLine(data []byte) (string, string, error) {
	parts := bytes.SplitN(data, []byte{':'}, 2)
	if len(parts) != 2 {
		return "", "", ErrInvalidHeaderLine
	}
	key := bytes.TrimPrefix(parts[0], []byte{' '})
	if bytes.HasSuffix(key, []byte{' '}) {
		return "", "", ErrInvalidWhitespaceAfterHeaderKey
	}

	if !isValidHeaderKey(key) {
		return "", "", ErrInvalidCharacter
	}

	value := bytes.TrimSpace(parts[1])
	return string(key), string(value), nil
}

func isValidHeaderKey(key []byte) bool {
	// Valid characters:
	// Uppercase letters: A-Z
	// Lowercase letters: a-z
	// Digits: 0-9
	// Special characters: !, #, $, %, &, ', *, +, -, ., ^, _, `, |, ~
	for i := range key {
		c := key[i]
		isValid := (c >= 'A' && c <= 'Z') ||
			(c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') ||
			(c == '!' || c == '#' || c == '$' || c == '%' || c == '&' || c == '\'' ||
				c == '*' || c == '+' || c == '-' || c == '.' || c == '^' || c == '_' ||
				c == '`' || c == '|' || c == '~')

		if !isValid {
			return false
		}
	}
	return true
}
