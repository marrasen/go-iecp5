// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package clog

import (
	"log"
	"os"
	"sync/atomic"
)

// LogProvider RFC5424 log message levels only Debug Warn and Error
type LogProvider interface {
	Critical(format string, v ...interface{})
	Error(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Debug(format string, v ...interface{})
}

// Level represents the logging severity.
// Ordering: Off < Critical < Error < Warn < Debug
// Setting a level enables logging for that level and all more critical levels.
type Level uint32

const (
	LevelOff Level = iota
	LevelCritical
	LevelError
	LevelWarn
	LevelDebug
)

// Clog internal logging implementation with level control
type Clog struct {
	provider LogProvider
	// level stores the current logging level (atomic)
	level uint32
}

// NewLogger creates a new logger using the specified prefix
// Default level is Off (no logs) to preserve previous behavior.
func NewLogger(prefix string) Clog {
	return Clog{
		defaultLogger{
			log.New(os.Stdout, prefix, log.LstdFlags),
		},
		uint32(LevelOff),
	}
}

// SetLogLevel sets the logging level. LevelOff disables all logs; higher levels allow more verbose logs.
func (sf *Clog) SetLogLevel(lvl Level) {
	atomic.StoreUint32(&sf.level, uint32(lvl))
}

// SetLogProvider set provider provider
func (sf *Clog) SetLogProvider(p LogProvider) {
	if p != nil {
		sf.provider = p
	}
}

func (sf Clog) allowed(required Level) bool {
	return atomic.LoadUint32(&sf.level) >= uint32(required)
}

// Critical Log CRITICAL level message.
func (sf Clog) Critical(format string, v ...interface{}) {
	if sf.allowed(LevelCritical) {
		sf.provider.Critical(format, v...)
	}
}

// Error Log ERROR level message.
func (sf Clog) Error(format string, v ...interface{}) {
	if sf.allowed(LevelError) {
		sf.provider.Error(format, v...)
	}
}

// Warn Log WARN level message.
func (sf Clog) Warn(format string, v ...interface{}) {
	if sf.allowed(LevelWarn) {
		sf.provider.Warn(format, v...)
	}
}

// Debug Log DEBUG level message.
func (sf Clog) Debug(format string, v ...interface{}) {
	if sf.allowed(LevelDebug) {
		sf.provider.Debug(format, v...)
	}
}

// default log
type defaultLogger struct {
	*log.Logger
}

var _ LogProvider = (*defaultLogger)(nil)

// Critical Log CRITICAL level message.
func (sf defaultLogger) Critical(format string, v ...interface{}) {
	sf.Printf("[C]: "+format, v...)
}

// Error Log ERROR level message.
func (sf defaultLogger) Error(format string, v ...interface{}) {
	sf.Printf("[E]: "+format, v...)
}

// Warn Log WARN level message.
func (sf defaultLogger) Warn(format string, v ...interface{}) {
	sf.Printf("[W]: "+format, v...)
}

// Debug Log DEBUG level message.
func (sf defaultLogger) Debug(format string, v ...interface{}) {
	sf.Printf("[D]: "+format, v...)
}
