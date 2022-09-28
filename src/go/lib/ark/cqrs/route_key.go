package cqrs

import "strings"

// RouteKey a stringer used for constructing conventional topic, source, and event names
type RouteKey string

func (r RouteKey) String() string {
	return string(r)
}

// With creates a copy of the route key with the additional keys applied
// Example:
//
//	RouteKey("com.service").With("events")
func (r RouteKey) With(keys ...RouteKey) RouteKey {
	sb := new(strings.Builder)
	sb.WriteString(r.String())
	for _, key := range keys {
		sb.WriteString("." + key.String())
	}
	return RouteKey(sb.String())
}

func (r RouteKey) StringPtr() *string {
	s := r.String()
	return &s
}
