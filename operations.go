package thermoprinter

import (
	"strings"
	"time"
)

func (p *Printer) Write(s string) error {
	var (
		err error
	)

	// write character by character
	for _, c := range strings.Split(s, "") {
		err = p.writeBytes([]byte(c))
		if err != nil {
			return err
		}

		if c == "\n" || p.column == MaxColumn {
			p.column = 0
		} else {
			p.column++
		}
	}
	return nil
}

// feed n lines.
func (p *Printer) Feed(n int) error {
	err := p.writeBytes([]byte{ASCII_ESC, byte('d'), byte(n)})
	if err != nil {
		return err
	}

	return nil
}

// writing bytes is used for settings modes on the printer, e.g. bold
// text weight modes, etc. It is not exposed as a public method since
// there are higher level methods which wrap this.
func (p *Printer) writeBytes(bytes []byte) error {
	var (
		nBytes int
		err    error
	)

	// block until we can write
	<-p.writeReady

	// write to serial stream
	nBytes, err = p.stream.Write(bytes)
	if err != nil {
		p.readyAfter <- 0
		return err
	}

	// reset time to wait before next write
	p.readyAfter <- nBytes

	return nil
}

// runs inside a goroutine and handles write timing, allowing for the printer
// to complete mechanical processes before moving on to the next operation.
func (p *Printer) handleWriteTiming() {
	// initialize timer. writes should be immediately ready, so
	// the timer should immediately fire.
	timer := time.AfterFunc(
		0,
		func() {
			// enable writes to printer once again
			p.writeReady <- true
		},
	)

	for {
		// block until we want to reset the timer.
		n := <-p.readyAfter

		// we've just written n bytes, we now want to reset
		// a timeout for n * WaitTimeAfterWrite
		timer.Reset(time.Duration(n) * p.WaitTimeAfterWrite)
	}
}
