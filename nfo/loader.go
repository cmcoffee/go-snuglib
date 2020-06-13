package nfo

import (
	"fmt"
	"github.com/cmcoffee/go-snuglib/bitflag"
	"sync"
	"sync/atomic"
	"time"
)

func init() {
	PleaseWait = new(loader)
	ProgressBar = new(progressBar)
	PleaseWait.Set(func() string { return "Please wait ..." }, []string{"[>  ]", "[>> ]", "[>>>]", "[ >>]", "[  >]", "[  <]", "[ <<]", "[<<<]", "[<< ]", "[<  ]"})
	Defer(func() { PleaseWait.Hide() })
}

// PleaseWait is a wait prompt to display between requests.
var PleaseWait *loader

type loader struct {
	flag     bitflag.BitFlag
	message  func() string
	loader_1 []string
	loader_2 []string
	mutex    sync.Mutex
}

const (
	loader_running = 1 << iota
	loader_stop
	loader_show
)

// Specify a "Please wait" animated PleaseWait line.
func (L *loader) Set(message func() string, loader ...[]string) {
	L.mutex.Lock()
	defer L.mutex.Unlock()

	if len(loader) == 0 {
		return
	}

	// Disable exisiting PleaseWait
	if L.flag.Has(loader_running) {
		L.flag.Set(loader_stop)
	}

	for {
		if L.flag.Has(loader_running) {
			time.Sleep(time.Millisecond * 125)
			continue
		}
		break
	}

	var loader_1, loader_2 []string

	loader_1 = loader[0]
	if len(loader) > 1 {
		loader_2 = loader[1]
	}

	if loader_2 == nil || len(loader_2) < len(loader_1) {
		loader_2 = make([]string, len(loader_1))
	}

	L.message = message
	L.loader_1 = loader_1
	L.loader_2 = loader_2

	L.flag.Unset(loader_stop)
	L.flag.Set(loader_running)

	go func(message func() string, loader_1 []string, loader_2 []string) {
		for {
			if L.flag.Has(loader_stop) {
				L.flag.Unset(loader_running)
				break
			}
			for i, str := range loader_1 {
				if L.flag.Has(loader_show) {
					Flash("%s %s %s", str, message(), loader_2[i])
				}
				time.Sleep(125 * time.Millisecond)
			}
		}
	}(message, loader_1, loader_2)
}

// Displays loader. "[>>>] Working, Please wait."
func (L *loader) Show() {
	L.flag.Set(loader_show)
}

// Hides display loader.
func (L *loader) Hide() {
	L.flag.Unset(loader_show)
	Flash("")
}

type progressBar struct {
	existing func() string
	cur      int32
	max      int32
	working  bool
	name     string
	loader_1 []string
	loader_2 []string
}

var ProgressBar *progressBar

// Produces progress bar for information on update.
func (p *progressBar) draw() string {
	num := int((float64(atomic.LoadInt32(&p.cur)) / float64(atomic.LoadInt32(&p.max))) * 100)
	sz := termWidth() - len(p.name) - 42
	if sz > 10 {
		display := make([]rune, sz)
		x := num * sz / 100
		for n := range display {
			if n < x {
				display[n] = '░'
			} else {
				display[n] = '.'
			}
		}
		return fmt.Sprintf("%d%% [%s]", int(num), string(display[0:]))
	} else {
		return fmt.Sprintf("%d%%", int(num))
	}
}

func (p *progressBar) updateMessage() string {
	return fmt.Sprintf("%s (%d/%d %s)", p.draw(), p.cur, p.max, p.name)
}

func (p *progressBar) New(name string, max int) {
	if p.working {
		return
	}

	p.cur = 0
	p.max = int32(max)
	p.existing = PleaseWait.message
	p.name = name
	p.loader_1 = PleaseWait.loader_1
	p.loader_2 = PleaseWait.loader_2
	PleaseWait.Set(p.updateMessage, p.loader_1)
	p.working = true
}

func (p *progressBar) Add(num int) {
	if !p.working {
		return
	}
	atomic.StoreInt32(&p.cur, atomic.LoadInt32(&p.cur)+int32(num))
}

func (p *progressBar) Sub(num int) {
	if !p.working {
		return
	}
	atomic.StoreInt32(&p.cur, atomic.LoadInt32(&p.cur)-int32(num))
}

func (p *progressBar) Done() {
	if !p.working {
		return
	}
	PleaseWait.Set(p.existing, p.loader_1, p.loader_2)
	p.working = false
}
