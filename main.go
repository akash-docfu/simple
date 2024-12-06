package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/storage"
	"github.com/joho/godotenv"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

type FileResponse struct {
	Content string `json:"content"`
	Status  int    `json:"status"`
}

var storageClient *storage.Client

func getSecretFromManager(projectID, secretName, version string) (string, error) {
	ctx := context.Background()

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create secretmanager client: %v", err)
	}
	defer client.Close()

	// Build the secret name
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/%s", projectID, secretName, version)

	// Access the secret
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret: %v", err)
	}

	return string(result.Payload.Data), nil
}

func initStorageClient() error {
	ctx := context.Background()

	// Get required environment variables
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	secretName := os.Getenv("CREDENTIALS_SECRET_NAME")
	if projectID == "" || secretName == "" {
		return fmt.Errorf("GOOGLE_CLOUD_PROJECT and CREDENTIALS_SECRET_NAME must be set")
	}

	// Get credentials from Secret Manager
	credentials, err := getSecretFromManager(projectID, secretName, "latest")
	if err != nil {
		return fmt.Errorf("failed to get credentials: %v", err)
	}

	// Write credentials to a temporary file
	tmpFile := "/tmp/credentials.json"
	if err := os.WriteFile(tmpFile, []byte(credentials), 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %v", err)
	}

	// Set the credentials file path
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tmpFile)

	// Initialize storage client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %v", err)
	}
	storageClient = client
	return nil
}

func readFileFromGCS(bucketName, filePath string) (string, error) {
	ctx := context.Background()

	bucket := storageClient.Bucket(bucketName)
	obj := bucket.Object(filePath)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create reader: %v", err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	return string(content), nil
}

// ... [previous middleware and handler functions remain the same] ...

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize storage client with credentials from Secret Manager
	if err := initStorageClient(); err != nil {
		log.Fatalf("Failed to initialize storage client: %v", err)
	}

	// Check if required env vars are set
	if os.Getenv("API_KEY") == "" {
		log.Fatal("API_KEY must be set in .env file")
	}
	if os.Getenv("BUCKET_NAME") == "" {
		log.Fatal("BUCKET_NAME must be set in .env file")
	}

	// Define routes
	http.HandleFunc("/api/protected", validateAPIKey(protected))
	http.HandleFunc("/api/read-file", validateAPIKey(readFile))

	// Start server
	port := os.Getenv("PORT_NUM")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
