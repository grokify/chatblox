package chatblox

import (
	"log"
	"regexp"
	"strings"

	"github.com/grokify/mogo/encoding/jsonutil"
	hum "github.com/grokify/mogo/net/http/httputilmore"
	"github.com/grokify/mogo/regexp/regexputil"
	"github.com/grokify/mogo/type/stringsutil"
)

/*
	type EventResponse struct {
		StatusCode int               `json:"statusCode,omitempty"`
		Headers    map[string]string `json:"headers,omitempty"`
		Message    string            `json:"message,omitempty"`
	}

	func (er *EventResponse) ToJson() []byte {
		if len(er.Message) == 0 {
			er.Message = ""
		}
		msgJson, err := json.Marshal(er)
		if err != nil {
			return []byte(`{"statusCode":500,"message":"Cannot Marshal to JSON"}`)
		}
		return msgJson
	}
*/
type IntentRouter struct {
	Intents []Intent
}

func NewIntentRouter() IntentRouter {
	return IntentRouter{Intents: []Intent{}}
}

func (ir *IntentRouter) ProcessRequest(bot *Bot, glipPostEventInfo *GlipPostEventInfo) (*hum.ResponseInfo, error) {
	tryCmdsNotMatched := []string{}
	intentResponses := []*hum.ResponseInfo{}

	regexps := []*regexp.Regexp{
		regexp.MustCompile(`[^a-zA-Z0-9\-]+`),
		regexp.MustCompile(`\s+`)}

	tryCmdsLc := stringsutil.SliceCondenseRegexps(
		glipPostEventInfo.TryCommandsLc,
		regexps,
		" ")

Commands:
	for _, tryCmdLc := range tryCmdsLc {
		matched := false
		for _, intent := range ir.Intents {
			if intent.Type == MatchStringLowerCase {
				for _, try := range intent.Strings {
					if try == tryCmdLc {
						matched = true
						evtResp, err := intent.HandleIntent(bot, map[string]string{}, glipPostEventInfo)
						if err == nil {
							intentResponses = append(intentResponses, evtResp)
							break Commands
						}
					}
				}
			} else if intent.Type == MatchRegexpCapture {
				log.Print("TRY_CMD_REGEXP_CAPTURE_MULTI_USER")
				for _, try := range intent.Regexps {
					resMss := regexputil.FindStringSubmatchNamedMap(try, tryCmdLc)
					if len(resMss) > 0 {
						matched = true
						log.Printf("TRY_CMD_REGEXP_CAPTURE_MULTI_USER__MATCH_TRUE [%v]", resMss)
						evtResp, err := intent.HandleIntent(bot, resMss, glipPostEventInfo)
						if err == nil {
							intentResponses = append(intentResponses, evtResp)
							break Commands
						}
					}
				}
			}
		}
		if !matched {
			tryCmdsNotMatched = append(tryCmdsNotMatched, tryCmdLc)
		}
	}

	tryCmdsNotMatched = stringsutil.SliceCondenseRegexps(
		tryCmdsNotMatched,
		regexps,
		" ",
	)

	if len(tryCmdsNotMatched) > 0 {
		log.Print("TRY_CMDS_NOT_MATCHED " + jsonutil.MustMarshalString(tryCmdsNotMatched, true))
		glipPostEventInfo.TryCommandsLc = tryCmdsNotMatched
		for _, intent := range ir.Intents {
			if intent.Type == MatchAny {
				return intent.HandleIntent(bot, map[string]string{}, glipPostEventInfo)
			}
		}
	}

	if len(intentResponses) > 0 {
		return intentResponses[0], nil
	} else {
		return nil, nil
	}
}

func (ir *IntentRouter) ProcessRequestSingle(bot *Bot, textNoBotMention string, glipPostEventInfo *GlipPostEventInfo) (*hum.ResponseInfo, error) {
	textNoBotMention = strings.TrimSpace(textNoBotMention)
	textNoBotMentionLc := strings.ToLower(textNoBotMention)
	for _, intent := range ir.Intents {
		if intent.Type == MatchStringLowerCase {
			for _, try := range intent.Strings {
				if try == textNoBotMentionLc {
					return intent.HandleIntent(bot, map[string]string{}, glipPostEventInfo)
				}
			}
		} else if intent.Type == MatchRegexpCapture {
			log.Print("TRY_CMD_REGEXP_CAPTURE_SINGLE_USER")
			for _, try := range intent.Regexps {
				resMss := regexputil.FindStringSubmatchNamedMap(try, textNoBotMentionLc)
				if len(resMss) > 0 {
					return intent.HandleIntent(bot, resMss, glipPostEventInfo)
				}
			}
		} else if intent.Type == MatchAny {
			return intent.HandleIntent(bot, map[string]string{}, glipPostEventInfo)
		}
	}
	return &hum.ResponseInfo{}, nil
}

type IntentType int

const (
	MatchString IntentType = iota
	MatchStringLowerCase
	MatchRegexp
	MatchRegexpCapture
	MatchAny
)

type Intent struct {
	Type         IntentType
	Strings      []string
	Regexps      []*regexp.Regexp
	HandleIntent func(bot *Bot, matchResults map[string]string, glipPostEventInfo *GlipPostEventInfo) (*hum.ResponseInfo, error)
}
