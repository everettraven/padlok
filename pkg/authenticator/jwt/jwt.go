
package jwt

import (
	"context"
	"errors"

	"github.com/everettraven/padlok/pkg/authenticator/jwt/config"
	"github.com/spf13/pflag"
	"github.com/everettraven/padlok/pkg/internal/third_party/kubernetes/apiserver/pkg/authentication/authenticator"
)

type Configurator interface {
	TokenAuthenticator() authenticator.Token
	Run(context.Context) error
	AddFlags(*pflag.FlagSet)
}

func New() *JWT {
	return &JWT{
		configurator: config.NewConfigurator(),
	}
}

type JWT struct {
	configurator Configurator
}

func (j *JWT) AddFlags(fs *pflag.FlagSet) {
	j.configurator.AddFlags(fs)
}

func (j *JWT) AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error) {
	authenticator := j.configurator.TokenAuthenticator()
	if authenticator != nil {
		return authenticator.AuthenticateToken(ctx, token)
	}

	return nil, false, errors.New("jwt token authenticator is not initialized")
}

func (j *JWT) Run(ctx context.Context) error {
	return j.configurator.Run(ctx)
}
