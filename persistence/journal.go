package persistence

import (
	"bufio"
	"io"
	"os"
	"sync"
	"time"
)

const (
	// defaultFlushInterval is the maximum time between automatic flushes.
	defaultFlushInterval = 10 * time.Millisecond
	// defaultBufSize is the initial write-buffer size for the journal.
	defaultBufSize = 64 * 1024 // 64 KiB
)

// Journal is a thread-safe, append-only Write-Ahead Log.
//
// Events are buffered in a bufio.Writer and flushed to disk either when the
// buffer fills up or every defaultFlushInterval (whichever comes first), so
// that the number of fsync system calls is minimised.
type Journal struct {
	mu     sync.Mutex
	file   *os.File
	writer *bufio.Writer

	ticker *time.Ticker
	done   chan struct{}
	wg     sync.WaitGroup
}

// OpenJournal opens (or creates) the journal file at path and starts the
// background flush goroutine.
func OpenJournal(path string) (*Journal, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}

	j := &Journal{
		file:   f,
		writer: bufio.NewWriterSize(f, defaultBufSize),
		ticker: time.NewTicker(defaultFlushInterval),
		done:   make(chan struct{}),
	}

	j.wg.Add(1)
	go j.flushLoop()
	return j, nil
}

// Append writes a MatchingEvent to the journal buffer. It is safe to call from
// multiple goroutines concurrently.
func (j *Journal) Append(event MatchingEvent) error {
	record, err := encodeEvent(event)
	if err != nil {
		return err
	}

	j.mu.Lock()
	defer j.mu.Unlock()

	_, err = j.writer.Write(record)
	return err
}

// Flush forces all buffered data to be written to disk (fsync).
func (j *Journal) Flush() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.flush()
}

// flush must be called with j.mu held.
func (j *Journal) flush() error {
	if err := j.writer.Flush(); err != nil {
		return err
	}
	return j.file.Sync()
}

// Close flushes remaining data, stops the background goroutine, and closes the
// underlying file.
func (j *Journal) Close() error {
	j.ticker.Stop()
	close(j.done)
	j.wg.Wait()

	j.mu.Lock()
	defer j.mu.Unlock()
	if err := j.flush(); err != nil {
		_ = j.file.Close()
		return err
	}
	return j.file.Close()
}

// flushLoop periodically flushes the write buffer.
func (j *Journal) flushLoop() {
	defer j.wg.Done()
	for {
		select {
		case <-j.ticker.C:
			j.mu.Lock()
			_ = j.flush()
			j.mu.Unlock()
		case <-j.done:
			return
		}
	}
}

// ReadAll opens the journal at path in read-only mode and decodes every
// record it contains.  It returns all successfully decoded events and the
// first unrecoverable error (io.EOF is never returned to the caller).
func ReadAll(path string) ([]MatchingEvent, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var events []MatchingEvent
	r := bufio.NewReader(f)
	for {
		e, err := decodeEvent(r)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break // truncated tail is tolerated (crash during write)
			}
			return events, err
		}
		events = append(events, e)
	}
	return events, nil
}
