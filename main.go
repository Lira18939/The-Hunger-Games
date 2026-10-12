package main

import (
	"fmt"
	"net/http"
)

func main() {
	// Создаём маршрутизатор — он решает, какой обработчик вызвать для какого пути
	mux := http.NewServeMux()

	// Маршрут "/" — главная страница
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Привет! Это главная страница приложения «Учёт личных трат».")
	})

	// Маршрут "/about" — страница с описанием проекта
	mux.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "О проекте: это учебное веб-приложение для учёта личных трат, разработанное на Go.")
	})

	// Маршрут "/ping" — проверка работы сервера
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		// Если метод не GET — возвращаем 405
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintln(w, "405 Method Not Allowed")
			return
		}
		// Если метод GET — возвращаем 200 и слово pong
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "pong")
	})

	// Выводим сообщение в консоль, что сервер запускается
	fmt.Println("Сервер запущен на http://localhost:8080")

	// Запускаем сервер на порту 8080
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
	}
}