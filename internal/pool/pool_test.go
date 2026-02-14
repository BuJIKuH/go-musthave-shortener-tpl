package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStruct — структура для теста пула
type TestStruct struct {
	A int
	B string
	C []int
	D map[string]string
}

// Reset реализует Resettable для TestStruct
func (t *TestStruct) Reset() {
	t.A = 0
	t.B = ""
	t.C = t.C[:0]
	if t.D != nil {
		for k := range t.D {
			delete(t.D, k)
		}
	}
}

func TestPool_GetPut(t *testing.T) {
	// Создаём пул для *TestStruct
	p := New(func() *TestStruct {
		return &TestStruct{
			D: make(map[string]string),
		}
	})

	// Получаем объект из пула
	obj := p.Get()
	assert.NotNil(t, obj)

	// Заполняем его данными
	obj.A = 42
	obj.B = "hello"
	obj.C = append(obj.C, 1, 2, 3)
	obj.D["key"] = "value"

	// Возвращаем объект в пул
	p.Put(obj)

	// Берём объект снова — должен быть сброшен
	obj2 := p.Get()
	assert.NotNil(t, obj2)
	assert.Equal(t, 0, obj2.A)
	assert.Equal(t, "", obj2.B)
	assert.Len(t, obj2.C, 0)
	assert.Len(t, obj2.D, 0)
}
