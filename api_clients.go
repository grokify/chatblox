package chatblox

import (
	"context"
	"errors"

	"fmt"
	"net/http"
	"strings"

	rc "github.com/grokify/go-ringcentral-client/office/v1/client"
	ru "github.com/grokify/go-ringcentral-client/office/v1/util"
	om "github.com/grokify/goauth"
	gu "github.com/grokify/goauth/google"
	"github.com/grokify/mogo/config"
	log "github.com/sirupsen/logrus"

	// "github.com/algolia/algoliasearch-client-go/algoliasearch"
	algoliasearch "github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/grokify/algoliautil"

	"github.com/grokify/googleutil/sheetsutil/v4/sheetsmap"
	"google.golang.org/api/sheets/v4"
)

const (
	CharQuoteLeft  = "“"
	CharQuoteRight = "”"
)

type BotConfig struct {
	Port                              int64  `env:"PORT"`
	BotbloxRequestFuzzyAtMentionMatch bool   `env:"CHATBLOX_REQUEST_FUZZY_AT_MENTION_MATCH"`
	BotbloxResponseAutoAtMention      bool   `env:"CHATBLOX_RESPONSE_AUTO_AT_MENTION"`
	BotbloxPostSuffix                 string `env:"CHATBLOX_POST_SUFFIX"`
	BotbloxCharQuoteLeft              string `env:"CHATBLOX_CHAR_QUOTE_LEFT"`
	BotbloxCharQuoteRight             string `env:"CHATBLOX_CHAR_QUOTE_RIGHT"`
	RingCentralToken                  string `env:"RINGCENTRAL_TOKEN"`
	RingCentralTokenJSON              string `env:"RINGCENTRAL_TOKEN_JSON"`
	RingCentralServerURL              string `env:"RINGCENTRAL_SERVER_URL"`
	RingCentralWebhookDefinitionJSON  string `env:"RINGCENTRAL_WEBHOOK_DEFINITION_JSON"`
	RingCentralBotId                  string `env:"RINGCENTRAL_BOT_ID"`
	RingCentralBotName                string `env:"RINGCENTRAL_BOT_NAME"`
	GoogleSvcAccountJWT               string `env:"GOOGLE_SERVICE_ACCOUNT_JWT"`
	GoogleSpreadsheetId               string `env:"GOOGLE_SPREADSHEET_ID"`
	GoogleSheetTitleRecords           string `env:"GOOGLE_SHEET_TITLE_RECORDS"`
	GoogleSheetTitleMetadata          string `env:"GOOGLE_SHEET_TITLE_METADATA"`
	AlgoliaAppCredentialsJSON         string `env:"ALGOLIA_APP_CREDENTIALS_JSON"`
	AlgoliaIndex                      string `env:"ALGOLIA_INDEX"`
}

func (ac *BotConfig) Inflate() {
	ac.RingCentralToken = config.JoinEnvNumbered("RINGCENTRAL_TOKEN", "", 2, true)
	log.Info(fmt.Sprintf("TOKEN_TOKEN_TOKEN [%v]\n", ac.RingCentralToken))
}

func (ac *BotConfig) AppendPostSuffix(s string) string {
	suffix := strings.TrimSpace(ac.BotbloxPostSuffix)
	if len(suffix) > 0 {
		return s + " " + suffix
	}
	return s
}

func (ac *BotConfig) Quote(s string) string {
	return ac.BotbloxCharQuoteLeft + strings.TrimSpace(s) + ac.BotbloxCharQuoteRight

}

func GetAlgoliaApiClient(botConfig BotConfig) (*algoliasearch.Client, error) {
	client, err := algoliautil.NewClientJSON(
		[]byte(botConfig.AlgoliaAppCredentialsJSON))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func GetRingCentralApiClient(botConfig BotConfig) (*rc.APIClient, error) {
	botConfig.Inflate()
	//fmt.Println(botConfig.RingCentralTokenJSON)
	/*
		rcHttpClient, err := om.NewClientTokenJSON(
			context.Background(),
			[]byte(botConfig.RingCentralTokenJSON))
	*/
	if len(strings.TrimSpace(botConfig.RingCentralToken)) <= 0 {
		return nil, errors.New("E_NO_RINGCENTRAL_TOKEN")
	}
	rcHttpClient := om.NewClientToken(
		om.TokenBearer,
		botConfig.RingCentralToken,
		false)
	/*
		url := "https://platform.ringcentral.com/restapi/v1.0/glip/groups"
		url = "https://platform.ringcentral.com/restapi/v1.0/subscription"

		resp, err := rcHttpClient.Get(url)
		if err != nil {
			log.Fatal(err)
		} else if resp.StatusCode >= 300 {
			log.Fatal(fmt.Errorf("API Error %v", resp.StatusCode))
		}
	*/
	return ru.NewApiClientHttpClientBaseURL(
		rcHttpClient, botConfig.RingCentralServerURL,
	)
}

func GetGoogleApiClient(botConfig BotConfig) (*http.Client, error) {
	jwtString := botConfig.GoogleSvcAccountJWT
	if len(jwtString) <= 0 {
		return nil, fmt.Errorf("No JWT")
	}

	return gu.NewClientFromJWTJSON(
		context.TODO(),
		[]byte(jwtString),
		sheets.DriveScope,
		sheets.SpreadsheetsScope)
}

func GetSheetsMap(googClient *http.Client, spreadsheetId string, sheetTitle string) (sheetsmap.SheetsMap, error) {
	sm, err := sheetsmap.NewSheetsMapTitle(googClient, spreadsheetId, sheetTitle)
	if err != nil {
		return sm, err
	}
	err = sm.FullRead()
	return sm, err
}
