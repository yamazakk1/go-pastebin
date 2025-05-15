package server

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/yamazakk1/go-pastebin/internal/app"
)

// Структура сервера
type Server struct {
	HTTPListener net.Listener
	ControlChan  chan struct{}
}

// Конструктор сервера
func NewServer() *Server {
	return &Server{
		ControlChan: make(chan struct{}),
	}
}

// Начало работы HTTP сервера
func (s *Server) Start(httpAddr string) error {
	var err error

	// Запуск HTTP сервера
	s.HTTPListener, err = net.Listen("tcp", httpAddr)
	if err != nil {
		return errors.New("ошибка запуска HTTP сервера")
	}
	defer s.HTTPListener.Close()

	fmt.Println("HTTP сервер запустился на порте ", httpAddr)

	// Инициализация HTTP обработчиков
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/paste/", s.handlePaste)
	mux.HandleFunc("/create", s.handleCreate)

	httpServer := &http.Server{
		Handler: mux,
	}

	// Запуск фоновой очистки паст
	app.Start()

	// Фоновая работа консоли
	go s.cmdReceiver()

	// Запуск HTTP сервера
	err = httpServer.Serve(s.HTTPListener)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

// HTTP обработчики

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Welcome to Pastebin\n\n")
	fmt.Fprintf(w, "Total pastes: %d\n", app.GetPasteCount())
	fmt.Fprintf(w, "Endpoints:\n")
	fmt.Fprintf(w, "POST /create - Create new paste\n")
	fmt.Fprintf(w, "GET /paste/{slug} - View paste\n")
	fmt.Fprintf(w, "Parameters for /create:\n")
	fmt.Fprintf(w, "  text: (required) Paste content\n")
	fmt.Fprintf(w, "  expires: (optional) Expiration time (e.g. 1h, 30m)\n")
}

func (s *Server) handlePaste(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	slug := strings.TrimPrefix(r.URL.Path, "/paste/")
	if slug == "" {
		http.Error(w, "Slug is required", http.StatusBadRequest)
		return
	}

	text, err := app.GetPaste(slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(text))
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "To create paste, send POST request with form data:\n")
		fmt.Fprintf(w, "text=Your+text+here\n")
		fmt.Fprintf(w, "expires=1h (optional)\n")
		return

	case http.MethodPost:
		text := r.FormValue("text")
		expiresStr := r.FormValue("expires")

		if text == "" {
			http.Error(w, "Text is required", http.StatusBadRequest)
			return
		}

		expires, err := time.ParseDuration(expiresStr)
		if err != nil || expires <= 0 {
			expires = 24 * time.Hour // дефолтное значение
		}

		slug, err := app.CreatePaste(text, expires)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Paste created successfully!\n")
		fmt.Fprintf(w, "Slug: %s\n", slug)
		fmt.Fprintf(w, "View URL: /paste/%s\n", slug)
		return

	default:
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Работа консоли на фоне
func (s *Server) cmdReceiver() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		cmd := scanner.Text()
		switch cmd {
		case "stop":
			close(s.ControlChan)
			fmt.Println("Shutting down server...")
			os.Exit(0)
		case "count":
			fmt.Printf("Total pastes: %d\n", app.GetPasteCount())
		default:
			fmt.Println("Available commands: stop, count")
		}
		fmt.Print("> ")
	}
}
