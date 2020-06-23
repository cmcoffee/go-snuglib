# options
--
    import "github.com/cmcoffee/go-snuglib/options"


## Usage

#### type Options

```go
type Options interface {
	Register(input Value)
	Select(seperate_last bool) (changed bool)
	String(desc string, default_value string, help string, mask_value bool) *string
	StringVar(p *string, desc string, value string, help string, mask_value bool)
	Bool(desc string, value bool) *bool
	BoolVar(p *bool, desc string, value bool)
	Int(desc string, value int, help string, min int, max int) *int
	IntVar(p *int, desc string, value int, help string, min int, max int)
	Options(desc string, value Options, seperate_last bool)
	Func(desc string, value func() bool)
}
```

Options Menu Interface

#### func  NewOptions

```go
func NewOptions(header, footer string, exit_char rune) Options
```
Creates new Options Menu

#### type Value

```go
type Value interface {
	Set() bool
	Get() interface{}
	String() string
}
```

Options Value
