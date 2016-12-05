// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package zap

import (
	"fmt"
	"os"
	"time"
)

// Logger is the logger implementation.
//
// TODO: note any concurrency concerns.
type Logger struct {
	LevelEnabler
	Facility

	Development bool
	Hooks       []Hook
	ErrorOutput WriteSyncer
}

// MakeMeta returns a new meta struct with sensible defaults: logging at
// InfoLevel, development mode off, and writing to standard error and standard
// out.
func New(fac Facility, options ...Option) *Logger {
	log := &Logger{
		Facility:     fac,
		ErrorOutput:  newLockedWriteSyncer(os.Stderr),
		LevelEnabler: InfoLevel,
	}
	for _, opt := range options {
		opt.apply(log)
	}
	return log
}

// InternalError prints an internal error message to the configured
// ErrorOutput. This method should only be used to report internal logger
// problems and should not be used to report user-caused problems.
func (log *Logger) InternalError(cause string, err error) {
	fmt.Fprintf(log.ErrorOutput, "%v %s error: %v\n", time.Now().UTC(), cause, err)
	log.ErrorOutput.Sync()
}

// Encode runs any Hook functions, returning a possibly modified
// time, message, and level.
func (log *Logger) Encode(enc Encoder, ent Entry, fields []Field) (string, Encoder) {
	clone := &Logger{
		LevelEnabler: log.LevelEnabler,
		Facility:     log.Facility,
		Development:  log.Development,
		Hooks:        log.Hooks,
		ErrorOutput:  log.ErrorOutput,
	}
	entry.enc = enc.Clone()
	addFields(entry.enc, fields)
	for _, hook := range log.Hooks {
		if err := hook(&entry); err != nil {
			log.InternalError("hook", err)
		}
	}
	return entry.Message, entry.enc
}

// With creates a new child *Logger with the given fields added to all child
// log sites.
func (log *Logger) With(fields ...Field) *Logger {
	clone := &Logger{
		Logger: log.Logger.Clone(),
	}
	addFields(clone.Encoder, fields)
	addFields(clone.Encoder, fields)
	return clone
}

// Check returns a CheckedMessage logging the given message is Enabled, nil
// otherwise.
func (log *Logger) Check(lvl Level, msg string) *CheckedMessage {
	switch lvl {
	case PanicLevel, FatalLevel:
		// Panic and Fatal should always cause a panic/exit, even if the level
		// is disabled.
		break
	case DPanicLevel:
		if log.Development {
			break
		}
		fallthrough
	default:
		if !log.Enabled(lvl) || !log.Facility.Enabled(Entry{
			Level:   lvl,
			Message: msg,
		}) {
			return nil
		}
	}
	return NewCheckedMessage(log.Facility, Entry{
		Time:    time.Now().UTC(),
		Level:   lvl,
		Message: msg,
	})
}

// Debug logs at DebugLevel.
func (log *Logger) Debug(msg string, fields ...Field) {
	log.Facility.Log(Entry{
		Time:    time.Now().UTC(),
		Level:   DebugLevel,
		Message: msg,
	}, fields...)
}

// Info logs at InfoLevel.
func (log *Logger) Info(msg string, fields ...Field) {
	log.Facility.Log(Entry{
		Time:    time.Now().UTC(),
		Level:   InfoLevel,
		Message: msg,
	}, fields...)
}

// Warn logs at WarnLevel.
func (log *Logger) Warn(msg string, fields ...Field) {
	log.Facility.Log(Entry{
		Time:    time.Now().UTC(),
		Level:   WarnLevel,
		Message: msg,
	}, fields...)
}

// Error logs at ErrorLevel.
func (log *Logger) Error(msg string, fields ...Field) {
	log.Facility.Log(Entry{
		Time:    time.Now().UTC(),
		Level:   ErrorLevel,
		Message: msg,
	}, fields...)
}

// DPanic logs at DPanicLevel and then calls panic(msg) if in Development mode.
func (log *Logger) DPanic(msg string, fields ...Field) {
	log.Facility.Log(Entry{
		Time:    time.Now().UTC(),
		Level:   DPanicLevel,
		Message: msg,
	}, fields...)
	if log.Development {
		panic(msg)
	}
}

// Panic logs at PanicLevel and then calls panic(msg).
func (log *Logger) Panic(msg string, fields ...Field) {
	log.Facility.Log(Entry{
		Time:    time.Now().UTC(),
		Level:   PanicLevel,
		Message: msg,
	}, fields...)
	panic(msg)
}

// Fatal logs at FataLevel and then calls os.Exit(1).
func (log *Logger) Fatal(msg string, fields ...Field) {
	log.Facility.Log(Entry{
		Time:    time.Now().UTC(),
		Level:   FatalLevel,
		Message: msg,
	}, fields...)
	_exit(1)
}
