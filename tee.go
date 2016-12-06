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

// Tee creates a Facility that duplicates log entries into two or more
// facilities; if you call it with less than two, you get back the one facility
// you passed (or nil in the pathological case).
func Tee(facs ...Facility) Facility {
	switch len(facs) {
	case 0:
		return nil
	case 1:
		return facs[0]
	default:
		return multiFacility(facs)
	}
}

type multiFacility []Facility

func (mf multiFacility) With(fields ...Field) Facility {
	clone := make(multiFacility, len(mf))
	for i := range mf {
		clone[i] = mf[i].With(fields...)
	}
	return clone
}

func (mf multiFacility) Log(ent Entry, fields ...Field) {
	for _, log := range mf {
		log.Log(ent, fields...)
	}
}

func (mf multiFacility) Enabled(ent Entry) bool {
	for i := range mf {
		if mf[i].Enabled(ent) {
			return true
		}
	}
	return false
}
