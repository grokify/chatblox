# Chatblox

[![Build Status][build-status-svg]][build-status-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![License][license-svg]][license-url]

 [build-status-svg]: https://github.com/grokify/chatblox/workflows/test/badge.svg
 [build-status-url]: https://github.com/grokify/chatblox/actions
 [goreport-svg]: https://goreportcard.com/badge/github.com/grokify/chatblox
 [goreport-url]: https://goreportcard.com/report/github.com/grokify/chatblox
 [docs-godoc-svg]: https://img.shields.io/badge/docs-godoc-blue.svg
 [docs-godoc-url]: https://godoc.org/github.com/grokify/chatblox
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/grokify/chatblox/blob/master/LICENSE

## Overview

Chatblox is framework for building chatbots in Go.

This is a work in progress app and not ready for use.

## Configuration

Set the following environment variables:

| Variable | Type | Required | Notes |
|----------|------|----------|-------|
| `CHATBLOX_ENGINE`             | string | y | `aws` or `nethttp` |
| `CHATBLOX_PORT`               | integer | n | local port number for `net/http` |
| `CHATBLOX_REQUEST_FUZZY_AT_MENTION_MATCH` | boolean | n | Match non-completed at mentions. |
| `CHATBLOX_RESPONSE_AUTO_AT_MENTION`   | boolean | n | |
| `CHATBLOX_POST_SUFFIX`        | string | n | |
| `GOOGLE_SERVICE_ACCOUNT_JWT`  | JSON string | y |  |
| `GOOGLE_SPREADSHEET_ID`       | string | y | ID as in URL |
| `GOOGLE_SHEET_TITLE_RECORDS`  | string | y | sheet title for data records, e.g. `Records` |
| `GOOGLE_SHEET_TITLE_METADATA` | string | y | sheet title for metadata, e.g. `Metadata` |
| `RINGCENTRAL_BOT_ID`          | string | y | bot `personId` in Glip |
| `RINGCENTRAL_BOT_NAME`        | string | y | bot name in Glip for fuzzy at matching |
| `RINGCENTRAL_SERVER_URL`      | string | y | Base API URL, e.g. https://platform.ringcentral.com |
| `RINGCENTRAL_TOKEN_JSON`      | JSON string | y | JSON token as returned by `/oauth/token` endpoint |

## Using the AWS Engine

To use the AWS Lambda engine, you need an AWS account. If you don't hae one, the [free trial account](https://aws.amazon.com/s/dm/optimization/server-side-test/free-tier/free_np/) includes 1 million free Lambda requests per month forever and 1 million free API Gateway requests per month for the first year.

### Installation via AWS Lambda

See the AWS docs for deployment:

https://docs.aws.amazon.com/lambda/latest/dg/lambda-go-how-to-create-deployment-package.html

Using the `aws-cli` you can use the following approach:

```
$ cd ./apps/server
$ GOOS=linux go build main.go
$ zip main.zip ./main
# --handler is the path to the executable inside the .zip
$ aws lambda create-function --region us-east-1 --function-name Databot --memory 128 --role arn:aws:iam::account-id:role/execution_role --runtime go1.x --zip-file fileb://main.zip --handler main
```

### Installation via AWS UI

#### API Gateway

1. "Actions" > "Create Resource"
  1. Click "Configure as proxy resource"
  1. Click "Enable API Gateway CORS"
  1. Click "Create Resource"
1. Add "Lambda Function"
  1. Click "Save"
1. Click "Actions" > "Deploy API"
  1. Add "Stage name", e.g. `prod`, `v1`, etc.
1. Create Webhook URL for `https://myId.execute-api.myRegion.amazonaws.com/v1/webhook`


### Keepalive

In production, there are are reasons why a RingCentral webhook may fail and become blacklisted. These should be tracked down an eliminated. If there are reasons to reenable the webhook, you can deploy the [`rchooks`] RingCentral Lambda keepalive function:

* [`rchooks/apps/keepalive_lambda`](https://github.com/grokify/rchooks/tree/master/apps/keepalive_lambda)
