package tokengetters

import "context"

type Anonymous struct{}

func (a *Anonymous) GetAccessToken(_ context.Context) (string, error) {
	return "", nil
}
