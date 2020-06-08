package nfo

import (
	"os"
	"os/signal"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"syscall"
)

var (
	// Signal Notification Channel. (ie..nfo.Signal<-os.Kill will initiate a shutdown.)
	signalChan  = make(chan os.Signal)
	globalDefer []func() error
	defLock     sync.Mutex
	errCode     = 0
	wait        sync.WaitGroup
	exit_lock   = make(chan struct{})
)

// Global wait group, allows running processes to finish up tasks before app shutdown
func BlockShutdown() {
	wait.Add(1)
}

// Task completed, carry on with shutdown.
func UnblockShutdown() {
	wait.Done()
}

// This is a way of removing the global defer and instead locally defering to the function.
func LocalDefer(closer func() error) {
	defLock.Lock()
	defer defLock.Unlock()

	my_func := reflect.ValueOf(closer)
	tmp := globalDefer[:0]
	for _, v := range globalDefer {
		if reflect.ValueOf(v) != my_func {
			tmp = append(tmp, v)
		}
	}
	globalDefer = tmp
	closer()
}

// Adds a function to the global defer, function must take no arguments and either return nothing or return an error.
func Defer(closer interface{}) func() error {
	defLock.Lock()
	defer defLock.Unlock()

	errorWrapper := func(closerFunc func()) func() error {
		return func() error {
			closerFunc()
			return nil
		}
	}

	switch closer := closer.(type) {
	case func():
		e := errorWrapper(closer)
		globalDefer = append([]func() error{e}, globalDefer[0:]...)
		return e
	case func() error:
		globalDefer = append([]func() error{closer}, globalDefer[0:]...)
		return closer
	}
	return nil
}

// Intended to be a defer statement at the begining of main, but can be called at anytime with an exit code.
// Tries to catch a panic if possible and log it as a fatal error,
// then proceeds to send a signal to the global defer/shutdown handler
func Exit(exit_code int) {
	if r := recover(); r != nil {
		Fatal("(panic) %s", string(debug.Stack()))
	} else {
		atomic.StoreInt32(&fatal_triggered, 2) // Ignore any Fatal() calls, we've been told to exit.
		signalChan <- os.Kill
		<-exit_lock
		os.Exit(exit_code)
	}
}

// Sets the signals that we listen for.
func SetSignals(sig ...os.Signal) {
	mutex.Lock()
	defer mutex.Unlock()
	signal.Stop(signalChan)
	signal.Notify(signalChan, sig...)
}

// Set a callback function(no arguments) to run after receiving a specific syscall, function returns true to continue shutdown process.
func SignalCallback(signal os.Signal, callback func() (continue_shutdown bool)) {
	mutex.Lock()
	defer mutex.Unlock()
	callbacks[signal] = callback
}

var callbacks = make(map[os.Signal]func() bool)

func init() {
	SetSignals(syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		var err error
		for {
			s := <-signalChan

			write2log(_no_logging | _flash_txt | _bypass_lock)

			mutex.Lock()
			cb := callbacks[s]
			mutex.Unlock()

			if cb != nil {
				if !cb() {
					continue
				}
			}

			atomic.CompareAndSwapInt32(&fatal_triggered, 0, 2)

			switch s {
			case syscall.SIGINT:
				errCode = 130
			case syscall.SIGHUP:
				errCode = 129
			case syscall.SIGTERM:
				errCode = 143
			}

			break
		}

		defLock.Lock()
		defer defLock.Unlock()

		// Run through all globalDefer functions.
		for _, x := range globalDefer {
			if err = x(); err != nil {
				write2log(ERROR|_bypass_lock, err.Error())
			}
		}

		// Wait on any process that have access to wait.
		wait.Wait()

		// Close out all open files.
		for name := range open_files {
			Close(name)
		}

		// Finally exit the application
		select {
		case exit_lock <- struct{}{}:
		default:
			os.Exit(errCode)
		}
	}()
}
