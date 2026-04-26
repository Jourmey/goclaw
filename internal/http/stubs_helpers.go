package http

import "context"

// lookupProviderByNameWithMasterFallback stub - provider lookup removed
func lookupProviderByNameWithMasterFallback(ctx context.Context, providerStore interface{}, name string, allowMaster bool) (interface{}, error) {
	return nil, nil
}

// validateChatGPTOAuthAgentRouting stub - ChatGPT OAuth routing removed
func validateChatGPTOAuthAgentRouting(ctx context.Context, providerStore interface{}, providerName string, routing interface{}) error {
	return nil
}

// marshalJSONRaw stub - JSON marshaling helper removed
func marshalJSONRaw(v interface{}) ([]byte, error) {
	return nil, nil
}

// validateChatGPTOAuthProviderCandidate stub - ChatGPT OAuth provider validation removed
func validateChatGPTOAuthProviderCandidate(ctx context.Context, store interface{}, id interface{}, provider interface{}) error {
	return nil
}

// registeredCodexPoolProviders stub - Codex pool providers removed
func registeredCodexPoolProviders(providerReg interface{}, tenantID interface{}, candidates interface{}) map[string]interface{} {
	return make(map[string]interface{})
}

// maxInt stub - max int helper removed
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// codexPoolRuntimeHealthSampleSize stub - Codex pool runtime health sample size removed
const codexPoolRuntimeHealthSampleSize = 100

// providerInPool stub - Codex pool provider check removed
func providerInPool(pool map[string]interface{}, provider interface{}) bool {
	return false
}
