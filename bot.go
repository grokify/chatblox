package chatblox

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/caarlos0/env/v6"
	rc "github.com/grokify/go-ringcentral-client/office/v1/client"
	ru "github.com/grokify/go-ringcentral-client/office/v1/util"
	"github.com/grokify/googleutil/sheetsutil/v4/sheetsmap"
	"github.com/grokify/mogo/encoding/jsonutil"
	hum "github.com/grokify/mogo/net/http/httputilmore"
	"github.com/grokify/mogo/type/stringsutil"
)

const ValidationTokenHeader = "Validation-Token"

type Bot struct {
	BotConfig         BotConfig
	IntentRouter      IntentRouter
	RingCentralClient *rc.APIClient
	GoogleClient      *http.Client
	SheetsMap         sheetsmap.SheetsMap
	SheetsMapMeta     sheetsmap.SheetsMap
}

type GlipPostEventInfo struct {
	PostEvent        *rc.GlipPostEvent
	GroupMemberCount int64
	CreatorInfo      *rc.GlipPersonInfo
	TryCommandsLc    []string
}

func (bot *Bot) Initialize() (hum.ResponseInfo, error) {
	botCfg := BotConfig{}
	err := env.Parse(&botCfg)
	if err != nil {
		log.Printf("Initialize Error: Cannot Parse Config: %v", err.Error())
		return hum.ResponseInfo{
			StatusCode: 500,
			Body:       fmt.Sprintf("Initialize Error: Cannot Parse Config: %v", err.Error()),
		}, err
	}
	botCfg.BotbloxCharQuoteLeft = CharQuoteLeft
	botCfg.BotbloxCharQuoteRight = CharQuoteRight
	bot.BotConfig = botCfg

	log.Printf("BOT_ID: %v", bot.BotConfig.RingCentralBotId)

	rcApiClient, err := GetRingCentralAPIClient(botCfg)
	if err != nil {
		log.Printf("Initialize Error: RC Client: %v", err.Error())
		return hum.ResponseInfo{
			StatusCode: 500,
			Body:       fmt.Sprintf("Initialize Error: RC Client: %v", err.Error()),
		}, err
	}
	bot.RingCentralClient = rcApiClient

	if 1 == 0 {
		googHttpClient, err := GetGoogleApiClient(botCfg)
		if err != nil {
			log.Printf("Initialize Error: Google Client: %v", err.Error())
			return hum.ResponseInfo{
				StatusCode: 500,
				Body:       fmt.Sprintf("Initialize Error: Google Client: %v", err.Error()),
			}, err
		}
		bot.GoogleClient = googHttpClient

		sm, err := GetSheetsMap(googHttpClient,
			bot.BotConfig.GoogleSpreadsheetID,
			bot.BotConfig.GoogleSheetTitleRecords)
		if err != nil {
			log.Printf("Initialize Error: Google Sheets: %v", err.Error())
			return hum.ResponseInfo{
				StatusCode: 500,
				Body:       fmt.Sprintf("Initialize Error: Google Sheets: %v", err.Error()),
			}, err
		}
		bot.SheetsMap = sm

		sm2, err := GetSheetsMap(googHttpClient,
			bot.BotConfig.GoogleSpreadsheetID,
			bot.BotConfig.GoogleSheetTitleMetadata)
		if err != nil {
			log.Printf("Initialize Error: Google Sheets: %v", err.Error())
			return hum.ResponseInfo{
				StatusCode: 500,
				Body:       fmt.Sprintf("Initialize Error: Google Sheets: %v", err.Error()),
			}, err
		}
		bot.SheetsMapMeta = sm2
	}

	return hum.ResponseInfo{
		StatusCode: 200,
		Body:       "Initialize success",
	}, nil
}

func (bot *Bot) HandleAwsLambda(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Print("Handling Lambda Request")
	log.Printf("REQ_BODY: %v", req.Body)
	/*
		vt := req.Header.Get(ValidationTokenHeader)
		if len(strings.TrimSpace(vt)) > 0 {
			res.Header().Set(ValidationTokenHeader, vt)
			return
		}
	*/
	/*
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    map[string]string{},
			Body:       `{"statusCode":200,"body":"Testing."}`,
		}, nil
	*/
	_, err := bot.Initialize()
	if err != nil {
		body := `{"statusCode":500,"body":"Cannot initialize."}`
		log.Print(body)
		evtResp := hum.ResponseInfo{
			StatusCode: 500,
			Body:       "Cannot initialize: " + err.Error(),
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{},
			Body:       string(evtResp.ToJSON()),
		}, nil
	}

	if vt, ok := req.Headers[ValidationTokenHeader]; ok {
		body := `{"statusCode":200}`
		log.Print(body)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{ValidationTokenHeader: vt},
			Body:       `{"statusCode":200}`,
		}, nil
	}
	evtResp, _ := bot.ProcessEvent([]byte(req.Body))

	awsRespBody := strings.TrimSpace(string(evtResp.ToJSON()))
	log.Printf("RESP_BODY: %v", awsRespBody)
	if len(awsRespBody) == 0 ||
		strings.Index(awsRespBody, "{") != 0 {
		awsRespBody = `{"statusCode":500}`
	}

	awsResp := events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{},
		Body:       awsRespBody}
	return awsResp, nil
}

func (bot *Bot) HandleNetHTTP(res http.ResponseWriter, req *http.Request) {
	// Check for RingCentral Validation-Token setup
	vt := req.Header.Get(ValidationTokenHeader)
	if len(strings.TrimSpace(vt)) > 0 {
		res.Header().Set(ValidationTokenHeader, vt)
		res.Header().Set("Content-Security-Policy", "default-src 'self'")
		res.Header().Set("Referrer-Policy", "origin-when-cross-origin, strict-origin-when-cross-origin")
		res.Header().Set("Vary", "Origin")
		res.Header().Set("X-Content-Type-Options", "nosniff")
		res.Header().Set("X-Frame-Options", "DENY")
		res.Header().Set("X-Permitted-Cross-Domain-Policies", "master-only")
		res.Header().Set("X-XSS-Protection", "1; mode=block")
		fmt.Fprint(res, "")
		return
	}
	_, err := bot.Initialize()
	if err != nil {
		log.Print(err) // Warn
	}

	if 1 == 1 {
		err := req.ParseForm()
		if err != nil {
			log.Print(err) // Warn
		}
		log.Print("WEBHOOK_ACCOUNT_ID")
		log.Print(req.FormValue("state"))
	}

	if 1 == 0 {
		rcApi := bot.RingCentralClient
		info, _, err := rcApi.CompanySettingsApi.LoadAccount(context.Background(), "~")
		if err != nil {
			log.Print(err)
		} else {
			bytes, _ := json.Marshal(info)
			log.Print("ACCOUNT_INFO")
			log.Print(string(bytes))
		}
	}

	reqBodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Print(err) // Warn
	}
	log.Print("REQ_BODY--------->")
	log.Print(string(reqBodyBytes))

	evtResp, err := bot.ProcessEvent(reqBodyBytes)

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	} else {
		res.WriteHeader(evtResp.StatusCode)
	}
}

func (bot *Bot) ProcessEvent(reqBodyBytes []byte) (*hum.ResponseInfo, error) {
	evt := &ru.Event{}
	err := json.Unmarshal(reqBodyBytes, evt)
	log.Print(string(reqBodyBytes))
	if err != nil {
		log.Printf("Request Bytes: %v", string(reqBodyBytes))    // Warn
		log.Printf("Cannot Unmarshal to Event: %s", err.Error()) // Warn
		return &hum.ResponseInfo{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf("400 Cannot Unmarshal to Event: %s", err.Error()),
		}, fmt.Errorf("JSON Unmarshal Error: %s", err.Error())
	}

	if !evt.IsEventType(ru.GlipPostEvent) {
		return &hum.ResponseInfo{
			StatusCode: http.StatusOK,
		}, nil
	}

	glipPostEvent, err := evt.GetGlipPostEventBody()
	if err != nil {
		log.Print(err) // Warn
		return &hum.ResponseInfo{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf("400 Cannot unmarshal to GlipPostEvent: %v", err.Error()),
		}, nil
	}
	log.Print(string(jsonutil.MustMarshal(glipPostEvent, true)))
	if (glipPostEvent.EventType != "PostAdded" &&
		glipPostEvent.EventType != "PostChanged") ||
		glipPostEvent.Type != "TextMessage" ||
		glipPostEvent.CreatorId == bot.BotConfig.RingCentralBotId {

		log.Print("POST_EVENT_TYPE_NOT_IN [PostAdded, TextMessage]")
		return &hum.ResponseInfo{
			StatusCode: http.StatusOK,
			Body:       "200 Not a relevant post: Not PostAdded|PostChanged && TextMessage",
		}, nil
	}

	glipApiUtil := ru.GlipApiUtil{ApiClient: bot.RingCentralClient}
	groupMemberCount, err := glipApiUtil.GlipGroupMemberCount(glipPostEvent.GroupId)
	if err != nil {
		groupMemberCount = -1
	}
	log.Print(fmt.Sprintf("GROUP_MEMBER_COUNT [%v]", groupMemberCount))

	info := ru.GlipInfoAtMentionOrGroupOfTwoInfo{
		PersonId:       bot.BotConfig.RingCentralBotId,
		PersonName:     bot.BotConfig.RingCentralBotName,
		FuzzyAtMention: bot.BotConfig.BotbloxRequestFuzzyAtMentionMatch,
		AtMentions:     glipPostEvent.Mentions,
		GroupId:        glipPostEvent.GroupId,
		TextRaw:        glipPostEvent.Text}

	log.Print("AT_MENTION_INPUT: " + string(jsonutil.MustMarshal(info, true)))
	log.Print("CONFIG: " + string(jsonutil.MustMarshal(bot.BotConfig, true)))

	atMentionedOrGroupOfTwo, err := glipApiUtil.AtMentionedOrGroupOfTwoFuzzy(info)

	if err != nil {
		log.Print("AT_MENTION_ERR: " + err.Error())
		return &hum.ResponseInfo{
			StatusCode: http.StatusBadRequest,
			Body:       "500 AtMentionedOrGroupOfTwo error",
		}, nil
	}
	if !atMentionedOrGroupOfTwo {
		log.Print("E_NO_MENTION")
		return &hum.ResponseInfo{
			StatusCode: http.StatusOK,
			Body:       "200 Not Mentioned in a Group != 2 members",
		}, nil
	}

	creator, resp, err := bot.RingCentralClient.GlipApi.LoadPerson(
		context.Background(), glipPostEvent.CreatorId)
	if err != nil {
		msg := fmt.Errorf("Glip API Load Person Error: %v", err.Error())
		log.Print(msg.Error()) // Warn
		return &hum.ResponseInfo{
			StatusCode: http.StatusInternalServerError,
			Body:       msg.Error()}, err
	} else if resp.StatusCode >= 300 {
		msg := fmt.Errorf("Glip API Status Error: %v", resp.StatusCode)
		log.Print(msg.Error()) // Warn
		return &hum.ResponseInfo{
			StatusCode: http.StatusInternalServerError,
			Body:       "500 " + msg.Error()}, err
	}

	name := strings.Join([]string{creator.FirstName, creator.LastName}, " ")
	email := creator.Email
	log.Printf("Poster [%v][%v]", name, email)

	log.Printf("TEXT_PREP [%v]", glipPostEvent.Text)
	//text := ru.StripAtMention(bot.BotConfig.RingCentralBotId, glipPostEvent.Text)
	text := ru.StripAtMentionAll(bot.BotConfig.RingCentralBotId,
		bot.BotConfig.RingCentralBotName,
		glipPostEvent.Text)
	texts := regexp.MustCompile(`[,\n]`).Split(strings.ToLower(text), -1)
	log.Print("TEXTS_1 " + jsonutil.MustMarshalString(texts, true))
	log.Print("TEXTS_2 " + jsonutil.MustMarshalString(stringsutil.SliceTrimSpace(texts, true), true))

	postEventInfo := GlipPostEventInfo{
		PostEvent:        glipPostEvent,
		GroupMemberCount: groupMemberCount,
		CreatorInfo:      &creator,
		TryCommandsLc:    texts}

	evtResp, err := bot.IntentRouter.ProcessRequest(bot, &postEventInfo)
	return evtResp, err
}

/*
func (bot *Bot) SendGlipPosts(glipPostEventInfo *GlipPostEventInfo, reqBodies []rc.GlipCreatePost) (*hum.ResponseInfo, error) {
	res := &hum.ResponseInfo{}
	var err error

	for _, reqBody := range reqBodies {
		res, err = bot.SendGlipPost(GlipPostEventInfo, reqBody)
		if err != nil {
			return res, err
		}
	}
	return res, err
}
*/
func (bot *Bot) SendGlipPost(glipPostEventInfo *GlipPostEventInfo, reqBody rc.GlipCreatePost) (*hum.ResponseInfo, error) {
	if bot.BotConfig.BotbloxResponseAutoAtMention && glipPostEventInfo.GroupMemberCount > 2 {
		atMentionId := strings.TrimSpace(glipPostEventInfo.PostEvent.CreatorId)
		reqBody.Text = ru.PrefixAtMentionUnlessMentioned(atMentionId, reqBody.Text)
	}

	reqBody.Text = bot.BotConfig.AppendPostSuffix(reqBody.Text)

	_, resp, err := bot.RingCentralClient.GlipApi.CreatePost(
		context.Background(), glipPostEventInfo.PostEvent.GroupId, reqBody,
	)
	if err != nil {
		msg := fmt.Errorf("Cannot Create Post: [%v]", err.Error())
		log.Print(msg.Error()) // Warn
		return &hum.ResponseInfo{
			StatusCode: http.StatusInternalServerError,
			Body:       "500 " + msg.Error(),
		}, err
	} else if resp.StatusCode >= 300 {
		msg := fmt.Errorf("Cannot Create Post, API Status [%v]", resp.StatusCode)
		log.Print(msg.Error()) // Warn
		return &hum.ResponseInfo{
			StatusCode: http.StatusInternalServerError,
			Body:       "500 " + msg.Error(),
		}, err
	}
	return &hum.ResponseInfo{}, nil
}
