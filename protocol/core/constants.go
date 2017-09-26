package core

/*
	Core components for protocol developers.
*/

import "github.com/lthibault/portal"

const (
	// REQEndpt identifies the REQ endpoint that originated a message
	REQEndpt portal.HeaderKey = iota
)
