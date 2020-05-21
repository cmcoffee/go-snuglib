package nfo

import (
	"sync/atomic"
	"time"
)

var start_time time.Time

func init() {
	start_time = time.Now()

	go func() {
		loader_1 := []string{"[\\]", "[|]", "[/]", "[-]"}
		for {
			for _, str := range loader_1 {

				if x := atomic.LoadInt32((*int32)(&PleaseWait)); x == 1 {
					Flash("%s Please wait ... ", str)
				} else if x == 2 {
					atomic.StoreInt32((*int32)(&PleaseWait), 3)
					return
				}
				time.Sleep(125 * time.Millisecond)
			}
		}
	}()
	Defer(func() {
		PleaseWait.Hide()
		Stdout("\n")
	})
}

type _loader int32

// PleaseWait is a wait prompt to display between requests.
var PleaseWait _loader

// Specify a "Please wait" animated PleaseWait line.
func (L *_loader) Set(message string, include_runtime bool, loader ...[]string) {

	if len(loader) == 0 {
		return
	}

	existing := (int32)(PleaseWait)

	// Disable exisiting PleaseWait
	atomic.StoreInt32((*int32)(&PleaseWait), 2)
	for {
		if atomic.LoadInt32((*int32)(&PleaseWait)) == 2 {
			time.Sleep(time.Millisecond * 125)
			continue
		}
		break
	}
	atomic.StoreInt32((*int32)(&PleaseWait), existing)

	var loader_1, loader_2 []string

	loader_1 = loader[0]
	if len(loader) > 1 {
		loader_2 = loader[1]
	}

	if loader_2 == nil || len(loader_2) < len(loader_1) {
		loader_2 = make([]string, len(loader_1))
	}

	go func(message string, include_runtime bool, loader_1 []string, loader_2 []string) {
		for {
			if atomic.LoadInt32((*int32)(&PleaseWait)) == 2 {
				atomic.StoreInt32((*int32)(&PleaseWait), 3)
				break
			}
			for i, str := range loader_1 {
				if atomic.LoadInt32((*int32)(&PleaseWait)) == 1 {
					if include_runtime {
						Flash("%s %s (%s) %s", str, message, time.Now().Sub(start_time).Round(time.Second).String(), loader_2[i])
					} else {
						Flash("%s %s %s", str, message, loader_2[i])
					}
				}
				time.Sleep(125 * time.Millisecond)
			}
		}
	}(message, include_runtime, loader_1, loader_2)
}

// Displays loader. "[>>>] Working, Please wait."
func (L *_loader) Show() {
	atomic.CompareAndSwapInt32((*int32)(&PleaseWait), 0, 1)
}

// Hides display loader.
func (L *_loader) Hide() {
	atomic.CompareAndSwapInt32((*int32)(&PleaseWait), 1, 0)
}
