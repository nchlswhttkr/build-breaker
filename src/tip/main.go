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

// TODO: Check signing on Travis responses

type Build struct {
	Provider     string `json:"provider"`     // https://travis-ci.org
	ProjectName  string `json:"projectName"`  // build-breaker
	ProjectOwner string `json:"projectOwner"` // nchlswhttkr
	ProjectHost  string `json:"projectHost"`  // https://github.com
	Message      string `json:"message"`      // Build #3 failed at stage "Test"
}

func Handler(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	provider := e.PathParameters["provider"]
	var err error
	var build Build

	if provider == "travis" {
		err = InterpretTravisResponse(e, &build)
	} else if provider == "bitbucket" {
		err = InterpretBitbucketResponse(e, &build)
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

	fmt.Printf("Processed build from %s for %s/%s/%s\n", build.Provider, build.ProjectHost, build.ProjectOwner, build.ProjectName)
	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: 201,
	}, nil

}

func InterpretTravisResponse(e events.APIGatewayProxyRequest, build *Build) error {
	// https://docs.travis-ci.com/user/notifications/#webhooks-delivery-format
	type TravisResponse struct {
		Number        string `json:"number"`
		StatusMessage string `json:"status_message"` // duplicated by ResultMessage?
		Repository    struct {
			Name      string `json:"name"`
			OwnerName string `json:"owner_name"`
		} `json:"repository"`
	}

	var err error
	var unescaped string
	var travis TravisResponse

	// Find the percent-encoded, stringified JSON from the payload field
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
	// See status_message in https://docs.travis-ci.com/user/notifications/#webhooks-delivery-format
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

	build.Provider = "https://travis-ci.org"
	build.ProjectName = travis.Repository.Name
	build.ProjectOwner = travis.Repository.OwnerName
	build.ProjectHost = "https://github.com"
	build.Message = fmt.Sprintf("Build #%s failed with status \"%s\"", travis.Number, travis.StatusMessage)
	return nil
}

func InterpretBitbucketResponse(e events.APIGatewayProxyRequest, build *Build) error {
	// https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-Buildstatuscreated
	// https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-Buildstatusupdated
	type BitbucketResponse struct {
		CommitStatus struct {
			State string `json:"state"`
			Name  string `json:"name"`
		} `json:"commit_status"`
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
	}

	var err error
	var bitbucket BitbucketResponse

	// Must be a build-related notification
	switch eventKey := e.Headers["X-Event-Key"]; eventKey {
	case "repo:commit_status_updated":
	case "repo:commit_status_created":
		break
	default:
		return errors.New("The Bitbucket notification is not related to a build")
	}

	err = json.Unmarshal([]byte(e.Body), &bitbucket)
	if err != nil {
		return err
	}

	if bitbucket.CommitStatus.State != "FAILED" {
		return errors.New("The Bitbucket build has not failed, nothing will be recorded")
	}

	// Both the owner name and repository name can be obtained from the FullName
	names := strings.SplitN(bitbucket.Repository.FullName, "/", 2)
	if len(names) != 2 {
		return errors.New("Something went wrong while getting the repository owner/name")
	}
	projectOwner, projectName := names[0], names[1]

	build.Provider = "https://bitbucket.org"
	build.ProjectName = projectName
	build.ProjectOwner = projectOwner
	build.ProjectHost = "https://bitbucket.org"
	build.Message = bitbucket.CommitStatus.Name
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
