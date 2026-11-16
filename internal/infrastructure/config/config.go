// Package config loads Hederium's settings once, at start-up.
//
// Each setting is looked up in four places, strongest first:
//
//  1. the environment variable the Hedera JSON-RPC relay uses for it
//     (MIRROR_NODE_URL), so relay deployments can reuse their environment;
//  2. Hederium's own environment variable, derived from the YAML key
//     (MIRRORNODE_BASEURL, or the older dotted form MIRRORNODE.BASEURL);
//  3. the optional file configs/config.yaml;
//  4. the default.
//
// An environment variable that is set but empty counts as not set.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	defaultConfigDir = "./configs"
	configFileName   = "config"
	redacted         = "[REDACTED]"
)

// Secret holds a value that must not reach logs or error messages. Printing or
// encoding it gives "[REDACTED]"; only Reveal returns the real value.
type Secret string

// String implements fmt.Stringer.
func (Secret) String() string { return redacted }

// GoString implements fmt.GoStringer, which %#v uses.
func (Secret) GoString() string { return redacted }

// MarshalText implements encoding.TextMarshaler, which JSON and YAML encoders use.
func (Secret) MarshalText() ([]byte, error) { return []byte(redacted), nil }

// Reveal returns the secret itself.
func (s Secret) Reveal() string { return string(s) }

// Config is every setting Hederium reads, validated and in its final type.
type Config struct {
	Application Application
	Server      Server
	Hedera      Hedera
	MirrorNode  MirrorNode
	Logging     Logging
	Features    Features
	Cache       Cache

	// APIKeys and Limiter are handed to the limiter package as they appear in
	// the YAML file. They have no environment variables.
	APIKeys interface{}
	Limiter map[string]interface{}
}

type Application struct {
	Version string
}

type Server struct {
	Port int
}

type Hedera struct {
	Network string
	// ChainID is 0x-prefixed lower-case hex, whichever form it was given in.
	ChainID     string
	OperatorID  string
	OperatorKey Secret
	HbarBudget  int
}

type MirrorNode struct {
	BaseURL string
	Timeout time.Duration
}

type Logging struct {
	// Level is a zap level name.
	Level string
}

type Features struct {
	EnforceAPIKey       bool
	EnableBatchRequests bool
}

type Cache struct {
	DefaultExpiration time.Duration
	CleanupInterval   time.Duration
}

// setting names one value: its path in config.yaml and, where the relay has the
// same setting, the relay's environment variable for it.
type setting struct {
	key   string
	relay string
}

var (
	applicationVersion  = setting{key: "application.version"}
	serverPort          = setting{key: "server.port", relay: "SERVER_PORT"}
	hederaNetwork       = setting{key: "hedera.network", relay: "HEDERA_NETWORK"}
	hederaChainID       = setting{key: "hedera.chainId", relay: "CHAIN_ID"}
	hederaOperatorID    = setting{key: "hedera.operatorId", relay: "OPERATOR_ID_MAIN"}
	hederaOperatorKey   = setting{key: "hedera.operatorKey", relay: "OPERATOR_KEY_MAIN"}
	hederaHbarBudget    = setting{key: "hedera.hbarBudget"}
	mirrorNodeBaseURL   = setting{key: "mirrorNode.baseUrl", relay: "MIRROR_NODE_URL"}
	mirrorNodeTimeout   = setting{key: "mirrorNode.timeoutSeconds", relay: "MIRROR_NODE_TIMEOUT"}
	loggingLevel        = setting{key: "logging.level", relay: "LOG_LEVEL"}
	enforceAPIKey       = setting{key: "features.enforceApiKey"}
	enableBatchRequests = setting{key: "features.enableBatchRequests", relay: "BATCH_REQUESTS_ENABLED"}
	cacheExpiration     = setting{key: "cache.defaultExpiration", relay: "CACHE_TTL"}
	cacheCleanup        = setting{key: "cache.cleanupInterval"}

	settings = []setting{
		applicationVersion, serverPort, hederaNetwork, hederaChainID, hederaOperatorID,
		hederaOperatorKey, hederaHbarBudget, mirrorNodeBaseURL, mirrorNodeTimeout, loggingLevel,
		enforceAPIKey, enableBatchRequests, cacheExpiration, cacheCleanup,
	}
)

// ownEnvNames returns Hederium's environment variables for the setting: the
// YAML key in upper case with underscores, then the dotted form that worked
// before relay names were supported.
func (s setting) ownEnvNames() []string {
	dotted := strings.ToUpper(s.key)
	return []string{strings.ReplaceAll(dotted, ".", "_"), dotted}
}

// envName is the environment variable to recommend for the setting.
func (s setting) envName() string {
	if s.relay != "" {
		return s.relay
	}
	return s.ownEnvNames()[0]
}

// EnvNames lists every environment variable Load reads.
func EnvNames() []string {
	var names []string
	seen := make(map[string]bool)
	for _, s := range settings {
		for _, name := range append([]string{s.relay}, s.ownEnvNames()...) {
			if name != "" && !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
	}
	return names
}

// Load reads the configuration from the environment and ./configs/config.yaml.
func Load() (*Config, error) {
	return LoadFrom(defaultConfigDir)
}

// LoadFrom is Load with the directory that may hold config.yaml.
func LoadFrom(dir string) (*Config, error) {
	file, err := readFile(dir)
	if err != nil {
		return nil, err
	}
	r := &resolver{lookupEnv: os.LookupEnv, file: file}

	cfg := &Config{
		Application: Application{
			Version: r.text(applicationVersion, "0.1.0"),
		},
		Server: Server{
			Port: r.integer(serverPort, 7546, 1, 65535),
		},
		Hedera: Hedera{
			Network:     r.mandatory(hederaNetwork),
			ChainID:     r.chainID(hederaChainID),
			OperatorID:  r.mandatory(hederaOperatorID),
			OperatorKey: Secret(r.mandatory(hederaOperatorKey)),
			HbarBudget:  r.integer(hederaHbarBudget, 0, 0, maxInt),
		},
		MirrorNode: MirrorNode{
			BaseURL: r.httpURL(mirrorNodeBaseURL),
			Timeout: r.timeout(mirrorNodeTimeout, wholeSeconds, 10*time.Second),
		},
		Logging: Logging{
			Level: r.logLevel(loggingLevel, "trace"),
		},
		Features: Features{
			EnforceAPIKey:       r.boolean(enforceAPIKey, false),
			EnableBatchRequests: r.boolean(enableBatchRequests, true),
		},
		Cache: Cache{
			DefaultExpiration: r.duration(cacheExpiration, goDuration, time.Hour),
			CleanupInterval:   r.duration(cacheCleanup, goDuration, 30*time.Minute),
		},
		APIKeys: file.Get("apiKeys"),
		Limiter: file.GetStringMap("limiter"),
	}

	if err := r.err(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// readFile reads config.yaml from dir. A missing file is not an error: every
// setting can come from the environment.
func readFile(dir string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName(configFileName)
	v.SetConfigType("yaml")
	v.AddConfigPath(dir)

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
	}
	return v, nil
}

// resolver finds each setting's value and collects every problem, so one
// start-up reports them all.
type resolver struct {
	lookupEnv func(string) (string, bool)
	file      *viper.Viper
	problems  []string
}

// found is a setting's raw text and where it came from.
type found struct {
	raw string
	// source is the name to use in error messages.
	source string
	// relay is true when the value came from the relay's environment variable,
	// which decides the unit of durations.
	relay bool
}

// find returns the strongest value set for s. Values are never put in error
// messages: a secret pasted into the wrong setting must not be echoed.
func (r *resolver) find(s setting) (found, bool) {
	if raw, ok := r.env(s.relay); ok {
		return found{raw: raw, source: s.relay, relay: true}, true
	}
	for _, name := range s.ownEnvNames() {
		if raw, ok := r.env(name); ok {
			return found{raw: raw, source: name}, true
		}
	}
	if !r.file.IsSet(s.key) {
		return found{}, false
	}
	source := s.key + " in config.yaml"
	raw, ok := scalarText(r.file.Get(s.key))
	if !ok {
		r.problem("%s must be a single value", source)
		return found{}, false
	}
	return found{raw: raw, source: source}, raw != ""
}

func (r *resolver) env(name string) (string, bool) {
	if name == "" {
		return "", false
	}
	raw, ok := r.lookupEnv(name)
	raw = strings.TrimSpace(raw)
	return raw, ok && raw != ""
}

// scalarText gives the text of a YAML scalar. Lists and maps are refused.
func scalarText(value interface{}) (string, bool) {
	switch v := value.(type) {
	case nil:
		return "", true
	case string:
		return strings.TrimSpace(v), true
	case bool, int, int64, uint64, float64:
		return fmt.Sprint(v), true
	default:
		return "", false
	}
}

func (r *resolver) problem(format string, args ...interface{}) {
	r.problems = append(r.problems, fmt.Sprintf(format, args...))
}

func (r *resolver) err() error {
	if len(r.problems) == 0 {
		return nil
	}
	return errors.New("invalid configuration:\n  - " + strings.Join(r.problems, "\n  - "))
}

func (r *resolver) text(s setting, def string) string {
	if f, ok := r.find(s); ok {
		return f.raw
	}
	return def
}

// need finds a setting that has no default. A missing one is a problem, unless
// find already reported why the value it saw is unusable.
func (r *resolver) need(s setting) (found, bool) {
	before := len(r.problems)
	f, ok := r.find(s)
	if !ok && len(r.problems) == before {
		r.problem("%s is mandatory: set the %s environment variable or %s in config.yaml", s.envName(), s.envName(), s.key)
	}
	return f, ok
}

func (r *resolver) mandatory(s setting) string {
	f, _ := r.need(s)
	return f.raw
}

func (r *resolver) boolean(s setting, def bool) bool {
	f, ok := r.find(s)
	if !ok {
		return def
	}
	switch strings.ToLower(f.raw) {
	case "true":
		return true
	case "false":
		return false
	}
	r.problem("%s must be true or false", f.source)
	return def
}

const maxInt = int(^uint(0) >> 1)

func (r *resolver) integer(s setting, def, lowest, highest int) int {
	f, ok := r.find(s)
	if !ok {
		return def
	}
	n, err := strconv.Atoi(f.raw)
	if err != nil || n < lowest || n > highest {
		if highest == maxInt {
			r.problem("%s must be a whole number, %d or more", f.source, lowest)
		} else {
			r.problem("%s must be a whole number from %d to %d", f.source, lowest, highest)
		}
		return def
	}
	return n
}
