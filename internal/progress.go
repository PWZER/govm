package internal

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

func ndigits(i int64) int {
	var n int
	for ; i != 0; i /= 10 {
		n++
	}
	return n
}

func fmtSize(size int64) string {
	const (
		byte_unit = 1 << (10 * iota)
		kilobyte_unit
		megabyte_unit
	)

	unit := "B"
	value := float64(size)

	switch {
	case size >= megabyte_unit:
		unit = "MB"
		value = value / megabyte_unit
	case size >= kilobyte_unit:
		unit = "KB"
		value = value / kilobyte_unit
	}
	formatted := strings.TrimSuffix(strconv.FormatFloat(value, 'f', 1, 64), ".0")
	return fmt.Sprintf("%s %s", formatted, unit)
}

type progressWriter struct {
	w         io.Writer
	n         int64
	total     int64
	last      time.Time
	formatted bool
	output    io.Writer
}

func (p *progressWriter) update() {
	end := " ..."
	if p.n == p.total {
		end = ""
	}
	if p.formatted {
		fmt.Fprintf(p.output, "Downloaded %5.1f%% (%s / %s)%s\n",
			(100.0*float64(p.n))/float64(p.total),
			fmtSize(p.n), fmtSize(p.total), end)
	} else {
		fmt.Fprintf(p.output, "Downloaded %5.1f%% (%*d / %d bytes)%s\n",
			(100.0*float64(p.n))/float64(p.total),
			ndigits(p.total), p.n, p.total, end)
	}
}

func (p *progressWriter) Write(buf []byte) (n int, err error) {
	n, err = p.w.Write(buf)
	p.n += int64(n)
	if now := time.Now(); now.Unix() != p.last.Unix() {
		p.update()
		p.last = now
	}
	return
}
