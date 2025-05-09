package server

import (
	"errors"
	"fmt"
	"net"
	"time"

	//"sync"
	"bufio"
	"os"

	"github.com/yamazakk1/go-pastebin/internal/app"
)

// Структура сервера
type Server struct {
	Storage     map[string]app.Paste
	Listener    net.Listener
	ControlChan chan struct{}
}

// Конструктор сервера
func NewServer() *Server {
	return &Server{Storage: make(map[string]app.Paste), ControlChan: make(chan struct{})}
}

// Начало работы сервера
func (s *Server) Start(addr string) error {
	var err error
	s.Listener, err = net.Listen("tcp", addr)
	if err != nil {
		return errors.New("ошибка запуска сервера")
	}
	defer s.Listener.Close()
	fmt.Println("Сервер запустился на порте ", addr)
	// фоновая работа консоли
	go s.cmdReceiver()

	for {
		select {
		case _, ok := <-s.ControlChan:
			if !ok {
				fmt.Println("Получена команда остановки. Завершение работы...")
				return nil
			}

		default:
			// Устанавливаем таймаут для Accept, чтобы не блокироваться навсегда
			s.Listener.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second))
			conn, err := s.Listener.Accept()
			if err != nil {
				continue
			}
			fmt.Println("Сервер установил соединение")
			conn.Close()
		}
	}
}

// Работа консоли на фоне. Общение с сервером через ControlChan
func (s *Server) cmdReceiver() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(">")
	for scanner.Scan() {
		if scanner.Text() == "stop" {
			close(s.ControlChan)
			return
		}
		fmt.Print(">")
	}
}

// func (s *Server) handleRequest(conn net.Conn) {

// 	buf := make([]byte, 1024)
// 	_, err := conn.Read(buf)
// 	if err != nil {
// 		fmt.Printf("Ошибка чтения: %v\n", err)
// 		return
// 	}
// 	conn.Write([]byte("OK\n"))
// }
