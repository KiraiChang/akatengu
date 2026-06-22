package event

import (
	"fmt"
	"strings"
)

// ------------------------------
// Helper
// ------------------------------

func joinErrors(errs []string) error {
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("validation failed: %s", strings.Join(errs, "; "))
}
