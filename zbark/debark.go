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

package zbark

import (
	"fmt"

	"github.com/uber-go/zap"
	"github.com/uber-go/zap/zwrap"

	"github.com/uber-common/bark"
)

type zapperBarkFields zwrap.KeyValueMap

// Debarkify wraps bark.Logger to make it compatible with zap.logger.
func Debarkify(bl bark.Logger) *zap.Logger {
	if wrapper, ok := bl.(*barker); ok {
		return wrapper.log
	}
	return zap.New(&barkFacility{
		bl: bl,
	})
}

type barkFacility struct {
	bl bark.Logger
}

// Create a child logger, and optionally add some context to that logger.
func (bf *barkFacility) With(fields ...zap.Field) zap.Facility {
	return &barkFacility{
		bl: z.bl.WithFields(zapToBark(fields)),
	}
}

func (bf *barkFacility) Enabled(ent Entry) bool { return true }

func (bf *barkFacility) Log(ent zap.Entry, fields ...zap.Field) {
	// NOTE: logging at panic and fatal level actually panic and exit the
	// process, meaning that bark loggers cannot compose well.
	bl := bf.bl.WithFields(zapToBark(fields))
	switch ent.Level {
	case zap.DebugLevel:
		bl.Debug(ent.Message)
	case zap.InfoLevel:
		bl.Info(ent.Message)
	case zap.WarnLevel:
		bl.Warn(ent.Message)
	case zap.ErrorLevel:
		bl.Error(ent.Message)
	case zap.DPanicLevel:
		bl.Error(ent.Message)
	case zap.PanicLevel:
		bl.Panic(ent.Message)
	case zap.FatalLevel:
		bl.Fatal(ent.Message)
	default:
		// TODO: panic seems a bit strong
		panic(fmt.Errorf("passed an unknown zap.Level: %v", l))
	}
}

func (zbf zapperBarkFields) Fields() map[string]interface{} {
	return zbf
}

func zapToBark(zfs []zap.Field) bark.LogFields {
	zbf := make(zwrap.KeyValueMap, len(zfs))
	for _, zf := range zfs {
		zf.AddTo(zbf)
	}
	return zapperBarkFields(zbf)
}
