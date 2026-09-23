package main

import (
	"sync"
	"testing"
)

func TestMemoryStoreCreate(t *testing.T) {
	s := NewMemoryStore()
	task := s.Create("buy milk")

	if task.ID != 0 {
		t.Fatalf("expected first task id=0, got %d", task.ID)
	}
	if task.Title != "buy milk" {
		t.Fatalf("expected title 'buy milk', got %q", task.Title)
	}
	if task.Done {
		t.Fatalf("new task should not be done")
	}
}

func TestMemoryStoreList(t *testing.T) {
	s := NewMemoryStore()
	s.Create("a")
	s.Create("b")
	s.Create("c")

	list := s.List()
	if len(list) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(list))
	}
}

func TestMemoryStoreDelete(t *testing.T) {
	s := NewMemoryStore()
	s.Create("x")

	if !s.Delete(0) {
		t.Fatalf("expected delete to succeed")
	}
	if s.Delete(0) {
		t.Fatalf("expected second delete to fail")
	}
}

func TestMemoryStoreConcurrency(t *testing.T) {
	s := NewMemoryStore()
	var wg sync.WaitGroup
	n := 100

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s.Create("task")
		}(i)
	}
	wg.Wait()

	list := s.List()
	if len(list) != n {
		t.Fatalf("expected %d tasks, got %d", n, len(list))
	}
}

func TestMemoryStoreStats(t *testing.T) {
	s := NewMemoryStore()
	s.Create("a")
	s.Create("b")
	done := true
	s.Update(0, nil, &done)

	total, doneCount := s.Stats()
	if total != 2 {
		t.Fatalf("expected total=2, got %d", total)
	}
	if doneCount != 1 {
		t.Fatalf("expected done=1, got %d", doneCount)
	}
}
