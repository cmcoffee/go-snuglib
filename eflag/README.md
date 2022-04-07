# eflag
--
    import "github.com/cmcoffee/go-snuglib/eflag"

Package 'eflag' is a wrapper around Go's standard flag, it provides enhancments
for: Adding Header's and Footer's to Usage. Adding Aliases to flags. (ie.. -d,
--debug) Enhances formatting for flag usage. Aside from that everything else is
standard from the flag library.

## Usage

```go
var (
	CLIArgs       = cmd.CLIArgs
	SyntaxName    = cmd.SyntaxName
	SetOutput     = cmd.SetOutput
	PrintDefaults = cmd.PrintDefaults
	Shorten       = cmd.Shorten
	String        = cmd.String
	StringVar     = cmd.StringVar
	Arg           = cmd.Arg
	Args          = cmd.Args
	Bool          = cmd.Bool
	BoolVar       = cmd.BoolVar
	Duration      = cmd.Duration
	DurationVar   = cmd.DurationVar
	Float64       = cmd.Float64
	Float64Var    = cmd.Float64Var
	Int           = cmd.Int
	IntVar        = cmd.IntVar
	Int64         = cmd.Int64
	Int64Var      = cmd.Int64Var
	Lookup        = cmd.Lookup
	Multi         = cmd.Multi
	MultiVar      = cmd.MultiVar
	NArg          = cmd.NArg
	NFlag         = cmd.NFlag
	Name          = cmd.Name
	Output        = cmd.Output
	Parsed        = cmd.Parsed
	Uint          = cmd.Uint
	UintVar       = cmd.UintVar
	Uint64        = cmd.Uint64
	Uint64Var     = cmd.Uint64Var
	Var           = cmd.Var
	Visit         = cmd.Visit
	VisitAll      = cmd.VisitAll
)
```

```go
var ErrHelp = flag.ErrHelp
```

#### func  Footer

```go
func Footer(input string)
```
Sets the footer for usage info.

#### func  Header

```go
func Header(input string)
```
Sets the header for usage info.

#### func  Parse

```go
func Parse() (err error)
```
Parse flags

#### func  Usage

```go
func Usage()
```
Shows usage.

#### type EFlagSet

```go
type EFlagSet struct {
	Header     string // Header presented at start of help.
	Footer     string // Footer presented at end of help.
	AdaptArgs  bool   // Reorders flags and arguments so flags come first, non-flag arguments second, unescapes arguments with '\' escape character.
	ShowSyntax bool   // Display Usage: line, CLIArgs will automatically display usage info.

	*flag.FlagSet
}
```

A EFlagSet is a set of defined flags.

#### func  NewFlagSet

```go
func NewFlagSet(name string, errorHandling ErrorHandling) (output *EFlagSet)
```
Load a flag created with flag package.

#### func (*EFlagSet) Args

```go
func (s *EFlagSet) Args() []string
```
Returns extra arguments.

#### func (*EFlagSet) Bool

```go
func (E *EFlagSet) Bool(name string, usage string) *bool
```
Bool defines a bool flag with specified name, default and usage string. The
return value is the address of a bool variable that stores the value of the
flag.

#### func (*EFlagSet) BoolVar

```go
func (E *EFlagSet) BoolVar(p *bool, name string, usage string)
```
BoolVar defines a bool flag with specified name, and usage string. The argument
p points to a bool variable in which to store the value of the flag.

#### func (*EFlagSet) CLIArgs

```go
func (E *EFlagSet) CLIArgs(name ...string)
```
Maps CLI Args not set to flags, to flags in order of addition.

#### func (*EFlagSet) IsSet

```go
func (s *EFlagSet) IsSet(name string) bool
```

#### func (*EFlagSet) Multi

```go
func (E *EFlagSet) Multi(name string, value string, usage string) *[]string
```
Array variable, ie.. comma-seperated values --flag="test","test2"

#### func (*EFlagSet) MultiVar

```go
func (E *EFlagSet) MultiVar(p *[]string, name string, value string, usage string)
```
Array variable, ie.. comma-seperated values --flag="test","test2"

#### func (*EFlagSet) Order

```go
func (s *EFlagSet) Order(name ...string)
```
Specifies the order in which flags are displayed.

#### func (*EFlagSet) Parse

```go
func (s *EFlagSet) Parse(args []string) (err error)
```
Wraps around the standard flag Parse, adds header and footer.

#### func (*EFlagSet) PrintDefaults

```go
func (s *EFlagSet) PrintDefaults()
```
Reads through all flags available and outputs with better formatting.

#### func (*EFlagSet) ResolveAlias

```go
func (s *EFlagSet) ResolveAlias(name string) string
```
Resolves Alias name to fullname

#### func (*EFlagSet) SetOutput

```go
func (s *EFlagSet) SetOutput(output io.Writer)
```
Change where output will be directed.

#### func (*EFlagSet) Shorten

```go
func (s *EFlagSet) Shorten(name string, ch rune)
```
Adds a single charachter alias to the command, ie.. --help h

#### func (*EFlagSet) SyntaxName

```go
func (E *EFlagSet) SyntaxName(name string)
```
Specifies the name that will be shown for the usage/syntax.

#### func (*EFlagSet) VisitAll

```go
func (s *EFlagSet) VisitAll(fn func(*Flag))
```
Provide same order.

#### type ErrorHandling

```go
type ErrorHandling int
```

Duplicate flag's ErrorHandling.

```go
const (
	ContinueOnError ErrorHandling = iota
	ExitOnError
	PanicOnError
	ReturnErrorOnly
)
```

#### type Flag

```go
type Flag = flag.Flag
```
