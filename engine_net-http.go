package chatblox

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/grokify/mogo/log/logutil"
	"github.com/grokify/mogo/net/http/httputilmore"
)

func ServeNetHTTP(intentRouter IntentRouter, mux *http.ServeMux) {
	bot := Bot{}
	_, err := bot.Initialize()
	logutil.FatalErr(err)
	bot.IntentRouter = intentRouter

	//mux := http.NewServeMux()
	mux.HandleFunc("/webhook", http.HandlerFunc(bot.HandleNetHTTP))
	mux.HandleFunc("/webhook/", http.HandlerFunc(bot.HandleNetHTTP))

	log.Printf("Starting server on port [%v]", bot.BotConfig.Port)

	svr := httputilmore.NewServerTimeouts(fmt.Sprintf(":%v", bot.BotConfig.Port), mux, 10*time.Second)
	log.Fatal(svr.ListenAndServe())
}
