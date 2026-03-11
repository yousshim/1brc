package main

import "slices"

type key[K any] struct {
	key K
	hash uint64
}

type HashMap[K, V any] struct {
	keys []key[K]
	data []*V
	hash func(K) uint64
	equal func(K, K) bool
}

func (m HashMap[K, V]) lookup(k K) *V {
	h := m.hash(k)
	l := uint64(len(m.data))
	idx := h & (l - 1)
	for m.data[idx] != nil {
		if m.keys[idx].hash == h && m.equal(m.keys[idx].key, k) {
			return m.data[idx]
		}
		idx = (idx + 1) & (l - 1)
	}
	return nil
}

func (m HashMap[K, V]) insert(k K, v *V) {
	h := m.hash(k)
	l := uint64(len(m.data))
	idx := h & (l - 1)
	for m.data[idx] != nil {
		if m.keys[idx].hash == h && m.equal(m.keys[idx].key, k) {
			break
		}
		idx = (idx + 1) & (l - 1)
	}
	m.data[idx] = v
	m.keys[idx] = key[K]{
		key: k,
		hash: h,
	}
}

func (m HashMap[K, V]) values() []*V {
	return slices.Clone(m.data)
}
