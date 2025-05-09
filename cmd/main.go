package main

import (
	"fmt"
	"time"

	"github.com/yamazakk1/go-pastebin/internal/app"
)

func main() {
	app.Start()
	defer app.Stop()

	fmt.Println("Программа запущена. Команды: create, count, stop")

	var input string
	for {
		fmt.Print("> ")
		fmt.Scanln(&input)

		switch input {
		case "stop":
			fmt.Println("Завершение работы программы")
			return
		case "create":
			slug, err := app.CreatePaste("Тестовая паста", 10*time.Second)
			if err != nil {
				fmt.Println("Ошибка:", err)
			} else {
				fmt.Println("Создана паста с slug:", slug)
			}
		case "count":
			fmt.Println("Текущее количество паст:", app.GetPasteCount())
		default:
			fmt.Println("Неизвестная команда")
		}
	}
}
