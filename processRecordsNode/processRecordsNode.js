'use strict';
const newrelic = require('newrelic');

/**
 * A Lambda function that logs the payload received from Kinesis.
 */
exports.handler = async (event, context) => {
    console.info(JSON.stringify(event));
    const transaction = newrelic.getTransaction();

    event.Records.forEach(function(record) {
      // Kinesis data is base64 encoded so decode here
      var payload = Buffer.from(record.kinesis.data, 'base64').toString('ascii');
      console.log('Decoded payload:', payload);

      let payloadJSON = JSON.parse(payload);

      let traceContext = JSON.parse(payloadJSON.nrDt);
      transaction.acceptDistributedTraceHeaders("Other", traceContext);

			// This attribute gets added to the normal AwsLambdaInvocation event
			newrelic.addCustomAttributes({
        "myCustomData": payloadJSON.message
      });

			console.log(`${record.eventName} Data = ${payloadJSON.message} ; ${payloadJSON.nrDt} \n`)

    });

    // transaction.end();

}
