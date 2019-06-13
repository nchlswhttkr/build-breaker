# [Build Breaker](https://twitter.com/nchlswhttkr/status/1121322470592499713)

Make charitable contributions for the builds you break!

[We have a theme song!](https://youtu.be/YPG5ASujyZg)

You will need to create a blank role for the Lambda to execute with (`$LAMBDA_EXECUTION_ROLE`) and to make a bucket (`$BUCKET_NAME`) to upload your code to AWS.

```
# Configure your Access Key ID/Secret and region, role and code bucket
export BUCKET_NAME
export LAMBDA_EXECUTION_ROLE
aws configure

# Build the project and copy to S3
aws s3 mb s3://$BUCKET_NAME
make
aws s3 cp handlers/** s3://$BUCKET_NAME/handlers/

# Deploy the stack
aws cloudformation deploy \
    --stack-name BuildBreaker \
    --template-file cloudformation.yml \
    --parameter-overrides LambdaExecutionRole=$LAMBDA_EXECUTION_ROLE LambdaCodeBucket=$BUCKET_NAME

# You can find the API of your URL as a stack output
aws cloudformation describe-stacks --stack-name BuildBreaker --query "Stacks[0].Outputs[?OutputKey=='HelloWorld'].OutputValue" --output text
```
