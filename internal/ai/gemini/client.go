package gemini

import (
	"context"
	"os"

	"google.golang.org/genai"
)

type Service struct {
	Client  *genai.Client
	Model   string
	Gateway Gateway
}

type Gateway interface {
	Execute(ctx context.Context, name string, args map[string]any) (any, error)
}

func NewService(ctx context.Context, gateway Gateway) (*Service, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, err
	}
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.5-flash"
	}
	return &Service{Client: client, Model: model, Gateway: gateway}, nil
}

func (s *Service) Generate(ctx context.Context, prompt string, tools []*genai.FunctionDeclaration) (*genai.GenerateContentResponse, error) {
	return s.Client.Models.GenerateContent(ctx, s.Model, genai.Text(prompt), &genai.GenerateContentConfig{
		Tools: []*genai.Tool{{FunctionDeclarations: tools}},
	})
}
