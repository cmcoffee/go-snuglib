/* Package 'cfg' provides functions for reading and writing configuration files and their coresponding string values.
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
*/
package cfg

import (
	"bufio"
<<<<<<< HEAD
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

type Store struct {
	file     string
	mutex    *sync.RWMutex
	cfgStore map[string]map[string][]string
	readOnly bool
=======
	"fmt"
	"os"
	"bytes"
	"strings"
	"io/ioutil"
	"sync"
	"io"
)

type Store struct {
	file string
	mutex *sync.RWMutex
	cfgStore map[string]map[string][]string
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
}

const (
	cfg_HEADER = 1 << iota
	cfg_KEY
	cfg_COMMA
	cfg_ESCAPE
)

// Returns array of all retrieved string values under section with key.
<<<<<<< HEAD
func (s *Store) Get(section, key string) []string {
=======
func (s *Store) Get(section, key string) ([]string) {
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	section = strings.ToLower(section)
	key = strings.ToLower(key)
<<<<<<< HEAD
	if result, found := s.cfgStore[section][key]; !found {
		return []string{""}
	} else {
		if len(result) == 0 {
			return []string{""}
		}

		// Remove escape characters.
		for i, val := range result {
			result[i] = strings.Replace(val, "\\", "", -1)
		}
=======
	if result, found := s.cfgStore[section][key]; !found { 
		return []string{""}
	} else {
		if len(result) == 0 { return []string{""} }

		// Remove escape characters.
		for i, val := range result { result[i] = strings.Replace(val, "\\", "", -1) }
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
		return result
	}
}

<<<<<<< HEAD
=======

>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
// Returns array of all sections in config file.
func (s *Store) ListSections() (out []string) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for section, _ := range s.cfgStore {
		out = append(out, section)
	}
	return
}

// Returns keys of section specified.
func (s *Store) ListKeys(section string) (out []string) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
<<<<<<< HEAD
	if v, ok := s.cfgStore[section]; !ok {
		return nil
	} else {
		for key, _ := range v {
=======
	if v, ok := s.cfgStore[section]; !ok { 
		return nil 
	} else {
		for key, _ := range v { 
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
			out = append(out, key)
		}
	}
	return
}

// Returns true if section or section and key exists.
<<<<<<< HEAD
func (s *Store) Exists(input ...string) (found bool) {
=======
func (s *Store) Exists(input...string) (found bool) {
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	inlen := len(input)
<<<<<<< HEAD
	if inlen == 0 {
		return false
	}
	if inlen > 0 {
		if _, found = s.cfgStore[input[0]]; found {
			return
		}
=======
	if inlen == 0 { return false }
	if inlen > 0 {
		if _, found = s.cfgStore[input[0]]; found { return } 
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
	}
	if inlen > 1 {
		if found == true {
			_, found = s.cfgStore[input[0]][input[1]]
			return
		}
	}
	return
}

// Sets key = values under [section], updates Store and saves to file.
<<<<<<< HEAD
func (s *Store) Set(section, key string, value ...string) (err error) {
=======
func (s *Store) Set(section, key string, value...string) (err error) {
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
	s.mutex.Lock()
	defer s.mutex.Unlock()
	section = strings.ToLower(section)
	key = strings.ToLower(key)
	var newValue []string
<<<<<<< HEAD
	for _, val := range value {
		newValue = append(newValue, val)
	}

	// If read-only, do not write to file.
	if !s.readOnly {
		if err := SetFile(s.file, section, key, newValue[0:]...); err != nil {
			return err
		}
	}
=======
	for _, val := range value { newValue = append(newValue, val) }

	if err := SetFile(s.file, section, key, newValue[0:]...); err != nil { return err }
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600

	// Create new map if one doesn't exist.
	if _, ok := s.cfgStore[section]; !ok {
		s.cfgStore[section] = make(map[string][]string)
	}

	s.cfgStore[section][key] = newValue
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

<<<<<<< HEAD
func cfgErr(file string, line int) error {
	return fmt.Errorf("Syntax error found in %s on line %d.", file, line)
}

// Creates a new empty config file & Store, overwriting an existing file with comments if specified.
func Create(file string, comment ...string) (out *Store, err error) {
	f, err := os.Create(file)
	if err != nil {
		return nil, err
	}
=======
func cfgErr(file string, line int) error { return fmt.Errorf("Syntax error found in %s on line %d.", file, line) }

// Creates a new empty config file & Store, overwriting an existing file with comments if specified.
func Create(file string, comment...string) (out *Store, err error) {
	f, err := os.Create(file)
	if err != nil { return nil, err }
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
	defer f.Close()
	out = &Store{
		file,
		new(sync.RWMutex),
		make(map[string]map[string][]string),
<<<<<<< HEAD
		false,
	}
	if len(comment) > 0 {
		for _, c := range comment {
			f.WriteString("# " + c + "\n")
=======
	}
	if len(comment) > 0 {
		for _, c := range comment {
			f.WriteString("# " + c + "\n");
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
		}
	}
	return
}

<<<<<<< HEAD
// Reads configuration file and returns Store, changes are not saved to disk.
func ReadOnly(file string) (out *Store, err error) {
	out, err = Load(file)
	if out != nil {
		out.readOnly = true
	}
	return
}

// Reads configuration file and returns Store.
func Load(file string) (out *Store, err error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)

	var flag, line, last int

=======
// Reads configuration file and returns Store.
func Load(file string) (out *Store, err error) {
	f, err := os.Open(file)
	if err != nil { return nil, err }
	defer f.Close()
	s := bufio.NewScanner(f)
	
	var flag, line, last int
	
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
	buf := &bytes.Buffer{}
	var section, key string
	var val []string
	out = &Store{
		file,
		new(sync.RWMutex),
		make(map[string]map[string][]string),
<<<<<<< HEAD
		false,
	}

scanLoop:
=======
	}
	
	scanLoop:
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
	for s.Scan() {
		line++
		txt := s.Text() + "\n"
		l := len(txt)
<<<<<<< HEAD
		if l < 2 {
			continue
		}

		for i, ch := range txt {
			if (flag & cfg_ESCAPE) > 0 {
				if i == l-1 && buf.Len() != 0 {
=======
		if l < 2 { continue }
		
		for i, ch := range txt {
			if (flag & cfg_ESCAPE) > 0 {
				if i == l - 1 && buf.Len() != 0 {
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
					return out, cfgErr(file, line)
				}
				buf.WriteRune(ch)
				flag &^= cfg_ESCAPE
				continue
			}
			switch ch {
			case '[':
<<<<<<< HEAD
				if flag&cfg_KEY != 0 {
					return out, cfgErr(file, last)
				}
				last = line
				if l > 2 && strings.ContainsAny(txt, "[ & ]") {
					section = txt[1 : l-2]
=======
				if (flag & cfg_KEY != 0) { return out, cfgErr(file, last) }
				last = line
				if l > 2 && strings.ContainsAny(txt, "[ & ]") {
					section = txt[1:l-2]
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
					flag |= cfg_HEADER
					section = strings.ToLower(section)
					out.cfgStore[section] = make(map[string][]string)
					continue scanLoop
<<<<<<< HEAD
				} else {
					return out, cfgErr(file, line)
				}
			case '#':
				continue scanLoop
			case '=':
				if flag&cfg_KEY != 0 {
					return out, cfgErr(file, line)
				}
=======
				} else { return out, cfgErr(file, line) }
			case '#': 
				continue scanLoop;
			case '=':
				if (flag & cfg_KEY != 0) { return out, cfgErr(file, line) }
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
				flag |= cfg_KEY
				key = setKey(buf)
				last = line
			case ',':
<<<<<<< HEAD
				if flag&cfg_KEY == 0 {
					return out, cfgErr(file, line)
				}
=======
				if (flag & cfg_KEY == 0) { return out, cfgErr(file, line) }
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
				addVal(buf, &val)
				last = line
				flag |= cfg_COMMA
			case '\n':
<<<<<<< HEAD
				if flag&cfg_KEY == 0 {
					return out, cfgErr(file, line)
				}
				if flag&cfg_COMMA != 0 {
=======
				if (flag & cfg_KEY == 0) { return out, cfgErr(file, line) }
				if (flag & cfg_COMMA != 0) {
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
					flag &^= cfg_COMMA
					continue
				}
				flag &^= cfg_HEADER
				flag &^= cfg_KEY
				addVal(buf, &val)
				out.cfgStore[section][key] = val
				val = nil
				last = line
				continue scanLoop
			case '\\':
				flag |= cfg_ESCAPE
				fallthrough
			default:
				if buf.Len() == 0 {
					switch ch {
<<<<<<< HEAD
					case ' ':
						fallthrough
					case '\t':
						continue
					}
				}
				flag &^= cfg_COMMA
				if i == l-1 && buf.Len() != 0 {
=======
						case ' ':
							fallthrough
						case '\t':
							continue
					}
				}
				flag &^= cfg_COMMA
				if i == l - 1 && buf.Len() != 0 {
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
					return out, cfgErr(file, line)
				}
				buf.WriteRune(ch)
			}
<<<<<<< HEAD

		}
	}
	if flag&cfg_KEY != 0 {
		return out, cfgErr(file, last)
	}
	return out, nil
=======
				
		}
	}
	if (flag & cfg_KEY != 0) {
		return out, cfgErr(file, last)
	}
	return out, nil	
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
}

//Lists all Sections in config file.
func ListSections(file string) (out []string, err error) {
	f, err := os.Open(file)
<<<<<<< HEAD
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)

	for s.Scan() {
		txt := s.Text()
		l := len(txt)

		if l > 1 && txt[0] == '#' || l == 1 {
			continue
		}
=======
	if err != nil { return nil, err }
	defer f.Close()
	s := bufio.NewScanner(f)
		
	for s.Scan() {
		txt := s.Text()
		l := len(txt)
		
		if l > 1 && txt[0] == '#' || l == 1 { continue }
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
		if l > 2 && txt[0] == '[' && txt[l-1] == ']' {
			out = append(out[0:], txt[1:l-1])
		}
	}
	return out, err
}

// Returns map of specific [section] within configuration file.
func ReadFile(file, section string) (out map[string][]string, err error) {
	section = strings.ToLower(section)
	f, err := os.Open(file)
<<<<<<< HEAD
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)

	var flag, line, last int

=======
	if err != nil { return nil, err }
	defer f.Close()
	s := bufio.NewScanner(f)
	
	var flag, line, last int
	
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
	buf := &bytes.Buffer{}
	var key string
	var val []string
	out = make(map[string][]string)
<<<<<<< HEAD

=======
		
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
	for s.Scan() {
		line++
		txt := s.Text() + "\n"
		l := len(txt)
<<<<<<< HEAD

		if l > 1 && txt[0] == '#' || l == 1 {
			continue
		}

		if flag&cfg_HEADER == 0 {

			// Skip to section sections only.
			if l > 1 {
				if !strings.ContainsAny(txt, "[ & ]") {
					continue
				}
			} else {
				continue
			}

			txt = strings.ToLower(txt)

			if strings.HasPrefix(txt, "["+section+"]") {
				flag |= cfg_HEADER
			}

		} else {
			for i, ch := range txt {
				if (flag & cfg_ESCAPE) > 0 {
					if i == l-1 && buf.Len() != 0 {
						return out, cfgErr(file, line)
					}
=======
		
		if l > 1 && txt[0] == '#' || l == 1 { continue }
		
		if (flag & cfg_HEADER == 0) {
	
			// Skip to section sections only.
			if l > 1 { 
				if !strings.ContainsAny(txt, "[ & ]") { continue }
			} else { continue }
			
			txt = strings.ToLower(txt)
			
			if (strings.HasPrefix(txt, "[" + section + "]")) {
				flag |= cfg_HEADER 
			}
			
		} else {
			for i, ch := range txt {
				if (flag & cfg_ESCAPE) > 0 {
					if i == l - 1 && buf.Len() != 0 {
						return out, cfgErr(file, line)
					}	
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
					buf.WriteRune(ch)
					flag &^= cfg_ESCAPE
					continue
				}
				switch ch {
				case '=':
<<<<<<< HEAD
					if flag&cfg_KEY != 0 {
						return out, cfgErr(file, line)
					}
=======
					if (flag & cfg_KEY != 0) { return out, cfgErr(file, line) }
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
					flag |= cfg_KEY
					key = setKey(buf)
					last = line
				case ',':
<<<<<<< HEAD
					if flag&cfg_KEY == 0 {
						return out, cfgErr(file, line)
					}
=======
					if (flag & cfg_KEY == 0) { return out, cfgErr(file, line) }
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
					addVal(buf, &val)
					last = line
					flag |= cfg_COMMA
				case '\n':
<<<<<<< HEAD
					if flag&cfg_KEY == 0 {
						return out, cfgErr(file, line)
					}
					if flag&cfg_COMMA > 0 {
=======
					if (flag & cfg_KEY == 0) { return out, cfgErr(file, line) }
					if (flag & cfg_COMMA > 0) {
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
						flag &^= cfg_COMMA
						continue
					}
					flag &^= cfg_KEY
					addVal(buf, &val)
					out[key] = val
					val = nil
					last = line
				case '[':
<<<<<<< HEAD
					if flag&cfg_KEY != 0 {
						return out, cfgErr(file, last)
					}
=======
					if (flag & cfg_KEY != 0) { return out, cfgErr(file, last) }
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
					return
				case '\\':
					flag |= cfg_ESCAPE
					fallthrough
				default:
					if buf.Len() == 0 {
						switch ch {
<<<<<<< HEAD
						case ' ':
							fallthrough
						case '\t':
							continue
						}
					}
					flag &^= cfg_COMMA
					if i == l-1 && buf.Len() != 0 {
=======
							case ' ':
								fallthrough
							case '\t':
								continue
					 	}
					}
					flag &^= cfg_COMMA
					if i == l - 1 && buf.Len() != 0 {
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
						return out, cfgErr(file, line)
					}
					buf.WriteRune(ch)
				}
<<<<<<< HEAD

=======
				
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
			}
		}
	}
	return out, nil
}

// Writes key = values under [section] to File.
<<<<<<< HEAD
func SetFile(file, section, key string, value ...string) error {
	for _, val := range value {
		for _, ch := range val {
			switch ch {
			case '[':
				fallthrough
			case ']':
				fallthrough
			case ',':
				return fmt.Errorf("Invalid character found in value: '%c' found in \"%s\".", ch, val)
			}
		}
	}

	section = strings.ToLower(section)
	key = strings.ToLower(key)
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	// Generate temp file, then close it, reopen it with append.
	tmp, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s.temp_conf.", os.Args[0]))
	if err != nil {
		return err
	}
	tmpfname := tmp.Name()

=======
func SetFile(file, section, key string, value...string) error {
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
	
	section = strings.ToLower(section)
	key = strings.ToLower(key)
	f, err := os.Open(file)
	if err != nil { return err }
	defer f.Close()
	
	// Generate temp file, then close it, reopen it with append.
	tmp, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s.temp_conf.", os.Args[0]))
	if err != nil { return err }
	tmpfname := tmp.Name()
	
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
	no_end_comma := func(input string) (no_comma bool) {
		no_comma = true
		for _, ch := range input {
			switch ch {
<<<<<<< HEAD
			case ',':
				no_comma = false
			case '\t':
				fallthrough
			case ' ':
				fallthrough
			default:
				no_comma = true
=======
				case ',':
					no_comma = false
				case '\t':
					fallthrough
				case ' ':
					fallthrough
				default:
					no_comma = true
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
			}
		}
		return
	}
<<<<<<< HEAD

	// cfgSeek returns first half and bottom half of file, excluding the key = value.
	cfgSeek := func(section, key string, f *os.File) (upper int, lower int, flag int) {
		f.Seek(0, 0)
		s := bufio.NewScanner(f)

		var line int

=======
	
	// cfgSeek returns first half and bottom half of file, excluding the key = value.
	cfgSeek := func(section, key string, f *os.File) (upper int, lower int, flag int) {
		f.Seek(0,0)
		s := bufio.NewScanner(f)
		
		var line int
	
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
		for s.Scan() {
			line++
			b := s.Text()
			b = strings.ToLower(b)
			l := len(b)
<<<<<<< HEAD

			if l > 0 && b[0] == '#' || l == 0 {
				continue
			}

			if flag&cfg_HEADER == 0 {
				if strings.HasPrefix(b, "["+section+"]") {
					flag |= cfg_HEADER
=======
		
			if l > 0 && b[0] == '#' || l == 0 { continue }
		
			if (flag & cfg_HEADER == 0) {
				if strings.HasPrefix(b, "[" + section + "]") { 
					flag |= cfg_HEADER 
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
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

<<<<<<< HEAD
			if flag&cfg_HEADER > 0 {
				if (flag&cfg_KEY == 0) && strings.HasPrefix(b, key) {
					pfx := strings.Split(b, "=")
					if strings.TrimSpace(pfx[0]) == key {
						upper = line - 1
						flag |= cfg_KEY
					}
				}
				if (flag&cfg_KEY > 0) && no_end_comma(b) {
					lower = line + 1
					return
				}
=======
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
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
			}
		}
		return line, -1, flag
	}
<<<<<<< HEAD

	head, tail, flag := cfgSeek(section, key, f)

	// Copys line start to line end of src file to dst file.
	copyFile := func(src, dst *os.File, start, end int) error {
		_, err := src.Seek(0, 0)
		if err != nil {
			return err
		}

		s := bufio.NewScanner(src)
		var line int

=======
	
	head, tail, flag := cfgSeek(section, key, f)
	
	// Copys line start to line end of src file to dst file.
	copyFile := func(src, dst *os.File, start, end int) error {
		_, err := src.Seek(0, 0)
		if err != nil { return err }
		
		s := bufio.NewScanner(src)
		var line int
		
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
		for line < start {
			s.Scan()
			line++
		}

		for (line < end || end == -1) && s.Scan() {
			line++
			_, err := dst.WriteString(s.Text() + "\n")
<<<<<<< HEAD
			if err != nil {
				return err
			}
=======
			if err != nil { return err }
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
		}
		return nil
	}

	var txt []string
<<<<<<< HEAD

	if flag&cfg_HEADER == 0 {
		txt = append(txt, "["+section+"]")
	}

	var spacer []byte

	for i, str := range value {
		if str == "" {
			break
		}
		if i == 0 {
			txt = append(txt, key+" = "+str)
			spacer = make([]byte, len(key+" = "))
=======
	
	if (flag & cfg_HEADER == 0) { txt = append(txt, "[" + section + "]") }
	
	var spacer []byte
	
	for i, str := range value {
		if str == "" { break }
		if i == 0 {
			txt = append(txt, key + " = " + str)
			spacer = make([]byte, len(key + " = "))
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
			for ch := range spacer {
				spacer[ch] = ' '
			}
			continue
		}
<<<<<<< HEAD
		txt = append(txt, string(spacer)+str)
	}

	// Appends first half of the file.
	err = copyFile(f, tmp, 0, head)
	if err != nil {
		return err
	}

	// Inject new section when needed, and key = values.
	txtL := len(txt) - 1
	for i, out := range txt {
		if i == 0 {
			if flag&cfg_HEADER == 0 {
				_, err = tmp.WriteString("\n" + out + "\n")
				if err != nil {
					return err
				}
				continue
			}
		}
		if i < txtL {
			_, err = tmp.WriteString(out + ",\n")
			if err != nil {
				return err
			}
		} else {
			_, err = tmp.WriteString(out + "\n")
			if err != nil {
				return err
			}
		}
	}

	// Appends second half of file.
	if tail != -1 {
		err = copyFile(f, tmp, tail-1, -1)
		if err != nil {
			return err
		}
=======
		txt = append(txt, string(spacer) + str)
	}
	
	// Appends first half of the file.
	err = copyFile(f, tmp, 0, head)
	if err != nil { return err }
	
	// Inject new section when needed, and key = values.	
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
>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600
	}

	// Sync and close everything.
	err = tmp.Sync()
<<<<<<< HEAD
	if err != nil {
		return err
	}

	err = tmp.Close()
	if err != nil {
		return err
	}

	tmp, err = os.Open(tmpfname)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	destfile, err := os.OpenFile(file, os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer destfile.Close()

	_, err = io.Copy(destfile, tmp)
	if err != nil {
		return err
	}

	err = destfile.Sync()
	if err != nil {
		return err
	}

	err = tmp.Close()
	if err != nil {
		return err
	}

	err = os.Remove(tmpfname)
	if err != nil {
		return err
	}
=======
	if err != nil { return err }

	err = tmp.Close()
	if err != nil { return err }

	tmp, err = os.Open(tmpfname)
	if err != nil { return err }

	err = f.Close()
	if err != nil { return err }
	
	destfile, err := os.OpenFile(file, os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil { return err }
	defer destfile.Close()

	_, err = io.Copy(destfile, tmp)
	if err != nil { return err }

	err = destfile.Sync()
	if err != nil {return err }

	err = tmp.Close()
	if err != nil { return err }

	err = os.Remove(tmpfname)
	if err != nil { return err }

>>>>>>> eb4e0ca105f39f9d6110b121ff1cfe5079ab6600

	return nil
}
