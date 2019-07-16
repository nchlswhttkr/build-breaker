package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Build struct {
	Provider     string `json:"provider"`
	ProjectName  string `json:"projectName"`
	ProjectOwner string `json:"projectOwner"`
	ProjectHost  string `json:"projectHost"`
	BuildNumber  string `json:"buildNumber"` // builds can be retried, this may not be unique

}

type TravisResponse struct {
	Number        string `json:"number"`
	StatusMessage string `json:"status_message"` // duplicated by ResultMessage?
	Repository    struct {
		Name      string `json:"name"`
		OwnerName string `json:"owner_name"`
	} `json:"repository"`
}

func Handler(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	provider := e.PathParameters["provider"]
	var err error
	var build Build

	if provider == "travis" {
		err = InterpretTravisResponse(e, &build)
	} else {
		return GenerateBadRequestResponse(errors.New("A known build provider must be specified"))
	}
	if err != nil {
		return GenerateErrorResponse(err)
	}

	body, err := json.Marshal(build)
	if err != nil {
		return GenerateErrorResponse(err)
	}

	fmt.Printf("Processed build #%s from %s %s/%s/%s\n", build.BuildNumber, build.Provider, build.ProjectHost, build.ProjectOwner, build.ProjectName)
	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: 201,
	}, nil

}

func InterpretTravisResponse(e events.APIGatewayProxyRequest, build *Build) error {
	var err error
	var unescaped string
	var travis TravisResponse

	// find the percent-encoded, stringified JSON from the payload field
	// https://developer.mozilla.org/en-US/docs/Glossary/percent-encoding
	tuples := strings.Split(e.Body, "&")
	for _, tuple := range tuples {
		if strings.HasPrefix(tuple, "payload=") {
			unescaped, err = url.QueryUnescape(tuple[8:])
		}
	}
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(unescaped), &travis)
	if err != nil {
		return err
	}

	fmt.Printf("Received notification from Travis CI (%s/%s#%s - %s)\n", travis.Repository.OwnerName, travis.Repository.Name, travis.Number, travis.StatusMessage)

	// Travis may notify for all build status changes, not just failures
	// See status_message in https://docs.travis-ci.com/user/notifications/#configuring-webhook-notifications
	switch status := travis.StatusMessage; status {
	case "Broken":
	case "Failed":
	case "Still Failing":
	case "Errored":
		break
	default:
		err = errors.New("The build has not failed, nothing will be recorded")
	}
	if err != nil {
		return err
	}

	build.Provider = "travis"
	build.ProjectName = travis.Repository.Name
	build.ProjectOwner = travis.Repository.OwnerName
	build.ProjectHost = "https://github.com"
	build.BuildNumber = travis.Number
	return nil
}

func GenerateErrorResponse(err error) (events.APIGatewayProxyResponse, error) {
	if err != nil {
		fmt.Println(err.Error())
	}
	return events.APIGatewayProxyResponse{
		Body:       "An error occured",
		StatusCode: 500,
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
	}, nil
}

func GenerateBadRequestResponse(err error) (events.APIGatewayProxyResponse, error) {
	if err != nil {
		fmt.Println(err.Error())
	}
	return events.APIGatewayProxyResponse{
		Body:       err.Error(),
		StatusCode: 400,
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
	}, nil
}

func main() {
	lambda.Start(Handler)
}
