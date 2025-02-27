package deepseek

// AIRequest represents the payload sent to the Deepseek API for chat completions.
type AIRequest struct {
	// Model specifies the identifier of the model to use for generating the completion.
	// Example: "deepseek-chat" or "deepseek-reasoner"
	Model string `json:"model"`

	// Messages is an array of message objects that constitute the conversation history.
	// Each message should include a role ("system", "user", or "assistant") and content.
	Messages []Message `json:"messages"`

	// Stream specifies if the response is streamed or not.
	// The default is false
	Stream bool `json:"stream"`
}

type ResponseDeepSeek struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role             string `json:"role"`
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
		} `json:"message"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens        int `json:"prompt_tokens"`
		CompletionTokens    int `json:"completion_tokens"`
		TotalTokens         int `json:"total_tokens"`
		PromptTokensDetails struct {
			CachedTokens int `json:"cached_tokens"`
		} `json:"prompt_tokens_details"`
		CompletionTokensDetails struct {
			ReasoningTokens int `json:"reasoning_tokens"`
		} `json:"completion_tokens_details"`
		PromptCacheHitTokens  int `json:"prompt_cache_hit_tokens"`
		PromptCacheMissTokens int `json:"prompt_cache_miss_tokens"`
	} `json:"usage"`
	SystemFingerprint string `json:"system_fingerprint"`
}

// Message represents a single message in the conversation.
type Message struct {
	// Role specifies the role of the message author.
	// Acceptable values are "system", "user", or "assistant".
	Role string `json:"role"`

	// Content contains the text content of the message.
	Content string `json:"content"`
}

func NewDeepseekPayload() *AIRequest {
	return &AIRequest{
		Model:  "deepseek-chat",
		Stream: false,
	}
}

func (d *AIRequest) NewMessage(message Message) {
	d.Messages = append(d.Messages, message)
}
