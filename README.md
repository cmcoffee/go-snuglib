
# cfg
    import "github.com/cmcoffee/go-cfg"

Package 'cfg' provides functions for reading and writing configuration files and their coresponding string values.


	Ignores '#' as comment lines, ','s denote multiple values.
	
	# Example config file.
	[section]
	key = value
	key2 = value1, value2
	key3 = value1,
	       value2,
	       value3
	
	[section2]
	key = value1,
	      value2,
	      value3






## func ListSections
``` go
func ListSections(file string) (out []string, err error)
```
Lists all Sections in config file.


## func ReadFile
``` go
func ReadFile(file, section string) (out map[string][]string, err error)
```
Returns map of specific [section] within configuration file.


## func SetFile
``` go
func SetFile(file, section, key string, value ...string) error
```
Writes key = values under [section] to File.



## type Store
``` go
type Store struct {
    // contains filtered or unexported fields
}
```








### func Create
``` go
func Create(file string, comment ...string) (out *Store, err error)
```
Creates a new empty config file & Store, overwriting an existing file with comments if specified.


### func Load
``` go
func Load(file string) (out *Store, err error)
```
Reads configuration file and returns Store.


<<<<<<< HEAD
### func ReadOnly
``` go
func ReadOnly(file string) (out *Store, err error)
```
Reads configuration file and returns Store, changes are not saved to disk.


=======
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600


### func (\*Store) Exists
``` go
func (s *Store) Exists(input ...string) (found bool)
```
Returns true if section or section and key exists.



### func (\*Store) Get
``` go
func (s *Store) Get(section, key string) []string
```
Returns array of all retrieved string values under section with key.



### func (\*Store) ListKeys
``` go
func (s *Store) ListKeys(section string) (out []string)
```
Returns keys of section specified.



### func (\*Store) ListSections
``` go
func (s *Store) ListSections() (out []string)
```
Returns array of all sections in config file.



### func (\*Store) Set
``` go
func (s *Store) Set(section, key string, value ...string) (err error)
```
Sets key = values under [section], updates Store and saves to file.









- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)