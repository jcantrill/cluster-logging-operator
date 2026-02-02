package exporters

import "time"

type Retry struct {
	Enabled         bool          `json:"enabled,omitempty"`
	InitialInterval time.Duration `json:"initial_interval,omitempty"`
	MaxInterval     time.Duration `json:"max_interval,omitempty"`
	MaxElapsedTime  time.Duration `json:"max_elapsed_time,omitempty"`
	Multiplier      float64       `json:"multiplier,omitempty"`
}

func NewDefaultRetry() *Retry {
	return &Retry{
		Enabled:         true,
		InitialInterval: 5 * time.Second,
		MaxInterval:     30 * time.Second,
		MaxElapsedTime:  300 * time.Second,
		Multiplier:      1.5,
	}
}
