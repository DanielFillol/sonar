package deepseek

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"os"
)

func Search(system, question, model string) (*ResponseDeepSeek, error) {
	// DeepSeek API endpoint (compatible with OpenAI's API format)
	url := "https://api.deepseek.com/chat/completions"

	// Provide a default system prompt if not set.
	if system == "" {
		system = "Be precise and concise."
	}

	payload := NewDeepseekPayload()
	if model != "" {
		payload.Model = model
	} else {
		payload.Model = "deepseek-chat"
	}
	payload.NewMessage(Message{
		Role:    "system",
		Content: system,
	})
	payload.NewMessage(Message{
		Role:    "user",
		Content: question,
	})

	// Marshal the payload to JSON.
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.New("error marshalling JSON: " + err.Error())
	}

	// Create the HTTP request.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.New("error creating request: " + err.Error())
	}

	// Load .env file to fetch environment variables.
	err = godotenv.Load(".env")
	if err != nil {
		return nil, errors.New("error loading .env file: " + err.Error())
	}

	// Retrieve the DeepSeek API key from the environment.
	token := os.Getenv("DEEPSEEK_API_KEY")
	if token == "" {
		return nil, errors.New("missing DEEPSEEK_API_KEY in environment")
	}

	// Set the required headers.
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	// Send the HTTP request.
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.New("error making request: " + err.Error())
	}
	defer res.Body.Close()

	// Read the response body.
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("error reading response: " + err.Error())
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("error on response, current status: " + res.Status)
	}

	// Unmarshal the JSON response into the OpenAIResponse struct.
	var result ResponseDeepSeek
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
