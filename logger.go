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

import "os"

// For tests.
var _exit = os.Exit

type Facility interface {
	With(...Field) Facility
	Enabled(Entry) bool
	Log(Entry, ...Field)
}

type ioFacility struct {
	Logger
	Encoder
	Output WriteSyncer
}

func newIOFacility(log Logger, enc Encode, out WriteSyncer) *ioFacility {
	if out == nil {
		out = newLockedWriteSyncer(os.Stdout)
	}
	return &ioFacility{
		Logger:  log,
		Encoder: enc,
		Output:  out,
	}
}

func (iof *ioFacility) With(...Field) Facility {
	enc := iof.enc.Clone()
	addFields(enc, fields)
	return &ioFacility{
		Logger:  iof.log,
		Encoder: enc,
		Output:  iof.out,
	}
}

func (*ioFacility) Enabled(Entry) bool { return true }

func (iof *ioFacility) Log(ent Entry, fields ...Field) {
	if !iof.Logger.Enabled(ent.Level) {
		return
	}
	msg, enc := log.Encode(ent.Time, ent.Level, ent.Message, fields)
	if err := enc.WriteEntry(iof.Output, msg, ent.Level, ent.Time); err != nil {
		log.InternalError("encoder", err)
	}
	enc.Free()
	if ent.Level > ErrorLevel {
		// Sync on Panic and Fatal, since they may crash the program.
		iof.Output.Sync()
	}
}
