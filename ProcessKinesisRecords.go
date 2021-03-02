package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/newrelic/go-agent/v3/integrations/nrlambda"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func handler(ctx context.Context, kinesisEvent events.KinesisEvent) {
	for _, record := range kinesisEvent.Records {
		kinesisRecord := record.Kinesis
		dataBytes := kinesisRecord.Data
		dataText := string(dataBytes)

		if txn := newrelic.FromContext(ctx); nil != txn {
			txn.Application().RecordCustomEvent("MyEvent", map[string]interface{}{
				"zip": "zap",
			})

			// This attribute gets added to the normal AwsLambdaInvocation event
			txn.AddAttribute("myCustomData", dataText)
		}

		fmt.Printf("%s Data = %s \n", record.EventName, dataText)
	}
}

func main() {
	// Here we are in cold start. Anything you do in main happens once.
	// In main, we initialize the agent.
	app, err := newrelic.NewApplication(nrlambda.ConfigOption())
	if nil != err {
		fmt.Println("error creating app (invalid config):", err)
	}
	// Then we start the lambda handler using `nrlambda` rather than `lambda`
	nrlambda.Start(handler, app)
}
