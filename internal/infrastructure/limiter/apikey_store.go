package limiter

type APIKeyStore struct {
	keys map[string]string
}

func NewAPIKeyStore(apiKeys interface{}) *APIKeyStore {
	keys := make(map[string]string)
	if kArr, ok := apiKeys.([]interface{}); ok {
		for _, kv := range kArr {
			keyMap, ok := kv.(map[interface{}]interface{})
			if !ok {
				continue
			}
			apikey, keyOk := keyMap["key"].(string)
			tier, tierOk := keyMap["tier"].(string)
			if !keyOk || !tierOk {
				continue
			}
			keys[apikey] = tier
		}
	}
	return &APIKeyStore{keys: keys}
}

func (s *APIKeyStore) GetTierForKey(apiKey string) (string, bool) {
	tier, exists := s.keys[apiKey]
	return tier, exists
}
