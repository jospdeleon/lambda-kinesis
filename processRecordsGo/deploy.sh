#!/bin/bash

accountId=$1

region=$2
echo "region set to ${region}"

# The Go1.x runtime does not support Lambda Extensions. Instead, Go Lambdas should be written
# against the "provided" runtime. The aws-lambda-go SDK provides a build tag that makes this easy.
runtime="provided"

echo "Building stand-alone lambda"
build_tags="-tags lambda.norpc"

# Custom runtimes need a bootstrap executable. See https://docs.aws.amazon.com/lambda/latest/dg/runtimes-custom.html
handler="bootstrap"

env GOARCH=amd64 GOOS=linux go build ${build_tags} -ldflags="-s -w" -o ${handler}
zip ProcessKinesisRecords-NR.zip "${handler}"

bucket="process-kinesis-records-${region}-${accountId}"
aws s3 mb --region ${region} s3://${bucket}
aws s3 cp ProcessKinesisRecords-NR.zip s3://${bucket}
aws cloudformation deploy --region ${region} \
  --template-file template.yaml \
  --stack-name ProcessKinesisRecords-NR \
  --capabilities CAPABILITY_IAM \
  --parameter-overrides "NRAccountId=${accountId}"
