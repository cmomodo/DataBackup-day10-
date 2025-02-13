package main

import (
	"log"
	"os/exec"
	"time"
)

// runCommand logs and executes a Go script using "go run <path>".
func runCommand(name string, path string) error {
    log.Printf("Starting command: go run %s", path)
    cmd := exec.Command("go", "run", path)
    output, err := cmd.CombinedOutput()
    log.Printf("Output for %s:\n%s", name, output)
    if err != nil {
        log.Printf("Error running %s: %v", name, err)
        return err
    }
    log.Printf("Completed command: go run %s", path)
    return nil
}

func main() {
    log.Println("===== Orchestrator Start =====")

    // Run fetch.go
    if err := runCommand("fetch", "./fetch/fetch.go"); err != nil {
        log.Fatalf("Fetch command failed: %v", err)
    }

    // Wait 10 seconds
    log.Println("Waiting for 10 seconds before next command...")
    time.Sleep(10 * time.Second)

    // Run video_process.go
    if err := runCommand("video_process", "./process_video/video_process.go"); err != nil {
        log.Fatalf("Video process command failed: %v", err)
    }

    // Wait 10 seconds
    log.Println("Waiting for 10 seconds before next command...")
    time.Sleep(10 * time.Second)

    // Run media_convert.go
    if err := runCommand("media_convert", "./media_convert/media_convert.go"); err != nil {
        log.Fatalf("Media convert command failed: %v", err)
    }

    log.Println("===== Orchestrator Completed =====")
}