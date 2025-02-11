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
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

func main() {
    // Load .env file
    err := godotenv.Load(".env")
    if err != nil {
        fmt.Printf("Error loading .env file: %v\n", err)
        return
    }

    // Get the API URL, key, and host from the environment variables
    apiURL := os.Getenv("API_URL")
    rapidapiKey := os.Getenv("RAPIDAPI_KEY")
    rapidapiHost := os.Getenv("RAPIDAPI_HOST")

    req, err := http.NewRequest("GET", apiURL, nil)
    if err != nil {
        fmt.Println("Error creating the request:", err)
        return
    }

    req.Header.Add("x-rapidapi-key", rapidapiKey)
    req.Header.Add("x-rapidapi-host", rapidapiHost)

    res, err := http.DefaultClient.Do(req)
    if err != nil {
        fmt.Println("Error making the request:", err)
        return
    }

    defer res.Body.Close()
    body, err := io.ReadAll(res.Body)
    if err != nil {
        fmt.Println("Error reading the response body:", err)
        return
    }

    // Parse the JSON response into a map
    var responseData map[string]interface{}
    if err := json.Unmarshal(body, &responseData); err != nil {
        fmt.Printf("Error parsing JSON: %v\n", err)
        return
    }

    fmt.Println("Fetching highlights...")
    
    // Upload the parsed data to S3
    fmt.Println("Saving highlights to S3...")
    if err := UploadToS3(responseData); err != nil {
        fmt.Printf("Error saving to S3: %v\n", err)
        return
    }
}

func UploadToS3(data interface{}) error {
    // Start logging
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
    // Use the INPUT_KEY directly from environment variable
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