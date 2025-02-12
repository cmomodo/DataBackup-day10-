package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
	"github.com/joho/godotenv"
)

func init() {
    log.Println("Initializing application...")
    // Load environment variables from .env file
    err := godotenv.Load(".env")
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }
    log.Println("Successfully loaded .env file. (Ensure .env is in your .gitignore)")
}

// CreateMediaConvertJob creates a MediaConvert job to process a video from S3.
func CreateMediaConvertJob() error {
    log.Println("Starting MediaConvert job creation...")

    // Extract environment variables
    region := os.Getenv("AWS_REGION")
    endpoint := os.Getenv("MEDIACONVERT_ENDPOINT")
    bucketName := os.Getenv("S3_BUCKET_NAME")
    roleArn := os.Getenv("MEDIACONVERT_ROLE_ARN")

    // Define input and output S3 paths
    inputPath := fmt.Sprintf("s3://%s/videos/first_video.mp4", bucketName)
    outputPath := fmt.Sprintf("s3://%s/processed_videos/", bucketName)

    log.Printf("AWS_REGION: %s", region)
    log.Printf("MEDIACONVERT_ENDPOINT: %s", endpoint)
    log.Printf("Input video: %s", inputPath)
    log.Printf("Output location: %s", outputPath)

    // Initialize AWS session with custom endpoint
    sess, err := session.NewSession(&aws.Config{
        Region:   aws.String(region),
        Endpoint: aws.String(endpoint),
    })
    if err != nil {
        log.Printf("❌ Failed to create AWS session: %v", err)
        return err
    }
    log.Println("AWS session created successfully.")

    // Initialize MediaConvert client
    mcClient := mediaconvert.New(sess)
    log.Println("MediaConvert client initialized.")

    // Set up MediaConvert job settings
    jobSettings := &mediaconvert.JobSettings{
        Inputs: []*mediaconvert.Input{
            {
                FileInput: aws.String(inputPath),
                AudioSelectors: map[string]*mediaconvert.AudioSelector{
                    "Audio Selector 1": {
                        DefaultSelection: aws.String("DEFAULT"),
                    },
                },
                VideoSelector: &mediaconvert.VideoSelector{},
            },
        },
        OutputGroups: []*mediaconvert.OutputGroup{
            {
                Name: aws.String("File Group"),
                OutputGroupSettings: &mediaconvert.OutputGroupSettings{
                    Type: aws.String("FILE_GROUP_SETTINGS"),
                    FileGroupSettings: &mediaconvert.FileGroupSettings{
                        Destination: aws.String(outputPath),
                    },
                },
                Outputs: []*mediaconvert.Output{
                    {
                        ContainerSettings: &mediaconvert.ContainerSettings{
                            Container: aws.String("MP4"),
                            Mp4Settings: &mediaconvert.Mp4Settings{},
                        },
                        VideoDescription: &mediaconvert.VideoDescription{
                            CodecSettings: &mediaconvert.VideoCodecSettings{
                                Codec: aws.String("H_264"),
                                H264Settings: &mediaconvert.H264Settings{
                                    Bitrate:            aws.Int64(5000000),
                                    RateControlMode:    aws.String("CBR"),
                                    QualityTuningLevel: aws.String("SINGLE_PASS"),
                                    CodecProfile:       aws.String("MAIN"),
                                },
                            },
                            Width:  aws.Int64(1920),
                            Height: aws.Int64(1080),
                        },
                        AudioDescriptions: []*mediaconvert.AudioDescription{
                            {
                                CodecSettings: &mediaconvert.AudioCodecSettings{
                                    Codec: aws.String("AAC"),
                                    AacSettings: &mediaconvert.AacSettings{
                                        Bitrate:    aws.Int64(96000),
                                        CodingMode: aws.String("CODING_MODE_2_0"),
                                        SampleRate: aws.Int64(48000),
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }

    log.Println("Job settings defined:")
    settingsJSON, _ := json.MarshalIndent(jobSettings, "", "  ")
    log.Println(string(settingsJSON))

    // Create MediaConvert job input
    inputJob := &mediaconvert.CreateJobInput{
        Role:                 aws.String(roleArn),
        Settings:             jobSettings,
        AccelerationSettings: &mediaconvert.AccelerationSettings{Mode: aws.String("DISABLED")},
        StatusUpdateInterval: aws.String("SECONDS_60"),
    }

    log.Println("Creating MediaConvert job...")
    result, err := mcClient.CreateJob(inputJob)
    if err != nil {
        log.Printf("❌ Failed to create MediaConvert job: %v", err)
        return err
    }

    log.Println("MediaConvert job created successfully. Response:")
    resultJSON, _ := json.MarshalIndent(result, "", "  ")
    log.Println(string(resultJSON))

    return nil
}

func main() {
    log.Println("===== MediaConvert Job Process Start =====")
    if err := CreateMediaConvertJob(); err != nil {
        log.Fatalf("Error in MediaConvert job creation: %v", err)
    }
    log.Println("===== MediaConvert Job Process Completed =====")
}