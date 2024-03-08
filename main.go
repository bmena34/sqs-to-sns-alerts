package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func main() {
	// To run as a lambda uncomment the below line //
	// lambda.Start(handler)
	// To run locally uncomment the below line //
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	handler(ctx)
}

const message = "The logs have not updated for more than 48 hours"

// Environment variables //
var topicArn = os.Getenv("SNS_ALERT_TOPIC_URL") //TODO: Add your topic to send alerts to
var sqsUrlList = os.Getenv("SQS_LIST")             //TODO: Add your SQS list delimited by "," to process.

var sqsUrls = strings.Split(sqsUrlList, ",")

func handler(ctx context.Context) {
	cfg, err := getConfig(ctx)
	if err != nil {
		log.Printf("unable to load SDK config : %v", err)
	}

	wg := sync.WaitGroup{}

	for _, queueURL := range sqsUrls {
		wg.Add(1)
		go func(queueURL string) {
			defer wg.Done()

			err := getSqsMessages(ctx, queueURL, cfg)
			if err != nil {
				log.Printf("failed to get messages from SQS and alert, %v", err)
			}
		}(queueURL)
	}
	wg.Wait()

	log.Printf("Processing complete.")
}

func getConfig(ctx context.Context) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, err
	}
	return cfg, err
}

func getSqsMessages(ctx context.Context, queueURL string, cfg aws.Config) error {

	sqsClient := sqs.NewFromConfig(cfg)

	for {
		msg, err := sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     5,
			AttributeNames:      []types.QueueAttributeName{"SentTimestamp"},
		})
		if err != nil {
			log.Printf("failed to receive messages, %v", err)
		}

		if len(msg.Messages) == 0 {
			log.Println("No messages left to process.")
			break
		}

		for _, m := range msg.Messages {
			sentTimeStr := m.Attributes["SentTimestamp"]
			sentTimeInt, err := strconv.ParseInt(sentTimeStr, 10, 64)
			if err != nil {
				log.Printf("failed to parse sent time, %v", err)
			}

			sentTime := time.Unix(sentTimeInt/1000, 0)
			if time.Since(sentTime) > 48*time.Hour {
				err := alertSns(ctx, message, topicArn)
				if err != nil {
					log.Printf("failed to alert, %v", err)
				}
			}

			_, err = sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(queueURL),
				ReceiptHandle: m.ReceiptHandle,
			})
			if err != nil {
				log.Printf("failed to delete message, %v", err)
			}
		}
	}
	return nil
}

func alertSns(ctx context.Context, message string, topicArn string) error {

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("unable to load SDK config : %v", err)
	}

	snsClient := sns.NewFromConfig(cfg)

	_, err = snsClient.Publish(ctx, &sns.PublishInput{
		Message:  aws.String(message),
		TopicArn: aws.String(topicArn),
	})
	if err != nil {
		log.Printf("failed to publish message, %v", err)
	}
	return err
}
