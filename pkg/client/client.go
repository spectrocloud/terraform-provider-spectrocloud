package client

import (
	"context"
	"errors"
	"log"
	"time"

	openapiclient "github.com/go-openapi/runtime/client"

	"github.com/go-openapi/strfmt"

	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	authC "github.com/spectrocloud/hapi/auth/client/v1"

	hashboardC "github.com/spectrocloud/hapi/hashboard/client/v1"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
	userC "github.com/spectrocloud/hapi/user/client/v1"
)

const (
	//UriTemplate    string = "%s:%s"
	authTokenInput string = "header"
	authTokenKey   string = "Authorization"
	authApiKey     string = "ApiKey"
)

var hubbleUri string

var AuthClient authC.ClientService

var schemes []string

var authToken *AuthToken
var tokenExpiry = 10 * time.Minute

//var hubbleRestClientLatch sync.Mutex

type AuthToken struct {
	token  *models.V1UserToken
	expiry time.Time
}

type V1Client struct {
	Ctx            context.Context
	email          string
	password       string
	apikey         string
	transportDebug bool
	retryAttempts  int
}

func New(hubbleHost, email, password, projectUID string, apikey string, transportDebug bool, retryAttempts int) *V1Client {
	ctx := context.Background()
	if projectUID != "" {
		ctx = GetProjectContextWithCtx(ctx, projectUID)
	}

	hubbleUri = hubbleHost
	schemes = []string{"https"}
	authHttpTransport := hapitransport.New(hubbleUri, "", schemes)
	authHttpTransport.RetryAttempts = 0
	//authHttpTransport.Debug = true
	AuthClient = authC.New(authHttpTransport, strfmt.Default)
	return &V1Client{ctx, email, password, apikey, transportDebug, retryAttempts}
}

func (h *V1Client) getNewAuthToken() (*AuthToken, error) {
	//httpClient, err := certs.GetHttpClient()
	//if err != nil {
	//	return nil, err
	//}
	authParam := authC.NewV1AuthenticateParams().
		WithBody(&models.V1AuthLogin{
			EmailID:  h.email,
			Password: strfmt.Password(h.password),
		})
	res, err := AuthClient.V1Authenticate(authParam)
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}

	if len(res.Payload.Authorization) == 0 {
		errMsg := "authorization auth token is empty in hubble payload"
		return nil, errors.New(errMsg)
	}

	authToken = &AuthToken{
		token:  res.Payload,
		expiry: time.Now().Add(tokenExpiry),
	}
	return authToken, nil
}

func GetProjectContextWithCtx(c context.Context, projectUid string) context.Context {
	return context.WithValue(c, hapitransport.CUSTOM_HEADERS, hapitransport.Values{
		HeaderMap: map[string]string{
			"ProjectUid": projectUid,
		}})
}

func (h *V1Client) getTransport() (*hapitransport.Runtime, error) {
	if h.apikey == "" && (authToken == nil || authToken.expiry.Before(time.Now())) {
		if tkn, err := h.getNewAuthToken(); err != nil {
			log.Fatal("Failed to get auth token ", err.Error())
			return nil, err
		} else {
			authToken = tkn
		}
	}

	httpTransport := hapitransport.New(hubbleUri, "", schemes)
	if h.apikey != "" {
		httpTransport.DefaultAuthentication = openapiclient.APIKeyAuth(authApiKey, authTokenInput, h.apikey)
	} else {
		httpTransport.DefaultAuthentication = openapiclient.APIKeyAuth(authTokenKey, authTokenInput, authToken.token.Authorization)
	}
	httpTransport.RetryAttempts = h.retryAttempts
	httpTransport.Debug = h.transportDebug
	return httpTransport, nil
}

// Clients
func (h *V1Client) GetClusterClient() (clusterC.ClientService, error) {
	httpTransport, err := h.getTransport()
	if err != nil {
		return nil, err
	}

	return clusterC.New(httpTransport, strfmt.Default), nil
}

func (h *V1Client) GetUserClient() (userC.ClientService, error) {
	httpTransport, err := h.getTransport()
	if err != nil {
		return nil, err
	}

	return userC.New(httpTransport, strfmt.Default), nil
}

func (h *V1Client) GetHashboard() (hashboardC.ClientService, error) {
	httpTransport, err := h.getTransport()
	if err != nil {
		return nil, err
	}

	return hashboardC.New(httpTransport, strfmt.Default), nil
}
