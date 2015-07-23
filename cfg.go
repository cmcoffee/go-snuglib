// Package 'cfg' provides functions for reading and writing configuration files and their coresponding string values.
// Ignores '#' as comment lines, ','s denote multiple values.
package cfg

import (
	"bufio"
	"fmt"
	"os"
	"bytes"
	"strings"
	"io/ioutil"
	"sync"
)

type Store struct {
	file string
	mutex *sync.RWMutex
	cfgStore map[string]map[string][]string
}

const (
	cfg_HEADER = 1 << iota
	cfg_KEY
	cfg_COMMA
)

type Value struct {
	num uint
	val []string
}

// Output string value.
func (v *Value) String() string {
	if len(v.val) == 0 { return "" }
	return v.val[v.num]
}

// Go to next value if available.
func (v *Value) Next() bool {
	if len(v.val) < int(v.num) + 2 { return false }
	v.num++
	return true
}

// Returns array of all retrieved string values under header with key.
func (s *Store) Get(header, key string) (*Value, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	header = strings.ToLower(header)
	key = strings.ToLower(key)
	result, found := s.cfgStore[header][key]
	return &Value {
		0,
		result,
	}, found
}

// Sets key = values under [header], updates Store and saves to file.
func (s *Store) Set(header, key string, value...string) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	header = strings.ToLower(header)
	key = strings.ToLower(key)
	var newValue []string
	for _, val := range value { newValue = append(newValue, val) }

	if err := SetFile(s.file, header, key, newValue[0:]...); err != nil { return err }

	// Create new map if one doesn't exist.
	if _, ok := s.cfgStore[header]; !ok {
		s.cfgStore[header] = make(map[string][]string)
	}

	s.cfgStore[header][key] = newValue
	return
}

func setKey(buf *bytes.Buffer) (key string) {
	key = strings.ToLower(strings.TrimSpace(buf.String()))
	buf.Reset()
	return
}

func addVal(buf *bytes.Buffer, val *[]string) {
	*val = append(*val, strings.TrimSpace(buf.String()))
	buf.Reset()
}

func cfgErr(file string, line int) error { return fmt.Errorf("Syntax error found in %s on line %d.", file, line) }

// Creates a new empty config file & Store, overwriting an existing file with comments if specified.
func Create(file string, comment...string) (out *Store, err error) {
	f, err := os.Create(file)
	if err != nil { return nil, err }
	defer f.Close()
	out = &Store{
		file,
		new(sync.RWMutex),
		make(map[string]map[string][]string),
	}
	if len(comment) > 0 {
		for _, c := range comment {
			f.WriteString("# " + c + "\n");
		}
	}
	return
}

// Reads configuration file and returns Store.
func Load(file string) (out *Store, err error) {
	f, err := os.Open(file)
	if err != nil { return nil, err }
	defer f.Close()
	s := bufio.NewScanner(f)
	
	var flag, line, last int
	
	buf := &bytes.Buffer{}
	var header, key string
	var val []string
	out = &Store{
		file,
		new(sync.RWMutex),
		make(map[string]map[string][]string),
	}
	
	scanLoop:
	for s.Scan() {
		line++
		txt := s.Text() + "\n"
		l := len(txt)
		if l < 2 { continue }
		
		for i, ch := range txt {
			switch ch {
			case '[':
				if (flag & cfg_KEY != 0) { return out, cfgErr(file, last) }
				last = line
				if l > 2 && strings.ContainsAny(txt, "[ & ]") {
					header = txt[1:l-2]
					flag |= cfg_HEADER
					header = strings.ToLower(header)
					out.cfgStore[header] = make(map[string][]string)
					continue scanLoop
				} else { return out, cfgErr(file, line) }
			case '#': 
				continue scanLoop;
			case '=':
				if (flag & cfg_KEY != 0) { return out, cfgErr(file, line) }
				flag |= cfg_KEY
				key = setKey(buf)
				last = line
			case ',':
				if (flag & cfg_KEY == 0) { return out, cfgErr(file, line) }
				addVal(buf, &val)
				last = line
				flag |= cfg_COMMA
			case '\n':
				if (flag & cfg_KEY == 0) { return out, cfgErr(file, line) }
				if (flag & cfg_COMMA != 0) {
					flag &^= cfg_COMMA
					continue
				}
				flag &^= cfg_HEADER
				flag &^= cfg_KEY
				addVal(buf, &val)
				out.cfgStore[header][key] = val
				val = nil
				last = line
				continue scanLoop
			default:
				if buf.Len() == 0 {
					switch ch {
						case ' ':
							fallthrough
						case '\t':
							continue
					}
				}
				flag &^= cfg_COMMA
				if i == l - 1 && buf.Len() != 0 {
					return out, cfgErr(file, line)
				}
				buf.WriteRune(ch)
			}
				
		}
	}
	if (flag & cfg_KEY != 0) {
		return out, cfgErr(file, last)
	}
	return out, nil	
}

// Returns map of specific [section] within configuration file.
func ReadFile(file, section string) (out map[string][]string, err error) {
	section = strings.ToLower(section)
	f, err := os.Open(file)
	if err != nil { return nil, err }
	defer f.Close()
	s := bufio.NewScanner(f)
	
	var flag, line, last int
	
	buf := &bytes.Buffer{}
	var key string
	var val []string
	out = make(map[string][]string)
		
	for s.Scan() {
		line++
		txt := s.Text() + "\n"
		l := len(txt)
		
		if l > 1 && txt[0] == '#' || l == 1 { continue }
		
		if (flag & cfg_HEADER == 0) {
	
			// Skip to section headers only.
			if l > 1 { 
				if !strings.ContainsAny(txt, "[ & ]") { continue }
			} else { continue }
			
			txt = strings.ToLower(txt)
			
			if (strings.HasPrefix(txt, "[" + section + "]")) {
				flag |= cfg_HEADER 
			}
			
		} else {
			for i, ch := range txt {
				switch ch {
				case '=':
					if (flag & cfg_KEY != 0) { return out, cfgErr(file, line) }
					flag |= cfg_KEY
					key = setKey(buf)
					last = line
				case ',':
					if (flag & cfg_KEY == 0) { return out, cfgErr(file, line) }
					addVal(buf, &val)
					last = line
					flag |= cfg_COMMA
				case '\n':
					if (flag & cfg_KEY == 0) { return out, cfgErr(file, line) }
					if (flag & cfg_COMMA > 0) {
						flag &^= cfg_COMMA
						continue
					}
					flag &^= cfg_KEY
					addVal(buf, &val)
					out[key] = val
					val = nil
					last = line
				case '[':
					if (flag & cfg_KEY != 0) { return out, cfgErr(file, last) }
					return
				default:
					if buf.Len() == 0 {
						switch ch {
							case ' ':
								fallthrough
							case '\t':
								continue
					 	}
					}
					flag &^= cfg_COMMA
					if i == l - 1 && buf.Len() != 0 {
						return out, cfgErr(file, line)
					}
					buf.WriteRune(ch)
				}
				
			}
		}
	}
	return out, nil
}

// Writes key = values under [header] to File.
func SetFile(file, header, key string, value...string) error {
	for _, val := range value {
		for _, ch := range val {
			switch ch {
				case '[':
					fallthrough
				case ']':
					fallthrough
				case ',':
					return fmt.Errorf("Invalid character found in value: '%c' found in \"%s\".", ch, val )
			}
		}
	}
	
	header = strings.ToLower(header)
	key = strings.ToLower(key)
	f, err := os.Open(file)
	if err != nil { return err }
	defer f.Close()
	
	// Generate temp file, then close it, reopen it with append.
	tmp, err := ioutil.TempFile(os.TempDir(), "snugconf.tmp")
	if err != nil { return err }
	tmpfname := tmp.Name()
	tmp.Close()
	tmp, err = os.OpenFile(tmpfname, os.O_RDWR|os.O_APPEND, 0600)
	if err != nil { return err }
	defer tmp.Close()
	
	no_end_comma := func(input string) (no_comma bool) {
		no_comma = true
		for _, ch := range input {
			switch ch {
				case ',':
					no_comma = false
				case '\t':
					fallthrough
				case ' ':
					fallthrough
				default:
					no_comma = true
			}
		}
		return
	}
	
	// cfgSeek returns first half and bottom half of file, excluding the key = value.
	cfgSeek := func(header, key string, f *os.File) (upper int, lower int, flag int) {
		f.Seek(0,0)
		s := bufio.NewScanner(f)
		
		var line int
	
		for s.Scan() {
			line++
			b := s.Text()
			b = strings.ToLower(b)
			l := len(b)
		
			if l > 0 && b[0] == '#' || l == 0 { continue }
		
			if (flag & cfg_HEADER == 0) {
				if strings.HasPrefix(b, "[" + header + "]") { 
					flag |= cfg_HEADER 
					upper = line
					continue
				}
			} else {
				// if we hit the next [section], we didn't find the key to replace, which means its new.
				if b[0] == '[' {
					lower = upper + 1
					return
				}
			}

			if (flag & cfg_HEADER > 0) {
				if (flag & cfg_KEY == 0) && strings.HasPrefix(b, key) {
					pfx := strings.Split(b, "=")
					if strings.TrimSpace(pfx[0]) == key { 
						upper = line - 1
						flag |= cfg_KEY 
					}
				}
				if (flag & cfg_KEY > 0) && no_end_comma(b) {
					lower = line + 1
					return
				 }
			}
		}
		return line, -1, flag
	}
	
	head, tail, flag := cfgSeek(header, key, f)
	
	// Copys line start to line end of src file to dst file.
	copyFile := func(src, dst *os.File, start, end int) error {
		_, err := src.Seek(0, 0)
		if err != nil { return err }
		
		s := bufio.NewScanner(src)
		var line int
		
		for line < start {
			s.Scan()
			line++
		}

		for (line < end || end == -1) && s.Scan() {
			line++
			_, err := dst.WriteString(s.Text() + "\n")
			if err != nil { return err }
		}
		return nil
	}

	var txt []string
	
	if (flag & cfg_HEADER == 0) { txt = append(txt, "[" + header + "]") }
	
	var spacer []byte
	
	for i, str := range value {
		if str == "" { break }
		if i == 0 {
			txt = append(txt, key + " = " + str)
			spacer = make([]byte, len(key + " = "))
			for ch := range spacer {
				spacer[ch] = ' '
			}
			continue
		}
		txt = append(txt, string(spacer) + str)
	}
	
	// Appends first half of the file.
	err = copyFile(f, tmp, 0, head)
	if err != nil { return err }
	
	// Inject new header when needed, and key = values.	
	txtL := len(txt) - 1
	for i, out := range txt {
		if i == 0 {
			if flag & cfg_HEADER == 0 {
				_, err = tmp.WriteString("\n" + out + "\n")
				if err != nil { return err }
				continue
			}
		}
		if i < txtL { 
			_, err = tmp.WriteString(out + ",\n") 
			if err != nil { return err }
		} else { 
			_, err = tmp.WriteString(out + "\n")
			if err != nil { return err }
		}
	}
	
	// Appends second half of file.
	if tail != -1 { 
		err = copyFile(f, tmp, tail-1, -1)
		if err != nil { return err }
	}

	// Sync and close everything.
	err = tmp.Sync()
	if err != nil { return err }
	err = tmp.Close()
	if err != nil { return err }
	err = f.Close()
	if err != nil { return err }
	err = os.Remove(file)
	if err != nil { return err }
	
	// Move temp file to config file.
	err = os.Rename(tmpfname, file)
	if err != nil { return err }

	return nil
}
