package gpt

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	maxRetries     = 5
	baseDelay      = 1 * time.Second
	maxDelay       = 30 * time.Second
	defaultTimeout = 30 * time.Second
)

type StreamProcessor func(string)

type RateLimiter struct {
	RequestsReset time.Time
	TokensReset   time.Time
}

func StreamSearch(system, question, model string, processor StreamProcessor) error {
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	token := os.Getenv("OPENAI_API_KEY")
	if token == "" {
		return errors.New("missing OPENAI_API_KEY in environment")
	}

	// Create request payload
	payload := NewGPTPayload()
	payload.Stream = true
	if model != "" {
		payload.Model = model
	}
	payload.NewMessage(Message{
		Role:    "developer",
		Content: system,
	})
	payload.NewMessage(Message{
		Role:    "user",
		Content: question,
	})

	var lastError error
	rateLimiter := &RateLimiter{}

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Calcular delay com backoff exponencial
		delay := calculateBackoff(attempt)
		time.Sleep(delay)

		// Verificar limites de rate
		if time.Now().Before(rateLimiter.RequestsReset) || time.Now().Before(rateLimiter.TokensReset) {
			waitTime := time.Until(rateLimiter.RequestsReset)
			if waitTime > maxDelay {
				waitTime = maxDelay
			}
			time.Sleep(waitTime)
			continue
		}

		// Criar cliente com timeout
		client := &http.Client{Timeout: defaultTimeout}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			lastError = fmt.Errorf("error marshalling JSON: %w", err)
			continue
		}

		req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
		if err != nil {
			lastError = fmt.Errorf("error creating request: %w", err)
			continue
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			lastError = fmt.Errorf("error making request: %w", err)
			continue
		}
		defer res.Body.Close()

		// Atualizar rate limiter com headers
		updateRateLimiter(res.Header, rateLimiter)

		switch res.StatusCode {
		case http.StatusOK:
			return processStreamResponse(res.Body, processor)
		case http.StatusTooManyRequests:
			lastError = handleRateLimitError(res.Header, attempt)
		case http.StatusBadRequest:
			lastError = fmt.Errorf("invalid request: check model and parameters")
		default:
			lastError = fmt.Errorf("unexpected status code: %d", res.StatusCode)
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastError)
}

func calculateBackoff(attempt int) time.Duration {
	delay := time.Duration(math.Pow(2, float64(attempt))) * baseDelay
	if delay > maxDelay {
		return maxDelay
	}
	return delay
}

func updateRateLimiter(headers http.Header, rl *RateLimiter) {
	if reset := headers.Get("x-ratelimit-reset-requests"); reset != "" {
		if resetTime, err := time.Parse(time.RFC3339, reset); err == nil {
			rl.RequestsReset = resetTime
		}
	}
	if reset := headers.Get("x-ratelimit-reset-tokens"); reset != "" {
		if resetTime, err := time.Parse(time.RFC3339, reset); err == nil {
			rl.TokensReset = resetTime
		}
	}
}

func handleRateLimitError(headers http.Header, attempt int) error {
	resetTime := time.Now().Add(20 * time.Second) // Default fallback
	if reset := headers.Get("x-ratelimit-reset-requests"); reset != "" {
		if parsed, err := time.Parse(time.RFC3339, reset); err == nil {
			resetTime = parsed
		}
	}
	return fmt.Errorf("rate limit exceeded (attempt %d), reset at: %v",
		attempt+1, resetTime.Format("15:04:05"))
}

func processStreamResponse(body io.ReadCloser, processor StreamProcessor) error {
	reader := bufio.NewReader(body)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading stream: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var response OpenAIStreamResponse
		if err := json.Unmarshal([]byte(data), &response); err != nil {
			continue
		}

		if len(response.Choices) > 0 && response.Choices[0].Delta.Content != "" {
			content := response.Choices[0].Delta.Content
			if processor != nil {
				processor(content)
			}
		}
	}
	return nil
}

func Search(system, question, model string) (*OpenAIResponse, error) {
	url := "https://api.openai.com/v1/chat/completions"

	// Struct to properly format JSON payload
	if system == "" {
		system = "Be precise and concise."
	}
	if model != "" {

	}

	// Create request payload
	payload := NewGPTPayload()
	if model != "" {
		payload.Model = model
	}
	payload.NewMessage(Message{
		Role:    "developer",
		Content: system,
	})
	payload.NewMessage(Message{
		Role:    "user",
		Content: question,
	})

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

	token := os.Getenv("OPENAI_API_KEY")
	if token == "" {
		return nil, errors.New("missing OPENAI_API_KEY in environment")
	}

	// Set Headers
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	// Send request
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

	var result OpenAIResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
