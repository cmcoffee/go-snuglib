package nfo

import (
	"fmt"
	. "github.com/cmcoffee/go-snuglib/xsync"
	"golang.org/x/crypto/ssh/terminal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// For displaying multiple simultaneous transfers
var transferDisplay struct {
	update_lock sync.RWMutex
	display     int64
	monitors    []*tmon
}

// ReadSeekCloser interface
type ReadSeekCloser interface {
	Seek(offset int64, whence int) (int64, error)
	Read(p []byte) (n int, err error)
	Close() error
}

func termWidth() int {
	width, _, _ := terminal.GetSize(int(syscall.Stderr))
	width--
	if width < 1 {
		width = 0
	}
	return width
}

const (
	LeftToRight = 1 << iota // Display progress bar left to right.
	RightToLeft             // Display progress bar right to left.
	NoRate                  // Do not show transfer rate, left to right.
	trans_active
	trans_closed
	trans_complete
	trans_error
)

type readSeekCounter struct {
	counter func(int)
	ReadSeekCloser
}

func (r readSeekCounter) Read(p []byte) (n int, err error) {
	n, err = r.ReadSeekCloser.Read(p)
	r.counter(n)
	return
}

// TransferCounter allows you to add a counter callback function to add bytes added during read.
func TransferCounter(input ReadSeekCloser, counter func(int)) ReadSeekCloser {
	return readSeekCounter{
		counter,
		input,
	}
}

// Add Transfer to transferDisplay.
// Parameters are "name" displayed for file transfer, "limit_sz" for when to pause transfer (aka between calls/chunks), and "total_sz" the total size of the transfer.
func TransferMonitor(name string, total_size int64, flag int, source ReadSeekCloser, optional_prefix ...string) ReadSeekCloser {
	transferDisplay.update_lock.Lock()
	defer transferDisplay.update_lock.Unlock()

	var (
		short_name  []rune
		target_size int
		prefix      string
	)

	if len(optional_prefix) > 0 {
		prefix = optional_prefix[0]
	}

	b_flag := BitFlag(flag)
	if b_flag.Has(LeftToRight) || b_flag <= 0 {
		b_flag.Set(LeftToRight)
	}

	if !b_flag.Has(NoRate) {
		target_size = 18
	} else {
		target_size = 36
	}

	for i, v := range name {
		if i < target_size {
			short_name = append(short_name, v)
		} else {
			short_name = append(short_name, []rune("..")[0:]...)
			break
		}
	}

	if len(short_name) < target_size {
		x := len(short_name) - 1
		var y []rune
		for i := 0; i <= target_size-x; i++ {
			y = append(y, ' ')
		}
		short_name = append(y[0:], short_name[0:]...)
	}

	b_flag.Set(trans_active)

	tm := &tmon{
		flag:        b_flag,
		name:        name,
		prefix:      prefix,
		short_name:  string(short_name),
		total_size:  total_size,
		transferred: 0,
		offset:      0,
		rate:        "0.0bps",
		start_time:  time.Now(),
		source:      source,
	}

	var spin_index int
	spin_txt := []string{"\\", "|", "/", "-"}

	spinner := func() string {
		if spin_index < len(spin_txt)-1 {
			spin_index++
		} else {
			spin_index = 0
		}
		return fmt.Sprintf(spin_txt[spin_index])
	}

	transferDisplay.monitors = append(transferDisplay.monitors, tm)

	if len(transferDisplay.monitors) == 1 {
		PleaseWait.flag.Set(transfer_monitor_active)
		transferDisplay.display = 1

		go func() {
			for {
				transferDisplay.update_lock.Lock()

				var monitors []*tmon

				// Clean up transfers.
				for i := len(transferDisplay.monitors) - 1; i >= 0; i-- {
					if transferDisplay.monitors[i].flag.Has(trans_closed) {
						transferDisplay.monitors = append(transferDisplay.monitors[:i], transferDisplay.monitors[i+1:]...)
					} else {
						monitors = append(monitors, transferDisplay.monitors[i])
					}
				}

				if len(transferDisplay.monitors) == 0 {
					PleaseWait.flag.Unset(transfer_monitor_active)
					transferDisplay.update_lock.Unlock()
					return
				}

				transferDisplay.update_lock.Unlock()

				// Display transfers.
				for _, v := range monitors {
					for i := 0; i < 10; i++ {
						if v.flag.Has(trans_active) {
							Flash("[%s] %s", spinner(), v.showTransfer(false))
						} else {
							break
						}
						time.Sleep(time.Millisecond * 200)
					}
				}
			}
		}()

	}

	return tm
}

// Wrapper Seeker
func (tm *tmon) Seek(offset int64, whence int) (int64, error) {
	o, err := tm.source.Seek(offset, whence)
	tm.transferred = o
	tm.offset = o
	return o, err
}

// Wrapped Reader
func (tm *tmon) Read(p []byte) (n int, err error) {
	n, err = tm.source.Read(p)
	atomic.StoreInt64(&tm.transferred, atomic.LoadInt64(&tm.transferred)+int64(n))
	if err != nil {
		if tm.flag.Has(trans_closed) {
			return
		}
		tm.flag.Set(trans_closed | trans_error)
		if tm.transferred == 0 {
			return
		}
	}
	return
}

// Close out speicfic transfer monitor
func (tm *tmon) Close() error {
	tm.flag.Set(trans_closed)
	if !tm.flag.Has(NoRate) {
		if tm.transferred > 0 || tm.total_size == 0 {
			Log(tm.showTransfer(true))
		}
	}
	return tm.source.Close()
}

func spacePrint(min int, input string) string {
	output := make([]rune, min)
	for i := 0; i < len(output); i++ {
		output[i] = ' '
	}
	return string(append(output[len(input)-1:], []rune(input)[0:]...))
}

// Transfer Monitor
type tmon struct {
	flag        BitFlag
	prefix      string
	name        string
	short_name  string
	total_size  int64
	transferred int64
	offset      int64
	rate        string
	chunk_size  int64
	start_time  time.Time
	source      ReadSeekCloser
}

// Outputs progress of TMonitor.
func (t *tmon) showTransfer(summary bool) string {
	transferred := atomic.LoadInt64(&t.transferred)
	rate := t.showRate()

	var name string

	if summary {
		t.flag.Unset(trans_active)
		name = t.name
	} else {
		name = t.short_name
	}

	// 35 + 8 +8 + 8 + 8
	if t.total_size > -1 {
		if !t.flag.Has(NoRate) {
			return fmt.Sprintf("%s", t.progressBar(name))
		} else {
			return DrawProgressBar(19, t.transferred, t.total_size, t.name)
		}
	} else {
		return fmt.Sprintf("%s: %s (%s) ", t.name, rate, HumanSize(transferred))
	}
}

// Provides average rate of transfer.
func (t *tmon) showRate() (rate string) {

	transferred := atomic.LoadInt64(&t.transferred)
	if transferred == 0 || t.flag.Has(trans_complete) {
		return t.rate
	}

	since := time.Since(t.start_time).Seconds()
	if since < 0.1 {
		since = 0.1
	}

	sz := float64(transferred-t.offset) * 8 / since

	names := []string{
		"bps",
		"kbps",
		"mbps",
		"gbps",
	}

	suffix := 0

	for sz >= 1000 && suffix < len(names)-1 {
		sz = sz / 1000
		suffix++
	}

	if sz != 0.0 {
		rate = fmt.Sprintf("%.1f%s", sz, names[suffix])
	} else {
		if t.flag.Has(trans_active) {
			rate = "0.0bps"
		} else {
			rate = "\b"
		}
	}

	t.rate = rate

	if !t.flag.Has(trans_complete) && atomic.LoadInt64(&t.transferred)+t.offset == t.total_size {
		t.flag.Set(trans_complete)
	}

	if !t.flag.Has(trans_closed) {
		return string(append([]rune{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '}[len(rate)-1:], []rune(rate)[0:]...))
	} else {
		return t.rate
	}
}

// Draws a progress bar using sz as the size.
func DrawProgressBar(sz int, current, max int64, text string) string {
	var num int
	if max > 0 {
		num = int(float64(current) / float64(max) * 100)
	} else {
		num = 0
	}

	display := make([]rune, sz)
	x := num * sz / 100

	for n := range display {
		if n < x {
			display[n] = '░'
		} else {
			display[n] = '.'
		}
	}

	perc := strconv.Itoa(num)

	return fmt.Sprintf("[%s]%s%%: %s", string(display[0:]), string(append([]rune{' ', ' '}[len(perc)-1:], []rune(perc)[0:]...)), text)

}

// Produces progress bar for information on update.
func (t *tmon) progressBar(name string) string {
	num := int((float64(atomic.LoadInt64(&t.transferred)) / float64(t.total_size)) * 100)

	if t.total_size == 0 {
		num = 100
	}

	sz := termWidth() - 3

	first_half := fmt.Sprintf("%s: %s", name, t.showRate())
	second_half := fmt.Sprintf("(%s/%s)", HumanSize(t.transferred), HumanSize(t.total_size))

	sz = sz - len(first_half) - 35

	if t.flag.Has(trans_closed) && !t.flag.Has(NoRate) || sz <= 0 {
		sz = 10
	}

	display := make([]rune, sz)
	x := num * sz / 100

	if t.flag.Has(LeftToRight) {
		for n := range display {
			if n < x {
				if n+1 < x {
					display[n] = '='
				} else {
					display[n] = '>'
				}
			} else {
				display[n] = ' '
			}
		}
	} else {
		x = sz - x - 1
		for n := range display {
			if n > x {
				if n-1 > x {
					display[n] = '='
				} else {
					display[n] = '<'
				}
			} else {
				display[n] = ' '
			}
		}
	}

	if sz > 10 {
		return fmt.Sprintf("%s [%s] %d%% %s ", first_half, string(display[0:]), int(num), second_half)
	} else {
		if t.flag.Has(trans_closed) {
			return fmt.Sprintf("%s%s %d%% %s", t.prefix, first_half, int(num), second_half)
		} else {
			return fmt.Sprintf("%s %d%% %s", first_half, int(num), second_half)
		}
	}
}

// Provides human readable file sizes.
func HumanSize(bytes int64) string {

	names := []string{
		"Bytes",
		"KB",
		"MB",
		"GB",
	}

	suffix := 0
	size := float64(bytes)

	for size >= 1024 && suffix < len(names)-1 {
		size = size / 1024
		suffix++
	}

	return fmt.Sprintf("%.1f%s", size, names[suffix])
}
