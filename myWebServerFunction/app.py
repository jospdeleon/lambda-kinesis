import json
import boto3
import newrelic
import os

GET_RESPONSE = """
<html>
<head>
<title>NRDT Demo App</title>
</head>
<body>
<form id="post_me" name="post_me" method="POST" action="">
    <label for="message">Message</label>
    <input id="message" name="message" type="text" value="Hello world" />
    <select id="stream" name="stream">
        <option value="go">Go</option>
        <option value="node">Node</option>
    </select>
    <button type="submit" name="submit">Submit</button>
</form>
<div id="output" style="white-space: pre-wrap; font-family: monospace;">
</div>
<script>
const formElem = document.getElementById("post_me");
const messageElem = document.getElementById("message");
const streamElem = document.getElementById("stream");
formElem.addEventListener("submit", (ev) => {
    let data = {
        message: messageElem.value,
        stream: streamElem.value
    }
    fetch(location.href, {
        "method": "POST",
        headers: {
            'Content-Type': 'application/json'
        },
        "body": JSON.stringify(data)
    })
    .then(resp => resp.text())
    .then(body => {
        document.getElementById("output").innerText = body;
    });
    ev.preventDefault();
});
</script>
</body>
</html>
"""

def nr_trace_context_json():
    """Generate a distributed trace context as a JSON document"""
    # The Python agent expects a list as an out-param
    dt_headers = []
    newrelic.agent.insert_distributed_trace_headers(headers=dt_headers)
    # At this point, dt_headers is a list of tuples. We first convert it to a dict, then serialize as a JSON object.
    # The resulting string can be used as a kinesis record attribute string value.
    return json.dumps(dict(dt_headers))

def send_kinesis_message(message, stream):
    nrcontext = newrelic.agent.get_linking_metadata()

    # Get the Kinesis client
    kinesis = boto3.client("kinesis")

    # a Python object (dict):
    nrData = {
      "message": message,
      "nrDt": nr_trace_context_json()
    }

    # Logs in context example using the agent API
    log_message = {"message": 'RECORD: ' + nrData['message']}
    log_message.update(nrcontext)
    print(json.dumps(log_message))

    streamArn = ''
    if stream == 'go':
        streamArn = os.environ.get('GO_STREAM')
    elif stream == 'node':
        streamArn = os.environ.get('NODE_STREAM')

    return kinesis.put_record(
      StreamName = streamArn,
      Data=json.dumps(nrData),
      PartitionKey='1'
    )

def lambda_handler(event, context):
    nrcontext = newrelic.agent.get_linking_metadata()

    if event['httpMethod'] == 'GET':
        print('inside GET')

        # For our example, we return a static HTML page in response to GET requests
        return {
            "statusCode": 200,
            "headers": {
                "Content-Type": "text/html"
            },
            "isBase64Encoded": False,
            "body": GET_RESPONSE
        }
    elif event['httpMethod'] == 'POST':
        # Logs in context example using the agent API
        log_message = {"message": "inside POST"}
        log_message.update(nrcontext)
        print(json.dumps(log_message))

        data = json.loads(event['body'])
        message = data['message']
        stream = data['stream']
        newrelic.agent.add_custom_parameter('myMessage', message)
        
        # Handle POST requests by sending the message into a kinesis stream
        send_status = send_kinesis_message(message, stream)
        # Returns the raw batch status. A real application would want to process the API response.
        return {
            "statusCode": 200,
            "headers": {
                "Content-Type": "application/json"
            },
            "isBase64Encoded": False,
            "body": json.dumps(send_status),
        }
