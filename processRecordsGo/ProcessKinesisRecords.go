package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/newrelic/go-agent/v3/integrations/logcontext/nrlogrusplugin"
	"github.com/newrelic/go-agent/v3/integrations/nrlambda"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
)

type payload struct {
	Message string `json:"message"`
	NrDt    string `json:"nrDt"`
}
type nrpayload struct {
	Newrelic string `json:"newrelic"`
}

func handler(ctx context.Context, kinesisEvent events.KinesisEvent) error {
	for _, record := range kinesisEvent.Records {
		kinesisRecord := record.Kinesis
		// dataBytes := kinesisRecord.Data
		// dataText := string(dataBytes)

		if txn := newrelic.FromContext(ctx); nil != txn {

			// unmarshal data
			var p payload
			err := json.Unmarshal(kinesisRecord.Data, &p)
			if err != nil {
				fmt.Printf("json unmarshal error: %v", err)
				return err
			}

			//after unmarshalling into p...
			var np nrpayload
			err = json.Unmarshal([]byte(p.NrDt), &np)
			if err != nil {
				fmt.Printf("json unmarshal error: %v", err)
				return err
			}

			hdrs := http.Header{}
			hdrs.Set(newrelic.DistributedTraceNewRelicHeader, np.Newrelic)

			fmt.Printf("HDRS value = %s \n", hdrs)
			txn.AcceptDistributedTraceHeaders(newrelic.TransportOther, hdrs)

			txn.InsertDistributedTraceHeaders(hdrs)

			// If you need to record a custom event
			// txn.Application().RecordCustomEvent("MyEvent", map[string]interface{}{
			// 	"zip": "zap",
			// })

			// This attribute gets added to the normal AwsLambdaInvocation event
			txn.AddAttribute("myCustomData", p.Message)

			// Logs in context example using logrus and nrlogrusplugin
			log := logrus.New()
			log.SetFormatter(nrlogrusplugin.ContextFormatter{})
			log.WithContext(ctx).Info("Data from " + record.EventName + " = " + p.Message)

			// fmt.Printf("%s Data = %s %s \n", record.EventName, p.Message, p.NrDt)
		}
	}
	return nil
}

func main() {
	// Here we are in cold start. Anything you do in main happens once.
	// In main, we initialize the agent.
	app, err := newrelic.NewApplication(nrlambda.ConfigOption(), newrelic.ConfigDebugLogger(os.Stdout))
	if nil != err {
		fmt.Println("error creating app (invalid config):", err)
	}
	// Then we start the lambda handler using `nrlambda` rather than `lambda`
	nrlambda.Start(handler, app)
}
