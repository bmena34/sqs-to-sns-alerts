# sqs-to-sns-alerts
This is a quick Go project that was intended to be deployed as a container to Lambda or ECS/EKS as a cron to pull timestamps from an AWS SQS and send a message to SNS if any timestamps are greater than 48hrs old. The SNS was intended to be subscribed to by OpsGenie or Email for alerting using event notifications setup on S3 buckets used for retaining audit logs. Basically, no put requests for 48 hours in S3 generates an alert. While I've gone another route for this use case this may be a useful template for another use case.

# TO-DO:
This uses environmental variables to provide the necessary arns:

SNS_ALERT_TOPIC_URL = is the SNS you intend to base alerts on.
SQS_LIST = is the list of SQS urls split by "," as this can process one queue or many