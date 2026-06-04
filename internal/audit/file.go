package audit

import (
	"encoding/json"
	"os"
	"sync"
)

type FileReceiver struct {
	path string
	file *os.File
	mu   sync.Mutex
}

func NewFileReceiver(path string) (*FileReceiver, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &FileReceiver{path: path, file: f}, nil
}

func (r *FileReceiver) Notify(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	_, err = r.file.Write(append(data, '\n'))
	return err
}
