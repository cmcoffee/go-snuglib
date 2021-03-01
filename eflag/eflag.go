// Package 'eflag' is a wrapper around Go's standard flag, it provides enhancments for:
// Adding Header's and Footer's to Usage.
// Adding Aliases to flags. (ie.. -d, --debug)
// Enhances formatting for flag usage.
// Aside from that everything else is standard from the flag library.
package eflag

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

// Duplicate flag's ErrorHandling.
type ErrorHandling int

const (
	ContinueOnError ErrorHandling = iota
	ExitOnError
	PanicOnError
	ReturnErrorOnly
)

var ErrHelp = flag.ErrHelp

type Flag = flag.Flag

// Write to nothing, to remove standard output of flag.
type _voidText struct{}

var voidText _voidText

func (s _voidText) Write(p []byte) (n int, err error) {
	return len(p), nil
}

type multiValue struct {
	value *[]string
}

// removes quotation marks from examples.
func remove_quotes(input string) string {
	if len(input) > 2 {
		if input[0] == '"' && input[len(input)-1] == '"' {
			return input[1 : len(input)-1]
		}
	}
	return input
}

func (A *multiValue) String() string {
	if *A.value != nil && len(*A.value) > 0 {
		return escape_array(*A.value)
	} else {
		return ""
	}
}

func (A *multiValue) Set(value string) error {
	*A.value = string_split(value)
	return nil
}

func string_split(input string) (output []string) {
	if len(input) == 0 {
		return
	}
	var escaped bool
	var temp []rune
	for _, c := range input {
		switch c {
		case '\\':
			if escaped {
				escaped = false
			} else {
				escaped = true
			}
		case ',':
			if !escaped {
				output = append(output, string(temp[0:]))
				temp = temp[0:0]
			} else {
				escaped = false
				temp = append(temp, c)
			}
		case '"':
			escaped = false
			temp = append(temp, c)
		default:
			if escaped {
				temp = append(temp, '\\')
			}
			temp = append(temp, c)
			escaped = false
		}
	}
	output = append(output, string(temp[0:]))
	return
}

func escape_array(input []string) string {
	var (
		temp   []rune
		output []string
	)

	for _, str := range input {
		for _, v := range str {
			switch v {
			case '"':
				temp = append(temp, '\\', '"')
			case ',':
				temp = append(temp, '\\', ',')
			default:
				temp = append(temp, v)
			}
		}
		output = append(output, fmt.Sprintf("\"%s\"", string(temp[0:])))
		temp = temp[0:0]
	}
	return strings.Join(output, ",")
}

func (A *multiValue) Get() interface{} { return []string(*A.value) }

// Array variable, ie.. comma-seperated values --flag="test","test2"
func (E *EFlagSet) Multi(name string, value string, usage string) *[]string {
	output := new([]string)
	E.MultiVar(output, name, value, usage)
	return output
}

// Array variable, ie.. comma-seperated values --flag="test","test2"
func (E *EFlagSet) MultiVar(p *[]string, name string, value string, usage string) {
	*p = string_split(value)

	v := multiValue{
		value: p,
	}

	if len(usage) > 0 {
		usage = fmt.Sprintf("%s (multi: comma-seperated)", usage)
	}
	E.Var(&v, name, usage)
}

// Specifies the name that will be shown for the usage/syntax.
func (E *EFlagSet) SyntaxName(name string) {
	E.syntaxName = name
}

// BoolVar defines a bool flag with specified name, and usage string. The argument p points to a bool variable in which to store the value of the flag.
func (E *EFlagSet) BoolVar(p *bool, name string, usage string) {
	E.FlagSet.BoolVar(p, name, *p, usage)
}

// Bool defines a bool flag with specified name, default and usage string. The return value is the address of a bool variable that stores the value of the flag.
func (E *EFlagSet) Bool(name string, usage string) *bool {
	return E.FlagSet.Bool(name, false, usage)
}

// Maps CLI Args not set to flags, to flags in order of addition.
func (E *EFlagSet) CLIArgs(name ...string) {
	fmap := make(map[string]*flag.Flag)

	E.VisitAll(func(input *Flag) {
		fmap[input.Name] = input
	})

	for _, v := range name {
		if flag, ok := fmap[v]; ok {
			E.argMap = append(E.argMap, flag)
		}
	}
}

// A EFlagSet is a set of defined flags. 
type EFlagSet struct {
	name          string
	Header        string // Header presented at start of help.
	Footer        string // Footer presented at end of help.
	AdaptArgs     bool   // Reorders flags and arguments so flags come first, non-flag arguments second, unescapes arguments with '\' escape character.
	ShowSyntax    bool   // Display Usage: line, CLIArgs will automatically display usage info.
	alias         map[string]string
	out           io.Writer
	errorHandling ErrorHandling
	setFlags      []string
	order         []string
	argMap        []*flag.Flag
	syntaxName    string
	*flag.FlagSet
}

var cmd = EFlagSet{
	os.Args[0],
	"",
	"",
	false,
	false,
	make(map[string]string),
	os.Stderr,
	ExitOnError,
	make([]string, 0),
	make([]string, 0),
	make([]*flag.Flag, 0),
	os.Args[0],
	flag.NewFlagSet(os.Args[0], flag.ContinueOnError),
}

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

// Sets the header for usage info.
func Header(input string) {
	cmd.Header = input
}

// Sets the footer for usage info.
func Footer(input string) {
	cmd.Footer = input
}

// Parse flags
func Parse() (err error) {
	if len(os.Args) > 1 {
		return cmd.Parse(os.Args[1:])
	} else {
		return cmd.Parse([]string{})
	}
}

// Shows usage.
func Usage() {
	//if !cmd.Parsed() {
		cmd.Parse([]string{"--help"})
	//} 
}

// Specifies the order in which flags are displayed.
func (s *EFlagSet) Order(name ...string) {
	if name != nil {
		s.order = name[0:]
	}
}

// Returns extra arguments.
func (s *EFlagSet) Args() []string {
	args := s.FlagSet.Args()
	if s.AdaptArgs {
		for i, v := range args {
			if strings.HasPrefix(v, "\\-") {
				args[i] = strings.TrimPrefix(v, "\\")
			}
		}
	}
	return args
}

// Change where output will be directed.
func (s *EFlagSet) SetOutput(output io.Writer) {
	s.out = output
}

// Load a flag created with flag package.
func NewFlagSet(name string, errorHandling ErrorHandling) (output *EFlagSet) {
	output = &EFlagSet{
		name,
		"",
		"",
		false,
		false,
		make(map[string]string),
		os.Stderr,
		errorHandling,
		make([]string, 0),
		make([]string, 0),
		make([]*flag.Flag, 0),
		name,
		flag.NewFlagSet(name, flag.ContinueOnError),
	}
	output.Usage = func() {
		output.Parse([]string{"--help"})
	}
	return output
}

// Reads through all flags available and outputs with better formatting.
func (s *EFlagSet) PrintDefaults() {

	output := tabwriter.NewWriter(s.out, 1, 1, 3, ' ', 0)

	flag_text := make(map[string]string)
	var flag_order []string
	var alias_order []string

	argMap := make(map[string]struct{})
	for _, v := range s.argMap {
		argMap[v.Name] = struct{}{}
	}

	s.VisitAll(func(flag *flag.Flag) {
		if flag.Usage == "" {
			return
		}
		if _, ok := argMap[flag.Name]; ok {
			return
		}
		var text []string
		name := flag.Name
		alias := s.alias[flag.Name]
		if alias != "" {
			if len(alias) > 1 {
				text = append(text, fmt.Sprintf("  --%s,", alias))
			} else {
				text = append(text, fmt.Sprintf("  -%s,", alias))
			}
		}
		space := " "
		if alias == "" {
			space = "  "
		}
		if len(name) > 1 {
			text = append(text, fmt.Sprintf("%s--%s", space, name))
		} else {
			text = append(text, fmt.Sprintf("%s-%s", space, name))
		}

		switch flag.DefValue[0] {
		case '"':
			if strings.HasPrefix(flag.DefValue, "\"<") && strings.HasSuffix(flag.DefValue, ">\"") {
				text = append(text, fmt.Sprintf("=%q", flag.DefValue[2:len(flag.DefValue)-2]))
			} else {
				text = append(text, fmt.Sprintf("=%s", flag.DefValue))
			}
		case '<':
			if flag.DefValue[len(flag.DefValue)-1] == '>' {
				text = append(text, fmt.Sprintf("=%q", flag.DefValue[1:len(flag.DefValue)-1]))
			} else {
				text = append(text, fmt.Sprintf("=%s", flag.DefValue))
			}
		default:
			if flag.DefValue != "true" && flag.DefValue != "false" {
				text = append(text, fmt.Sprintf("=%s", flag.DefValue))
			}
		}

		text = append(text, fmt.Sprintf("\t%s\n", flag.Usage))

		if alias == "" {
			flag_text[name] = strings.Join(text[0:], "")
			flag_order = append(flag_order, name)
		} else {
			flag_text[name] = strings.Join(text[0:], "")
			alias_order = append(alias_order, name)
		}
	})

	// Place Aliases first.
	flag_order = append(alias_order, flag_order[0:]...)

	//OutterLoop:
	for _, v := range flag_order {
		for _, o := range s.order {
			if o == v {
				for _, name := range s.order {
					txt, ok := flag_text[name]
					if ok {
						fmt.Fprintf(output, txt)
						delete(flag_text, name)
					}
				}
			}
		}
	}
	for _, v := range flag_order {
		if txt, ok := flag_text[v]; ok {
			fmt.Fprintf(output, txt)
		}
	}
	fmt.Fprintf(output, "  --help\tDisplays this usage information.\n")
	output.Flush()
}

// Adds a single charachter alias to the command, ie.. --help h
func (s *EFlagSet) Shorten(name string, ch rune) {
	flag := s.Lookup(name)
	if flag == nil {
		return
	}
	s.Var(flag.Value, string(ch), "")
	s.alias[name] = string(ch)

	// Create reverse lookup
	s.alias[fmt.Sprintf("-%s-", string(ch))] = name
}

// Resolves Alias name to fullname
func (s *EFlagSet) ResolveAlias(name string) string {
	if v, ok := s.alias[fmt.Sprintf("-%s-", name)]; ok {
		return v
	} else {
		return name
	}
}

func (s *EFlagSet) IsSet(name string) bool {
	for _, k := range s.setFlags {
		if k == name {
			return true
		}
	}
	return false
}

// Wraps around the standard flag Parse, adds header and footer.
func (s *EFlagSet) Parse(args []string) (err error) {
	// set usage to empty to prevent unessisary work as we dump the output of flag.
	s.Usage = func() {}

	var (
		tmp      []string
		trailing []string
	)

	// Split bool flags so that '-abc' becomes '-a -b -c' before being parsed.
	for _, a := range args {
		if !strings.HasPrefix(a, "-") {
			if !s.AdaptArgs {
				tmp = append(tmp, a)
			} else {
				trailing = append(trailing, a)
			}
			continue
		}
		if strings.HasPrefix(a, "--") {
			tmp = append(tmp, a)
			continue
		}
		if strings.Contains(a, "=") {
			tmp = append(tmp, a)
			continue
		}
		a = strings.TrimPrefix(a, "-")
		if len(a) == 0 {
			continue

		}
		tmp = append(tmp, fmt.Sprintf("-%c", a[0]))
		for _, ch := range a[1:] {
			tmp = append(tmp, fmt.Sprintf("-%c", ch))
		}
	}

	args = tmp[0:]
	if s.AdaptArgs {
		args = append(args, trailing[0:]...)
	}

	// Remove normal error message printing.
	s.FlagSet.SetOutput(voidText)

	// Harvest error message, conceal flag.Parse() output, then reconstruct error message.
	stdOut := s.out
	s.out = voidText

	err = s.FlagSet.Parse(args)
	s.out = stdOut

	val_map := make(map[string]*flag.Value)

	// Remove example text from strings, ie.. <server to connect with>
	clear_examples := func(f *flag.Flag) {
		val := f.Value.String()
		if (strings.HasPrefix(val, "<") || strings.HasPrefix(val, "\"<")) && (strings.HasSuffix(val, ">") || strings.HasSuffix(val, ">\"")) {
			f.Value.Set("")
			val_map[f.Name] = &f.Value
		}
	}

	s.FlagSet.VisitAll(clear_examples)

	mark_set_flags := func(f *flag.Flag) {
		s.setFlags = append(s.setFlags, f.Name)
	}

	num := 0
	txt_args := s.FlagSet.Args()
	multi_set := false

	for i, f := range s.argMap {
		if val, ok := val_map[f.Name]; ok {
			v := *val
			if v.String() == "" {
				if _, ok := v.(*multiValue); ok && !multi_set {
					multi_set = true
					txt_len := len(txt_args)
					// First Argument
					if i == 0 {
						if txt_len == 1 {
							v.Set(txt_args[0])
							num++
						} else if txt_len > 1 {
							if e := txt_len - (len(s.argMap) - 1); e > 0 {
								v.Set(strings.Join(txt_args[0:e], ","))
								num = e
							} else {
								v.Set(txt_args[num])
								num++
							}
						}
					// Last Argument
					} else if i == len(s.argMap)-1 {
						v.Set(strings.Join(txt_args[num:], ","))
						num = txt_len - 1
					// Somwhere in the middle.
					} else {
						if x := txt_len - num; x > 1 {
							v.Set(strings.Join(txt_args[num:txt_len-1], ","))
							num = txt_len - 1
						} else if x > 0 {
							v.Set(txt_args[txt_len-1])
							num++
						}
					}
				} else if str := s.FlagSet.Arg(num); str != "" {
					v.Set(str)
					num++
				}
			}
			if v.String() != "" {
				mark_set_flags(f)
			}
		}
	}

	s.FlagSet.Visit(mark_set_flags)

	// Implement new Usage function.
	s.Usage = func() {
		if s.Header != "" {
			fmt.Fprintf(s.out, "%s\n", s.Header)
		}
		var (
			arg_names []string
			has_multi bool
		)

		for _, v := range s.argMap {
			if val, ok := val_map[v.Name]; ok {
				if _, ok := (*(val)).(*multiValue); ok && !has_multi {
					def := remove_quotes(v.DefValue)
					has_multi = true
					arg_names = append(arg_names, fmt.Sprintf("%s...", def))
				} else {
					arg_names = append(arg_names, remove_quotes(v.DefValue))
				}
			}
		}
		if s.name == "" {
			fmt.Fprintf(s.out, "Options:\n")
		} else {
			if len(arg_names) > 0 {
				fmt.Fprintf(s.out, "Usage: %s [options] %s\n\n", s.syntaxName, strings.Join(arg_names, " "))
			} else if s.ShowSyntax {
				fmt.Fprintf(s.out, "Usage: %s [options]\n\n", s.syntaxName)
			}
			fmt.Fprintf(s.out, "Available '%s' options:\n", s.name)
		}
		s.PrintDefaults()
		if s.Footer != "" {
			fmt.Fprintf(s.out, "%s\n", s.Footer)
		}
	}

	// Implement a new error message.
	if err != nil {
		if err != flag.ErrHelp {
			errStr := err.Error()
			cmd := strings.Split(errStr, "-")
			if len(cmd) > 1 {
				for _, arg := range args {
					if strings.Contains(arg, cmd[1]) {
						err = fmt.Errorf("%s%s", cmd[0], arg)
						if s.errorHandling != ReturnErrorOnly {
							fmt.Fprintf(s.out, "%s\n\n", errStr)
						}
						break
					}
				}
			} else {
				if s.errorHandling != ReturnErrorOnly {
					fmt.Fprintf(s.out, "%s\n\n", errStr)
				}
			}
		}

		// Errorflag handling.
		switch s.errorHandling {
		case ReturnErrorOnly:
		case ContinueOnError:
			s.Usage()
		case ExitOnError:
			s.Usage()
			os.Exit(2)
		case PanicOnError:
			panic(err)
		}
	}
	return
}
