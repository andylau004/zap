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

package spy

import (
	"sync"

	"github.com/uber-go/zap"
)

// A Log is an encoding-agnostic representation of a log message.
type Log struct {
	Level  zap.Level
	Msg    string
	Fields []zap.Field
}

// A Sink stores Log structs.
type Sink struct {
	sync.Mutex

	logs []Log
}

// WriteLog writes a log message to the LogSink.
func (s *Sink) WriteLog(lvl zap.Level, msg string, fields []zap.Field) {
	s.Lock()
	log := Log{
		Msg:    msg,
		Level:  lvl,
		Fields: fields,
	}
	s.logs = append(s.logs, log)
	s.Unlock()
}

// Logs returns a copy of the sink's accumulated logs.
func (s *Sink) Logs() []Log {
	var logs []Log
	s.Lock()
	logs = append(logs, s.logs...)
	s.Unlock()
	return logs
}

type spyFacility struct {
	sync.Mutex
	sink    *Sink
	context []zap.Field
}

func (sf *spyFacility) With(fields ...zap.Field) zap.Facility {
	return &spyFacility{
		sink:    sf.sink,
		context: append(sf.context, fields...),
	}
}

func (*spyFacility) Enabled(zap.Entry) bool { return true }

func (sf *spyFacility) Log(ent zap.Entry, fields ...zap.Field) {
	all := make([]zap.Field, 0, len(fields)+len(sf.context))
	all = append(all, sf.context...)
	all = append(all, fields...)
	sf.sink.WriteLog(ent.Level, ent.Message, all)
}

// New constructs a spy logger that collects spy.Log records to a Sink. It
// returns the logger and its sink.
//
// Options can change things like log level and initial fields, but any output
// related options will not be honored.
func New(options ...zap.Option) (*zap.Logger, *Sink) {
	s := &Sink{}
	return zap.New(&spyFacility{sink: s}, options...), s
}
