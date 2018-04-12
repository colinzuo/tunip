package logp

import (
	"fmt"

	"go.uber.org/zap"
)

// HasSelector returns true if the given selector was explicitly set.
func HasSelector(selector string) bool {
	_, found := loadLogger().selectors[selector]
	return found
}

// Recover stops a panicking goroutine and logs an Error.
func Recover(msg string) {
	if r := recover(); r != nil {
		msg := fmt.Sprintf("%s. Recovering, but please report this.", msg)
		globalLogger().WithOptions(zap.AddCallerSkip(2)).
			Error(msg, zap.Any("panic", r), zap.Stack("stack"))
	}
}
