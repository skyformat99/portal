package proto

import "github.com/lthibault/portal"

type (
	rg struct{ portal.ReadOnly }
	wg struct{ portal.WriteOnly }
)

// ReadGuard prevents type-switches/assertions from recasting a portal as writable
func ReadGuard(p portal.Portal) portal.ReadOnly { return &rg{ReadOnly: p} }

// WriteGuard prevents type-switches/assertions from recasting a portal as readable
func WriteGuard(p portal.Portal) portal.WriteOnly { return &wg{WriteOnly: p} }
