package handler

import (
	"net/http"
	"strings"
)

func headersFromEnv(headersEnv string) http.Header {
	values := http.Header{}
	if headersEnv == "" {
		return values
	}

	for h := range strings.SplitSeq(headersEnv, ",") {
		parts := strings.SplitN(h, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		http.Header.Add(values, key, value)
	}

	return values
}
