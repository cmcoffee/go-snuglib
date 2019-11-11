# iotimeout
--
    import "github.com/cmcoffee/go-iotimeout"


## Usage

```go
var ErrReadTimeout = errors.New("IO timeout exceeded waiting for bytes.")
```

#### type ReadCloser

```go
type ReadCloser struct {
	*Reader
}
```

Timeout ReadCloser

#### func  NewReadCloser

```go
func NewReadCloser(readcloser io.ReadCloser, timeout time.Duration) *ReadCloser
```
Timeout ReadCloser: Adds a timer to io.ReadCloser

#### func (*ReadCloser) Close

```go
func (t *ReadCloser) Close() (err error)
```
Close function for ReadCloser.

#### type Reader

```go
type Reader struct {
}
```

Timeout Reader.

#### func  NewReader

```go
func NewReader(reader io.Reader, timeout time.Duration) *Reader
```
Timeout ReadCloser: Adds a timer to io.Reader

#### func (*Reader) Read

```go
func (t *Reader) Read(p []byte) (n int, err error)
```
Time Sensitive Read function.
