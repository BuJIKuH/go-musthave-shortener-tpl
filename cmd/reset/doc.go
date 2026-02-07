// Package reset реализует утилиту генерации методов Reset() для структур,
// помеченных комментарием // generate:reset.
//
// Утилита сканирует все пакеты проекта, находит структуры с пометкой
// и создаёт файл reset.gen.go с методами Reset(), которые сбрасывают
// поля структуры к нулевым значениям.
//
// Пример использования:
//
//	// Запуск генератора из командной строки
//	go run ./cmd/reset
//
// Для структуры:
//
//	// generate:reset
//	type ResetableStruct struct {
//	    i     int
//	    str   string
//	    strP  *string
//	    s     []int
//	    m     map[string]string
//	    child *ResetableStruct
//	}
//
// Утилита создаст метод:
//
//	func (s *ResetableStruct) Reset() {
//	    if s == nil {
//	        return
//	    }
//	    s.i = 0
//	    s.str = ""
//	    if s.strP != nil {
//	        *s.strP = ""
//	    }
//	    s.s = s.s[:0]
//	    if s.m != nil {
//	        for k := range s.m {
//	            delete(s.m, k)
//	        }
//	    }
//	    if s.child != nil {
//	        if resetter, ok := interface{}(s.child).(interface{ Reset() }); ok {
//	            resetter.Reset()
//	        }
//	    }
//	}
package main
