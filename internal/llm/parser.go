package llm

import (
	"context"
	"encoding/json"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type ParsedQuery struct {
	SearchQuery string                 `json:"search_query"`
	Filters     map[string]interface{} `json:"filters"`
}

func ParseNaturalLanguageQuery(ctx context.Context, client *openai.Client, userInput string) (ParsedQuery, error) {
	systemPrompt := `You are a search query parser for a travel database.
Extract the semantic search query and any structured filters from the user's input.
The database has these metadata fields for filtering: "country" (string), "city" (string), "kinds" (string, comma-separated tags), "rate" (number, 1-7).
If the user mentions a country, put it in filters.country.
If the user mentions a specific city, put it in filters.city.
If the user mentions specific types of places (museums, nature, food), put them in the search query, but you can also try to map them to 'kinds' if they are very specific categories.
The 'search_query' should be the core topic for vector search (e.g. "hiking", "modern art", "coffee").
Return ONLY a JSON object with keys "search_query" and "filters".
Example: "I want to go hiking in France" -> {"search_query": "hiking", "filters": {"country": "France"}}
Example: "Museums in Tokyo with high rating" -> {"search_query": "museums", "filters": {"city": "Tokyo", "rate": {"$gte": 5}}} (Note: Pinecone filters use mongo-style operators like $gte, $eq, $in. For simple equality use the value directly.)
For this task, only use simple equality for strings. For 'rate', you can use $gte if the user asks for "good" or "high" rating (assume >= 5).`

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userInput,
				},
			},
		},
	)

	if err != nil {
		return ParsedQuery{}, err
	}

	content := resp.Choices[0].Message.Content
	// Clean up potential markdown code blocks
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result ParsedQuery
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// Fallback to using input as query if JSON parsing fails
		return ParsedQuery{SearchQuery: userInput}, nil
	}

	return result, nil
}
