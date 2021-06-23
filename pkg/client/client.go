package client

import (
	"context"
	"errors"
	"time"

	openapiclient "github.com/go-openapi/runtime/client"
	"github.com/prometheus/common/log"

	"github.com/go-openapi/strfmt"

	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	authC "github.com/spectrocloud/hapi/auth/client/v1alpha1"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1alpha1"
	userC "github.com/spectrocloud/hapi/user/client/v1alpha1"
)

const (
	//UriTemplate    string = "%s:%s"
	authTokenInput string = "header"
	authTokenKey   string = "Authorization"
)

var hubbleUri string

var AuthClient authC.ClientService

var schemes []string

var authToken *AuthToken
var tokenExpiry = 10 * time.Minute

//var hubbleRestClientLatch sync.Mutex

type AuthToken struct {
	token  *models.V1alpha1UserToken
	expiry time.Time
}

type V1alpha1Client struct {
	ctx      context.Context
	email    string
	password string
}

func New(hubbleHost, email, password, projectUID string) *V1alpha1Client {
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
	return &V1alpha1Client{ctx, email, password}
}

func (h *V1alpha1Client) getNewAuthToken() (*AuthToken, error) {
	//httpClient, err := certs.GetHttpClient()
	//if err != nil {
	//	return nil, err
	//}
	authParam := authC.NewV1alpha1AuthenticateParams().
		WithBody(&models.V1alpha1AuthLogin{
			EmailID:  h.email,
			Password: strfmt.Password(h.password),
		})
	res, err := AuthClient.V1alpha1Authenticate(authParam)
	if err != nil {
		log.Error("Error", err)
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

func (h *V1alpha1Client) getTransport() (*hapitransport.Runtime, error) {
	if authToken == nil || authToken.expiry.Before(time.Now()) {
		if tkn, err := h.getNewAuthToken(); err != nil {
			log.Error("Failed to get auth token ", err)
			return nil, err
		} else {
			authToken = tkn
		}
	}

	httpTransport := hapitransport.New(hubbleUri, "", schemes)
	httpTransport.DefaultAuthentication = openapiclient.APIKeyAuth(authTokenKey, authTokenInput, authToken.token.Authorization)
	httpTransport.RetryAttempts = 0
	//httpTransport.Debug = true
	return httpTransport, nil
}

// Clients
func (h *V1alpha1Client) getClusterClient() (clusterC.ClientService, error) {
	httpTransport, err := h.getTransport()
	if err != nil {
		return nil, err
	}

	return clusterC.New(httpTransport, strfmt.Default), nil
}

func (h *V1alpha1Client) getUserClient() (userC.ClientService, error) {
	httpTransport, err := h.getTransport()
	if err != nil {
		return nil, err
	}

	return userC.New(httpTransport, strfmt.Default), nil
}
