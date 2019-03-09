# nfo
--
    import "github.com/cmcoffee/go-nfo"


## Usage

```go
const (
	INFO   = 1 << iota // Log Information
	AUX                // Auxilary Log
	ERROR              // Log Errors
	WARN               // Log Warning
	NOTICE             // Log Notices
	DEBUG              // Debug Logging
	TRACE              // Trace Logging
	FATAL              // Fatal Logging

)
```

```go
const STD = INFO | AUX | ERROR | WARN | NOTICE | FATAL
```
Standard Loggers, minus debug and trace.

```go
var (
	FatalOnOutError    = true // Fatal on Output logging error.
	FatalOnFileError   = true // Fatal on log file or file rotation errors.
	FatalOnExportError = true // Fatal on export/syslog error.

)
```

```go
var None dummyWriter
```
False writer for discarding output.

#### func  Aux

```go
func Aux(vars ...interface{})
```
Log as Info, as auxilary output.

#### func  BlockShutdown

```go
func BlockShutdown()
```
Global wait group, allows running processes to finish up tasks before app
shutdown

#### func  Close

```go
func Close(filename string) (err error)
```
Closes logging file, removes file from all loggers, removes file from open
files.

#### func  Debug

```go
func Debug(vars ...interface{})
```
Log as Debug.

#### func  Defer

```go
func Defer(closer interface{})
```
Adds a function to the global defer, function must take no arguments and either
return nothing or return an error.

#### func  Err

```go
func Err(vars ...interface{})
```
Log as Error.

#### func  Exit

```go
func Exit(exit_code int)
```
Intended to be a defer statement at the begining of main, but can be called at
anytime with an exit code. Tries to catch a panic if possible and log it as a
fatal error, then proceeds to send a signal to the global defer/shutdown handler

#### func  Fatal

```go
func Fatal(vars ...interface{})
```
Log as Fatal, then quit.

#### func  File

```go
func File(l_file_flag int, filename string, max_size_mb uint, max_rotation uint) (err error)
```
Opens a new log file for writing, max_size is threshold for rotation,
max_rotation is number of previous logs to hold on to. Set max_size_mb to 0 to
disable file rotation.

#### func  Flash

```go
func Flash(vars ...interface{})
```
Don't log, write text to standard error which will be overwritten on the next
output.

#### func  HideTS

```go
func HideTS()
```
Hide timestamps.

#### func  HookSyslog

```go
func HookSyslog(syslog_writer SyslogWriter)
```
Send messages to syslog

#### func  LTZ

```go
func LTZ()
```
Switches timestamps to local timezone. (Default Setting)

#### func  Log

```go
func Log(vars ...interface{})
```
Log as Info.

#### func  Notice

```go
func Notice(vars ...interface{})
```
Log as Notice.

#### func  SetExports

```go
func SetExports(flag int)
```
Specify which logs to send to syslog.

#### func  SetLoggers

```go
func SetLoggers(flag int)
```
Specify which loggers to enable, STD is enabled by default.

#### func  SetOutput

```go
func SetOutput(logger int, w io.Writer)
```
Change output for logger(s).

#### func  SetPrefix

```go
func SetPrefix(logger int, prefix_str string)
```
Change prefix for specified logger.

#### func  SetSignals

```go
func SetSignals(sig ...os.Signal)
```
Sets the signals that we listen for.

#### func  ShowTS

```go
func ShowTS()
```
Show timestamps. (Default Enabled)

#### func  SignalCallback

```go
func SignalCallback(signal os.Signal, callback func() (continue_shutdown bool))
```
Set a callback function(no arguments) to run after receiving a specific syscall,
function returns true to continue shutdown process.

#### func  Stderr

```go
func Stderr(vars ...interface{})
```
Don't log, just print text to standard error.

#### func  Stdout

```go
func Stdout(vars ...interface{})
```
Don't log, just print text to standard out.

#### func  Trace

```go
func Trace(vars ...interface{})
```
Log as Trace.

#### func  UTC

```go
func UTC()
```
Switches logger to use UTC instead of local timezone.

#### func  UnblockShutdown

```go
func UnblockShutdown()
```
Task completed, carry on with shutdown.

#### func  UnhookSyslog

```go
func UnhookSyslog()
```
Disconnect form syslog

#### func  Warn

```go
func Warn(vars ...interface{})
```
Log as Warn.

#### type SyslogWriter

```go
type SyslogWriter interface {
	Alert(string) error
	Crit(string) error
	Debug(string) error
	Emerg(string) error
	Err(string) error
	Info(string) error
	Notice(string) error
	Warning(string) error
}
```

Interface for log/syslog/Writer.
