# [Build Breaker](https://twitter.com/nchlswhttkr/status/1121322470592499713)

Make charitable contributions for the builds you break!

[We have a theme song!](https://youtu.be/YPG5ASujyZg)

You will need to create a blank role for the Lambda to execute with (`$LAMBDA_EXECUTION_ROLE`) and to make a bucket (`$BUCKET_NAME`) to upload your code to AWS.

```
# Configure your Access Key ID/Secret and region, role and code bucket
export BUCKET_NAME
export LAMBDA_EXECUTION_ROLE
aws configure

# Build the project and sync our binaries to S3, with an expiration lifecycle
aws s3 mb s3://$BUCKET_NAME
aws s3api put-bucket-lifecycle-configuration --bucket $BUCKET_NAME --lifecycle-configuration file://lifecycle-configuration.json
make
aws s3 sync handlers/ s3://$BUCKET_NAME/handlers

# Deploy the stack
aws cloudformation deploy \
    --stack-name BuildBreaker \
    --template-file cloudformation.yml \
    --parameter-overrides LambdaExecutionRole=$LAMBDA_EXECUTION_ROLE LambdaCodeBucket=$BUCKET_NAME

# You can find the API of your URL as a stack output
aws cloudformation describe-stacks --stack-name BuildBreaker --query "Stacks[0].Outputs[?OutputKey=='HelloWorldUrl'].OutputValue" --output text
```

To ensure that future deployments of our Lamdba functions use up-to-date code (CloudFormation only updates the stack where it changes), you can use the `BB_VERSION` environment variable when building and deploying. The `Makefile` is set up to use this variable if it is set, and our CloudFormation template allows a `BuildBreakerVersion` parameter override to be used.

```
...
make
aws s3 sync handlers/ s3://$BUCKET_NAME/handlers
aws cloudformation deploy \
    --stack-name BuildBreaker \
    --template-file cloudformation.yml \
    --parameter-overrides LambdaExecutionRole=$LAMBDA_EXECUTION_ROLE LambdaCodeBucket=$BUCKET_NAME BuildBreakerVersion=$BB_VERSION
```

If you add new method after deploying, you'll need to replace the existing deployment. The easiest way to do this is to deploy once with the `BuildBreakerApiDeployment` resource commented out, and then to deploy with it commented back in. Remember to specify that it `DependsOn` on all the methods that invoke your lambdas.

I'll be moving this all over to the AWS::Serverless transform soon, so this shouldn't be a problem much longer. It's because CloudFormation is only meant to _provision_ your stack, and not _update_ once it is deployed. To make sure our deployment includes all of the app's methods, it needs to be _updated_ if it has already been deployed.
