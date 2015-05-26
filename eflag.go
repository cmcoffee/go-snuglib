// Package 'eflag' is a wrapper around Go's standard flag, it provides enhancments for:
// Adding Header's and Footer's to Usage.
// Adding Aliases to flags. (ie.. -d, --debug)
// Enhances formatting for flag usage.
// Aside from that everything else is standard from the flag library.
package eflag

import (
	"flag"
	"fmt"
	"os"
	"io"
	"time"
	"text/tabwriter"
	"strings"
)

// Duplicate flag's ErrorHandling.
type ErrorHandling int

const (
   		ContinueOnError ErrorHandling = iota
		ExitOnError
		PanicOnError
)

// Write to nothing, to remove standard output of flag.
type _voidText struct{}
var voidText _voidText 
func (self _voidText) Write(p []byte) (n int, err error) {
		return len(p), nil
}

type EFlagSet struct {
	name string
	Header string
	Footer string
	alias map[string]string
	stringVars map[string]bool
	*flag.FlagSet
	out io.Writer
	errorHandling ErrorHandling
}

// Allows for quick command line argument parsing, flag's default usage.
func CommandLine() *EFlagSet {
	return NewFlagSet(os.Args[0], ExitOnError)
}

// Change where output will be directed.
func (self *EFlagSet) SetOutput(output io.Writer) {
	self.out = output
}

// Load a flag created with flag package.
func NewFlagSet(name string, errorHandling ErrorHandling) *EFlagSet {
	return &EFlagSet{
		name,
		"",
		"",
		make(map[string]string),
		make(map[string]bool),
		flag.NewFlagSet(name, flag.ContinueOnError),
		os.Stderr,
		errorHandling,
	}
}

// Reads through all flags available and outputs with better formatting.
func (self *EFlagSet) PrintDefaults() {
	
	output := tabwriter.NewWriter(self.out, 34, 8, 1, ' ', 0)
	
	fmt.Fprintf(output, "  -h, --help\tPrints usage\n")
		
	self.VisitAll(func(flag *flag.Flag) {
		if flag.Usage == "" { return }
		var text []string
		name := flag.Name
		alias := self.alias[flag.Name]
		is_string := self.stringVars[flag.Name]
		if len(name) > 1 {
			text = append(text, fmt.Sprintf("  --%s", name))
		} else {
			text = append(text, fmt.Sprintf("  -%s", name))
		}
		if alias != "" {
			text = append(text, fmt.Sprintf(", --%s", alias))
		}
		if is_string == true {
			text = append(text, fmt.Sprintf("=%q", flag.DefValue))
		} else {
			if flag.DefValue != "true" && flag.DefValue != "false" {
				text = append(text, fmt.Sprintf("=%s", flag.DefValue))
			}
		}
		text = append(text, fmt.Sprintf("\t%s\n", flag.Usage))
		
		fmt.Fprintf(output, strings.Join(text[0:], ""))
		output.Flush()
	})
}

// Adds an alias to an existing flag, requires a pointer to the variable, the current name and the new alias name.
func (self *EFlagSet) Alias(val interface{}, name string, alias string) {
	flag := self.Lookup(name)
	if flag == nil { return }
	switch v := val.(type) {
		case *bool:
			self.BoolVar(v, alias, *v, "")
		case *time.Duration:
			self.DurationVar(v, alias, *v, "")
		case *float64:
			self.Float64Var(v, alias, *v, "")
		case *int:
			self.IntVar(v, alias, *v, "")
		case *int64:
			self.Int64Var(v, alias, *v, "")
		case *string:
			self.StringVar(v, alias, *v, "")
		case *uint:
			self.UintVar(v, alias, *v, "")
		case *uint64:
			self.Uint64Var(v, alias, *v, "")
		default:
			self.Var(flag.Value, alias, "")
			
	}
	self.alias[name] = alias
}

// Wraps around the standard flag Parse, adds header and footer.
func (self *EFlagSet) Parse(args...string) (err error) {
	if len(args) == 0 {
		args = os.Args[1:]
	}
	
	// set usage to empty to prevent unessisary work as we dump the output of flag.
	self.Usage = func() {}

	// Allows for multiple switches when single '-' is used.
	for n, arg := range args {
		if !strings.HasPrefix(arg, "-") { break } // Is not a modifier, but a command.
		if strings.HasPrefix(arg, "--") { continue }
		if self.Lookup(arg) != nil { continue } 
		if _, ok := self.alias[arg]; ok { continue }
		arg = strings.TrimLeft(arg, "-")
		var newArgs []string
		
		// Break multiple flags into individual flags.
		for _, ch := range arg {
			if ch == '=' {
				newArgs[len(newArgs)-1] = fmt.Sprintf("%s=%s", newArgs[len(newArgs) - 1], strings.Split(arg, "=")[1])
				break
			}
			newArgs = append(newArgs, fmt.Sprintf("-%c", ch))
		}
		copy(args[n:], args[n+1:])
		args = args[:len(args)-1]
		args = append(newArgs, args[0:]...)   
	}
	
	// Remove normal error message printing.
	self.FlagSet.SetOutput(voidText)
	
	// Harvest error message, conceal flag.Parse() output, then reconstruct error message.
	stdOut := self.out
	self.out = voidText
	err = self.FlagSet.Parse(args)
	self.out = stdOut

	// Implement new Usage function.
	self.Usage = func() {
		if self.Header != "" {
			fmt.Fprintf(self.out, "%s\n\n", self.Header)
		} else {
			if self.name == "" {
				fmt.Fprintf(self.out, "Available modifiers:\n")
			} else {
				if self.name == os.Args[0] {
					fmt.Fprintf(self.out, "Available %s modifiers:\n", os.Args[0])
				} else {
					fmt.Fprintf(self.out, "Available '%s %s' modifiers:\n", os.Args[0],self.name)
				}
			}
		}
		self.PrintDefaults()
		if self.Footer != "" { fmt.Fprintf(self.out, "%s\n", self.Footer) }
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
						fmt.Fprintf(self.out, "%s\n\n", err.Error())
						break
					}
				}
			} else {
				fmt.Fprintf(self.out, "%s\n\n", errStr)
			}
		}
		
		// Errorflag handling.
		switch self.errorHandling {
			case ContinueOnError:
				self.Usage()
			case ExitOnError:
				self.Usage()
				os.Exit(2)
			case PanicOnError:
				panic(err)
		}
	}	
	return
}

func (self *EFlagSet) String(name string, value string, usage string) *string {
	self.stringVars[name] = true
	return self.FlagSet.String(name, value, usage)
}

func (self *EFlagSet) StringVar(p *string, name string, value string, usage string) {
	self.stringVars[name] = true
	self.FlagSet.StringVar(p, name, value, usage)
}
