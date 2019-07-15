# [Build Breaker](https://twitter.com/nchlswhttkr/status/1121322470592499713)

Track the builds you break, make charitable contributions to atone for your crimes!

Is your [![Build Status](https://travis-ci.org/nchlswhttkr/build-breaker.svg?branch=master&style=flat-square)](https://travis-ci.org/nchlswhttkr/build-breaker)?

[We have a theme song!](https://youtu.be/YPG5ASujyZg)

## Usage

Build Breaker can be set up to receive web requests from your CI provider, and to record when a failed build occurs.

### [Travis CI](https://travis-ci.org)

Travis CI supports webhooks for notifications when a build status changes.

```
notifications:
  webhooks:
    urls:
      - https://lavutnnx0l.execute-api.ap-southeast-2.amazonaws.com/production/tip/travis

```

## Setting up your own instance

You will need to make a bucket (`$BUCKET_NAME`) to upload your code to AWS.

Here we create our bucket (with an expiration policy), compile and upload the binaries and deploy.

```sh
# Configure your Access Key ID/Secret and region, role and code bucket
export BUCKET_NAME="Your bucket name here"
aws configure

# Build the project and sync our binaries to S3, with an expiration lifecycle
aws s3 mb s3://$BUCKET_NAME
aws s3api put-bucket-lifecycle-configuration --bucket $BUCKET_NAME --lifecycle-configuration file://lifecycle-configuration.json
make
aws s3 sync handlers/ s3://$BUCKET_NAME/handlers

# Deploy the stack, with the permission to create IAM entities (--capabilities)
aws cloudformation deploy \
    --stack-name BuildBreaker \
    --template-file cloudformation.yml \
    --capabilities CAPABILITY_IAM \
    --parameter-overrides LambdaCodeBucket=$BUCKET_NAME

# You can find the API of your URL as a stack output
aws cloudformation describe-stacks --stack-name BuildBreaker --query "Stacks[0].Outputs[?OutputKey=='HelloWorldUrl'].OutputValue" --output text
```

To ensure that future deployments of our Lamdba functions use up-to-date code (CloudFormation only updates the stack where it changes), you can use the `BB_VERSION` environment variable when building and deploying. The `Makefile` is set up to use this variable if it is set, and our CloudFormation template allows a `BuildBreakerVersion` parameter override to be used.

```
export BB_VERSION="A unique version identifier"
make
aws s3 sync handlers/ s3://$BUCKET_NAME/handlers
aws cloudformation deploy \
    --stack-name BuildBreaker \
    --template-file cloudformation.yml \
    --capabilities CAPABILITY_IAM \
    --parameter-overrides LambdaCodeBucket=$BUCKET_NAME BuildBreakerVersion=$BB_VERSION
```
