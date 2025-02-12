package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

// VideoRecord represents a single video record within the JSON file.
type VideoRecord struct {
    URL string `json:"url"`
    // other fields can be added if needed
}

// JSONData represents the structure of the JSON file stored in S3.
type JSONData struct {
    Data []VideoRecord `json:"data"`
}

func init() {
    log.Println("Initializing application...")
    // Load .env file
    err := godotenv.Load(".env")
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }
    log.Println("Successfully loaded .env file.")
}

// ProcessVideos retrieves the JSON from the S3 bucket, downloads video files from the URLs,
// and uploads each video back to S3 under a single key.
func ProcessVideos() error {
    log.Println("Starting video processing...")

    // Initialize AWS session and S3 client using AWS_REGION from .env
    region := os.Getenv("AWS_REGION")
    bucketName := os.Getenv("S3_BUCKET_NAME")
    inputKey := os.Getenv("INPUT_KEY")   // JSON file key
    outputKey := os.Getenv("OUTPUT_KEY") // Output key for video file

    log.Printf("Initializing AWS session for region %s...", region)
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(region),
    })
    if err != nil {
        log.Printf("❌ AWS session creation failed: %v", err)
        return fmt.Errorf("failed to create AWS session: %v", err)
    }
    s3Client := s3.New(sess)
    log.Println("✅ AWS session and S3 client initialized.")

    // Retrieve the JSON file from S3
    log.Printf("Retrieving JSON file from S3: bucket=%s, key=%s...", bucketName, inputKey)
    objOut, err := s3Client.GetObject(&s3.GetObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(inputKey),
    })
    if err != nil {
        log.Printf("❌ Failed to retrieve JSON file from S3: %v", err)
        return fmt.Errorf("failed to retrieve JSON: %v", err)
    }
    defer objOut.Body.Close()

    jsonBytes, err := io.ReadAll(objOut.Body)
    if err != nil {
        log.Printf("❌ Error reading JSON object body: %v", err)
        return fmt.Errorf("failed to read JSON body: %v", err)
    }
    log.Printf("✅ Successfully retrieved JSON file from S3 (size: %d bytes).", len(jsonBytes))

    // Parse the JSON content and extract video URLs from the data field.
    var jsonData JSONData
    if err := json.Unmarshal(jsonBytes, &jsonData); err != nil {
        log.Printf("❌ Error parsing JSON: %v", err)
        return fmt.Errorf("failed to unmarshal JSON: %v", err)
    }
    log.Printf("✅ Parsed JSON content and found %d video records.", len(jsonData.Data))

    // Loop through each video record and process it.
    for index, record := range jsonData.Data {
        // Check if a valid video URL is present.
        if record.URL == "" {
            log.Printf("Warning: skipped record %d due to missing video URL.", index)
            continue
        }
        log.Printf("Processing record %d with video URL: %s", index, record.URL)

        // Download the video using yt-dlp
        cmd := exec.Command("yt-dlp",
            "-f", "best[height<=360]", // Select the best quality under or equal to 360p
            "-o", "-", // Output to stdout
            record.URL)

        log.Printf("Executing yt-dlp command: %s", cmd.String())

        var videoBuffer bytes.Buffer
        cmd.Stdout = &videoBuffer
        cmd.Stderr = os.Stderr // Redirect stderr to the console

        err = cmd.Run()
        if err != nil {
            log.Printf("❌ yt-dlp download failed for URL '%s': %v", record.URL, err)
            continue
        }

        log.Printf("✅ Successfully downloaded video using yt-dlp (size: %d bytes).", videoBuffer.Len())

        // Upload the video back to the S3 bucket
        _, err = s3Client.PutObject(&s3.PutObjectInput{
            Bucket:      aws.String(bucketName),
            Key:         aws.String(outputKey),
            Body:        bytes.NewReader(videoBuffer.Bytes()),
            ContentType: aws.String("video/mp4"),
        })
        if err != nil {
            log.Printf("❌ Failed to upload video '%s' to S3: %v", outputKey, err)
            continue
        }
        log.Printf("✅ Successfully uploaded video '%s' to S3.", outputKey)
    }

    log.Println("Completed processing all video records.")
    return nil
}

func main() {
    log.Println("===== Video Process Start =====")
    if err := ProcessVideos(); err != nil {
        log.Fatalf("Error processing videos: %v", err)
    }
    log.Println("===== Video Process Completed =====")
}