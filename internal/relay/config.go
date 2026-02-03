package relay

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Env struct {
	ListenAddr      string
	ForwardURL      string
	ForwardToken    string
	WebhookSecret   string
	MetadataToken   string
	AllowedMonitors string
	MaxBodyBytes    string
	ReplayWindowSec string
	RPS             string
	Burst           string
}

type Config struct {
	ListenAddr      string
	ForwardURL      string
	ForwardToken    string
	WebhookSecret   string
	MetadataToken   string
	AllowedMonitors map[string]struct{}
	MaxBodyBytes    int64
	ReplayWindow    time.Duration
	RPS             float64
	Burst           float64
	Now             func() time.Time
}

func LoadConfigFromEnv(keys Env) (Config, error) {
	var cfg Config

	cfg.ListenAddr = strings.TrimSpace(os.Getenv(keys.ListenAddr))
	if cfg.ListenAddr == "" {
		if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
			cfg.ListenAddr = ":" + port
		} else {
			cfg.ListenAddr = ":8080"
		}
	}

	cfg.ForwardURL = strings.TrimSpace(os.Getenv(keys.ForwardURL))
	if cfg.ForwardURL == "" {
		return Config{}, fmt.Errorf("missing env %s", keys.ForwardURL)
	}

	cfg.ForwardToken = strings.TrimSpace(os.Getenv(keys.ForwardToken))
	cfg.WebhookSecret = strings.TrimSpace(os.Getenv(keys.WebhookSecret))
	if cfg.WebhookSecret == "" {
		return Config{}, fmt.Errorf("missing env %s", keys.WebhookSecret)
	}

	cfg.MetadataToken = strings.TrimSpace(os.Getenv(keys.MetadataToken))

	allow := strings.TrimSpace(os.Getenv(keys.AllowedMonitors))
	if allow != "" {
		cfg.AllowedMonitors = map[string]struct{}{}
		for _, s := range strings.Split(allow, ",") {
			id := strings.TrimSpace(s)
			if id == "" {
				continue
			}
			cfg.AllowedMonitors[id] = struct{}{}
		}
	}

	cfg.MaxBodyBytes = 1 << 20
	if v := strings.TrimSpace(os.Getenv(keys.MaxBodyBytes)); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n <= 0 {
			return Config{}, fmt.Errorf("invalid env %s", keys.MaxBodyBytes)
		}
		cfg.MaxBodyBytes = n
	}

	cfg.ReplayWindow = 10 * time.Minute
	if v := strings.TrimSpace(os.Getenv(keys.ReplayWindowSec)); v != "" {
		sec, err := strconv.Atoi(v)
		if err != nil || sec <= 0 {
			return Config{}, fmt.Errorf("invalid env %s", keys.ReplayWindowSec)
		}
		cfg.ReplayWindow = time.Duration(sec) * time.Second
	}

	cfg.RPS = 5
	if v := strings.TrimSpace(os.Getenv(keys.RPS)); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f <= 0 {
			return Config{}, fmt.Errorf("invalid env %s", keys.RPS)
		}
		cfg.RPS = f
	}

	cfg.Burst = 20
	if v := strings.TrimSpace(os.Getenv(keys.Burst)); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f <= 0 {
			return Config{}, fmt.Errorf("invalid env %s", keys.Burst)
		}
		cfg.Burst = f
	}

	cfg.Now = time.Now
	return cfg, nil
}
