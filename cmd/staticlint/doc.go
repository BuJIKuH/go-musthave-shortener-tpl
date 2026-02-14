// Package main реализует multichecker staticlint.
//
// staticlint — это статический анализатор кода, объединяющий:
//   - стандартные анализаторы из golang.org/x/tools/go/analysis/passes;
//   - все анализаторы класса SA пакета honnef.co/go/tools/staticcheck;
//   - дополнительные анализаторы классов ST и QF;
//   - публичные сторонние анализаторы (errcheck, ineffassign);
//   - собственный анализатор noosexit, запрещающий вызов os.Exit в функции main.
//
// Запуск:
//
//	go run ./cmd/staticlint ./...
//
// Или после сборки:
//
//	go build -o staticlint ./cmd/staticlint
//	./staticlint ./...
//
// Анализатор предназначен для проверки кода сервиса сокращения URL
// и должен завершаться без ошибок на корректном коде проекта.
package main
