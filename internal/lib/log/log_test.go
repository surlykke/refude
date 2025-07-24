package log_test

import (
	"github.com/surlykke/refude/internal/lib/log"
	"testing"
)

func TestInfo(t *testing.T) {
	log.Debug("debug statement")
	log.Info("info statement")
}
