package config

import (
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// unit is how a duration is written under Hederium's own names. Under a relay
// name a duration is always a whole number of milliseconds.
type unit int

const (
	// wholeSeconds is a whole number of seconds, such as 10.
	wholeSeconds unit = iota
	// goDuration is Go duration text, such as 90s or 1h.
	goDuration
)

// duration reads a length of time. Zero is allowed.
func (r *resolver) duration(s setting, own unit, def time.Duration) time.Duration {
	if d, _, ok := r.durationOf(s, own); ok {
		return d
	}
	return def
}

// timeout reads a length of time that must be above zero: a timeout of zero
// would expire every request before it is sent.
func (r *resolver) timeout(s setting, own unit, def time.Duration) time.Duration {
	d, source, ok := r.durationOf(s, own)
	if !ok {
		return def
	}
	if d == 0 {
		r.problem("%s must be above zero", source)
		return def
	}
	return d
}

// durationOf returns the duration set for s and the name of where it was set.
// It reports false when nothing is set or the value is unusable.
func (r *resolver) durationOf(s setting, own unit) (time.Duration, string, bool) {
	f, ok := r.find(s)
	if !ok {
		return 0, "", false
	}
	switch {
	case f.relay:
		if d, ok := wholeNumberOf(f.raw, time.Millisecond); ok {
			return d, f.source, true
		}
		r.problem("%s must be a whole number of milliseconds", f.source)
	case own == wholeSeconds:
		if d, ok := wholeNumberOf(f.raw, time.Second); ok {
			return d, f.source, true
		}
		r.problem("%s must be a whole number of seconds", f.source)
	default:
		if d, err := time.ParseDuration(f.raw); err == nil && d >= 0 {
			return d, f.source, true
		}
		r.problem("%s must be a duration such as 90s or 1h", f.source)
	}
	return 0, "", false
}

// wholeNumberOf reads raw as a count of step, refusing negative counts and
// counts too large for a time.Duration.
func wholeNumberOf(raw string, step time.Duration) (time.Duration, bool) {
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n < 0 || n > math.MaxInt64/int64(step) {
		return 0, false
	}
	return time.Duration(n) * step, true
}

// chainID accepts the chain ID in decimal or 0x-prefixed hex, as the relay
// does, and returns it as 0x-prefixed lower-case hex. Unlike the relay, which
// would turn a value that is not a number into "0xNaN", it refuses bad values.
func (r *resolver) chainID(s setting) string {
	f, ok := r.need(s)
	if !ok {
		return ""
	}
	digits, base := f.raw, 10
	if strings.HasPrefix(digits, "0x") || strings.HasPrefix(digits, "0X") {
		digits, base = digits[2:], 16
	}
	n, err := strconv.ParseUint(digits, base, 64)
	if err != nil || n == 0 {
		r.problem("%s must be a chain ID above zero, in decimal (298) or hex (0x12a)", f.source)
		return ""
	}
	return "0x" + strconv.FormatUint(n, 16)
}

func (r *resolver) httpURL(s setting) string {
	f, ok := r.need(s)
	if !ok {
		return ""
	}
	u, err := url.Parse(f.raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		r.problem("%s must be an http or https URL", f.source)
		return ""
	}
	return f.raw
}

// zapLevels maps the level names the relay accepts (pino's), and zap's own, to
// the zap level Hederium runs at. zap has no trace or silent level, so those
// run as the nearest one.
var zapLevels = map[string]string{
	"trace":  "debug",
	"debug":  "debug",
	"info":   "info",
	"warn":   "warn",
	"error":  "error",
	"dpanic": "dpanic",
	"panic":  "panic",
	"fatal":  "fatal",
	"silent": "fatal",
}

func (r *resolver) logLevel(s setting, def string) string {
	f, ok := r.find(s)
	if !ok {
		return zapLevels[def]
	}
	level, known := zapLevels[strings.ToLower(f.raw)]
	if !known {
		r.problem("%s must be one of trace, debug, info, warn, error, fatal, silent", f.source)
		return zapLevels[def]
	}
	return level
}
