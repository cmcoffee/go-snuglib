# swapreader
--
    import "github.com/cmcoffee/go-swapreader"


## Usage

#### type Reader

```go
type Reader struct {
}
```

Swap Reader allows for swapping the io.Reader backed []bytes

#### func (*Reader) Read

```go
func (r *Reader) Read(p []byte) (n int, err error)
```
swap_reader Read function.

#### func (*Reader) SetBytes

```go
func (r *Reader) SetBytes(in []byte)
```
Set []byte for reader

#### func (*Reader) SetReader

```go
func (r *Reader) SetReader(in io.Reader)
```
Set Reader to Reader
