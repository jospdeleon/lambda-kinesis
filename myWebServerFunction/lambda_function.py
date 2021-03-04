import json
import boto3
import newrelic

GET_RESPONSE = """
<html>
<head>
<title>NRDT Demo App</title>
</head>
<body>
<form id="post_me" name="post_me" method="POST" action="">
    <label for="message">Message</label>
    <input id="message" name="message" type="text" value="Hello world" />
    <button type="submit" name="submit">Submit</button>
</form>
<div id="output" style="white-space: pre-wrap; font-family: monospace;">
</div>
<script>
const formElem = document.getElementById("post_me");
const messageElem = document.getElementById("message");
formElem.addEventListener("submit", (ev) => {
    fetch(location.href, {
        "method": "POST",
        "body": messageElem.value
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

def send_kinesis_message(message):
    """Turns a list of strings into a batch of records for Kinesis stream"""
    # Get the Kinesis client
    kinesis = boto3.client("kinesis")

    return kinesis.put_record(
      StreamName='lambda-stream-NR',
      Data=message.encode('utf-8') ,
      PartitionKey='1'
    )

def lambda_handler(event, context):
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
        print('inside POST')
        message = event['body']
        newrelic.agent.add_custom_parameter('myMessage', message)
        
        # Handle POST requests by splitting the post body into words, and sending each as an SQS message
        send_status = send_kinesis_message(message)
        # Returns the raw batch status. A real application would want to process the API response.
        return {
            "statusCode": 200,
            "headers": {
                "Content-Type": "application/json"
            },
            "isBase64Encoded": False,
            "body": json.dumps(send_status),
        }
