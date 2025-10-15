package headers

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Get(key string) (string, bool) {
	keyLower := strings.ToLower(key)
	val, ok := h[keyLower]
	return val, ok
}

// Parse will parse a byte array header and
// return the number of bytes read and a done bool
func (h Headers) Parse(data []byte) (int, bool, error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return 2, true, nil
	}
	headerString := string(data[:idx])
	hdrSplitIdx := strings.Index(headerString, ":")
	if hdrSplitIdx == -1 || hdrSplitIdx == 0 {
		return 0, false, fmt.Errorf("no valid \":\" found in header line: %v", headerString)
	}
	if string(headerString[hdrSplitIdx-1]) == " " {
		return 0, false, fmt.Errorf("can't have any whitespace before \":\" value")
	}
	key := headerString[:hdrSplitIdx]
	val := headerString[hdrSplitIdx+1:]
	trimmedLowerKey := strings.ToLower(strings.TrimSpace(key))
	trimmedVal := strings.TrimSpace(val)

	if err := checkForValidKey(trimmedLowerKey); err != nil {
		return 0, false, fmt.Errorf("couldn't parse header: %w", err)
	}

	v, exists := h[trimmedLowerKey]
	if exists {
		v := fmt.Sprintf("%s, %s", v, trimmedVal)
		h[trimmedLowerKey] = v
	} else {
		h[trimmedLowerKey] = trimmedVal
	}

	// Check whether we have a carrage return right after previous one
	nextRtn := bytes.Index(data[idx+2:], []byte("\r\n"))
	if nextRtn == 0 {
		return idx + 2, true, nil
	}
	return idx + 2, false, nil
}

var allowedChars = []string{"!", "#", "$", "%", "&", "'", "*", "+", "-", ".", "^", "_", "`", "|", "~"}

func checkForValidKey(key string) error {
	if len(key) == 1 {
		return fmt.Errorf("key is too short")
	}
	for _, char := range key {
		if !(char >= 'A' && char <= 'z') &&
			!(char >= '0' && char <= '9') &&
			!slices.Contains(allowedChars, string(char)) {
			return fmt.Errorf("key %q has incorrect character: %s", key, string(char))
		}
	}
	return nil
}
