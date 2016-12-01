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

// Option is used to set options for the logger.
type Option interface {
	apply(*Logger)
}

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(*Logger)

func (f optionFunc) apply(m *Logger) {
	f(m)
}

// This allows any Level to be used as an option.
func (l Level) apply(m *Logger)         { m.LevelEnabler = l }
func (lvl AtomicLevel) apply(m *Logger) { m.LevelEnabler = lvl }

// Fields sets the initial fields for the logger.
func Fields(fields ...Field) Option {
	return optionFunc(func(m *Logger) {
		addFields(m.Encoder, fields)
	})
}

// Output sets the destination for the logger's output. The supplied WriteSyncer
// is automatically wrapped with a mutex, so it need not be safe for concurrent
// use.
func Output(w WriteSyncer) Option {
	return optionFunc(func(m *Logger) {
		m.Output = newLockedWriteSyncer(w)
	})
}

// ErrorOutput sets the destination for errors generated by the logger. The
// supplied WriteSyncer is automatically wrapped with a mutex, so it need not be
// safe for concurrent use.
func ErrorOutput(w WriteSyncer) Option {
	return optionFunc(func(m *Logger) {
		m.ErrorOutput = newLockedWriteSyncer(w)
	})
}

// Development puts the logger in development mode, which alters the behavior
// of the DPanic method.
func Development() Option {
	return optionFunc(func(m *Logger) {
		m.Development = true
	})
}
