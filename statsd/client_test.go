// Copyright (©) 2012-2013 Timothée Peignier <timothee.peignier@tryphon.org>
// Copyright (©) 2014 TJ Holowaychuk <tj@vision-media.ca>
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

package statsd

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)

var client = &Client{}

func assert(t *testing.T, value, control string) {
	if value != control {
		t.Errorf("incorrect command, want '%s', got '%s'", control, value)
	}
}

func TestPrefix(t *testing.T) {
	buf := new(bytes.Buffer)
	c := NewClient(buf)
	c.Prefix("foo.bar.baz.")
	err := c.Increment("incr", 1, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	c.Flush()
	assert(t, buf.String(), "foo.bar.baz.incr:1|c||T"+strconv.FormatInt(time.Now().Unix(), 10))
}

func TestIncr(t *testing.T) {
	buf := new(bytes.Buffer)
	c := NewClient(buf)
	err := c.Incr("incr", nil)
	if err != nil {
		t.Fatal(err)
	}
	c.Flush()
	assert(t, buf.String(), "incr:1|c||T"+strconv.FormatInt(time.Now().Unix(), 10))
}

func TestDecr(t *testing.T) {
	buf := new(bytes.Buffer)
	c := NewClient(buf)
	err := c.Decr("decr", nil)
	if err != nil {
		t.Fatal(err)
	}
	c.Flush()
	assert(t, buf.String(), "decr:-1|c||T"+strconv.FormatInt(time.Now().Unix(), 10))
}

func TestDuration(t *testing.T) {
	buf := new(bytes.Buffer)
	c := NewClient(buf)
	err := c.Duration("timing", time.Duration(123456789), nil)
	if err != nil {
		t.Fatal(err)
	}
	c.Flush()
	assert(t, buf.String(), "timing:123|d||T"+strconv.FormatInt(time.Now().Unix(), 10))
}

func TestGauge(t *testing.T) {
	buf := new(bytes.Buffer)
	c := NewClient(buf)
	err := c.Gauge("gauge", 300, nil)
	if err != nil {
		t.Fatal(err)
	}
	c.Flush()
	assert(t, buf.String(), "gauge:300|g||T"+strconv.FormatInt(time.Now().Unix(), 10))
}

var millisecondTests = []struct {
	duration time.Duration
	control  int
}{
	{
		duration: 350 * time.Millisecond,
		control:  350,
	},
	{
		duration: 5 * time.Second,
		control:  5000,
	},
	{
		duration: 50 * time.Nanosecond,
		control:  0,
	},
}

func TestMilliseconds(t *testing.T) {
	for i, mt := range millisecondTests {
		value := millisecond(mt.duration)
		if value != mt.control {
			t.Errorf("%d: incorrect value, want %d, got %d", i, mt.control, value)
		}
	}
}

func TestMultiPacket(t *testing.T) {
	buf := new(bytes.Buffer)
	c := NewClient(buf)
	err := c.Unique("unique", 765, 1, map[string]string{"foo": "bar", "baz": "foo"})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Unique("unique", 765, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	c.Flush()
	assert(t, buf.String(), fmt.Sprintf("unique:765|s|#foo:bar,baz:foo|T%d\nunique:765|s||T%d", time.Now().Unix(), time.Now().Unix()))
}

func TestMultiPacketOverflow(t *testing.T) {
	t.Skip()
	buf := new(bytes.Buffer)
	c := NewClient(buf)
	for i := 0; i < 40; i++ {
		err := c.Unique("unique", 765, 1, map[string]string{"foo": "bar"})
		if err != nil {
			t.Fatal(err)
		}
	}
	assert(t, buf.String(), strings.TrimSuffix(strings.Repeat(fmt.Sprintf("unique:765|s|#foo:bar|T%d\n", time.Now().Unix()), 30), "\n"))
	buf.Reset()
	c.Flush()
	assert(t, buf.String(), strings.TrimSuffix(strings.Repeat(fmt.Sprintf("unique:765|s|#foo:bar|T%d\n", time.Now().Unix()), 10), "\n"))
}
