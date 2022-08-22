package main

import "sync"

type AtomicCache[KeyType comparable, ValueType any] struct {
	mtx   *sync.Mutex
	cache map[KeyType]ValueType
}

func NewAtomicCache[KeyType comparable, ValueType any]() *AtomicCache[KeyType, ValueType] {
	return &AtomicCache[KeyType, ValueType]{
		mtx:   &sync.Mutex{},
		cache: map[KeyType]ValueType{},
	}
}

func (ac *AtomicCache[KeyType, ValueType]) Load(key KeyType) ValueType {
	ac.mtx.Lock()
	defer ac.mtx.Unlock()
	return ac.cache[key]
}

func (ac *AtomicCache[KeyType, ValueType]) Store(key KeyType, value ValueType) {
	ac.mtx.Lock()
	ac.cache[key] = value
	ac.mtx.Unlock()
}

func (ac *AtomicCache[KeyType, ValueType]) Exists(key KeyType) bool {
	ac.mtx.Lock()
	_, ok := ac.cache[key]
	ac.mtx.Unlock()
	return ok
}
