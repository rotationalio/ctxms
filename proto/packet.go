package api

import (
	"fmt"
	"strings"
	"time"
)

func (p *Packet) Repr() string {
	route := strings.Join(p.Route, " -> ")

	ts, err := time.Parse(time.RFC3339Nano, p.Timestamp)
	if err != nil {
		return fmt.Sprintf("%s (?) %s", p.Id, route)
	}

	latency := time.Since(ts)
	return fmt.Sprintf("%s (%s) %s", p.Id, latency, route)
}
