// Package 'nfo' is a simple central logging library with file log rotation as well as exporting to syslog.
// Additionally it provides a global defer for cleanly exiting applications and performing last minute tasks before application exits.

package nfo

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

import . "itoa"

const (
	INFO   = 1 << iota // Log Information
	AUX                // Auxilary Log
	ERROR              // Log Errors
	WARN               // Log Warning
	NOTICE             // Log Notices
	DEBUG              // Debug Logging
	TRACE              // Trace Logging
	FATAL              // Fatal Logging
	_flash_txt
	_print_txt
	_stderr_txt
)

// Standard Loggers, minus debug and trace.
const STD = INFO | AUX | ERROR | WARN | NOTICE | FATAL

var prefix = map[int]string{
	INFO:   "",
	AUX:    "",
	ERROR:  "[ERROR] ",
	WARN:   "[WARN] ",
	NOTICE: "[NOTICE] ",
	DEBUG:  "[DEBUG] ",
	TRACE:  "[TRACE] ",
	FATAL:  "[FATAL] ",
}

var (
	FatalOnOutError    = true // Fatal on Output logging error.
	FatalOnFileError   = true // Fatal on log file or file rotation errors.
	FatalOnExportError = true // Fatal on export/syslog error.
	flush_len          int
	flush_line         []rune
	flush_needed       bool
	fatal_triggered    int32
	msgBuffer          bytes.Buffer
	enabled_logging    = STD
	enabled_exports    = STD
	mutex              sync.Mutex
	use_ts             = true
	use_utc            = false
)

var l_map = map[int]*_logger{
	INFO:        {out1: os.Stdout, out2: None},
	AUX:         {out1: os.Stdout, out2: None},
	ERROR:       {out1: os.Stdout, out2: None},
	WARN:        {out1: os.Stdout, out2: None},
	NOTICE:      {out1: os.Stdout, out2: None},
	DEBUG:       {out1: os.Stdout, out2: None},
	TRACE:       {out1: os.Stdout, out2: None},
	FATAL:       {out1: os.Stdout, out2: None},
	_print_txt:  {out1: os.Stdout, out2: None},
	_stderr_txt: {out1: os.Stderr, out2: None},
}

type _logger struct {
	out1 io.Writer
	out2 io.Writer
}

// False writer for discarding output.
var None dummyWriter

type dummyWriter struct{}

func (dummyWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// Hide timestamps.
func HideTS() {
	mutex.Lock()
	defer mutex.Unlock()
	use_ts = false
}

// Show timestamps. (Default Enabled)
func ShowTS() {
	mutex.Lock()
	defer mutex.Unlock()
	use_ts = true
}

// Specify which loggers to enable, STD is enabled by default.
func SetLoggers(flag int) {
	mutex.Lock()
	defer mutex.Unlock()
	enabled_logging = flag
}

// Specify which logs to send to syslog.
func SetExports(flag int) {
	mutex.Lock()
	defer mutex.Unlock()
	enabled_exports = flag
}

// Switches timestamps to local timezone. (Default Setting)
func LTZ() {
	mutex.Lock()
	defer mutex.Unlock()
	use_utc = false
}

// Switches logger to use UTC instead of local timezone.
func UTC() {
	mutex.Lock()
	defer mutex.Unlock()
	use_utc = true
}

// Generate TS Bytes
func genTS(in *[]byte) {
	var CT time.Time

	if !use_utc {
		CT = time.Now()
	} else {
		CT = time.Now().UTC()
	}

	year, mon, day := CT.Date()
	hour, min, sec := CT.Clock()

	ts := in

	Itoa(ts, year, 4)
	*ts = append(*ts, '/')
	Itoa(ts, int(mon), 2)
	*ts = append(*ts, '/')
	Itoa(ts, day, 2)
	*ts = append(*ts, ' ')
	Itoa(ts, hour, 2)
	*ts = append(*ts, ':')
	Itoa(ts, min, 2)
	*ts = append(*ts, ':')
	Itoa(ts, sec, 2)
	*ts = append(*ts, ' ')
}

// Change output for logger(s).
func SetOutput(logger int, w io.Writer) {
	mutex.Lock()
	defer mutex.Unlock()
	for n, v := range l_map {
		if logger&n == n {
			v.out1 = w
		}
	}
}

// Change prefix for specified logger.
func SetPrefix(logger int, prefix_str string) {
	mutex.Lock()
	defer mutex.Unlock()
	for n := range prefix {
		if logger&n == n {
			prefix[n] = prefix_str
		}
	}
}

// Don't log, write text to standard error which will be overwritten on the next output.
func Flash(vars ...interface{}) {
	if atomic.LoadInt32(&fatal_triggered) == 0 {
		write2log(_flash_txt, vars...)
	}
}

// Don't log, just print text to standard out.
func Stdout(vars ...interface{}) {
	if atomic.LoadInt32(&fatal_triggered) == 0 {
		write2log(_print_txt|_flash_txt, vars...)
	}
}

// Don't log, just print text to standard error.
func Stderr(vars ...interface{}) {
	if atomic.LoadInt32(&fatal_triggered) == 0 {
		write2log(_stderr_txt|_flash_txt, vars...)
	}
}

// Log as Info.
func Log(vars ...interface{}) {
	if atomic.LoadInt32(&fatal_triggered) == 0 {
		write2log(INFO, vars...)
	}
}

// Log as Info, as auxilary output.
func Aux(vars ...interface{}) {
	if atomic.LoadInt32(&fatal_triggered) == 0 {
		write2log(AUX, vars...)
	}
}

// Log as Error.
func Err(vars ...interface{}) {
	if atomic.LoadInt32(&fatal_triggered) == 0 {
		write2log(ERROR, vars...)
	}
}

// Log as Warn.
func Warn(vars ...interface{}) {
	if atomic.LoadInt32(&fatal_triggered) == 0 {
		write2log(WARN, vars...)
	}
}

// Log as Notice.
func Notice(vars ...interface{}) {
	if atomic.LoadInt32(&fatal_triggered) == 0 {
		write2log(NOTICE, vars...)
	}
}

// Log as Fatal, then quit.
func Fatal(vars ...interface{}) {
	if atomic.CompareAndSwapInt32(&fatal_triggered, 0, 1) {
		// Defer fatal output, so it is the last log entry displayed.
		Defer(func() { write2log(FATAL, vars...) })
		signalChan <- os.Kill
		<-exit_lock
		os.Exit(1)
	}
}

// Log as Debug.
func Debug(vars ...interface{}) {
	if atomic.LoadInt32(&fatal_triggered) == 0 {
		write2log(DEBUG, vars...)
	}	
}

// Log as Trace.
func Trace(vars ...interface{}) {
	if atomic.LoadInt32(&fatal_triggered) == 0 {
		write2log(TRACE, vars...)
	}
}

// Prepares output text and sends to appropriate logging destinations.
func write2log(flag int, vars ...interface{}) {
	if enabled_logging&flag != flag && flag&_flash_txt != _flash_txt {
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	logger := l_map[flag]

	// Reset buffer.
	msgBuffer.Reset()

	var pre []byte

	if flag&_flash_txt != _flash_txt {
		if use_ts {
			genTS(&pre)
		}
		pre = append(pre, []byte(prefix[flag])[0:]...)
	}

	vlen := len(vars)

	if vlen == 0 {
		fmt.Fprintf(&msgBuffer, "")
		vlen = 1
	} else if vlen == 1 {
		if o, ok := vars[0].([]byte); ok {
			msgBuffer.Write(o)
		} else {
			fmt.Fprintf(&msgBuffer, "%v", vars[0])
		}
	} else {
		str, ok := vars[0].(string)
		if ok {
			fmt.Fprintf(&msgBuffer, str, vars[1:]...)
		} else {
			for n, item := range vars {
				if n == 0 || n == vlen-1 {
					fmt.Fprintf(&msgBuffer, "%v", item)
				} else {
					fmt.Fprintf(&msgBuffer, "%v, ", item)
				}
			}
		}
	}

	msg := msgBuffer.String()
	output := append(pre, msgBuffer.Bytes()[0:]...)
	bufferLen := utf8.RuneCount(output)

	if bufferLen > 0 && output[len(output)-1] != '\n' && flag != _flash_txt {
		output = append(output, '\n')
		bufferLen++
	}

	// Clear out last flash text.
	if flush_needed && (flag&_flash_txt == _flash_txt || logger.out1 != None) {
		fmt.Fprintf(os.Stderr, "\r%s\r", string(flush_line[0:flush_len]))
		flush_needed = false
	}

	// Flash text handler, make a line of text available to remove remnents of this text.
	if flag == _flash_txt {
		for i := len(flush_line); i < bufferLen; i++ {
			flush_line = append(flush_line[0:], ' ')
		}
		flush_len = bufferLen
		io.Copy(os.Stderr, bytes.NewReader(output))
		flush_needed = true
		return
	}

	if flag&_flash_txt == _flash_txt {
		if flag&_print_txt == _print_txt {
			flag = _print_txt
		} else {
			flag = _stderr_txt
		}
		io.Copy(l_map[flag].out1, bytes.NewReader(output))
		return
	}

	_, err := io.Copy(logger.out1, bytes.NewReader(output))
	if err != nil && FatalOnOutError {
		go Fatal(err)
	}

	// Preprend timestamp for file.
	if !use_ts {
		out_len := len(output)
		genTS(&output)
		out := output[out_len:]
		out = append(out, output[0:out_len]...)
		output = out
	}

	// Write to file.
	_, err = io.Copy(logger.out2, bytes.NewReader(output))
	if err != nil && FatalOnFileError {
		go Fatal(err)
	}

	if export_syslog != nil && enabled_exports&flag == flag {
		switch flag {
		case INFO:
		case AUX:
			err = export_syslog.Info(msg)
		case ERROR:
			err = export_syslog.Err(msg)
		case WARN:
			err = export_syslog.Warning(msg)
		case FATAL:
			err = export_syslog.Emerg(msg)
		case NOTICE:
			err = export_syslog.Notice(msg)
		case DEBUG:
			err = export_syslog.Debug(msg)
		case TRACE:
			err = export_syslog.Debug(msg)
		}
		if err != nil && FatalOnExportError {
			go Fatal(err)
		}
	}

}
