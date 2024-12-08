package service

import (
	"math/big"
)

func getFeeWeibars(s *EthService) (*big.Int, map[string]interface{}) {
	gasTinybars, err := s.mClient.GetNetworkFees()
	if err != nil {
		return nil, map[string]interface{}{
			"code":    -32000,
			"message": "Failed to fetch gas price",
		}
	}

	// Convert tinybars to weibars
	weibars := big.NewInt(gasTinybars).
		Mul(big.NewInt(gasTinybars), big.NewInt(100000000)) // 10^8 conversion factor

	return weibars, nil
}
