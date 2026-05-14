package tokengetters

import (
	"context"
	"errors"
	"fmt"

	k8soidc "github.com/everettraven/padlok/pkg/internal/third_party/kubernetes/apiserver/plugin/pkg/authenticator/token/oidc"
)

type RequestProvided struct{}

func (rpatg *RequestProvided) GetAccessToken(ctx context.Context) (string, error) {
	val := ctx.Value(k8soidc.RequestProvidedTokenContextKey)
	if val == nil {
		return "", errors.New("getting access token: no access token present in the request context")
	}

	strVal, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("getting access token: expected access token in the request context to be of type string but got %T", val)
	}

	if strVal == "" {
		return "", fmt.Errorf("getting access token: empty access token found in the request context")
	}

	return strVal, nil
}
