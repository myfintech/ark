package mnet

import (
	"net/http"
	"strings"
)

// ClientIPFromRequest looks up the client ip in a set of preferred headers
// If the IP can't be located we return an empty string
// We make no assumptions of weather or not the ip is valid
// That's up to you to determine
//
// True-Client-Ip is a header cloudflare allows you to enable (Preferred)
func ClientIPFromRequest(r *http.Request) string {
	for _, h := range []string{"True-Client-Ip", "X-Forwarded-For", "X-Real-Ip"} {
		for _, address := range strings.Split(r.Header.Get(h), ",") {
			if address == "" {
				continue
			}
			return strings.TrimSpace(address)
		}
	}
	return r.RemoteAddr
}
