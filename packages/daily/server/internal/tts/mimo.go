package tts

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	mimoAPIBase = "https://api.xiaomimimo.com/v1"
	mimoTimeout = 30 * time.Second
)

type mimoRequest struct {
	Model    string       `json:"model"`
	Messages []mimoMsg    `json:"messages"`
	Audio    mimoAudioCfg `json:"audio"`
}

type mimoMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type mimoAudioCfg struct {
	Format string `json:"format"`
	Voice  string `json:"voice"`
}

type mimoResponse struct {
	Choices []mimoChoice `json:"choices"`
}

type mimoChoice struct {
	Message mimoRespMsg `json:"message"`
}

type mimoRespMsg struct {
	Audio *mimoAudioData `json:"audio"`
}

type mimoAudioData struct {
	Data string `json:"data"` // base64 encoded WAV
}

// CallMiMoTTS calls the MiMo V2.5 TTS API and returns raw WAV bytes.
// text: the content to synthesize (placed in assistant message)
// apiKey: MiMo API key
// voice: voice name (e.g. "苏打", "白桦")
// style: optional style instruction (placed in user message), empty string if not used
func CallMiMoTTS(text, apiKey, voice, style string) ([]byte, error) {
	// Clean API key: remove characters invalid in HTTP header values (RFC 7230)
	// Only allow visible ASCII (VCHAR) and whitespace (SP/HTAB)
	cleaned := make([]byte, 0, len(apiKey))
	for _, b := range []byte(apiKey) {
		if b >= 0x21 && b <= 0x7E || b == ' ' || b == '\t' {
			cleaned = append(cleaned, b)
		}
	}
	apiKey = string(cleaned)
	if apiKey == "" {
		return nil, fmt.Errorf("TTS_API_KEY is required")
	}
	if text == "" {
		return nil, fmt.Errorf("text is empty")
	}

	messages := []mimoMsg{
		{Role: "assistant", Content: text},
	}
	if style != "" {
		// Prepend user message with style instruction
		messages = append([]mimoMsg{{Role: "user", Content: style}}, messages...)
	}

	reqBody := mimoRequest{
		Model:    "mimo-v2.5-tts",
		Messages: messages,
		Audio: mimoAudioCfg{
			Format: "wav",
			Voice:  voice,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", mimoAPIBase+"/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", apiKey)

	client := &http.Client{Timeout: mimoTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("TTS API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TTS API error %d: %s", resp.StatusCode, string(respBody))
	}

	var mimoResp mimoResponse
	if err := json.NewDecoder(resp.Body).Decode(&mimoResp); err != nil {
		return nil, fmt.Errorf("decode TTS response: %w", err)
	}

	if len(mimoResp.Choices) == 0 || mimoResp.Choices[0].Message.Audio == nil {
		return nil, fmt.Errorf("TTS API returned no audio data")
	}

	wavBytes, err := base64.StdEncoding.DecodeString(mimoResp.Choices[0].Message.Audio.Data)
	if err != nil {
		return nil, fmt.Errorf("decode base64 audio: %w", err)
	}

	return wavBytes, nil
}
