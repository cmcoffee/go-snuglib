package nfo

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type logFile struct {
	write_lock   sync.Mutex
	output       int
	buffer       bytes.Buffer
	file         *os.File
	filename     string
	max_rotation uint
	cur_size     int64
	max_size     int64
}

// Keep map of open files
var open_files = make(map[string]*logFile)
var open_files_mutex sync.Mutex

const (
	to_BUFFER = iota
	to_FILE
)

// Write function that switches between file output and buffers to memory when files is being rotated.
func (f *logFile) Write(p []byte) (n int, err error) {
	f.write_lock.Lock()
	defer f.write_lock.Unlock()

	if f.output == to_FILE && int64(len(p))+f.cur_size >= f.max_size && f.max_size > 0 {
		f.output = to_BUFFER

		// Rotate files in background while writing to memory.
		go func() {
			if err := f.rotator(); err != nil { 
				if FatalOnFileError {
					Fatal(err)
				} else {
					Close(f.filename)
				}
			}
		}()
	}

	if f.output == to_FILE {
		n, err = f.file.Write(p)
		f.cur_size = f.cur_size + int64(n)
		return
	} else {
		return f.buffer.Write(p)
	}
}

// Opens a new log file for writing, max_size is threshold for rotation, max_rotation is number of previous logs to hold on to.
// Set max_size_mb to 0 to disable file rotation.
func File(l_file_flag int, filename string, max_size_mb uint, max_rotation uint) (err error) {
	max_size := int64(max_size_mb * 1048576)
	fpath, _ := filepath.Split(filename)

	set_writer := func(l_file *logFile) {
		mutex.Lock()
		for k, v := range l_map {
			if l_file_flag&k == k {
				v.out2 = l_file
			}
		}
		mutex.Unlock()
	}

	open_files_mutex.Lock()
	defer open_files_mutex.Unlock()

	// Check if we already have the file open, if so use it, and adjust the max_size and max_rotation.
	if e, ok := open_files[filename]; ok {
		if e.max_rotation != max_rotation || e.max_size != max_size {
			e.write_lock.Lock()
			e.max_rotation = max_rotation
			e.max_size = max_size
			e.write_lock.Unlock()
		}
		set_writer(e)
		return
	}

	_, err = os.Stat(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			if strings.Contains(fpath, string(os.PathSeparator)) {
				err = os.Mkdir(fpath, 0766)
				if err != nil {
					return err
				}
			}
		} else {
			return err
		}
	}

	l_file := new(logFile)
	l_file.filename = filename
	l_file.max_rotation = max_rotation
	l_file.max_size = max_size
	l_file.output = to_FILE

	l_file.file, err = os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	finfo, err := l_file.file.Stat()
	if err != nil {
		return err
	}

	l_file.cur_size = finfo.Size()

	set_writer(l_file)
	open_files[filename] = l_file

	return nil
}

// Closes logging file, removes file from all loggers, removes file from open files.
func Close(filename string) (err error) {
	open_files_mutex.Lock()
	defer open_files_mutex.Unlock()
	f := open_files[filename]
	mutex.Lock()
	for _, v := range l_map {
		if v.out2 == f {
			v.out2 = None
		}
	}
	mutex.Unlock()
	delete(open_files, filename)
	return f.file.Close()
}

// Closes file, rotates and removes files greater than max rotations allow, opens new file, dumps buffer to disk and switches write function back to disk.
func (l_file *logFile) rotator() (err error) {
	BlockShutdown()
	defer UnblockShutdown()

	fpath, fname := filepath.Split(l_file.filename)

	l_file.file.Close()

	flist, err := ioutil.ReadDir(fpath)
	if err != nil {
		return
	}

	files := make(map[string]os.FileInfo)

	for _, v := range flist {
		if strings.Contains(v.Name(), fname) {
			files[v.Name()] = v
		}
	}

	file_count := uint(len(files))

	// Rename files
	for i := file_count; i > 0; i-- {
		target := fname

		if i > 1 {
			target = fmt.Sprintf("%s.%d", target, i-1)
		}

		if _, ok := files[target]; ok {
			if i > l_file.max_rotation {
				err = os.Remove(fmt.Sprintf("%s%s", fpath, target))
				if err != nil {
					return err
				}
			} else {
				err = os.Rename(fmt.Sprintf("%s%s", fpath, target), fmt.Sprintf("%s%s.%d", fpath, fname, i))
				if err != nil {
					return err
				}
			}
		}
	}

	// Open new file.
	l_file.file, err = os.OpenFile(l_file.filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return
	}

	// Switch Write function back to writing to file.
	l_file.write_lock.Lock()

	// Set l_files new size to new buffer.
	l_file.cur_size = int64(l_file.buffer.Len())

	// Copy buffer to new file.
	_, err = io.Copy(l_file.file, &l_file.buffer)
	if err != nil {
		return
	}

	l_file.buffer.Reset()
	l_file.output = to_FILE

	// Unlock mutex to allow writing to file.
	l_file.write_lock.Unlock()
	return
}
