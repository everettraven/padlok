package oidc

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc"

	"github.com/everettraven/padlok/pkg/apis/authentication"
	"github.com/everettraven/padlok/pkg/apis/authentication/conversion"
	externaloidccel "github.com/everettraven/padlok/pkg/oidc/externalclaims/cel"
	externalclaimsresolver "github.com/everettraven/padlok/pkg/oidc/externalclaims/resolver"
	"github.com/everettraven/padlok/pkg/oidc/externalclaims/tokengetters"

	authenticationcel "github.com/everettraven/padlok/pkg/internal/third_party/kubernetes/apiserver/pkg/authentication/cel"
	k8soidc "github.com/everettraven/padlok/pkg/internal/third_party/kubernetes/apiserver/plugin/pkg/authenticator/token/oidc"
)

type Options struct {
	// JWTAuthenticator is the authenticator that will be used to verify the JWT.
	JWTAuthenticator authentication.JWTAuthenticator

	// Optional KeySet to allow for synchronous initialization instead of fetching from the remote issuer.
	// Mutually exclusive with JWTAuthenticator.Issuer.DiscoveryURL.
	//
	// The following API server metrics for fetching JWKS and provider status will not be recorded if this is set.
	//  - apiserver_authentication_jwt_authenticator_jwks_fetch_last_timestamp_seconds
	//  - apiserver_authentication_jwt_authenticator_jwks_fetch_last_key_set_info
	KeySet oidc.KeySet

	// PEM encoded root certificate contents of the provider.  Mutually exclusive with Client.
	CAContentProvider k8soidc.CAContentProvider

	// Optional http.Client used to make all requests to the remote issuer.  Mutually exclusive with CAContentProvider and EgressLookup.
	Client *http.Client

	// Optional CEL compiler used to compile the CEL expressions. This is useful to use a shared instance
	// of the compiler as these compilers holding a CEL environment are expensive to create. If not provided,
	// a default compiler will be created.
	// Note: the compiler construction depends on feature gates and the compatibility version to be initialized.
	Compiler Compiler

	// SupportedSigningAlgs sets the accepted set of JOSE signing algorithms that
	// can be used by the provider to sign tokens.
	//
	// https://tools.ietf.org/html/rfc7518#section-3.1
	//
	// This value defaults to RS256, the value recommended by the OpenID Connect
	// spec:
	//
	// https://openid.net/specs/openid-connect-core-1_0.html#IDTokenValidation
	SupportedSigningAlgs []string

	DisallowedIssuers []string
}

type Compiler interface {
	authenticationcel.Compiler
	CompileExternalSourceExpression(expressionAccessor authenticationcel.ExpressionAccessor) (authenticationcel.CompilationResult, error)
}

func New(ctx context.Context, opts Options) (k8soidc.AuthenticatorTokenWithHealthCheck, error) {
	if opts.Compiler == nil {
		opts.Compiler = externaloidccel.NewCompiler()
	}

	externalClaimsExpanders, err := buildExternalClaimsExpanders(ctx, opts.Compiler, opts.JWTAuthenticator.ExternalClaimsSources...)
	if err != nil {
		return nil, fmt.Errorf("building external claims expanders: %w", err)
	}

	k8sOpts := k8soidc.Options{
		JWTAuthenticator:                     conversion.ConvertJWTAuthenticatorToApiserverJWTAuthenticator(opts.JWTAuthenticator),
		KeySet:                               opts.KeySet,
		CAContentProvider:                    opts.CAContentProvider,
		Client:                               opts.Client,
		Compiler:                             opts.Compiler,
		SupportedSigningAlgs:                 opts.SupportedSigningAlgs,
		DisallowedIssuers:                    opts.DisallowedIssuers,
		ClaimsExpanders:                      externalClaimsExpanders,
		UnknownCELValueTypesToStringListFunc: externalclaimsresolver.UnknownCELValueTypesToStringListHandler(),
	}

	return k8soidc.New(ctx, k8sOpts)
}

func buildExternalClaimsExpanders(ctx context.Context, compiler Compiler, externalClaimSources ...authentication.ExternalClaimsSource) ([]k8soidc.ClaimsExpander, error) {
	expanders := make([]k8soidc.ClaimsExpander, 0, len(externalClaimSources))
	for _, source := range externalClaimSources {
		tokenGetter, err := buildExternalClaimsTokenGetter(ctx, source.Authentication)
		if err != nil {
			return nil, fmt.Errorf("building access token getter: %w", err)
		}

		expander, err := externalclaimsresolver.New(compiler, tokenGetter, source)
		if err != nil {
			return nil, fmt.Errorf("creating new external claims resolver: %w", err)
		}

		expanders = append(expanders, expander)
	}

	return expanders, nil
}

func buildExternalClaimsTokenGetter(ctx context.Context, authn *authentication.Authentication) (externalclaimsresolver.AccessTokenGetter, error) {
	if authn == nil {
		return &tokengetters.Anonymous{}, nil
	}

	if authn.Type == nil {
		return nil, errors.New("no authentication type specified")
	}

	switch *authn.Type {
	case authentication.AuthenticationTypeRequestProvidedToken:
		return &tokengetters.RequestProvided{}, nil
	case authentication.AuthenticationTypeClientCredential:
		if authn.ClientCredential == nil {
			return nil, fmt.Errorf("client credential configuration is required when authentication type is ClientCredential")
		}
		return tokengetters.NewClientCredential(ctx, authn.ClientCredential)
	default:
		return nil, fmt.Errorf("unknown authentication type %q", *authn.Type)
	}
}
