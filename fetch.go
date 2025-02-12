package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

func init() {
    // Load .env file
    err := godotenv.Load(".env")
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }
}

func FetchHighlights() (map[string]interface{}, error) {
    log.Println("Fetching highlights from API...")
    apiURL := os.Getenv("API_URL")
    rapidapiKey := os.Getenv("RAPIDAPI_KEY")
    rapidapiHost := os.Getenv("RAPIDAPI_HOST")

    req, err := http.NewRequest("GET", apiURL, nil)
    if err != nil {
        log.Printf("Error creating the request: %v", err)
        return nil, err
    }

    req.Header.Add("x-rapidapi-key", rapidapiKey)
    req.Header.Add("x-rapidapi-host", rapidapiHost)

    res, err := http.DefaultClient.Do(req)
    if err != nil {
        log.Printf("Error making the request: %v", err)
        return nil, err
    }
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    if err != nil {
        log.Printf("Error reading the response body: %v", err)
        return nil, err
    }

    var result map[string]interface{}
    if err := json.Unmarshal(body, &result); err != nil {
        log.Printf("Error parsing JSON: %v", err)
        return nil, err
    }

    log.Println("Successfully fetched highlights from API")
    return result, nil
}

func UploadToS3(data interface{}) error {
    log.Println("Starting S3 upload process...")

    // Initialize AWS session
    log.Println("Initializing AWS session...")
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(os.Getenv("AWS_REGION")),
    })
    if err != nil {
        log.Printf("❌ AWS session creation failed: %v\n", err)
        return fmt.Errorf("failed to create AWS session: %v", err)
    }
    log.Println("✅ AWS session initialized successfully")

    // Create S3 client
    log.Println("Creating S3 client...")
    svc := s3.New(sess)
    bucketName := os.Getenv("S3_BUCKET_NAME")
    s3Key := os.Getenv("INPUT_KEY") // This should be "highlights/basketball_highlights.json"
    log.Printf("Using bucket: %s with key: %s\n", bucketName, s3Key)

    // Check if bucket exists
    log.Printf("Checking if bucket '%s' exists...\n", bucketName)
    _, err = svc.HeadBucket(&s3.HeadBucketInput{
        Bucket: aws.String(bucketName),
    })
    if err != nil {
        if aerr, ok := err.(awserr.Error); ok {
            switch aerr.Code() {
            case "NotFound", "NoSuchBucket":
                log.Printf("Bucket '%s' not found, creating new bucket...\n", bucketName)
                _, err = svc.CreateBucket(&s3.CreateBucketInput{
                    Bucket: aws.String(bucketName),
                })
                if err != nil {
                    log.Printf("❌ Failed to create bucket: %v\n", err)
                    return fmt.Errorf("failed to create bucket: %v", err)
                }
                log.Printf("✅ Created new S3 bucket: %s\n", bucketName)
            default:
                log.Printf("❌ Error checking bucket: %v\n", err)
                return fmt.Errorf("error checking bucket: %v", err)
            }
        }
    } else {
        log.Printf("✅ Bucket '%s' exists\n", bucketName)
    }

    // Convert data to JSON
    log.Println("Converting data to JSON format...")
    jsonData, err := json.MarshalIndent(data, "", "    ")
    if err != nil {
        log.Printf("❌ JSON marshaling failed: %v\n", err)
        return fmt.Errorf("failed to marshal JSON: %v", err)
    }
    log.Printf("✅ Data converted to JSON (size: %d bytes)\n", len(jsonData))

    // Upload to S3
    log.Printf("Uploading data to S3 with key: %s...\n", s3Key)
    startTime := time.Now()
    _, err = svc.PutObject(&s3.PutObjectInput{
        Bucket:      aws.String(bucketName),
        Key:         aws.String(s3Key),
        Body:        bytes.NewReader(jsonData),
        ContentType: aws.String("application/json"),
    })
    if err != nil {
        log.Printf("❌ Upload failed: %v\n", err)
        return fmt.Errorf("failed to upload to S3: %v", err)
    }

    duration := time.Since(startTime)
    log.Printf("✅ Successfully uploaded data to s3://%s/%s (took %v)\n", bucketName, s3Key, duration)
    return nil
}

func StoreHighlightsToDynamoDB(highlights map[string]interface{}) error {
    log.Println("Storing highlights in DynamoDB...")

    // Initialize AWS session
    log.Println("Initializing AWS session...")
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(os.Getenv("AWS_REGION")),
    })
    if err != nil {
        log.Printf("❌ AWS session creation failed: %v\n", err)
        return fmt.Errorf("failed to create AWS session: %v", err)
    }
    log.Println("✅ AWS session initialized successfully")

    // Create DynamoDB client
    log.Println("Creating DynamoDB client...")
    svc := dynamodb.New(sess)
    tableName := os.Getenv("DYNAMODB_TABLE")
    log.Printf("Using DynamoDB table: %s\n", tableName)

    // Loop through the records in highlights["data"]
    data, ok := highlights["data"].([]interface{})
    if !ok {
        return fmt.Errorf("invalid data format or missing data field")
    }

    for _, item := range data {
        record, ok := item.(map[string]interface{})
        if !ok {
            continue
        }

        // Extract a unique identifier (id or url)
        itemKey, ok := record["id"].(string)
        if !ok || itemKey == "" {
            itemKey, ok = record["url"].(string)
            if !ok || itemKey == "" {
                continue
            }
        }

        // Add a fetch_date to the record
        record["id"] = itemKey
        record["fetch_date"] = time.Now().Format(time.RFC3339)

        // Convert the record to a DynamoDB attribute value map
        av, err := dynamodbattribute.MarshalMap(record)
        if err != nil {
            log.Printf("❌ Failed to marshal record: %v\n", err)
            return fmt.Errorf("failed to marshal record: %v", err)
        }

        // Store the record in DynamoDB
        _, err = svc.PutItem(&dynamodb.PutItemInput{
            TableName: aws.String(tableName),
            Item:      av,
        })
        if err != nil {
            log.Printf("❌ Failed to store record in DynamoDB: %v\n", err)
            return fmt.Errorf("failed to store record in DynamoDB: %v", err)
        }

        log.Printf("✅ Stored record with key %s in DynamoDB\n", itemKey)
    }

    log.Println("Successfully stored highlights in DynamoDB")
    return nil
}

func main() {
    // Fetch the highlights
    highlights, err := FetchHighlights()
    if err != nil {
        log.Fatalf("Error fetching highlights: %v", err)
    }

    // Upload the parsed data to S3
    if err := UploadToS3(highlights); err != nil {
        log.Fatalf("Error uploading to S3: %v", err)
    }

    // Store the highlights in DynamoDB
    if err := StoreHighlightsToDynamoDB(highlights); err != nil {
        log.Fatalf("Error storing highlights in DynamoDB: %v", err)
    }
}