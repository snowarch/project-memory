package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	GroqAPIURL = "https://api.groq.com/openai/v1/chat/completions"
	Model      = "moonshotai/kimi-k2-instruct"
)

type GroqClient struct {
	apiKey     string
	httpClient *http.Client
}

type GroqRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GroqResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

func NewGroqClient(apiKey string) *GroqClient {
	return &GroqClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *GroqClient) Analyze(systemPrompt, userPrompt string) (string, int, error) {
	reqBody := GroqRequest{
		Model: Model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.3,
		MaxTokens:   2000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", GroqAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var groqResp GroqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return "", 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return "", 0, fmt.Errorf("no response from API")
	}

	return groqResp.Choices[0].Message.Content, groqResp.Usage.TotalTokens, nil
}

func (c *GroqClient) AnalyzeProject(projectName, description, technologies, readme string) (string, int, error) {
	systemPrompt := `You are a senior software engineer analyzing development projects. Provide concise, actionable insights about project status, progress, and next steps. Focus on technical accuracy.`

	userPrompt := fmt.Sprintf(`Analyze this project:

Project: %s
Description: %s
Technologies: %s

README excerpt:
%s

Provide:
1. Current state assessment (2-3 sentences)
2. Estimated completion percentage (0-100)
3. Key next steps (3-5 items)
4. Technical concerns or blockers

Format your response as clear, structured text.`, projectName, description, technologies, readme)

	return c.Analyze(systemPrompt, userPrompt)
}

func (c *GroqClient) SummarizeTODOs(todos string) (string, int, error) {
	systemPrompt := `You are analyzing TODO items from source code. Provide a brief summary of the main tasks, priorities, and overall progress.`

	userPrompt := fmt.Sprintf(`Summarize these TODO items:

%s

Provide:
1. Main themes (2-3 items)
2. Priority breakdown
3. Quick wins vs long-term tasks`, todos)

	return c.Analyze(systemPrompt, userPrompt)
}
