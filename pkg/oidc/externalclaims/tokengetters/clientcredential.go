package tokengetters

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/everettraven/padlok/pkg/apis/authentication"
)

// tokenEndpointRequestTimeout is the timeout for HTTP requests to the token endpoint.
const tokenEndpointRequestTimeout = time.Second

// ClientCredential is an AccessTokenGetter that uses the OAuth2 client credentials
// grant flow to obtain an access token for authenticating with external claims sources.
type ClientCredential struct {
	tokenSource oauth2.TokenSource
}

// NewClientCredential creates a new ClientCredential token getter from the provided configuration.
func NewClientCredential(ctx context.Context, cfg *authentication.ClientCredentialConfig) (*ClientCredential, error) {
	if cfg == nil {
		return nil, fmt.Errorf("client credential configuration must not be nil")
	}

	httpClient, err := httpClientForClientCredential(cfg.TLS)
	if err != nil {
		return nil, fmt.Errorf("building http client for client credential token endpoint: %w", err)
	}

	ccCfg := &clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     cfg.TokenEndpoint,
		Scopes:       cfg.Scopes,
	}

	wrappedCtx := context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	cc := &ClientCredential{
		tokenSource: ccCfg.TokenSource(wrappedCtx),
	}

	// do an initial access token fetch to prime
	// the cache and catch any configuration issues.
	_, err = cc.GetAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting initial access token: %v", err)
	}

	return cc, nil
}

func (cc *ClientCredential) GetAccessToken(_ context.Context) (string, error) {
	token, err := cc.tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("getting token: %w", err)
	}

	return token.AccessToken, nil
}

func httpClientForClientCredential(tlsCfg *authentication.TLS) (*http.Client, error) {
	client := &http.Client{
		Timeout: tokenEndpointRequestTimeout,
	}

	if tlsCfg == nil {
		return client, nil
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()

	err := modifyTransportRootCAs(transport, tlsCfg.CertificateAuthority)
	if err != nil {
		return nil, fmt.Errorf("setting client root CAs: %w", err)
	}

	client.Transport = transport

	return client, nil
}

func modifyTransportRootCAs(transport *http.Transport, certificateAuthority *string) error {
	if certificateAuthority == nil || len(*certificateAuthority) == 0 {
		return nil
	}

	if transport == nil {
		transport = http.DefaultTransport.(*http.Transport).Clone()
	}

	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{}
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM([]byte(*certificateAuthority)); !ok {
		return fmt.Errorf("certificate authority does not contain any valid PEM certificates")
	}

	transport.TLSClientConfig.RootCAs = caCertPool
	return nil
}
