package auth

import (
	"net/http"
	"testing"
)

func TestGetBearerToken_Success(t *testing.T) {
	token := "foo"
	headers := make(http.Header)
	headers.Add("Authorization", "Bearer "+token)

	result, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if result != token {
		t.Fatalf("expected %s, got %s", token, result)
	}
}

func TestGetBearerToken_BadHeader(t *testing.T) {
	token := "foo"
	headers := make(http.Header)
	// no space after Bearer
	headers.Add("Authorization", "Bearer"+token)

	_, err := GetBearerToken(headers)
	if err == nil {
		t.Fatalf("an error was expected")
	}
	if err.Error() != "invalid authorization header" {
		t.Fatalf("incorrect error received")
	}
}
