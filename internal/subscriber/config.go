package subscriber

import (
	"os"
	"strconv"
)

type TransactionSubscriberConfig struct {
	Interval int
}

func CreateFromEnv() (cfg *TransactionSubscriberConfig, err error) {
	var interval int

	intervalString := os.Getenv("NEON_SUBSCRIBER_INTERVAL")

	interval, err = strconv.Atoi(intervalString)
	if err != nil {
		interval = 5
	}

	return &TransactionSubscriberConfig{
		Interval: interval,
	}, nil
}
