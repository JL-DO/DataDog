package log

import "github.com/DataDog/datadog-agent/pkg/util/scrubber"

// scrubber wraps another ddLogger and scrubs all log messages
type scrubbingLogger struct {
	inner ddLogger
}

var _ ddLogger = (*scrubbingLogger)(nil)

func newScrubbingLogger(inner ddLogger) *scrubbingLogger {
	return &scrubbingLogger{inner}
}

func (l *scrubbingLogger) scrub(s string) string {
	if scrubbed, err := scrubber.ScrubBytes([]byte(s)); err == nil {
		return string(scrubbed)
	}

	return s
}

func (l *scrubbingLogger) trace(message string, context []interface{}, depth int) {
	l.inner.trace(l.scrub(message), context, depth)
}

func (l *scrubbingLogger) debug(message string, context []interface{}, depth int) {
	l.inner.debug(l.scrub(message), context, depth)
}

func (l *scrubbingLogger) info(message string, context []interface{}, depth int) {
	l.inner.info(l.scrub(message), context, depth)
}

func (l *scrubbingLogger) warn(message string, context []interface{}, depth int) {
	l.inner.warn(l.scrub(message), context, depth)
}

func (l *scrubbingLogger) error(message string, context []interface{}, depth int) {
	l.inner.error(l.scrub(message), context, depth)
}

func (l *scrubbingLogger) critical(message string, context []interface{}, depth int) {
	l.inner.critical(l.scrub(message), context, depth)
}

func (l *scrubbingLogger) flush() {
	l.inner.flush()
}
