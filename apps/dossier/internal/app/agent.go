package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/webshoppe/dossier/internal/dossier"
)

// Agent context selection (documented in STATUS-GAPFIX-T4.2.md):
// - FTS5 SearchLoose: stopwords stripped, remaining terms OR-matched with prefix *
// - Hits ordered by bm25() (best first); only the top agentSearchLimit docs/notes are used
// - Per-hit body cap agentPerHitBodyCap; total context agentMaxContext runes
const (
	// Top few most relevant hits only — OR matching can surface many weak matches.
	agentSearchLimit   = 4
	agentPerHitBodyCap = 1800
	agentMaxContext    = 12000
	agentHTTPTimeout   = 90 * time.Second
)

// AskDossierRequest is the UI payload for a chat turn.
type AskDossierRequest struct {
	Question string `json:"question"`
}

// AskDossierResponse is returned to the webview.
type AskDossierResponse struct {
	Answer      string              `json:"answer"`
	ContextUsed []map[string]string `json:"contextUsed"`
	Endpoint    string              `json:"endpoint"`
	Error       string              `json:"error,omitempty"`
}

// AgentConfigured reports whether the optional agent panel should appear.
func AgentConfigured(cfg *dossier.AppConfig) bool {
	if cfg == nil {
		return false
	}
	host := strings.TrimSpace(cfg.Settings.AgentHost)
	port := cfg.Settings.AgentPort
	return host != "" && port > 0
}

func agentBaseURL(cfg *dossier.AppConfig) string {
	host := strings.TrimSpace(cfg.Settings.AgentHost)
	port := cfg.Settings.AgentPort
	path := strings.TrimSpace(cfg.Settings.AgentPath)
	if path == "" {
		path = "/v1/chat/completions"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// Allow host to include scheme
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return strings.TrimRight(host, "/") + path
	}
	return fmt.Sprintf("http://%s:%d%s", host, port, path)
}

// AskDossier runs FTS-backed context selection then POSTs to the user-configured agent.
func (a *API) AskDossier(req AskDossierRequest) AskDossierResponse {
	q := strings.TrimSpace(req.Question)
	if q == "" {
		return AskDossierResponse{Error: "question required"}
	}
	cfg, _ := dossier.LoadConfig()
	if !AgentConfigured(cfg) {
		return AskDossierResponse{Error: "agent not configured"}
	}
	s, err := a.requireStore()
	if err != nil {
		return AskDossierResponse{Error: err.Error()}
	}

	ctxText, used, err := buildAgentContext(s, q)
	if err != nil {
		return AskDossierResponse{Error: "context: " + err.Error()}
	}

	endpoint := agentBaseURL(cfg)
	answer, err := callAgentHTTP(endpoint, cfg.Settings.AgentToken, q, ctxText)
	if err != nil {
		return AskDossierResponse{
			Error:       err.Error(),
			ContextUsed: used,
			Endpoint:    endpoint,
		}
	}
	return AskDossierResponse{
		Answer:      answer,
		ContextUsed: used,
		Endpoint:    endpoint,
	}
}

func buildAgentContext(s *dossier.Store, question string) (string, []map[string]string, error) {
	// Natural-language questions need OR + stopword stripping (strict AND fails on "what/is/the…").
	hits, err := s.SearchLoose(question, agentSearchLimit)
	if err != nil {
		return "", nil, err
	}
	var b strings.Builder
	var used []map[string]string
	total := 0
	for _, h := range hits {
		if total >= agentMaxContext {
			break
		}
		title := h.Title
		if title == "" {
			title = h.Kind + " " + h.ID
		}
		chunk := h.Snippet
		// Enrich documents/notes with a body slice when possible
		switch h.Kind {
		case "document":
			body, _, rerr := s.ReadDocumentBody(h.ID)
			if rerr == nil && body != "" {
				chunk = truncateRunes(body, agentPerHitBodyCap)
			}
		case "note":
			n, rerr := s.GetNote(h.ID)
			if rerr == nil && n != nil && n.Body != "" {
				chunk = truncateRunes(n.Body, agentPerHitBodyCap)
			}
		case "sticky", "kanban", "decision":
			if h.Snippet != "" {
				chunk = h.Snippet
			} else {
				chunk = h.Title
			}
		}
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		block := fmt.Sprintf("### %s (%s)\n%s\n\n", title, h.Kind, chunk)
		if total+utf8.RuneCountInString(block) > agentMaxContext {
			remain := agentMaxContext - total
			if remain < 80 {
				break
			}
			block = truncateRunes(block, remain) + "\n"
		}
		b.WriteString(block)
		total += utf8.RuneCountInString(block)
		used = append(used, map[string]string{
			"id":    h.ID,
			"kind":  h.Kind,
			"title": title,
		})
	}
	if used == nil {
		used = []map[string]string{}
	}
	if b.Len() == 0 {
		b.WriteString("(No matching dossier content found via full-text search. Answer from general knowledge only if appropriate, and say when context was empty.)\n")
	}
	return b.String(), used, nil
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}

func callAgentHTTP(endpoint, token, question, context string) (string, error) {
	system := "You are a helpful assistant answering questions about the user's local dossier (documents, notes, stickies, decisions). " +
		"Use the provided CONTEXT excerpts when relevant. If context is insufficient, say so. Be concise and concrete."
	user := "CONTEXT:\n" + context + "\n\nQUESTION:\n" + question

	// OpenAI-compatible chat completions body (widely used by local gateways).
	payload := map[string]interface{}{
		"model": "local",
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"temperature": 0.2,
		// Also send simple fields some Hermes-style agents accept
		"question": question,
		"context":  context,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	}
	client := &http.Client{Timeout: agentHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("agent request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(body))
		if len(msg) > 400 {
			msg = msg[:400] + "…"
		}
		return "", fmt.Errorf("agent HTTP %d: %s", resp.StatusCode, msg)
	}
	return parseAgentAnswer(body)
}

func parseAgentAnswer(body []byte) (string, error) {
	// Try OpenAI chat completion shape
	var oai struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Text string `json:"text"`
		} `json:"choices"`
		Answer   string `json:"answer"`
		Response string `json:"response"`
		Content  string `json:"content"`
		Output   string `json:"output"`
		Error    interface{} `json:"error"`
	}
	if err := json.Unmarshal(body, &oai); err == nil {
		if len(oai.Choices) > 0 {
			if c := strings.TrimSpace(oai.Choices[0].Message.Content); c != "" {
				return c, nil
			}
			if c := strings.TrimSpace(oai.Choices[0].Text); c != "" {
				return c, nil
			}
		}
		for _, s := range []string{oai.Answer, oai.Response, oai.Content, oai.Output} {
			if strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s), nil
			}
		}
	}
	// Plain text response
	t := strings.TrimSpace(string(body))
	if t == "" {
		return "", fmt.Errorf("empty agent response")
	}
	// If it looked like JSON without known fields, still return it
	return t, nil
}
