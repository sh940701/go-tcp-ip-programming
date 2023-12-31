package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
	"syscall"
)

func main() {
	fmt.Println("TCP Client")

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		log.Fatalln("Error in syscall.Socket:", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	ch := make(chan os.Signal, 1)
	go handleSignal(fd, ch, &wg)

	serverAddr := &syscall.SockaddrInet4{Port: 8080} // 8080 포트
	serverAddr.Addr = [4]byte{127, 0, 0, 1}          // localhost

	err = syscall.Connect(fd, serverAddr)
	if err != nil {
		log.Fatalln("Error in syscall.Connect:", err)
	}

	fmt.Println("Client socket", fd, "is connected")

	for {
		fmt.Print("Enter message: ")

		scanner := bufio.NewScanner(os.Stdin)

		var message string

		if scanner.Scan() {
			message = scanner.Text()
		}
		if err != nil {
			log.Fatalln("Error in fmt.Scanf", err)
		}

		_, err = syscall.Write(fd, []byte(message))
		if err != nil {
			log.Fatalln("Error in syscall.Write:", err)
		}

		if message == "quit" {
			ch <- syscall.SIGINT
			wg.Wait()
			break
		}

		buf := make([]byte, 1024)
		n, err := syscall.Read(fd, buf)
		if err != nil {
			log.Fatalln("Error in syscall.Read:", err)
		}

		fmt.Printf("Server: %s\n", string(buf[:n]))
	}
}

func handleSignal(fd int, ch chan os.Signal, wg *sync.WaitGroup) {
	defer wg.Done()

	sig := <-ch
	fmt.Println("\nReceived signal:", sig)
	err := syscall.Close(fd)
	if err != nil {
		log.Fatalln("Error in syscall.Close:", err)
	}

	fmt.Println("socket", fd, "is closed.")
	os.Exit(0)
}
