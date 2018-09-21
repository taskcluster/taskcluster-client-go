// The following code is AUTO-GENERATED. Please DO NOT edit.
// To update this generated code, run the following command:
// in the /codegenerator/model subdirectory of this project,
// making sure that `${GOPATH}/bin` is in your `PATH`:
//
// go install && go generate
//
// This package was generated from the schema defined at
// https://references.taskcluster.net/login/v1/api.json

// The Login service serves as the interface between external authentication
// systems and Taskcluster credentials.
//
// See: https://docs.taskcluster.net/reference/core/login/api-docs
//
// How to use this package
//
// First create a Login object:
//
//  login := tclogin.New(nil)
//
// and then call one or more of login's methods, e.g.:
//
//  err := login.Ping(.....)
//
// handling any errors...
//
//  if err != nil {
//  	// handle error...
//  }
//
// Taskcluster Schema
//
// The source code of this go package was auto-generated from the API definition at
// https://references.taskcluster.net/login/v1/api.json together with the input and output schemas it references, downloaded on
// Tue, 18 Sep 2018 at 16:23:00 UTC. The code was generated
// by https://github.com/taskcluster/taskcluster-client-go/blob/master/build.sh.
package tclogin

import (
	"net/url"

	tcclient "github.com/taskcluster/taskcluster-client-go"
)

type Login tcclient.Client

// New returns a Login client, configured to run against production. Pass in
// nil to create a client without authentication. The
// returned client is mutable, so returned settings can be altered.
//
//  login := tclogin.New(nil)                              // client without authentication
//  login.BaseURL = "http://localhost:1234/api/Login/v1"   // alternative API endpoint (production by default)
//  err := login.Ping(.....)                               // for example, call the Ping(.....) API endpoint (described further down)...
//  if err != nil {
//          // handle errors...
//  }
func New(credentials *tcclient.Credentials) *Login {
	return &Login{
		Credentials: credentials,
		Service:     "login",
		Version:     "v1",
	}
}

// NewFromEnv returns a Login client with credentials taken from the environment variables:
//
//  TASKCLUSTER_CLIENT_ID
//  TASKCLUSTER_ACCESS_TOKEN
//  TASKCLUSTER_CERTIFICATE
//  TASKCLUSTER_ROOT_URL
//
// If environment variable TASKCLUSTER_ROOT_URL is empty string or not set,
// https://taskcluster.net will be assumed.
//
// If environment variable TASKCLUSTER_CLIENT_ID is empty string or not set,
// authentication will be disabled.
func NewFromEnv() *Login {
	return &Login{
		Credentials: tcclient.CredentialsFromEnvVars(),
		Service:     "login",
		Version:     "v1",
	}
}

// Respond without doing anything.
// This endpoint is used to check that the service is up.
//
// See https://docs.taskcluster.net/reference/core/login/api-docs#ping
func (login *Login) Ping() error {
	cd := tcclient.Client(*login)
	_, _, err := (&cd).APICall(nil, "GET", "/ping", nil, nil)
	return err
}

// Stability: *** EXPERIMENTAL ***
//
// Given an OIDC `access_token` from a trusted OpenID provider, return a
// set of Taskcluster credentials for use on behalf of the identified
// user.
//
// This method is typically not called with a Taskcluster client library
// and does not accept Hawk credentials. The `access_token` should be
// given in an `Authorization` header:
// ```
// Authorization: Bearer abc.xyz
// ```
//
// The `access_token` is first verified against the named
// :provider, then passed to the provider's APIBuilder to retrieve a user
// profile. That profile is then used to generate Taskcluster credentials
// appropriate to the user. Note that the resulting credentials may or may
// not include a `certificate` property. Callers should be prepared for either
// alternative.
//
// The given credentials will expire in a relatively short time. Callers should
// monitor this expiration and refresh the credentials if necessary, by calling
// this endpoint again, if they have expired.
//
// See https://docs.taskcluster.net/reference/core/login/api-docs#oidcCredentials
func (login *Login) OidcCredentials(provider string) (*CredentialsResponse, error) {
	cd := tcclient.Client(*login)
	responseObject, _, err := (&cd).APICall(nil, "GET", "/oidc-credentials/"+url.QueryEscape(provider), new(CredentialsResponse), nil)
	return responseObject.(*CredentialsResponse), err
}
