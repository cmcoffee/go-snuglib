# cfg
--
    import "github.com/cmcoffee/go-cfg"

Package 'cfg' provides functions for reading and writing configuration files and
their coresponding string values. Ignores '#' as comment lines, ','s denote
multiple values.

## Usage

#### func  ReadFile

```go
func ReadFile(file, section string) (out map[string][]string, err error)
```
Returns map of specific [section] within configuration file.

#### func  SetFile

```go
func SetFile(file, header, key string, value ...string) error
```
Writes key = values under [header] to File.

#### type Store

```go
type Store struct {
}
```


#### func  Create

```go
func Create(file string, comment ...string) (out *Store, err error)
```
Creates a new empty config file & Store, overwriting an existing file with
comments if specified.

#### func  Load

```go
func Load(file string) (out *Store, err error)
```
Reads configuration file and returns Store.

#### func (*Store) Get

```go
func (s *Store) Get(header, key string) (*Value, bool)
```
Returns array of all retrieved string values under header with key.

#### func (*Store) Set

```go
func (s *Store) Set(header, key string, value ...string) (err error)
```
Sets key = values under [header], updates Store and saves to file.

#### type Value

```go
type Value struct {
}
```


#### func (*Value) Next

```go
func (v *Value) Next() bool
```
Go to next value if available.

#### func (*Value) String

```go
func (v *Value) String() string
```
Output string value.
