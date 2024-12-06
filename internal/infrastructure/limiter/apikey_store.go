package limiter

type APIKeyStore struct {
	keys map[string]string
}

func NewAPIKeyStore(apiKeys interface{}) *APIKeyStore {
	keys := make(map[string]string)
	if kArr, ok := apiKeys.([]interface{}); ok {
		for _, kv := range kArr {
			if keyMap, ok := kv.(map[interface{}]interface{}); ok {
				apikey := keyMap["key"].(string)
				tier := keyMap["tier"].(string)
				keys[apikey] = tier
			}
		}
	}
	return &APIKeyStore{keys: keys}
}

func (s *APIKeyStore) GetTierForKey(apiKey string) (string, bool) {
	tier, exists := s.keys[apiKey]
	return tier, exists
}
