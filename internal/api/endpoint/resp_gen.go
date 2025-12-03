package endpoint

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type StreamResponse struct {
	Model     string  `json:"model"`
	CreatedAt string  `json:"created_at"`
	Message   Message `json:"message"`
	Done      bool    `json:"done"`
	Choices   []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Think    string    `json:"think,omitempty"`
}

type Provider interface {
	GetEndpoint() string
	GetModel() string
	PrepareRequest(req Request) ([]byte, error)
	ParseStreamResponse(line string) (string, bool, error)
}

type OpenRouterProvider struct {
	APIKey string
	Model  string
}

func (p OpenRouterProvider) GetEndpoint() string {
	return "https://openrouter.ai/api/v1/chat/completions"
}

func (p OpenRouterProvider) GetModel() string {
	if p.Model == "" {
		return "anthropic/claude-3-haiku"
	}
	return p.Model
}

func (p OpenRouterProvider) PrepareRequest(req Request) ([]byte, error) {
	// Remove Think field for OpenAI compatibility
	req.Think = ""
	return json.Marshal(req)
}

func (p OpenRouterProvider) ParseStreamResponse(line string) (string, bool, error) {
	if line == "[DONE]" {
		return "", true, nil
	}

	var streamResp StreamResponse
	if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
		return "", false, err
	}

	if len(streamResp.Choices) > 0 {
		content := streamResp.Choices[0].Delta.Content
		done := streamResp.Choices[0].FinishReason != ""
		return content, done, nil
	}

	return "", false, nil
}

func getProvider() Provider {
	const apiKey = ""
	const model = "x-ai/grok-4.1-fast:free"
	return OpenRouterProvider{APIKey: apiKey, Model: model}
}

type Response struct {
	Message Message `json:"message"`
}

type ResponseTitle struct {
	Text string `json:"text"`
}

func GenerateTitle(content string) (string, error) {
	provider := getProvider()

	prompt := fmt.Sprintf(`Я отправлю тебе отврывок чата, выведи потенциальное краткое название чата онсновывась на содержании сообщения. Выведи только название чата без дополнительных комментариев. Содержание сообщения: --- %s ---`, content)

	conversation := []Message{
		{Role: "system", Content: ""},
		{Role: "user", Content: prompt},
	}

	req := Request{
		Model:    provider.GetModel(),
		Messages: conversation,
		Stream:   false,
	}

	jsonReq, err := provider.PrepareRequest(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest("POST", provider.GetEndpoint(), strings.NewReader(string(jsonReq)))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	if orProvider, ok := provider.(OpenRouterProvider); ok {
		httpReq.Header.Set("Authorization", "Bearer "+orProvider.APIKey)
		httpReq.Header.Set("HTTP-Referer", "https://github.com")
		httpReq.Header.Set("X-Title", "Document Fields Generator")
	}

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response *ResponseTitle

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", fmt.Errorf("ошибка при разборе JSON ответа: %w", err)
	}

	return response.Text, nil
}

type DocumentFields struct {
	Document string
	Fields   []string
}

func generateDocumentFields(documents []string) ([]DocumentFields, error) {
	provider := getProvider()

	var results []DocumentFields

	for _, doc := range documents {
		prompt := fmt.Sprintf(`For the document type "%s", generate a list of 2-4 essential fields that would need to be filled out to complete this document. 

Respond with ONLY a JSON array of field names, like: ["field1", "field2", "field3"]

Do not include any other text or explanation.`, doc)

		conversation := []Message{
			{Role: "system", Content: "You are a document analysis assistant. Always respond with only valid JSON arrays."},
			{Role: "user", Content: prompt},
		}

		req := Request{
			Model:    provider.GetModel(),
			Messages: conversation,
			Stream:   false,
		}

		jsonReq, err := provider.PrepareRequest(req)
		if err != nil {
			return nil, err
		}

		httpReq, err := http.NewRequest("POST", provider.GetEndpoint(), strings.NewReader(string(jsonReq)))
		if err != nil {
			return nil, err
		}
		httpReq.Header.Set("Content-Type", "application/json")

		if orProvider, ok := provider.(OpenRouterProvider); ok {
			httpReq.Header.Set("Authorization", "Bearer "+orProvider.APIKey)
			httpReq.Header.Set("HTTP-Referer", "https://github.com")
			httpReq.Header.Set("X-Title", "Document Fields Generator")
		}

		client := &http.Client{}
		resp, err := client.Do(httpReq)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		// Try to parse as OpenRouter format
		var openRouterResp struct {
			Choices []struct {
				Message Message `json:"message"`
			} `json:"choices"`
		}

		var content string
		if err := json.Unmarshal(bodyBytes, &openRouterResp); err == nil && len(openRouterResp.Choices) > 0 {
			content = openRouterResp.Choices[0].Message.Content
		} else {
			var response Response
			if err := json.Unmarshal(bodyBytes, &response); err != nil {
				return nil, fmt.Errorf("failed to parse response for document %s: %v", doc, err)
			}
			content = response.Message.Content
		}

		// Parse the JSON array of fields
		var fields []string
		if err := json.Unmarshal([]byte(content), &fields); err != nil {
			return nil, fmt.Errorf("failed to parse fields JSON for document %s: %v", doc, err)
		}

		// Ensure we have 2-4 fields
		if len(fields) < 2 {
			return nil, fmt.Errorf("document %s generated fewer than 2 fields", doc)
		}
		if len(fields) > 4 {
			fields = fields[:4] // Truncate to 4 fields
		}

		results = append(results, DocumentFields{
			Document: doc,
			Fields:   fields,
		})
	}

	return results, nil
}

func matchDocument(userMessage string, documents []string) (int, error) {
	if len(documents) == 0 {
		return 0, fmt.Errorf("no documents provided")
	}

	provider := getProvider()

	// Create numbered list of documents for the prompt
	var docList strings.Builder
	for i, doc := range documents {
		docList.WriteString(fmt.Sprintf("%d. %s\n", i+1, doc))
	}

	prompt := fmt.Sprintf(`You are a document matching system. Given a user message and a list of documents, identify which document best matches the user's intent.

User message: %s

Documents:
%s

Respond with ONLY the number (1-%d) of the best matching document. Do not include any other text or explanation.`, userMessage, docList.String(), len(documents))

	conversation := []Message{
		{Role: "system", Content: "You are a precise document matching assistant. Always respond with only a single number. You must think about it before responding to request."},
		{Role: "user", Content: prompt},
	}

	req := Request{
		Model:    provider.GetModel(),
		Messages: conversation,
		Stream:   false,
	}

	jsonReq, err := provider.PrepareRequest(req)
	if err != nil {
		return 0, err
	}

	httpReq, err := http.NewRequest("POST", provider.GetEndpoint(), strings.NewReader(string(jsonReq)))
	if err != nil {
		return 0, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	if orProvider, ok := provider.(OpenRouterProvider); ok {
		httpReq.Header.Set("Authorization", "Bearer "+orProvider.APIKey)
		httpReq.Header.Set("HTTP-Referer", "https://github.com")
		httpReq.Header.Set("X-Title", "Document Matcher")
	}

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Read the entire response body for debugging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	fmt.Printf("DEBUG: Raw response body: %s\n", string(bodyBytes))

	// Try to parse as OpenRouter format with choices
	var openRouterResp struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(bodyBytes, &openRouterResp); err == nil && len(openRouterResp.Choices) > 0 {
		fmt.Println("LLM:")
		fmt.Println(openRouterResp.Choices[0].Message.Content)

		// Extract the number from the response
		re := regexp.MustCompile(`\d+`)
		match := re.FindString(openRouterResp.Choices[0].Message.Content)
		if match == "" {
			return 0, fmt.Errorf("no number found in response")
		}

		docNum, err := strconv.Atoi(match)
		if err != nil {
			return 0, err
		}

		// Validate the range
		if docNum < 1 || docNum > len(documents) {
			return 0, fmt.Errorf("document number %d is out of range (1-%d)", docNum, len(documents))
		}

		return docNum, nil
	}

	// Fallback to original format
	var response Response
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return 0, fmt.Errorf("failed to parse response: %v", err)
	}

	fmt.Println("LLM:")
	fmt.Println(response.Message.Content)

	// Extract the number from the response
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(response.Message.Content)
	if match == "" {
		return 0, fmt.Errorf("no number found in response")
	}

	docNum, err := strconv.Atoi(match)
	if err != nil {
		return 0, err
	}

	// Validate the range
	if docNum < 1 || docNum > len(documents) {
		return 0, fmt.Errorf("document number %d is out of range (1-%d)", docNum, len(documents))
	}

	return docNum, nil
}

var docNumbers map[int]string = map[int]string{
	1: "claim",
}

var urlTemplateStorage = "C:\\Users\\salam\\quattroProject\\containers\\"

func collectingTags(docNum int) ([]string, string, error) {
	docName, exists := docNumbers[docNum]
	if !exists {
		return nil, "", fmt.Errorf("document with number %s not found", docNum)
	}

	fullPath := urlTemplateStorage + docName + ".txt"

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, "", fmt.Errorf("error reading file %s: %w", fullPath, err)
	}

	textContent := string(content)

	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(textContent, -1)

	uniqueTags := make(map[string]bool)
	var tags []string

	for _, match := range matches {
		if len(match) > 1 {
			tag := strings.TrimSpace(match[1])
			if !uniqueTags[tag] {
				uniqueTags[tag] = true
				tags = append(tags, tag)
			}
		}
	}

	return tags, docName, nil
}
