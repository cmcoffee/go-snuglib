// Package 'nfo' is a simple central logging library with file log rotation as well as exporting to syslog.
// Additionally it provides a global defer for cleanly exiting applications and performing last minute tasks before application exits.

package nfo

import (
	"bytes"
	"fmt"
	"github.com/cmcoffee/go-wrotate"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

import . "itoa"

const (
	INFO   = 1 << iota // Log Information
	ERROR              // Log Errors
	WARN               // Log Warning
	NOTICE             // Log Notices
	DEBUG              // Debug Logging
	TRACE              // Trace Logging
	FATAL              // Fatal Logging
	AUX                // Auxilary Log
	AUX2               // Auxilary Log
	AUX3               // Auxilary Log
	AUX4               // Auxilary Log
	_flash_txt
	_print_txt
	_stderr_txt
	_bypass_lock
	_no_logging
)

// Standard Loggers, minus debug and trace.
const (
	STD = INFO | ERROR | WARN | NOTICE | FATAL | AUX | AUX2 | AUX3 | AUX4
	ALL = INFO | ERROR | WARN | NOTICE | FATAL | AUX | AUX2 | AUX3 | AUX4 | DEBUG | TRACE
)

var prefix = map[int]string{
	INFO:   "",
	AUX:    "",
	AUX2:   "",
	AUX3:   "",
	AUX4:   "",
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
	piped_stdout       bool
	piped_stderr       bool
	fatal_triggered    int32
	msgBuffer          bytes.Buffer
	enabled_exports    = STD
	mutex              sync.Mutex
	use_ts             = true
	use_utc            = false
)

// False writer for discarding output.
var None dummyWriter

type dummyWriter struct{}

func (dummyWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func (dummyWriter) Close() error {
	return nil
}

var l_map = map[int]*_logger{
	INFO:        {os.Stdout, None, true},
	AUX:         {os.Stdout, None, true},
	AUX2:        {os.Stdout, None, true},
	AUX3:        {os.Stdout, None, true},
	AUX4:        {os.Stdout, None, true},
	ERROR:       {os.Stdout, None, true},
	WARN:        {os.Stdout, None, true},
	NOTICE:      {os.Stdout, None, true},
	DEBUG:       {None, None, true},
	TRACE:       {None, None, true},
	FATAL:       {os.Stdout, None, true},
	_flash_txt:  {os.Stderr, None, false},
	_print_txt:  {os.Stdout, None, false},
	_stderr_txt: {os.Stderr, None, false},
}

func init() {
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		piped_stdout = true
	}
	if !terminal.IsTerminal(int(os.Stderr.Fd())) {
		piped_stderr = true
	}
}

type _logger struct {
	out1   io.Writer
	out2   io.WriteCloser
	use_ts bool
}

// Keep map of open files
var open_files = make(map[string]io.WriteCloser)
var open_files_mutex sync.Mutex

// Creates folders.
func mkDir(name ...string) (err error) {
	for _, path := range name {
		subs := strings.Split(path, string(os.PathSeparator))
		for i := 0; i < len(subs); i++ {
			p := strings.Join(subs[0:i], string(os.PathSeparator))
			if p == "" {
				p = "."
			}
			_, err = os.Stat(p)
			if err != nil {
				if os.IsNotExist(err) {
					err = os.Mkdir(p, 0766)
					if err != nil {
						return err
					}
				} else {
					return err
				}
			}
		}
	}
	return nil
}

// Opens a new log file for writing, max_size is threshold for rotation, max_rotation is number of previous logs to hold on to.
// Set max_size_mb to 0 to disable file rotation.
func File(l_file_flag int, filename string, max_size_mb uint, max_rotation uint) (err error) {
	max_size := int64(max_size_mb * 1048576)
	fpath, _ := filepath.Split(filename)

	if err := mkDir(fpath); err != nil {
		return err
	}

	file, err := wrotate.OpenFile(filename, max_size, max_rotation)
	if err != nil {
		return err
	}

	open_files_mutex.Lock()
	defer open_files_mutex.Unlock()

	open_files[filename] = file
	SetFile(l_file_flag, file)

	return nil
}

// Closes out a log file.
func Close(filename string) (err error) {
	open_files_mutex.Lock()
	defer open_files_mutex.Unlock()
	mutex.Lock()
	defer mutex.Unlock()
	f := open_files[filename]
	for _, v := range l_map {
		if v.out2 == f {
			v.out2 = None
		}
	}
	delete(open_files, filename)
	return f.Close()
}

// Tacks an additional logger to an exising log file.
func LogFileAppend(existing_logger int, flag int) {
	logger := getLogger(existing_logger)
	updateLogger(flag, 2, logger.out2)
}

// Retrieve first matching logger.
func getLogger(flag int) *_logger {
	mutex.Lock()
	defer mutex.Unlock()
	for k, v := range l_map {
		if flag&k == k {
			return v
		}
	}
	return nil
}

// Updates logger.
func updateLogger(flag int, field int, input interface{}) {
	mutex.Lock()
	defer mutex.Unlock()
	for k, v := range l_map {
		if flag&k == k {
			switch field {
			case 1:
				if x, ok := input.(io.Writer); ok {
					v.out1 = x
				} else {
					return
				}
			case 2:
				if x, ok := input.(io.WriteCloser); ok {
					v.out2 = x
				} else {
					return
				}
			case 3:
				if x, ok := input.(bool); ok {
					v.use_ts = x
				} else {
					return
				}
			default:
				return
			}
		}
	}
}

// Hide timestamps in output.
func HideTS() {
	updateLogger(ALL, 3, false)
}

// Show timestamps. (Default Enabled)
func ShowTS() {
	updateLogger(ALL, 3, true)
}

// Enable/Disable Timestamp on output.
func SetTimestamp(flag int, use_ts bool) {
	updateLogger(flag, 3, use_ts)
}

// Enable a specific logger.
func SetOutput(flag int, w io.Writer) {
	updateLogger(flag, 1, w)
}

func SetFile(flag int, input io.Writer) {
	updateLogger(flag, 2, input)
}

// Disable a specific logger
func DisableOutput(flag int) {
	updateLogger(flag, 1, None)
}

// Specify which logs to send to syslog.
func EnableExport(flag int) {
	mutex.Lock()
	defer mutex.Unlock()
	enabled_exports = enabled_exports | flag
}

// Specific which logger to not export.
func DisableExport(flag int) {
	mutex.Lock()
	defer mutex.Unlock()
	enabled_exports = enabled_exports & ^flag
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
	write2log(_flash_txt|_no_logging, vars...)
}

// Don't log, just print text to standard out.
func Stdout(vars ...interface{}) {
	write2log(_print_txt|_no_logging, vars...)
}

// Don't log, just print text to standard error.
func Stderr(vars ...interface{}) {
	write2log(_stderr_txt|_no_logging, vars...)
}

// Log as Info.
func Log(vars ...interface{}) {
	write2log(INFO, vars...)
}

// Log as Error.
func Err(vars ...interface{}) {
	write2log(ERROR, vars...)
}

// Log as Warn.
func Warn(vars ...interface{}) {
	write2log(WARN, vars...)
}

// Log as Notice.
func Notice(vars ...interface{}) {
	write2log(NOTICE, vars...)
}

// Log as Info, as auxilary output.
func Aux(vars ...interface{}) {
	write2log(AUX, vars...)
}

// Log as Info, as auxilary output.
func Aux2(vars ...interface{}) {
	write2log(AUX2, vars...)
}

// Log as Info, as auxilary output.
func Aux3(vars ...interface{}) {
	write2log(AUX3, vars...)
}

// Log as Info, as auxilary output.
func Aux4(vars ...interface{}) {
	write2log(AUX4, vars...)
}

// Log as Fatal, then quit.
func Fatal(vars ...interface{}) {
	if atomic.CompareAndSwapInt32(&fatal_triggered, 0, 1) {
		// Defer fatal output, so it is the last log entry displayed.
		write2log(FATAL|_bypass_lock, vars...)
		signalChan <- os.Kill
		<-exit_lock
		os.Exit(1)
	}
}

// Log as Debug.
func Debug(vars ...interface{}) {
	write2log(DEBUG, vars...)
}

// Log as Trace.
func Trace(vars ...interface{}) {
	write2log(TRACE, vars...)
}

// sprintf
func outputFactory(buffer io.Writer, vars ...interface{}) {
	vlen := len(vars)

	if vlen == 0 {
		fmt.Fprintf(buffer, "")
		vlen = 1
	} else if vlen == 1 {
		if o, ok := vars[0].([]byte); ok {
			buffer.Write(o)
		} else {
			fmt.Fprintf(buffer, "%v", vars[0])
		}
	} else {
		str, ok := vars[0].(string)
		if ok {
			fmt.Fprintf(buffer, str, vars[1:]...)
		} else {
			for n, item := range vars {
				if n == 0 || n == vlen-1 {
					fmt.Fprintf(buffer, "%v", item)
				} else {
					fmt.Fprintf(buffer, "%v, ", item)
				}
			}
		}
	}
}

// Prepares output text and sends to appropriate logging destinations.
func write2log(flag int, vars ...interface{}) {

	if atomic.LoadInt32(&fatal_triggered) == 1 {
		if flag&_bypass_lock == _bypass_lock {
			flag ^= _bypass_lock
		} else {
			return
		}
	}

	flag = flag &^ _bypass_lock

	mutex.Lock()
	defer mutex.Unlock()

	logger := l_map[flag&^_no_logging]

	var pre []byte

	if flag&_no_logging != _no_logging {
		if logger.use_ts {
			genTS(&pre)
		}
		pre = append(pre, []byte(prefix[flag])[0:]...)
	}

	// Reset buffer.
	msgBuffer.Reset()

	outputFactory(&msgBuffer, vars...)

	output := msgBuffer.Bytes()
	msg := msgBuffer.String()
	output = append(pre, output[0:]...)
	bufferLen := utf8.RuneCount(output)

	if bufferLen > 0 {
		if output[len(output)-1] != '\n' && flag&_flash_txt != _flash_txt {
			output = append(output, '\n')
			bufferLen++
		}
	} else if flag&_flash_txt != _flash_txt {
		output = append(output, '\n')
		bufferLen++
	}

	// Clear out last flash text.
	if flush_needed && !piped_stderr && ((logger.out1 == os.Stdout && !piped_stdout) || logger.out1 == os.Stderr) {
		if bufferLen == 0 {
			fmt.Fprintf(os.Stderr, "\r%s  \r", string(flush_line[0:flush_len]))
		} else {
			fmt.Fprintf(os.Stderr, "\r%s\r", string(flush_line[0:flush_len]))
		}
		flush_needed = false
	}

	// Flash text handler, make a line of text available to remove remnents of this text.
	if flag&_flash_txt == _flash_txt {
		if !piped_stderr {
			for i := len(flush_line); i < bufferLen; i++ {
				flush_line = append(flush_line[0:], ' ')
			}
			flush_len = bufferLen
			io.Copy(os.Stderr, bytes.NewReader(output))
			flush_needed = true
			return
		}
		return
	}

	if flag&_no_logging == _no_logging {
		io.Copy(logger.out1, bytes.NewReader(output))
		return
	}

	var err error

	_, err = io.Copy(logger.out1, bytes.NewReader(output))
	if err != nil && FatalOnOutError {
		go Fatal(err)
		return
	}

	// Preprend timestamp for file.
	if !logger.use_ts {
		out_len := len(output)
		genTS(&output)
		out := output[out_len:]
		out = append(out, output[0:out_len]...)
		output = out
	}

	// Write to file.
	_, err = io.Copy(logger.out2, bytes.NewReader(output))
	// Launch fatal in a go routine, as the mutex is currently locked.
	if err != nil && FatalOnFileError {
		go Fatal(err)
	}

	if export_syslog != nil && enabled_exports&flag == flag {
		switch flag {
		case INFO:
			fallthrough
		case AUX:
			fallthrough
		case AUX2:
			fallthrough
		case AUX3:
			fallthrough
		case AUX4:
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
