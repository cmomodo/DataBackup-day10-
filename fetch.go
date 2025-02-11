package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
    fmt.Println("Starting S3 upload process...")
    
    // Initialize AWS session
    fmt.Println("Initializing AWS session...")
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(os.Getenv("AWS_REGION")),
    })
    if err != nil {
        fmt.Printf("❌ AWS session creation failed: %v\n", err)
        return fmt.Errorf("failed to create AWS session: %v", err)
    }
    fmt.Println("✅ AWS session initialized successfully")

    // Create S3 client
    fmt.Println("Creating S3 client...")
    svc := s3.New(sess)
    bucketName := os.Getenv("S3_BUCKET_NAME")
    fmt.Printf("Using bucket: %s\n", bucketName)

    // Check if bucket exists
    fmt.Printf("Checking if bucket '%s' exists...\n", bucketName)
    _, err = svc.HeadBucket(&s3.HeadBucketInput{
        Bucket: aws.String(bucketName),
    })
    if err != nil {
        if aerr, ok := err.(awserr.Error); ok {
            switch aerr.Code() {
            case "NotFound", "NoSuchBucket":
                fmt.Printf("Bucket '%s' not found, creating new bucket...\n", bucketName)
                // Create bucket if it doesn't exist
                _, err = svc.CreateBucket(&s3.CreateBucketInput{
                    Bucket: aws.String(bucketName),
                })
                if err != nil {
                    fmt.Printf("❌ Failed to create bucket: %v\n", err)
                    return fmt.Errorf("failed to create bucket: %v", err)
                }
                fmt.Printf("✅ Created new S3 bucket: %s\n", bucketName)
            default:
                fmt.Printf("❌ Error checking bucket: %v\n", err)
                return fmt.Errorf("error checking bucket: %v", err)
            }
        }
    } else {
        fmt.Printf("✅ Bucket '%s' exists\n", bucketName)
    }

    // Convert data to JSON
    fmt.Println("Converting data to JSON format...")
    jsonData, err := json.MarshalIndent(data, "", "    ")
    if err != nil {
        fmt.Printf("❌ JSON marshaling failed: %v\n", err)
        return fmt.Errorf("failed to marshal JSON: %v", err)
    }
    fmt.Printf("✅ Data converted to JSON (size: %d bytes)\n", len(jsonData))

    // Generate key with timestamp
    timestamp := time.Now().Format("2006-01-02-15-04-05")
    key := fmt.Sprintf("highlights/%s/data.json", timestamp)
    fmt.Printf("Generated S3 key: %s\n", key)

    // Upload to S3
    fmt.Println("Uploading data to S3...")
    startTime := time.Now()
    _, err = svc.PutObject(&s3.PutObjectInput{
        Bucket:      aws.String(bucketName),
        Key:         aws.String(key),
        Body:        bytes.NewReader(jsonData),
        ContentType: aws.String("application/json"),
    })
    if err != nil {
        fmt.Printf("❌ Upload failed: %v\n", err)
        return fmt.Errorf("failed to upload to S3: %v", err)
    }
    
    duration := time.Since(startTime)
    fmt.Printf("✅ Successfully uploaded data to s3://%s/%s (took %v)\n", bucketName, key, duration)
    return nil
}