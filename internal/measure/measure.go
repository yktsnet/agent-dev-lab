package measure

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Row struct {
	Condition string
	Run       int
	Lane      int
	Stage     string
	Seconds   float64
}

type Writer struct {
	mu sync.Mutex
	f  *os.File
	w  *csv.Writer
}

func Open(path string) (*Writer, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	_, statErr := os.Stat(path)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	w := csv.NewWriter(f)
	if os.IsNotExist(statErr) {
		w.Write([]string{"condition", "run", "lane", "stage", "seconds"})
	}
	return &Writer{f: f, w: w}, nil
}

func (x *Writer) Add(r Row) {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.w.Write([]string{r.Condition, fmt.Sprint(r.Run), fmt.Sprint(r.Lane), r.Stage, fmt.Sprintf("%.3f", r.Seconds)})
	x.w.Flush()
}

func (x *Writer) Close() error { x.w.Flush(); return x.f.Close() }

// Time は f にかかった秒数を返す。
func Time(f func() error) (float64, error) {
	t := time.Now()
	err := f()
	return time.Since(t).Seconds(), err
}
