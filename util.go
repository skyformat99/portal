package portal

import (
	"github.com/pkg/errors"
)

// EndpointsCompatible returns true if the Endpoints have compatible protocols
func EndpointsCompatible(sig0, sig1 ProtocolSignature) bool {
	return sig0.Number() == sig1.PeerNumber() && sig0.Number() == sig1.PeerNumber()
}

// MustBeCompatible panics if the Endpoints have incompatible protocols
func MustBeCompatible(sig0, sig1 ProtocolSignature) {
	if !EndpointsCompatible(sig0, sig1) {
		panic(errors.Errorf("%s incompatible with %s", sig0.Name(), sig1.Name()))
	}
}
