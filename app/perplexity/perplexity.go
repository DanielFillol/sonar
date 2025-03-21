package perplexity

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"os"
)

func Search(system, question, model string) (*ResponsePerplexity, error) {
	url := "https://api.perplexity.ai/chat/completions"

	// Struct to properly format JSON payload
	if system == "" {
		system = "Be precise and concise."
	}

	payload := NewPerplexityPayload()
	if model != "" {
		payload.Model = model
	}
	payload.NewMessage(Message{
		Role:    "system",
		Content: system,
	})
	payload.NewMessage(Message{
		Role:    "user",
		Content: question,
	})

	// Convert the payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.New("error marshalling JSON: " + err.Error())
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.New("error creating request: " + err.Error())
	}

	// Load .env file
	err = godotenv.Load(".env")
	if err != nil {
		return nil, errors.New("error loading .env file: " + err.Error())
	}

	token := os.Getenv("PPLX_API_KEY")
	if token == "" {
		return nil, errors.New("missing PPLX_API_KEY in environment")
	}

	// Set headers
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.New("error making request: " + err.Error())
	}
	defer res.Body.Close()

	// Read response
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("error reading response: " + err.Error())
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("error on response, current status: " + res.Status)
	}

	var result ResponsePerplexity
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
