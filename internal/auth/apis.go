package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	str := headers.Get("Authorization")
	if str == "" {
		return "", fmt.Errorf("there is no header authorization")
	}
	after, found := strings.CutPrefix(str, "ApiKey ")
	if !found {
		return "", fmt.Errorf("there is no 'ApiKey' in header authorization")
	}
	return after, nil
}