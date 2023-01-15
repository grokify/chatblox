package chatblox

import (
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var intentRouter_ = IntentRouter{}

func ServeAWSLambda(intentRouter IntentRouter) {
	log.Print("serveAwsLambda_S1")
	intentRouter_ = intentRouter
	lambda.Start(HandleAWSLambda)
}

func HandleAWSLambda(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Print("HandleAwsLambda_S1")
	log.Printf("HandleAwsLambda_S2_req_body: %v", req.Body)

	if val, ok := req.Headers[ValidationTokenHeader]; ok && len(val) > 0 {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    map[string]string{ValidationTokenHeader: val},
			Body:       `{"statusCode":200}`,
		}, nil
	}

	bot := Bot{IntentRouter: intentRouter_}
	return bot.HandleAwsLambda(req)
}
