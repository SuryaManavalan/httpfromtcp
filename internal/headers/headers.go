package headers

import (
	"bytes"
	"errors"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func isValidTokenChar(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '!' || c == '#' || c == '$' || c == '%' || c == '&' || c == '\'' ||
		c == '*' || c == '+' || c == '-' || c == '.' || c == '^' || c == '_' ||
		c == '`' || c == '|' || c == '~'
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, false, nil
	}

	if idx == 0 {
		return 2, true, nil
	}

	line := strings.TrimSpace(string(data[:idx]))

	colonIdx := strings.Index(line, ":")
	if colonIdx == -1 {
		return 0, false, errors.New("invalid header: no colon found")
	}

	key := line[:colonIdx]
	value := line[colonIdx+1:]

	if strings.TrimSpace(key) != key {
		return 0, false, errors.New("invalid header: space between field name and colon")
	}

	if key == "" {
		return 0, false, errors.New("invalid header: empty field name")
	}

	for i := 0; i < len(key); i++ {
		if !isValidTokenChar(key[i]) {
			return 0, false, errors.New("invalid header: field name contains invalid character")
		}
	}

	key = strings.ToLower(key)

	value = strings.TrimSpace(value)

	if existingValue, exists := h[key]; exists {
		h[key] = existingValue + ", " + value
	} else {
		h[key] = value
	}

	return idx + 2, false, nil
}
