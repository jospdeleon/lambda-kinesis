package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/newrelic/go-agent/v3/integrations/nrlambda"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type payload struct {
	Message string `json:"message"`
	NrDt    string `json:"nrDt"`
}

func handler(ctx context.Context, kinesisEvent events.KinesisEvent) error {
	for _, record := range kinesisEvent.Records {
		kinesisRecord := record.Kinesis
		// dataBytes := kinesisRecord.Data
		// dataText := string(dataBytes)

		if txn := newrelic.FromContext(ctx); nil != txn {

			fmt.Printf("Trace Metadata BEFORE = %s %s \n", txn.GetTraceMetadata().TraceID, txn.GetTraceMetadata().SpanID)

			// unmarshal data
			var p payload
			err := json.Unmarshal(kinesisRecord.Data, &p)
			if err != nil {
				fmt.Printf("json unmarshal error: %v", err)
				return err
			}

			hdrs := http.Header{}
			hdrs.Set(newrelic.DistributedTraceNewRelicHeader, p.NrDt)
			txn.AcceptDistributedTraceHeaders(newrelic.TransportOther, hdrs)

			fmt.Printf("Trace Metadata AFTER = %s %s \n", txn.GetTraceMetadata().TraceID, txn.GetTraceMetadata().SpanID)

			txn.Application().RecordCustomEvent("MyEvent", map[string]interface{}{
				"zip": "zap",
			})

			// This attribute gets added to the normal AwsLambdaInvocation event
			txn.AddAttribute("myCustomData", p.Message)

			fmt.Printf("%s Data = %s %s \n", record.EventName, p.Message, p.NrDt)
		}
	}
	return nil
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
