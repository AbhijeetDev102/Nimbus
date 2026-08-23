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
)

type ConnectorPayload struct {
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

func main() {
	// read the json file from /infra/development/k8s/register-postres-outbox.json

	content, err := os.ReadFile("infra/development/k8s/register-postgres-outbox.json")
	if err != nil {
		log.Fatalf("something went wrong while reading the config of kafka connector : %v", err)
	}

	var connector *ConnectorPayload
	// Unmarshal the file into ConnectorPayload
	err = json.Unmarshal(content, &connector)
	if err != nil {
		log.Fatalf("json unmarshal failed : %v", err)
	}

	// Marshal just the Config map into bytes for the PUT request body

	marshaledConfig, err := json.Marshal(connector.Config)
	if err != nil {
		log.Printf("Failed to marshal register config : %v", err)
	}

	// Build the PUT URL: "http://localhost:8083/connectors/" + payload.Name + "/config"

	URL := fmt.Sprintf("http://localhost:8083/connectors/%s/config", connector.Name)

	// Run a retry loop (e.g. 30 attempts, 2s sleep) using http.Client with a timeout:
	//    - If http.NewRequestWithContext/client.Do returns error or status >= 500: retry.
	//    - If status == 200 or 201: print success and exit cleanly!

	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	maxAttempts := 30
	retryInterval := 2 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequest(http.MethodPut, URL, bytes.NewReader(marshaledConfig))
		if err != nil {
			log.Fatalf("Failed to create HTTP request: %v", err)

		}

		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			log.Printf("[Attempt %d/%d] Kafka Connect not ready (%v). Retrying in %v...", attempt, maxAttempts, err, retryInterval)
			time.Sleep(retryInterval)
			continue
		}

		bodyBytes, _ := io.ReadAll(res.Body)
		res.Body.Close()

		if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
			log.Printf(" Successfully registered connector '%s' with Kafka Connect!", connector.Name)
			return
		}

		log.Printf("[Attempt %d/%d] Kafka Connect rejected config (Status %d): %s", attempt, maxAttempts, res.StatusCode, string(bodyBytes))
		time.Sleep(retryInterval)

	}

	log.Fatalf("❌ Failed to register connector '%s' after %d attempts", connector.Name, maxAttempts)

}
