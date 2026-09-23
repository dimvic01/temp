package main

import (
	"sync"
	"time"
)

// Store defines the data-access interface for tasks.
type Store interface {
	List() []Task
	Get(id int) (Task, bool)
	Create(title string) Task
	Update(id int, title *string, done *bool) (Task, bool)
	Delete(id int) bool
	Stats() (total int, done int)
}

// MemoryStore is an in-memory, concurrency-safe task store.
type MemoryStore struct {
	mu    sync.RWMutex
	tasks map[int]Task
	next  int
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{tasks: make(map[int]Task)}
}

func (s *MemoryStore) List() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		out = append(out, t)
	}
	return out
}

func (s *MemoryStore) Get(id int) (Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	return t, ok
}

func (s *MemoryStore) Create(title string) Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.next
	s.next++
	t := Task{
		ID:        id,
		Title:     title,
		Done:      false,
		CreatedAt: time.Now().UTC(),
	}
	s.tasks[id] = t
	return t
}

func (s *MemoryStore) Update(id int, title *string, done *bool) (Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	if !ok {
		return Task{}, false
	}
	if title != nil {
		t.Title = *title
	}
	if done != nil {
		t.Done = *done
	}
	s.tasks[id] = t
	return t, true
}

func (s *MemoryStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; !ok {
		return false
	}
	delete(s.tasks, id)
	return true
}

func (s *MemoryStore) Stats() (total, done int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.tasks {
		total++
		if t.Done {
			done++
		}
	}
	return
}
