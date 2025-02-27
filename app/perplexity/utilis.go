package perplexity

import "log"

type Perplexity struct {
	Response string
	Links    string
}

// RequestPerplexity represents a request payload for the Perplexity chat completions API.
type RequestPerplexity struct {
	// Model specifies the name of the model to use for generating the completion.
	// Refer to the supported models in the Perplexity API documentation.
	Model string `json:"model"`

	// Messages is a list of messages that make up the conversation so far.
	// Each message should have a role (system, user, or assistant) and content.
	Messages []Message `json:"messages"`

	// MaxTokens defines the maximum number of tokens to generate in the completion.
	// The total of prompt tokens and max_tokens must not exceed the model's context window.
	// If not specified, the model will generate tokens until it reaches a stop sequence or the context limit.
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature controls the randomness of the output.
	// Values range from 0 to 2, with higher values producing more random outputs.
	// Default is 0.2.
	Temperature float64 `json:"temperature,omitempty"`

	// TopP implements nucleus sampling, considering only the tokens with top_p probability mass.
	// Values range from 0 to 1. Default is 0.9.
	TopP float64 `json:"top_p,omitempty"`

	// SearchDomainFilter limits the citations used by the online model to URLs from the specified domains.
	// Up to 3 domains can be specified. Prefix a domain with '-' to blacklist it.
	// Only available in certain tiers.
	SearchDomainFilter []string `json:"search_domain_filter,omitempty"`

	// ReturnImages determines whether the response should include images.
	// Default is false. Only available in certain tiers.
	ReturnImages bool `json:"return_images,omitempty"`

	// ReturnRelatedQuestions determines whether the response should include related questions.
	// Default is false. Only available in certain tiers.
	ReturnRelatedQuestions bool `json:"return_related_questions,omitempty"`

	// SearchRecencyFilter restricts search results to a specified time frame.
	// Acceptable values are "hour", "day", "week", or "month".
	SearchRecencyFilter string `json:"search_recency_filter,omitempty"`

	// TopK specifies the number of highest probability tokens to consider for top-k filtering.
	// Values range from 0 to 2048. A value of 0 disables top-k filtering. Default is 0.
	TopK int `json:"top_k,omitempty"`

	// Stream determines whether to stream the response incrementally using server-sent events.
	// Default is false.
	Stream bool `json:"stream,omitempty"`

	// PresencePenalty applies a penalty to the likelihood of new tokens based on their presence in the text so far.
	// Values range from -2.0 to 2.0. Positive values increase the model's likelihood to discuss new topics.
	// Incompatible with FrequencyPenalty. Default is 0.
	PresencePenalty float64 `json:"presence_penalty,omitempty"`

	// FrequencyPenalty applies a penalty to the likelihood of new tokens based on their frequency in the text so far.
	// Values greater than 1.0 decrease the model's tendency to repeat the same line verbatim.
	// Incompatible with PresencePenalty. Default is 1.
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`

	// ResponseFormat enables structured outputs with a JSON or Regex schema.
	// Only available in certain tiers.
	ResponseFormat interface{} `json:"response_format,omitempty"`
}

// Message represents a single message in the conversation.
type Message struct {
	// Role specifies the role of the message author.
	// Acceptable values are "system", "user", or "assistant".
	Role string `json:"role"`

	// Content contains the text content of the message.
	Content string `json:"content"`
}

func NewPerplexityPayload() *RequestPerplexity {
	return &RequestPerplexity{
		Model: "sonar-pro",
		//MaxTokens:              123,
		Temperature:            0.2,
		TopP:                   0.9,
		SearchDomainFilter:     nil,
		ReturnImages:           false,
		ReturnRelatedQuestions: false,
		SearchRecencyFilter:    "",
		TopK:                   0,
		Stream:                 false,
		PresencePenalty:        0,
		FrequencyPenalty:       1,
		ResponseFormat:         nil,
	}
}

func (r *RequestPerplexity) NewMessage(message Message) {
	r.Messages = append(r.Messages, message)
}

// AddSearchFilter
//
//	Given a list of domains, limit the citations used by the online model to URLs from the specified domains.
//	Currently limited to only 3 domains for whitelisting and blacklisting.
//	For blacklisting add a - to the beginning of the domain string.
func (r *RequestPerplexity) AddSearchFilter(urls []string) {
	for _, url := range urls {
		r.SearchDomainFilter = append(r.SearchDomainFilter, url)
	}
}

// ResponsePerplexity represents the response payload from the Perplexity chat completions API.
type ResponsePerplexity struct {
	// Id is a unique identifier generated for each response.
	Id string `json:"id"`

	// Model specifies the name of the model used to generate the response.
	Model string `json:"model"`

	// Created is the Unix timestamp (in seconds) indicating when the completion was created.
	Created int `json:"created"`

	// Usage provides statistics about token usage for the completion request.
	Usage struct {
		// PromptTokens is the number of tokens provided in the request prompt.
		PromptTokens int `json:"prompt_tokens"`

		// CompletionTokens is the number of tokens generated in the response.
		CompletionTokens int `json:"completion_tokens"`

		// TotalTokens is the total number of tokens used (prompt + completion).
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`

	// Citations contains a list of URLs cited in the generated answer.
	Citations []string `json:"citations"`

	// Object specifies the object type, which is always "chat.completion" for this endpoint.
	Object string `json:"object"`

	// Choices is a list of completion choices generated for the input prompt.
	Choices []struct {
		// Index is the position of this completion choice in the list.
		Index int `json:"index"`

		// FinishReason indicates why the model stopped generating tokens. Possible values include:
		// - "stop": The model reached a natural stopping point.
		// - "length": The maximum number of tokens specified in the request was reached.
		FinishReason string `json:"finish_reason"`

		// Message contains the message generated by the model.
		Message struct {
			// Role specifies the role of the message author. Possible values are:
			// - "system"
			// - "user"
			// - "assistant"
			Role string `json:"role"`

			// Content is the text content of the message.
			Content string `json:"content"`
		} `json:"message"`

		// Delta contains the incrementally streamed next tokens. This field is only meaningful when streaming is enabled (`stream=true`).
		Delta struct {
			// Role specifies the role of the message author. Possible values are:
			// - "system"
			// - "user"
			// - "assistant"
			Role string `json:"role"`

			// Content is the text content of the message.
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

// ExtractLinks Concatenate the relevant citation links.
func (r *ResponsePerplexity) ExtractLinks() string {
	var links string
	for i, citation := range r.Citations {
		links += citation
		if i != len(r.Citations)-1 {
			links += ", "
		} else {
			links += "."
		}
	}
	return links
}

func SearchForQuotes(system, field, authors, promptJuris, prompt string) (*Perplexity, error) {
	log.Println("Estudando a doutrina")
	searchInput := "O ramo do direito é:\n" + field +
		"\nOs doutrinadores relevantes são:\n" + authors +
		"\nO resumo do prompt do usuário é\n" + promptJuris +
		"\nO prompt original do usuário é:\n" + prompt

	papers, err := Search(system, searchInput, "sonar")
	if err != nil {
		return nil, err
	}

	return &Perplexity{
		Response: papers.Choices[0].Message.Content,
		Links:    papers.ExtractLinks(),
	}, nil
}

func SearchForLaws(system, field, prompt, promptJuris string) (*Perplexity, error) {
	log.Println("Estudando as leis relevantes")
	lawInput := "O ramo do direito é:\n" + field +
		"\nO prompt original do usuário é:\n" + prompt +
		"\nO resumo do prompt do usuário é\n" + promptJuris

	laws, err := Search(system, "quais as leis relevantes no site do planalto.gov que se relacionam com: "+lawInput+". Por favor liste todos os links do site do planalto com as leis mencionadas", "sonar")
	if err != nil {
		return nil, err
	}

	return &Perplexity{
		Response: laws.Choices[0].Message.Content,
		Links:    laws.ExtractLinks(),
	}, nil
}
