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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"
)

const defaultBufSize = 256

// Client is statsd client representing a
// connection to a statsd server.
type Client struct {
	buf    *bufio.Writer
	m      sync.Mutex
	prefix string
}

func millisecond(d time.Duration) int {
	return int(d.Seconds() * 1000)
}

func parsetags(tags map[string]string) string {
	var b bytes.Buffer
	length := len(tags)
	if length == 0 {
		return ""
	}

	b.WriteString("#")
	i := 0
	for key, value := range tags {
		b.WriteString(key)
		b.WriteString(":")
		b.WriteString(value)
		if i != length-1 {
			b.WriteString(",")
		}

		i++
	}

	return b.String()
}

// NewClient returns a new client with the given writer,
// useful for testing.
func NewClient(w io.Writer) *Client {
	return &Client{
		buf: bufio.NewWriterSize(w, defaultBufSize),
	}
}

// Prefix adds a prefix to every stat string. The prefix is literal,
// so if you want "foo.bar.baz" from "baz" you should set the prefix
// to "foo.bar." not "foo.bar" as no delimiter is added for you.
func (c *Client) Prefix(s string) {
	c.prefix = s
}

// Increment increments the counter for the given bucket.
func (c *Client) Increment(name string, count int, rate float64, tags map[string]string) error {
	return c.send(name, rate, "%d|c|%s|T%d", count, parsetags(tags), time.Now().Unix())
}

// Incr increments the counter for the given bucket by 1 at a rate of 1.
func (c *Client) Incr(name string, tags map[string]string) error {
	return c.Increment(name, 1, 1, tags)
}

// IncrBy increments the counter for the given bucket by N at a rate of 1.
func (c *Client) IncrBy(name string, n int, tags map[string]string) error {
	return c.Increment(name, n, 1, tags)
}

// Decrement decrements the counter for the given bucket.
func (c *Client) Decrement(name string, count int, rate float64, tags map[string]string) error {
	return c.Increment(name, -count, rate, tags)
}

// Decr decrements the counter for the given bucket by 1 at a rate of 1.
func (c *Client) Decr(name string, tags map[string]string) error {
	return c.Increment(name, -1, 1, tags)
}

// DecrBy decrements the counter for the given bucket by N at a rate of 1.
func (c *Client) DecrBy(name string, value int, tags map[string]string) error {
	return c.Increment(name, -value, 1, tags)
}

// Duration records time spent for the given bucket with time.Duration.
func (c *Client) Duration(name string, duration time.Duration, tags map[string]string) error {
	return c.send(name, 1, "%d|d|%s|T%d", millisecond(duration), parsetags(tags), time.Now().Unix())
}

// Histogram is an alias of .Duration() until the statsd protocol figures its shit out.
func (c *Client) Histogram(name string, value uint64, tags map[string]string) error {
	return c.send(name, 1, "%d|h|%s|T%d", value, parsetags(tags), time.Now().Unix())
}

// Gauge records arbitrary values for the given bucket.
func (c *Client) Gauge(name string, value int64, tags map[string]string) error {
	return c.send(name, 1, "%d|g|%s|T%d", value, parsetags(tags), time.Now().Unix())
}

// Unique records unique occurences of events.
func (c *Client) Unique(name string, value int, rate float64, tags map[string]string) error {
	return c.send(name, rate, "%d|s|%s|T%d", value, parsetags(tags), time.Now().Unix())
}

// Flush flushes writes any buffered data to the network.
func (c *Client) Flush() error {
	return c.buf.Flush()
}

// Close closes the connection.
func (c *Client) Close() error {
	if err := c.Flush(); err != nil {
		return err
	}
	c.buf = nil
	return nil
}

// send stat.
func (c *Client) send(stat string, rate float64, format string, args ...interface{}) error {
	if c.prefix != "" {
		stat = c.prefix + stat
	}

	if rate < 1 {
		if rand.Float64() < rate {
			format = fmt.Sprintf("%s|@%g", format, rate)
		} else {
			return nil
		}
	}

	format = fmt.Sprintf("\n%s:%s", stat, format)

	c.m.Lock()
	defer c.m.Unlock()

	// Flush data if we have reach the buffer limit
	if c.buf.Available() < len(format) {
		if err := c.Flush(); err != nil {
			return nil
		}
	}

	// Buffer is not empty, start filling it
	if c.buf.Buffered() > 0 {
		format = fmt.Sprintf("%s", format)
	}

	_, err := fmt.Fprintf(c.buf, format, args...)
	return err
}
