package aof

import (
	"bufio"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/KhasarMunkh/Go-Redis-From-Scratch/resp" // Importing the resp package for RESP protocol handling
)

type Aof struct {
	path string
	file *os.File
	wr   *bufio.Writer
	mu   sync.Mutex
}

// maybe we should use a more sophisticated way to handle AOF, like using a ring buffer or a queue
// to avoid blocking the main thread when writing to the file.
// the goroutine will handle the file writing in the background, would ring buffer be better?
// goroutine vs ring buffer: // - goroutine is simpler to implement, but may block if the file writing is slow
func NewAof(p string) (*Aof, error) {
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	aof := &Aof{
		path: p,
		file: f,
		wr:   bufio.NewWriter(f),
	}
	go func() {
		for {
			aof.mu.Lock()
			if err := aof.file.Sync(); err != nil {
				log.Printf("Error syncing AOF file: %v", err)
			}
			aof.mu.Unlock()
			time.Sleep(time.Millisecond * 1000) // Sync every second, only <1s of data loss
		}
	}()
	return aof, nil
}

func (aof *Aof) Close() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()
	err := aof.file.Close()
	return err
}

func (aof *Aof) Write(v resp.Value) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	_, err := aof.file.Write(v.Marshal())
	if err != nil {
		return err
	}

	return nil
}

// must specify a callback function to handle each value read from the AOF file
func (aof *Aof) Replay(callback func(value resp.Value)) error {
	rf, err := os.Open(aof.path)
	if err != nil {
		return err
	}
	defer rf.Close()

	r := resp.NewResp(rf) // Create a new RESP reader
	for {
		val, err := r.Read()
		if err == nil {
			callback(val)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	// parse and apply
	return nil
}
