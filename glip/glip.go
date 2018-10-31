package glip

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/grokify/gostor"
	"github.com/grokify/gotilla/net/anyhttp"
	ro "github.com/grokify/oauth2more/ringcentral"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultBotTokenPrefix = "rcBotExtId-"
)

type RcOAuthManager struct {
	client       *http.Client
	appCreds     ro.ApplicationCredentials
	gostorClient gostor.Client
	TokenPrefix  string
}

func NewRcOAuthManager(client *http.Client, appCreds ro.ApplicationCredentials) RcOAuthManager {
	return RcOAuthManager{
		client:   client,
		appCreds: appCreds}
}

func (h RcOAuthManager) HandleOAuthNetHttp(res http.ResponseWriter, req *http.Request) {
	h.HandleOAuthAny(anyhttp.NewResReqNetHttp(res, req))
}

func (h RcOAuthManager) HandleOAuthAny(aRes anyhttp.Response, aReq anyhttp.Request) {
	err := aReq.ParseForm()
	if err != nil {
		log.Warn("E_CANNOT_PARSE_RC_OAUTH_FORM")
		return
	}
	args := aReq.QueryArgs()
	code := strings.TrimSpace(args.GetString("code"))

	if len(code) == 0 {
		log.Warn("E_RC_OAUTH_FORM__NO_CODE")
		return
	}

	log.Info(">>>CODE>>>\n" + code + "\n<<<CODE<<<\n")
	rcToken, err := h.appCreds.Exchange(code)
	if err != nil {
		log.Warn("E_CANNOT_EXCHANGE_CODE_FOR_TOKEN")
		return
	}
	fmt.Printf("%v\n", rcToken)
	//authEndpoints := ru.NewEndpoint()

	rcBotTokenKey := "rcBotId-" + rcToken.OwnerID

	h.StoreToken(rcBotTokenKey, *rcToken)
}

func (h RcOAuthManager) StoreToken(key string, rcToken ro.RcToken) error {
	data, err := json.Marshal(rcToken)
	if err != nil {
		return err
	}
	return h.gostorClient.SetString(key, string(data))
}

func (h RcOAuthManager) GetToken(key string) (ro.RcToken, error) {
	data := strings.TrimSpace(h.gostorClient.GetOrEmptyString(key))
	if data == "" {
		return ro.RcToken{}, fmt.Errorf("Could not find token for %s", key)
	}
	var tok ro.RcToken
	return tok, json.Unmarshal([]byte(data), &tok)
}

/*
func HandleNetHTTPRingCentralCodeToToken(res http.ResponseWriter, req *http.Request) {
	reqBodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Warn(err)
	}
	log.Info(">>>BODY>>>\n" + string(reqBodyBytes) + "\n<<<BODY<<<\n")

	res.WriteHeader(http.StatusOK)
}
*/
