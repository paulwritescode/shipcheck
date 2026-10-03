package advisor

import (
	"context"
	"os"
	"strings"
)

// ModelProvider is the provider-agnostic seam through which the advisor obtains
// a structured analysis. Concrete providers are selectable without changing
// product code (Req 16.1). Every provider is called server-side only; key-based
// providers read their keys from server-side configuration only (Req 16.2-16.4).
type ModelProvider interface {
	// Name identifies the provider (for disclosure/logging).
	Name() string
	// AnalyzeLaunch returns a structured analysis for the given input. The
	// backend validates the result against the schema before display.
	AnalyzeLaunch(ctx context.Context, in AdvisorInput) (LaunchAnalysis, error)
}

// SelectProvider chooses a provider from configuration. The default — and the
// provider the public demo runs on — is the free, deterministic FallbackProvider
// (no external model, no key, always available, Req 16.5). Opt-in providers are
// selected via SHIPCHECK_MODEL_PROVIDER and require their own server-side config.
//
// Selection values (case-insensitive):
//
//	"", "fallback"       -> FallbackProvider (default)
//	"nvidia"             -> NvidiaModelProvider (NVIDIA catalog, NIM endpoint)
//	"huggingface", "hf"  -> HuggingFaceModelProvider (small SLM)
//	"bedrock"            -> BedrockModelProvider (Amazon Nova Micro, IAM role)
//	"local"              -> LocalModelProvider (self-hosted / local model)
//
// Any opt-in provider that is misconfigured (e.g. missing key) falls back to
// the deterministic provider so the demo never breaks.
func SelectProvider() ModelProvider {
	fallback := NewFallbackProvider()
	switch strings.ToLower(strings.TrimSpace(os.Getenv("SHIPCHECK_MODEL_PROVIDER"))) {
	case "", "fallback":
		return fallback
	case "nvidia":
		if p, ok := NewNvidiaModelProvider(fallback); ok {
			return p
		}
		return fallback
	case "huggingface", "hf":
		if p, ok := NewHuggingFaceModelProvider(fallback); ok {
			return p
		}
		return fallback
	case "bedrock":
		return NewBedrockModelProvider(fallback)
	case "local":
		if p, ok := NewLocalModelProvider(fallback); ok {
			return p
		}
		return fallback
	default:
		return fallback
	}
}
