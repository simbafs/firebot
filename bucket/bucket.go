package bucket

import (
	"sync"
	"time"
)

type BucketItem[T any] struct {
	updateTime time.Time
	data       T
}

type PairItem[T any] struct {
	HasOld bool
	Old    T
	New    T
}

type Bucket[K comparable, T any] struct {
	mu        sync.RWMutex
	data      map[K]BucketItem[T]
	aliveTime time.Duration
}

func New[K comparable, T any](aliveTime time.Duration) *Bucket[K, T] {
	return &Bucket[K, T]{
		data:      map[K]BucketItem[T]{},
		aliveTime: aliveTime,
	}
}

func (b *Bucket[K, T]) Set(id K, item T) {
	b.mu.Lock()
	b.data[id] = BucketItem[T]{updateTime: time.Now(), data: item}
	b.mu.Unlock()
}

func (b *Bucket[K, T]) Get(id K) (T, bool) {
	b.mu.RLock()
	item, ok := b.data[id]
	b.mu.RUnlock()
	return item.data, ok
}

// GC delete the data older than aliveTime
func (b *Bucket[K, T]) GC() {
	b.mu.Lock()
	for k, v := range b.data {
		if time.Since(v.updateTime) > b.aliveTime {
			delete(b.data, k)
		}
	}
	b.mu.Unlock()
}

func (b *Bucket[K, T]) Len() int {
	b.mu.RLock()
	n := len(b.data)
	b.mu.RUnlock()
	return n
}
