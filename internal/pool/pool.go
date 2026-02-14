// Package pool предоставляет generic-пул объектов, которые имеют метод Reset().
// Это позволяет повторно использовать «тяжёлые» объекты без лишней аллокации,
// при этом гарантируя, что объект сбрасывается перед повторным использованием.
package pool

import (
	"sync"
)

// Resettable описывает интерфейс для объектов, которые можно сбросить.
// Любая структура, используемая с Pool, должна реализовывать этот метод.
type Resettable interface {
	Reset()
}

// Pool представляет собой generic-пул объектов T.
// T должен реализовывать интерфейс Resettable.
type Pool[T Resettable] struct {
	pool *sync.Pool
	new  func() T
}

// New создаёт новый объект Pool для типа T.
// Аргумент newFunc задаёт функцию создания нового объекта, если пул пуст.
func New[T Resettable](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		pool: &sync.Pool{
			New: func() interface{} {
				return newFunc()
			},
		},
		new: newFunc,
	}
}

// Get возвращает объект из пула.
// Если пул пуст, создаётся новый объект с помощью функции newFunc.
func (p *Pool[T]) Get() T {
	obj := p.pool.Get()
	if obj == nil {
		return p.new()
	}
	return obj.(T)
}

// Put помещает объект обратно в пул.
// Перед возвратом в пул вызывается метод Reset().
func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.pool.Put(obj)
}
