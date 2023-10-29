package glip

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/grokify/goauth"
	ro "github.com/grokify/goauth/ringcentral"
	"github.com/grokify/gohttp/anyhttp"
	"github.com/grokify/gostor"
	"golang.org/x/oauth2"
)

const (
	DefaultBotTokenPrefix = "rcBotExtId-"
)

type RcOAuthManager struct {
	client       *http.Client
	appCreds     goauth.CredentialsOAuth2
	gostorClient gostor.Client
	TokenPrefix  string
}

func NewRcOAuthManager(client *http.Client, appCreds goauth.CredentialsOAuth2) RcOAuthManager {
	return RcOAuthManager{
		client:   client,
		appCreds: appCreds}
}

func (h RcOAuthManager) HandleOAuthNetHTTP(res http.ResponseWriter, req *http.Request) {
	h.HandleOAuthAny(anyhttp.NewResReqNetHTTP(res, req))
}

func (h RcOAuthManager) HandleOAuthAny(aRes anyhttp.Response, aReq anyhttp.Request) {
	err := aReq.ParseForm()
	if err != nil {
		log.Print("E_CANNOT_PARSE_RC_OAUTH_FORM") // Warn
		return
	}
	args := aReq.QueryArgs()
	code := strings.TrimSpace(args.GetString("code"))

	if len(code) == 0 {
		log.Print("E_RC_OAUTH_FORM__NO_CODE") // Warn
		return
	}

	log.Print(">>>CODE>>>\n" + code + "\n<<<CODE<<<\n")
	rcToken, err := h.appCreds.Exchange(context.Background(), code, map[string][]string{})
	if err != nil {
		log.Print("E_CANNOT_EXCHANGE_CODE_FOR_TOKEN") // Warn
		return
	}
	fmt.Printf("%v\n", rcToken)
	//authEndpoints := ru.NewEndpoint()

	ownerID := rcToken.Extra("owner_id").(string)

	rcBotTokenKey := "rcBotId-" + ownerID //  rcToken.OwnerID

	err = h.StoreToken(rcBotTokenKey, *rcToken)
	if err != nil {
		log.Printf("cannot store token (%s)", err.Error()) // Warn
		return
	}
}

func (h RcOAuthManager) StoreToken(key string, rcToken oauth2.Token) error {
	data, err := json.Marshal(rcToken)
	if err != nil {
		return err
	}
	return h.gostorClient.SetString(key, string(data))
}

func (h RcOAuthManager) GetToken(key string) (ro.RcToken, error) {
	data := strings.TrimSpace(h.gostorClient.GetOrEmptyString(key))
	if data == "" {
		return ro.RcToken{}, fmt.Errorf("could not find token for key (%s)", key)
	}
	var tok ro.RcToken
	return tok, json.Unmarshal([]byte(data), &tok)
}

/*
func HandleNetHTTPRingCentralCodeToToken(res http.ResponseWriter, req *http.Request) {
	reqBodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Warn(err)
	}
	log.Info(">>>BODY>>>\n" + string(reqBodyBytes) + "\n<<<BODY<<<\n")

	res.WriteHeader(http.StatusOK)
}
*/
