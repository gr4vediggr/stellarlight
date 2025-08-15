package objects

import "sync"

type SharedCollection[T any] struct {
	objects map[uint64]T
	nextId  uint64
	mapMux  sync.Mutex
}

func NewSharedCollection[T any](capacity ...int) *SharedCollection[T] {
	var newMap map[uint64]T
	if len(capacity) > 0 {
		newMap = make(map[uint64]T, capacity[0])
	} else {
		newMap = make(map[uint64]T)
	}
	return &SharedCollection[T]{
		objects: newMap,
		nextId:  1,
	}
}

func (sc *SharedCollection[T]) Add(obj T, id ...uint64) uint64 {
	sc.mapMux.Lock()
	defer sc.mapMux.Unlock()

	thisId := sc.nextId
	if len(id) > 0 {
		thisId = id[0]
	}

	sc.objects[thisId] = obj
	sc.nextId++

	return thisId
}

func (sc *SharedCollection[T]) Get(id uint64) (T, bool) {
	sc.mapMux.Lock()
	defer sc.mapMux.Unlock()

	obj, exists := sc.objects[id]
	return obj, exists
}

func (sc *SharedCollection[T]) Remove(id uint64) {
	sc.mapMux.Lock()
	defer sc.mapMux.Unlock()

	delete(sc.objects, id)
}

func (sc *SharedCollection[T]) ForEach(fn func(id uint64, obj T)) {
	sc.mapMux.Lock()

	localCopy := make(map[uint64]T, len(sc.objects))
	for id, obj := range sc.objects {
		localCopy[id] = obj
	}
	sc.mapMux.Unlock()

	for id, obj := range localCopy {
		fn(id, obj)
	}

}

func (sc *SharedCollection[T]) Len() int {
	return len(sc.objects)
}
