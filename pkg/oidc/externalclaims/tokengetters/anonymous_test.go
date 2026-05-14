package tokengetters

import "testing"

func TestAnonymousGetAccessToken(t *testing.T) {
	anonymous := &Anonymous{}
	token, err := anonymous.GetAccessToken(t.Context())
	if err != nil {
		t.Fatalf("received an unexpected error getting token from anonymous tokengetter: %v", err)
	}

	if token != "" {
		t.Fatalf("expected anonymous tokengetter to return an empty string as the token but got %q", token)
	}
}
