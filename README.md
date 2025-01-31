# [Build Breaker](https://twitter.com/nchlswhttkr/status/1121322470592499713) [![GitLab CI Pipeline Status](https://gitlab.com/nchlswhttkr/build-breaker/badges/master/pipeline.svg?style=flat-square)](https://gitlab.com/nchlswhttkr/build-breaker/commits/master)

> :exclamation: Build Breaker is still under development, use at your own risk.

_Track the builds you break and atone for your crimes!_

Build Breaker can be set up with your CI process to record failed builds. Like a swear jar, you should try to make up for each failed build with a small donation.

[We have a theme song!](https://youtu.be/YPG5ASujyZg)

## Usage

Build Breaker receives web requests from your CI provider, and records failed builds. This varies between providers.

### [Travis CI](https://travis-ci.org) [![Travis CI Build Status](https://travis-ci.org/nchlswhttkr/build-breaker.svg?branch=master)](https://travis-ci.org/nchlswhttkr/build-breaker)

Travis CI supports [webhooks for notifications](https://docs.travis-ci.com/user/notifications#configuring-webhook-notifications) when a build's status changes. You can add them in your `.travis.yml` configuration.

```
notifications:
  webhooks:
    urls:
      - https://lavutnnx0l.execute-api.ap-southeast-2.amazonaws.com/production/tip/travis

```

### [Bitbucket](https://bitbucket.org) [![Bitbucket Pipelines Build Status](https://img.shields.io/bitbucket/pipelines/nchlswhttkr/build-breaker.svg)](https://bitbucket.org/nchlswhttkr/build-breaker/addon/pipelines/home)

Bitbucket can be configured to [send notifications about builds](https://confluence.atlassian.com/bitbucket/manage-webhooks-735643732.html#Managewebhooks-create_webhookCreatingwebhooks) for your projects.

- The URL for your webhook will be `https://lavutnnx0l.execute-api.ap-southeast-2.amazonaws.com/production/tip/bitbucket`
- The webhook should be triggered by the **Build status created** and **Build status updated** actions

---

## Setting up your own instance

You will need to make a bucket (`$BUCKET_NAME`) to upload your code to AWS.

Here we create our bucket (with an expiration policy), compile and upload the binaries and deploy.

```sh
# Get the repository
go get github.com/nchlswhttkr/build-breaker
cd $GOPATH/src/github.com/nchlswhttkr/build-breaker

# Configure your Access Key ID/Secret and region, role and code bucket
export BUCKET_NAME="build-breaker"
aws configure

# Build the project and sync our binaries to S3, with an expiration lifecycle
aws s3 mb s3://$BUCKET_NAME
aws s3api put-bucket-lifecycle-configuration \
    --bucket $BUCKET_NAME \
    --lifecycle-configuration file://lifecycle-configuration.json
make
aws s3 sync handlers/ s3://$BUCKET_NAME/handlers/initial

# Deploy the stack, with the permission to create IAM entities (--capabilities)
aws cloudformation deploy \
    --stack-name BuildBreaker \
    --template-file cloudformation.yml \
    --capabilities CAPABILITY_IAM \
    --parameter-overrides LambdaCodeBucket=$BUCKET_NAME

# You can find the API of your URL as a stack output
aws cloudformation describe-stacks \
    --stack-name BuildBreaker \
    --query "Stacks[0].Outputs[?OutputKey=='HelloWorldUrl'].OutputValue" \
    --output text
```

To ensure that future deployments of our Lamdba functions use up-to-date code (CloudFormation only updates the stack when the path to the code changes), you can use the an environment variable when building and deploying.

```shell
export BB_VERSION="2019-01-01-abc123"
make
aws s3 sync handlers/ s3://$BUCKET_NAME/handlers/$BB_VERSION
aws cloudformation deploy \
    --stack-name BuildBreaker \
    --template-file cloudformation.yml \
    --capabilities CAPABILITY_IAM \
    --parameter-overrides LambdaCodeBucket=$BUCKET_NAME BuildBreakerVersion=$BB_VERSION
```
