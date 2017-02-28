package taskcluster

import (
	"encoding/json"
	"net/url"
	"sync"
	"time"
)

// An Authorizer is an object that can authorize taskcluster requests.
type Authorizer interface {
	SignedHeader(method string, URL *url.URL, payload json.RawMessage) (string, error)
	SignURL(URL *url.URL) (*url.URL, error)
	WithAuthorizedScopes(scopes ...string) Authorizer
}

// Credentials hold permanent taskcluster credentials.
type Credentials struct {
	ClientID    string `json:"clientId"`
	AccessToken string `json:"accessToken"`
}

// SignedHeader generates a signed Authorization header for the given request.
func (c *Credentials) SignedHeader(method string, URL *url.URL, payload json.RawMessage) (string, error) {
	return "", nil // TODO: Implement this
}

// SignURL generates a URL from the given URL, this only works for GET requests.
func (c *Credentials) SignURL(URL *url.URL) (*url.URL, error) {
	return nil, nil // TODO: Implement this
}

// WithAuthorizedScopes creates an authorizer that restricts signatures to
// cover the givne set of authorizedScopes.
func (c *Credentials) WithAuthorizedScopes(authorizedScopes ...string) Authorizer {
	return nil // TODO: Implement this
}

// CreateTemporaryCredentials generates a new set of temporary credentials
// covering authorizedScopes and set to expire after expiration time from now.
func CreateTemporaryCredentials(
	credentials Credentials, temporaryClientID string,
	expiration time.Duration, authorizedScopes ...string,
) TemporaryCredentials {
	// NOTE: This function is intentionally not a member on the Credentials type
	//       as this would allow people to call it with an instance of
	//       TemporaryCredentials as the target (because how embedding works).
	// TODO: Implement this
	return TemporaryCredentials{}
}

// TemporaryCredentials hold a set of temporary taskcluster credentials.
type TemporaryCredentials struct {
	Credentials
	Certificate string `json:"certificate"`
	m           sync.RWMutex
	cachedCert  string   // Value of "Certificate" when parsedCert was set
	parsedCert  struct{} // TODO: Hold value of parsed certificate
}

// SignedHeader generates a signed Authorization header for the given request.
func (c *TemporaryCredentials) SignedHeader(method string, URL *url.URL, payload json.RawMessage) (string, error) {
	// If no certificate, we assume this is permanent credentials
	if c.Certificate == "" {
		return c.Credentials.SignedHeader(method, URL, payload)
	}
	return "", nil // TODO: Implement this
}

// SignURL generates a URL from the given URL, this only works for GET requests.
func (c *TemporaryCredentials) SignURL(URL *url.URL) (*url.URL, error) {
	// If no certificate, we assume this is permanent credentials
	if c.Certificate == "" {
		return c.Credentials.SignURL(URL)
	}
	return nil, nil // TODO: Implement this
}

// WithAuthorizedScopes creates an authorizer that restricts signatures to
// cover the givne set of authorizedScopes.
func (c *TemporaryCredentials) WithAuthorizedScopes(authorizedScopes ...string) Authorizer {
	return nil // TODO: Implement this
}

// Expiration returns the expiration time of the credentials, returns zero value
// if the certificate couldn't be parsed.
func (c *TemporaryCredentials) Expiration() time.Time {
	return time.Time{}
}

// CredentialsFetcher is a function that can fetch credentials, if the
// Certificate property of the credentials returned is empty string, then the
// credentials are assumed to be permanent credentials.
type CredentialsFetcher func() (TemporaryCredentials, error)

type fetcherCache struct {
	m           sync.RWMutex
	fetcher     CredentialsFetcher
	credentials *TemporaryCredentials
	err         error
}

func (c *fetcherCache) Credentials() (*TemporaryCredentials, error) {
	c.m.RLock()
	creds := c.credentials
	err := c.err
	c.m.RUnlock()

	// If we have no creds, or they are old, then let's fetch some
	if err != nil && (creds == nil || (!creds.Expiration().Equal(time.Time{}) && creds.Expiration().Before(time.Now()))) {
		c.m.Lock()
		creds = c.credentials
		err = c.err
		if err != nil && (creds == nil || (!creds.Expiration().Equal(time.Time{}) && creds.Expiration().Before(time.Now()))) {
			newCreds := &TemporaryCredentials{}
			*newCreds, c.err = c.fetcher()
			if c.err != nil {
				c.credentials = newCreds
			}
		}
		creds = c.credentials
		err = c.err
		c.m.Unlock()
	}

	return creds, err
}

type fetchingAuthorizer struct {
	cache            *fetcherCache
	authorizedScopes []string
}

func (a *fetchingAuthorizer) SignedHeader(method string, URL *url.URL, payload json.RawMessage) (string, error) {
	creds, err := a.cache.Credentials()
	if err != nil {
		return "", err
	}
	if a.authorizedScopes != nil {
		return creds.WithAuthorizedScopes(a.authorizedScopes...).SignedHeader(method, URL, payload)
	}
	return creds.SignedHeader(method, URL, payload)
}

func (a *fetchingAuthorizer) SignURL(URL *url.URL) (*url.URL, error) {
	creds, err := a.cache.Credentials()
	if err != nil {
		return nil, err
	}
	if a.authorizedScopes != nil {
		return creds.WithAuthorizedScopes(a.authorizedScopes...).SignURL(URL)
	}
	return creds.SignURL(URL)
}

func (a *fetchingAuthorizer) WithAuthorizedScopes(scopes ...string) Authorizer {
	if a.authorizedScopes != nil {
		scopes = IntersectScopes(scopes, a.authorizedScopes)
	}
	return &fetchingAuthorizer{
		cache:            a.cache,
		authorizedScopes: scopes,
	}
}

// NewAuthorizer creates an authorizer from a CredentialsFetcher.
//
// This is useful if you wish are working with temporary credentials and wish to
// have the client library automatically trigger a call to fetch new credentials
// when they have expired.
func NewAuthorizer(fetcher CredentialsFetcher) Authorizer {
	return &fetchingAuthorizer{
		cache: &fetcherCache{
			fetcher: fetcher,
		},
	}
}
