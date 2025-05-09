package app

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

type Paste struct {
	Text      string
	Slug      string
	ExpiresAt time.Time
}

var alphabet = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890"
var storage = make(map[string]Paste)
var mu sync.RWMutex
var isrunning bool

// CreatePaste - создать новую пасту
func CreatePaste(text string, expiresAt time.Duration) (string, error) {
	if text == "" {
		return "", errors.New("попытка записи пустой строки")
	}
	mu.Lock()
	defer mu.Unlock()
	slug, _ := CreateSlug(10)
	newPaste := Paste{Text: text, Slug: slug, ExpiresAt: time.Now().Add(expiresAt)}
	storage[slug] = newPaste
	return slug, nil
}

// CreateSlug - создание слага
func CreateSlug(n int) (string, error) {
	if n < 3 {
		return "", errors.New("слишком короткий slug")
	}
	bytes := make([]byte, n)
	for i := 0; i < n; i++ {
		bytes[i] = alphabet[(rand.Intn(len(alphabet)))]
	}
	return string(bytes), nil
}

// GetPaste - получение пасты
func GetPaste(slug string) (string, error) {
	mu.RLock()
	if _, ok := storage[slug]; !ok {
		return "", errors.New("такого slug-а не существует")
	}

	if time.Now().After(storage[slug].ExpiresAt) {
		go deleteExpiredPaste(slug)
		return "", errors.New("паста удалена")
	}

	defer mu.RUnlock()
	return storage[slug].Text, nil
}

// DeleteExpiredPaste - удаление просроченной пасты
func deleteExpiredPaste(slug string) {
	mu.Lock()
	defer mu.Unlock()
	delete(storage, slug)
}

// ExpirienceCheck - проверка паст на проссроченность
func ExpirienceCheck() {
	now := time.Now()
	mu.RLock()
	expired := make([]string, 0)
	for key, value := range storage {
		if now.After(value.ExpiresAt) {
			expired = append(expired, key)
		}
	}
	mu.RUnlock()

	for _, key := range expired {
		deleteExpiredPaste(key)
	}
}

// Запуск фоновой очистки паст с интервалом в 30 сек
func StartBackgroundCleanUp() {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ExpirienceCheck()
			}
		}
	}()
}

// Начало работы программы
func Start() {
	if isrunning {
		return
	}
	isrunning = true

	go StartBackgroundCleanUp()
}

// Конец работы программы к
func Stop() {
	if !isrunning {
		return
	}
	isrunning = false
}

// ВЫвод количества паст
func GetPasteCount() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(storage)
}
