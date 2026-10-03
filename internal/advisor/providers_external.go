package advisor

import (
	"context"
	"os"
)

// This file holds the opt-in, key-based providers that speak the
// OpenAI-compatible API: the NVIDIA catalog, Hugging Face, and a local model.
// Each reads its configuration from server-side environment variables only and
// degrades to the deterministic fallback on any error, so a misconfigured or
// unavailable model never breaks the product.

// --- NVIDIA catalog (build.nvidia.com, NIM OpenAI-compatible endpoint) ---

// NvidiaModelProvider targets NVIDIA's model catalog over its OpenAI-compatible
// endpoint. The specific model (e.g. a DeepSeek, Llama, Mistral, Qwen, or
// Nemotron id) is selected by env, so no code changes are needed to switch
// models. The API trial is for dev/testing only, which the README discloses.
type NvidiaModelProvider struct {
	core     *openAICompatibleProvider
	fallback ModelProvider
}

// NewNvidiaModelProvider builds the provider from env. Returns ok=false (so the
// caller uses fallback) when the API key is absent.
func NewNvidiaModelProvider(fallback ModelProvider) (*NvidiaModelProvider, bool) {
	key := os.Getenv("NVIDIA_API_KEY")
	if key == "" {
		return nil, false
	}
	model := envOr("NVIDIA_MODEL", "meta/llama-3.2-3b-instruct")
	base := envOr("NVIDIA_BASE_URL", "https://integrate.api.nvidia.com/v1/chat/completions")
	return &NvidiaModelProvider{
		core:     &openAICompatibleProvider{name: "nvidia", baseURL: base, apiKey: key, model: model, fallback: fallback},
		fallback: fallback,
	}, true
}

func (p *NvidiaModelProvider) Name() string { return "nvidia" }

func (p *NvidiaModelProvider) AnalyzeLaunch(ctx context.Context, in AdvisorInput) (LaunchAnalysis, error) {
	a, err := p.core.analyze(ctx, in)
	if err != nil {
		return p.fallback.AnalyzeLaunch(ctx, in)
	}
	return a, nil
}

// --- Hugging Face (small SLM via the Inference router) ---

// HuggingFaceModelProvider calls a small hosted SLM (e.g. Llama 3.2 1B or
// Qwen2.5 1.5B) via Hugging Face's OpenAI-compatible router. The free
// serverless allowance is rate-limited and non-production, which the README
// discloses. This is the opt-in used for the demo video.
type HuggingFaceModelProvider struct {
	core     *openAICompatibleProvider
	fallback ModelProvider
}

// NewHuggingFaceModelProvider builds the provider from env. Returns ok=false
// when the token is absent.
func NewHuggingFaceModelProvider(fallback ModelProvider) (*HuggingFaceModelProvider, bool) {
	key := os.Getenv("HF_TOKEN")
	if key == "" {
		return nil, false
	}
	model := envOr("HF_MODEL", "meta-llama/Llama-3.2-1B-Instruct")
	base := envOr("HF_BASE_URL", "https://router.huggingface.co/v1/chat/completions")
	return &HuggingFaceModelProvider{
		core:     &openAICompatibleProvider{name: "huggingface", baseURL: base, apiKey: key, model: model, fallback: fallback},
		fallback: fallback,
	}, true
}

func (p *HuggingFaceModelProvider) Name() string { return "huggingface" }

func (p *HuggingFaceModelProvider) AnalyzeLaunch(ctx context.Context, in AdvisorInput) (LaunchAnalysis, error) {
	a, err := p.core.analyze(ctx, in)
	if err != nil {
		return p.fallback.AnalyzeLaunch(ctx, in)
	}
	return a, nil
}

// --- Local model (self-hosted OpenAI-compatible server, e.g. Ollama/vLLM) ---

// LocalModelProvider calls a local OpenAI-compatible model server. No key is
// required; the base URL is configured by env.
type LocalModelProvider struct {
	core     *openAICompatibleProvider
	fallback ModelProvider
}

// NewLocalModelProvider builds the provider from env. Returns ok=false when no
// base URL is configured.
func NewLocalModelProvider(fallback ModelProvider) (*LocalModelProvider, bool) {
	base := os.Getenv("LOCAL_MODEL_URL")
	if base == "" {
		return nil, false
	}
	model := envOr("LOCAL_MODEL", "llama3.2:1b")
	return &LocalModelProvider{
		core:     &openAICompatibleProvider{name: "local", baseURL: base, apiKey: os.Getenv("LOCAL_MODEL_KEY"), model: model, fallback: fallback},
		fallback: fallback,
	}, true
}

func (p *LocalModelProvider) Name() string { return "local" }

func (p *LocalModelProvider) AnalyzeLaunch(ctx context.Context, in AdvisorInput) (LaunchAnalysis, error) {
	a, err := p.core.analyze(ctx, in)
	if err != nil {
		return p.fallback.AnalyzeLaunch(ctx, in)
	}
	return a, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
