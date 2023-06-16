package utils

import (
	"regexp"
	"strings"
)

// var adrRex = regexp.MustCompile("^0x[a-f0-9]{40}$")
var topicsFormat = regexp.MustCompile(`^[0-9a-fA-F]*$`)

// IsAddressValid checks if adr is valid
func IsAddressValid(adr string) bool {
	return regexp.MustCompile("^0x[a-f0-9]{40}$").MatchString(adr)
}

func IsTopicValid(topic string) bool {
	if !strings.HasPrefix(topic, "0x") {
		return false
	}

	addr := topic[2:]
	if len(addr) != 64 {
		return false
	}
	if !topicsFormat.MatchString(addr) {
		return false
	}

	return true
}
