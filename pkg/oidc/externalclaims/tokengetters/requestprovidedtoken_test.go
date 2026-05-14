package tokengetters

import (
	"context"
	"fmt"
	"strings"
	"testing"

	k8soidc "github.com/everettraven/padlok/pkg/internal/third_party/kubernetes/apiserver/plugin/pkg/authenticator/token/oidc"
)

func TestRequestProvidedGetAccessToken(t *testing.T) {
	type testcase struct {
		name  string
		ctx   context.Context
		err   string
		value string
	}

	testcases := []testcase{
		{
			name: "no key present in context, error",
			ctx:  context.Background(),
			err:  "getting access token: no access token present in the request context",
		},
		{
			name: "key present in context, non-string value, error",
			ctx:  context.WithValue(context.Background(), k8soidc.RequestProvidedTokenContextKey, 0),
			err:  fmt.Sprintf("getting access token: expected access token in the request context to be of type string but got %T", 0),
		},
		{
			name: "key present in context, empty string value, error",
			ctx:  context.WithValue(context.Background(), k8soidc.RequestProvidedTokenContextKey, ""),
			err:  "getting access token: empty access token found in the request context",
		},
		{
			name:  "key present in context, non-empty string value, no error",
			ctx:   context.WithValue(context.Background(), k8soidc.RequestProvidedTokenContextKey, "token"),
			value: "token",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rptg := &RequestProvided{}

			val, err := rptg.GetAccessToken(tc.ctx)
			switch {
			case err == nil && len(tc.err) > 0:
				t.Fatalf("expected error containing %q but got none", tc.err)
			case err != nil && len(tc.err) == 0:
				t.Fatalf("received an unexpected error: %v", err)
			case err != nil && len(tc.err) > 0 && !strings.Contains(err.Error(), tc.err):
				t.Fatalf("error %v does not contain expected substring %q", err, tc.err)
			}

			if tc.value != "" && tc.value != val {
				t.Fatalf("expected a return value of %q but got %q", tc.value, val)
			}
		})
	}
}
