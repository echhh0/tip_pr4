package task

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
)

var ErrNotFound = errors.New("task not found")

type Repo struct {
	mu       sync.RWMutex
	seq      int64
	items    map[int64]*Task
	filePath string
}

func NewRepo() *Repo {
	return &Repo{items: make(map[int64]*Task)}
}

func (r *Repo) WithFile(path string) *Repo {
	r.filePath = path
	return r
}

func (r *Repo) Load() error {
	if r.filePath == "" {
		return nil
	}
	b, err := os.ReadFile(r.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	var s snapshot
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq = s.Seq
	r.items = make(map[int64]*Task, len(s.Items))
	for _, t := range s.Items {
		// делаем копии
		tt := *t
		r.items[tt.ID] = &tt
	}
	return nil
}

func (r *Repo) saveLocked() error {
	if r.filePath == "" {
		return nil
	}
	s := snapshot{Seq: r.seq, Items: make([]*Task, 0, len(r.items))}
	for _, t := range r.items {
		tt := *t
		s.Items = append(s.Items, &tt)
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := r.filePath + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, r.filePath)
}

type snapshot struct {
	Seq   int64   `json:"seq"`
	Items []*Task `json:"items"`
}

func (r *Repo) List() []*Task {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Task, 0, len(r.items))
	for _, t := range r.items {
		out = append(out, t)
	}
	return out
}

func (r *Repo) Get(id int64) (*Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	return t, nil
}

func (r *Repo) Create(title string) *Task {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	now := time.Now()
	t := &Task{ID: r.seq, Title: title, CreatedAt: now, UpdatedAt: now, Done: false}
	r.items[t.ID] = t
	_ = r.saveLocked()
	return t
}

func (r *Repo) Update(id int64, title string, done bool) (*Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	t.Title = title
	t.Done = done
	t.UpdatedAt = time.Now()
	_ = r.saveLocked()
	return t, nil
}

func (r *Repo) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return ErrNotFound
	}
	delete(r.items, id)
	return r.saveLocked()
}
