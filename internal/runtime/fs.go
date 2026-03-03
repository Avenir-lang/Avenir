package runtime

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

type fsService struct {
	nextID uint64
	mu     sync.Mutex
	files  map[uint64]*os.File
}

func newFSService() *fsService {
	return &fsService{
		files: make(map[uint64]*os.File),
	}
}

func (f *fsService) Open(path string, mode string) ([]byte, error) {
	flags, err := parseOpenMode(mode)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, flags, 0o666)
	if err != nil {
		return nil, err
	}
	id := f.nextHandle()
	f.mu.Lock()
	f.files[id] = file
	f.mu.Unlock()
	return encodeHandle(id), nil
}

func (f *fsService) Read(handle []byte, n int) ([]byte, error) {
	if n < 0 {
		return nil, fmt.Errorf("invalid read size %d", n)
	}
	id, err := decodeHandle(handle)
	if err != nil {
		return nil, err
	}
	f.mu.Lock()
	file := f.files[id]
	f.mu.Unlock()
	if file == nil {
		return nil, fmt.Errorf("invalid file handle")
	}
	if n == 0 {
		return []byte{}, nil
	}
	buf := make([]byte, n)
	nread, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return buf[:nread], nil
}

func (f *fsService) Write(handle []byte, data []byte) (int, error) {
	id, err := decodeHandle(handle)
	if err != nil {
		return 0, err
	}
	f.mu.Lock()
	file := f.files[id]
	f.mu.Unlock()
	if file == nil {
		return 0, fmt.Errorf("invalid file handle")
	}
	return file.Write(data)
}

func (f *fsService) Close(handle []byte) error {
	id, err := decodeHandle(handle)
	if err != nil {
		return err
	}
	f.mu.Lock()
	file, ok := f.files[id]
	if ok {
		delete(f.files, id)
	}
	f.mu.Unlock()
	if !ok || file == nil {
		return fmt.Errorf("invalid file handle")
	}
	return file.Close()
}

func (f *fsService) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func (f *fsService) Remove(path string) error {
	return os.Remove(path)
}

func (f *fsService) Mkdir(path string) error {
	return os.Mkdir(path, 0o755)
}

func (f *fsService) nextHandle() uint64 {
	return atomic.AddUint64(&f.nextID, 1)
}

func parseOpenMode(mode string) (int, error) {
	switch mode {
	case "r":
		return os.O_RDONLY, nil
	case "w":
		return os.O_CREATE | os.O_TRUNC | os.O_WRONLY, nil
	case "a":
		return os.O_CREATE | os.O_APPEND | os.O_WRONLY, nil
	case "r+":
		return os.O_RDWR, nil
	case "w+":
		return os.O_CREATE | os.O_TRUNC | os.O_RDWR, nil
	case "a+":
		return os.O_CREATE | os.O_APPEND | os.O_RDWR, nil
	case "rw":
		return os.O_CREATE | os.O_RDWR, nil
	default:
		return 0, fmt.Errorf("invalid open mode %q", mode)
	}
}
