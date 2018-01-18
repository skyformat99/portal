package star

import (
	"testing"

	"github.com/lthibault/portal"
)

func TestIntegration(t *testing.T) {
	t.Errorf(starAddrFmt, portal.NewID())
}
