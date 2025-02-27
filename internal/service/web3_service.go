package service

import "go.uber.org/zap"

// Web3 interface remains the same.
type Web3Servicer interface {
	ClientVersion() string
}

type web3Service struct {
	log                *zap.Logger
	applicationVersion string
}

func NewWeb3Service(log *zap.Logger, applicationVersion string) Web3Servicer {
	return &web3Service{
		log:                log,
		applicationVersion: applicationVersion,
	}
}

// ClientVersion returns "relay/<version>" where version is read from application.version in config.
// If application.version is not set, returns "relay/unknown".
func (w *web3Service) ClientVersion() string {
	w.log.Debug("Getting client version")

	version := w.applicationVersion
	if version == "" {
		w.log.Warn("Application version not set, using 'unknown'")
		version = "unknown"
	}

	clientVersion := "relay/" + version
	w.log.Debug("Returning client version", zap.String("version", clientVersion))
	return clientVersion
}
