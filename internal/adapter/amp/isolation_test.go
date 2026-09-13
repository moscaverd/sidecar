package amp

import (
	"os"
	"testing"

	"github.com/marcus/sidecar/internal/testutil"
)

func TestMain(m *testing.M) { os.Exit(testutil.Run(m)) }
