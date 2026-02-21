package retry

import (
	"context"
	"math/rand"
	"time"
)

type Operation = func() error

type Config struct {
	MaxRetries    int
	BackoffFactor float64
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	Jitter        time.Duration
}

func NewDefaultConfig() *Config {
	return &Config{
		MaxRetries:    5,
		BackoffFactor: 2.15,
		InitialDelay:  300 * time.Millisecond,
		MaxDelay:      20 * time.Second,
		Jitter:        50 * time.Millisecond,
	}
}

type Retrier struct {
	config *Config
}

func NewRetrier(config *Config) *Retrier {
	return &Retrier{
		config: config,
	}
}

func NewDefaultRetrier() *Retrier {
	return NewRetrier(NewDefaultConfig())
}

func (r *Retrier) Do(ctx context.Context, op Operation) error {
	var err error
	delay := r.config.InitialDelay
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		err = op()
		if err == nil {
			return nil
		}

		if attempt == r.config.MaxRetries {
			return err
		}

		jitter := time.Duration(rnd.Float64() * float64(r.config.Jitter))
		nextDelay := delay + jitter
		if nextDelay > r.config.MaxDelay {
			nextDelay = r.config.MaxDelay + jitter
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(nextDelay):
		}

		delay = time.Duration(float64(delay) * r.config.BackoffFactor)
		if delay > r.config.MaxDelay {
			delay = r.config.MaxDelay
		}
	}
	return err
}
