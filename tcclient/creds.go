package tcclient

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/taskcluster/jsonschema2go/text"
	"github.com/taskcluster/slugid-go/slugid"
	"github.com/tent/hawk-go"
)

type Credentials interface {
	fmt.Stringer
	ConfigureAuth(httpRequest *http.Request) error
	getExtField() (string, error)
	HawkCredentials() *hawk.Credentials
}

// Credentials for communicating with taskcluster services (Queue, Index,
// Scheduler, etc).
type TemporaryCredentials struct {
	temporaryCredentials
}

type temporaryCredentials struct {
	ClientID    string `json:"clientId"`
	AccessToken string `json:"accessToken"`
	// Certificate used only for temporary credentials
	Certificate string `json:"certificate"`
	// AuthorizedScopes if set to nil, is ignored. Otherwise, it should be a
	// subset of the scopes that the ClientId already has, and restricts the
	// Credentials to only having these scopes. This is useful when performing
	// actions on behalf of a client which has more restricted scopes. Setting
	// to nil is not the same as setting to an empty array. If AuthorizedScopes
	// is set to an empty array rather than nil, this is equivalent to having
	// no scopes at all.
	// See http://docs.taskcluster.net/auth/authorized-scopes
	AuthorizedScopes []string `json:"authorizedScopes"`
}

type EmptyCredentials struct {
}

func NewPermanentCredentials(clientID string, accessToken string, authorizedScopes []string) *PermanentCredentials {
	return &PermanentCredentials{
		permanentCredentials: permanentCredentials{
			ClientID:         clientID,
			AccessToken:      accessToken,
			AuthorizedScopes: authorizedScopes,
		},
	}
}

func NewTemporaryCredentials(clientID string, accessToken string, certificate string, authorizedScopes []string) *TemporaryCredentials {
	return &TemporaryCredentials{
		temporaryCredentials: temporaryCredentials{
			ClientID:         clientID,
			AccessToken:      accessToken,
			Certificate:      certificate,
			AuthorizedScopes: authorizedScopes,
		},
	}
}

type PermanentCredentials struct {
	permanentCredentials
}

type permanentCredentials struct {
	ClientID    string `json:"clientId"`
	AccessToken string `json:"accessToken"`
	// Certificate used only for temporary credentials
	Certificate string `json:"certificate"`
	// AuthorizedScopes if set to nil, is ignored. Otherwise, it should be a
	// subset of the scopes that the ClientId already has, and restricts the
	// Credentials to only having these scopes. This is useful when performing
	// actions on behalf of a client which has more restricted scopes. Setting
	// to nil is not the same as setting to an empty array. If AuthorizedScopes
	// is set to an empty array rather than nil, this is equivalent to having
	// no scopes at all.
	// See http://docs.taskcluster.net/auth/authorized-scopes
	AuthorizedScopes []string `json:"authorizedScopes"`
}

func (creds *PermanentCredentials) String() string {
	return fmt.Sprintf(
		"Permanent credentials: ClientId: %q\nAccessToken: %q\nAuthorizedScopes: %q",
		creds.ClientID,
		text.StarOut(creds.AccessToken),
		creds.AuthorizedScopes,
	)
}

func (creds *TemporaryCredentials) String() string {
	return fmt.Sprintf(
		"Temporary credentials: ClientId: %q\nAccessToken: %q\nCertificate: %q\nAuthorizedScopes: %q",
		creds.ClientID,
		text.StarOut(creds.AccessToken),
		text.StarOut(creds.Certificate),
		creds.AuthorizedScopes,
	)
}

// BaseClient contains authentication credentials, and a service endpoint,
// which are required for all HTTP operations.
type BaseClient struct {
	Credentials Credentials
	// The URL of the API endpoint to hit.
	// Use "https://auth.taskcluster.net/v1" for production.
	// Please note calling auth.New(clientId string, accessToken string) is an
	// alternative way to create an Auth object with BaseURL set to production.
	BaseURL string
	// Whether authentication is enabled (e.g. set to 'false' when using taskcluster-proxy)
	// Please note calling auth.New(clientId string, accessToken string) is an
	// alternative way to create an Auth object with Authenticate set to true.
	Authenticate bool
}

func (creds PermanentCredentials) HawkCredentials() *hawk.Credentials {
	return &hawk.Credentials{
		ID:   creds.ClientID,
		Key:  creds.AccessToken,
		Hash: sha256.New,
	}
}

func (creds TemporaryCredentials) HawkCredentials() *hawk.Credentials {
	return &hawk.Credentials{
		ID:   creds.ClientID,
		Key:  creds.AccessToken,
		Hash: sha256.New,
	}
}

// Certificate represents the certificate attached to hawk's ext property as
// per http://docs.taskcluster.net/auth/temporary-credentials/
type Certificate struct {
	Version   int      `json:"version"`
	Scopes    []string `json:"scopes"`
	Start     int64    `json:"start"`
	Expiry    int64    `json:"expiry"`
	Seed      string   `json:"seed"`
	Signature string   `json:"signature"`
	Issuer    string   `json:"issuer,omitempty"`
}

// CreateNamedTemporaryCredentials generates temporary credentials from permanent
// credentials, valid for the given duration, starting immediately.  The
// temporary credentials' scopes must be a subset of the permanent credentials'
// scopes. The duration may not be more than 31 days. Any authorized scopes of
// the permanent credentials will be passed through as authorized scopes to the
// temporary credentials, but will not be restricted via the certificate.
//
// See http://docs.taskcluster.net/auth/temporary-credentials/
func (permCreds *PermanentCredentials) CreateNamedTemporaryCredentials(tempClientID string, duration time.Duration, scopes ...string) (tempCreds *TemporaryCredentials, err error) {
	if duration > 31*24*time.Hour {
		return nil, errors.New("Temporary credentials must expire within 31 days; however a duration of " + duration.String() + " was specified to (*tcclient.BaseClient).CreateTemporaryCredentials(...) method")
	}

	now := time.Now()
	start := now.Add(time.Minute * -5) // subtract 5 min for clock drift
	expiry := now.Add(duration)

	if permCreds.ClientID == "" {
		return nil, errors.New("Temporary credentials cannot be created from credentials that have an empty ClientId")
	}
	if permCreds.AccessToken == "" {
		return nil, errors.New("Temporary credentials cannot be created from credentials that have an empty AccessToken")
	}
	if permCreds.Certificate != "" {
		return nil, errors.New("Temporary credentials cannot be created from temporary credentials, only from permanent credentials")
	}

	cert := &Certificate{
		Version:   1,
		Scopes:    scopes,
		Start:     start.UnixNano() / 1e6,
		Expiry:    expiry.UnixNano() / 1e6,
		Seed:      slugid.V4() + slugid.V4(),
		Signature: "", // gets set in updateSignature() method below
	}
	// include the issuer iff this is a named credential
	if tempClientID != "" {
		cert.Issuer = permCreds.ClientID
	}

	cert.updateSignature(permCreds.AccessToken, tempClientID)

	certBytes, err := json.Marshal(cert)
	if err != nil {
		return
	}

	tempAccessToken, err := generateTemporaryAccessToken(permCreds.AccessToken, cert.Seed)
	if err != nil {
		return
	}

	tempCreds = NewTemporaryCredentials(
		permCreds.ClientID,
		tempAccessToken,
		string(certBytes),
		permCreds.AuthorizedScopes,
	)
	if tempClientID != "" {
		tempCreds.ClientID = tempClientID
	}

	return
}

// CreateTemporaryCredentials is an alias for CreateNamedTemporaryCredentials
// with an empty name.
func (permCreds *PermanentCredentials) CreateTemporaryCredentials(duration time.Duration, scopes ...string) (tempCreds *TemporaryCredentials, err error) {
	return permCreds.CreateNamedTemporaryCredentials("", duration, scopes...)
}

func (cert *Certificate) updateSignature(accessToken string, tempClientID string) (err error) {
	lines := []string{"version:" + strconv.Itoa(cert.Version)}
	// iff this is a named credential, include clientId and issuer
	if cert.Issuer != "" {
		lines = append(lines,
			"clientId:"+tempClientID,
			"issuer:"+cert.Issuer,
		)
	}
	lines = append(lines,
		"seed:"+cert.Seed,
		"start:"+strconv.FormatInt(cert.Start, 10),
		"expiry:"+strconv.FormatInt(cert.Expiry, 10),
		"scopes:",
	)
	lines = append(lines, cert.Scopes...)
	hash := hmac.New(sha256.New, []byte(accessToken))
	text := strings.Join(lines, "\n")
	_, err = hash.Write([]byte(text))
	if err != nil {
		return err
	}
	cert.Signature = base64.StdEncoding.EncodeToString(hash.Sum([]byte{}))
	return
}

func generateTemporaryAccessToken(permAccessToken, seed string) (tempAccessToken string, err error) {
	hash := hmac.New(sha256.New, []byte(permAccessToken))
	_, err = hash.Write([]byte(seed))
	if err != nil {
		return "", err
	}
	tempAccessToken = strings.TrimRight(base64.URLEncoding.EncodeToString(hash.Sum([]byte{})), "=")
	return
}

// Cert attempts to parse the (string) certificate of the credentials as JSON
// and return it as a Certificate object. If the certificate string is empty
// (e.g. in the case of permanent credentials) then a nil pointer is returned.
// If the certificate is a non-empty string which cannot be parsed, an error is
// returned, and the returned certificate is a pointer to an empty Certificate,
// rather than a nil pointer.
func (creds *TemporaryCredentials) Cert() (cert *Certificate, err error) {
	if creds.Certificate != "" {
		cert = new(Certificate)
		err = json.Unmarshal([]byte(creds.Certificate), cert)
	}
	return
}
