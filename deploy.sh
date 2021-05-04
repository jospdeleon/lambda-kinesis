#!/bin/bash

accountId=$1

region=$2
echo "region set to ${region}"

# The Go1.x runtime does not support Lambda Extensions. Instead, Go Lambdas should be written
# against the "provided" runtime. The aws-lambda-go SDK provides a build tag that makes this easy.
runtime="provided"

cd processRecordsGo/
echo "Building stand-alone lambda"
build_tags="-tags lambda.norpc"

# Custom runtimes need a bootstrap executable. See https://docs.aws.amazon.com/lambda/latest/dg/runtimes-custom.html
handler="bootstrap"

env GOARCH=amd64 GOOS=linux go build ${build_tags} -ldflags="-s -w" -o ${handler}
zip processRecordsGo.zip "${handler}"

bucket="process-records-go-${region}-${accountId}"
aws s3 mb --region ${region} s3://${bucket}
aws s3 cp processRecordsGo.zip s3://${bucket}

# Go back to root folder
cd ..

aws cloudformation deploy --region ${region} \
  --template-file template.yaml \
  --stack-name processRecords-Go \
  --capabilities CAPABILITY_IAM \
  --parameter-overrides "NRAccountId=${accountId}"
