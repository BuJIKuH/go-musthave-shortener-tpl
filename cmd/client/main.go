// Package main запускает CLI-клиент для сервиса сокращения URL.
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// run выполняет логику CLI-клиента и возвращает ошибку при неудаче.
func run() error {
	endpoint := "http://localhost:8080/"

	// контейнер данных для запроса
	data := url.Values{}

	// приглашение в консоли
	fmt.Println("Введите длинный URL:")

	// читаем строку из консоли
	reader := bufio.NewReader(os.Stdin)
	long, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("не удалось прочитать URL из консоли: %w", err)
	}
	long = strings.TrimSpace(long)

	// заполняем контейнер данными
	data.Set("url", long)

	// создаем HTTP-клиент
	client := &http.Client{}

	// создаем POST-запрос
	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("не удалось создать HTTP-запрос: %w", err)
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// выполняем запрос
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	defer response.Body.Close()

	// выводим код ответа
	fmt.Println("Статус-код:", response.Status)

	// читаем тело ответа
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("не удалось прочитать тело ответа: %w", err)
	}

	// печатаем тело ответа
	fmt.Println(string(body))

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Println("Ошибка:", err)
	}
}
