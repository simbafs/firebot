package bucket

import (
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
	b.data[id] = BucketItem[T]{updateTime: time.Now(), data: item}
}

func (b *Bucket[K, T]) Get(id K) (T, bool) {
	item, ok := b.data[id]
	return item.data, ok
}

// GC delete the data older than
func (b *Bucket[K, T]) GC() {
	for k, v := range b.data {
		if time.Since(v.updateTime) > b.aliveTime {
			delete(b.data, k)
		}
	}
}

func (b *Bucket[K, T]) Len() int {
	return len(b.data)
}
