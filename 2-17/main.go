package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

func main() {
	timeout := flag.Duration("timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Println("Usage: go run main.go [--timeout=10s] host port")
		os.Exit(1)
	}

	host := flag.Arg(0)
	port := flag.Arg(1)
	address := net.JoinHostPort(host, port)

	// Устанавливаем соединение с таймаутом
	conn, err := net.DialTimeout("tcp", address, *timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Connection error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("Connected to %s\n", address)

	var wg sync.WaitGroup
	wg.Add(2)

	// Канал для сигнализации завершения
	done := make(chan struct{})

	// === STDIN -> SOCKET ===
	go func() {
		defer wg.Done()

		reader := bufio.NewReader(os.Stdin)

		for {
			select {
			case <-done:
				return
			default:
				data, err := reader.ReadBytes('\n')
				if err != nil {
					if err == io.EOF {
						// Ctrl+D
						fmt.Println("EOF received, closing connection...")
						conn.Close()
						close(done)
						return
					}
					fmt.Fprintf(os.Stderr, "Read stdin error: %v\n", err)
					close(done)
					return
				}

				_, err = conn.Write(data)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Write to socket error: %v\n", err)
					close(done)
					return
				}
			}
		}
	}()

	// === SOCKET -> STDOUT ===
	go func() {
		defer wg.Done()

		reader := bufio.NewReader(conn)

		for {
			select {
			case <-done:
				return
			default:
				data, err := reader.ReadBytes('\n')
				if err != nil {
					if err == io.EOF {
						fmt.Println("\nServer closed connection")
						close(done)
						return
					}
					fmt.Fprintf(os.Stderr, "Read from socket error: %v\n", err)
					close(done)
					return
				}

				fmt.Print(string(data))
			}
		}
	}()

	// Ждём завершения обеих горутин
	wg.Wait()

	fmt.Println("Connection closed")
}
