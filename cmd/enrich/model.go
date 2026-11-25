package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"google.golang.org/genai"
)

// The models. Claude answers; Gemini on Vertex AI is the backup, asked only when the Claude API is
// unavailable (see unavailable).
const (
	claudeModel = "claude-opus-5-5"
	geminiModel = "gemini-3.8-flash"
	location    = "global"
)

// models asks Claude, and Gemini when Claude is unavailable. gemini is nil without OES_GCP_PROJECT.
type models struct {
	claude anthropic.Client
	gemini *genai.Client
}

// newClient builds the Claude client from ANTHROPIC_API_KEY (or an `ant auth login` profile) and,
// when OES_GCP_PROJECT is set, the Vertex AI backup from Application Default Credentials.
func newClient(ctx context.Context) (*models, error) {
	m := &models{claude: anthropic.NewClient()}
	project := os.Getenv("OES_GCP_PROJECT")
	if project == "" {
		return m, nil
	}
	gemini, err := genai.NewClient(ctx, &genai.ClientConfig{Backend: genai.BackendVertexAI, Project: project, Location: location})
	if err != nil {
		return nil, fmt.Errorf("vertex ai backup: %w", err)
	}
	m.gemini = gemini
	return m, nil
}

// ask sends no response schema: the shape is in the prompt and the validator enforces it.
func ask(ctx context.Context, m *models, system, prompt string) ([]byte, error) {
	var last error
	for attempt := range 5 {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*attempt) * 3 * time.Second)
		}
		text, err := m.answer(ctx, system, prompt)
		if err != nil {
			last = err
			fmt.Fprintf(os.Stderr, "attempt %d: %v\n", attempt+1, err)
			continue
		}
		if out := jsonBody(text); json.Valid(out) {
			return out, nil
		}
		last = errors.New("model returned invalid JSON")
	}
	return nil, last
}

// answer is one attempt: Claude, then Gemini if Claude is unavailable and a backup is configured.
func (m *models) answer(ctx context.Context, system, prompt string) (string, error) {
	text, err := m.askClaude(ctx, system, prompt)
	if err == nil || m.gemini == nil || ctx.Err() != nil || !unavailable(err) {
		return text, err
	}
	fmt.Fprintf(os.Stderr, "claude unavailable (%v); asking %s\n", err, geminiModel)
	return m.askGemini(ctx, system, prompt)
}

func (m *models) askClaude(ctx context.Context, system, prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	// Fallbacks: a request the model's safety classifiers decline is re-served by the API's
	// default fallback model inside the same call.
	resp, err := m.claude.Beta.Messages.New(ctx, anthropic.BetaMessageNewParams{
		Model:     claudeModel,
		MaxTokens: 16000,
		System:    []anthropic.BetaTextBlockParam{{Text: system}},
		Messages:  []anthropic.BetaMessageParam{anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(prompt))},
		Fallbacks: anthropic.BetaFallbacksParamOfDefault(),
		Betas:     []anthropic.AnthropicBeta{anthropic.AnthropicBetaServerSideFallback2026_07_01},
	})
	if err != nil {
		return "", fmt.Errorf("claude: %w", err)
	}
	if resp.StopReason != anthropic.BetaStopReasonEndTurn {
		return "", fmt.Errorf("claude: stopped with %s", resp.StopReason)
	}
	var text strings.Builder
	for _, block := range resp.Content {
		if t, ok := block.AsAny().(anthropic.BetaTextBlock); ok {
			text.WriteString(t.Text)
		}
	}
	return text.String(), nil
}

func (m *models) askGemini(ctx context.Context, system, prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	resp, err := m.gemini.Models.GenerateContent(ctx, geminiModel, genai.Text(prompt), &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(system, genai.RoleUser),
		ResponseMIMEType:  "application/json",
		Temperature:       genai.Ptr[float32](0),
	})
	if err != nil {
		return "", fmt.Errorf("gemini: %w", err)
	}
	return resp.Text(), nil
}

// unavailable reports whether a Claude error is an outage the backup should cover: a rate limit, a
// server error, a timeout or a transport failure. Any other API status (a bad request, a missing
// credential, a permission) is a fault the backup would only hide.
func unavailable(err error) bool {
	var apiErr *anthropic.Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 429 || apiErr.StatusCode >= 500
	}
	return true
}

// jsonBody is the answer without the Markdown code fence a model may wrap JSON in.
func jsonBody(text string) []byte {
	text = strings.TrimSpace(text)
	if rest, ok := strings.CutPrefix(text, "```"); ok {
		rest = strings.TrimPrefix(rest, "json")
		text = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(rest), "```"))
	}
	return []byte(text)
}
