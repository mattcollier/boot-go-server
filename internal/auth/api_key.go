package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	a := headers.Get("Authorization")

	parts := strings.Split(a, " ")

	if len(parts) != 2 || parts[0] != "ApiKey" {
		return "", fmt.Errorf("invalid authorization header")
	}

	return parts[1], nil
}
