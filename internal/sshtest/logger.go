package sshtest

import "log"

type TestLogger interface {
	Log(args ...any)
	Logf(format string, args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
}

type noopLogger struct{}

var _ TestLogger = noopLogger{}

func (noopLogger) Log(args ...any)                   {}
func (noopLogger) Logf(format string, args ...any)   {}
func (noopLogger) Error(args ...any)                 {}
func (noopLogger) Errorf(format string, args ...any) {}
func (noopLogger) Fatal(args ...any)                 {}
func (noopLogger) Fatalf(format string, args ...any) {}

type StdLogger struct{}

// TestLoggerStd is a standard logger that logs to the standard logger.
var TestLoggerStd TestLogger = StdLogger{}

func (StdLogger) Log(args ...any)                   { log.Print(args...) }
func (StdLogger) Logf(format string, args ...any)   { log.Printf(format, args...) }
func (StdLogger) Error(args ...any)                 { log.Print(args...) }
func (StdLogger) Errorf(format string, args ...any) { log.Printf(format, args...) }
func (StdLogger) Fatal(args ...any)                 { log.Fatal(args...) }
func (StdLogger) Fatalf(format string, args ...any) { log.Fatalf(format, args...) }
