package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Multi Thread Socket Server")

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		log.Fatalln("Error in syscall.Socket")
	}

	defer func() {
		err = syscall.Close(fd)
		if err != nil {
			log.Fatalln("Error in syscall.Close:", err)
		}
		fmt.Println("socket", fd, "is closed.")
	}()

	ch := make(chan os.Signal, 1)
	go handleSignal(fd, ch)

	sockAddr := &syscall.SockaddrInet4{Port: 8080}

	sockAddr.Addr = [4]byte{0, 0, 0, 0}

	err = syscall.Bind(fd, sockAddr)

	err = syscall.Listen(fd, 10)
	if err != nil {
		log.Fatalln("Error in syscall.Listen:", err)
	}

	fmt.Println("Listening on", "localhost:8080")
	fmt.Println("Server Socket:", fd)

	for {
		clientFd, sockAddr, err := syscall.Accept(fd)
		if err != nil {
			log.Fatalln("Error in syscall.Accept:", err)
		}
		fmt.Println("Client Socket:", clientFd)

		go handleConnection(clientFd, sockAddr.(*syscall.SockaddrInet4))
	}
}

func handleConnection(fd int, sockAddr *syscall.SockaddrInet4) {
	defer func() {
		err := syscall.Close(fd)
		if err != nil {
			log.Fatalln("Error in syscall.Close:", err)
		}
		fmt.Println("socket", fd, "is closed")
	}()

	buf := make([]byte, 1024)

	for {
		n, err := syscall.Read(fd, buf)
		if err != nil {
			log.Fatalln("Error in syscall.Read:", err)
		}

		if n == 0 {
			return
		}

		clientIP := fmt.Sprintf("%d.%d.%d.%d", sockAddr.Addr[0], sockAddr.Addr[1], sockAddr.Addr[2], sockAddr.Addr[3])

		fmt.Printf("Received: %s\nFrom %s - Socket: %d\n---\n\n", string(buf[:n]), clientIP, fd)

		_, err = syscall.Write(fd, buf[:n])
		if err != nil {
			log.Fatalln("Error in syscall.Write:", err)
		}
	}
}

func handleSignal(fd int, ch chan os.Signal) {
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	sig := <-ch
	fmt.Println("\nReceived signal:", sig)

	err := syscall.Close(fd)
	if err != nil {
		log.Fatalln("Error in syscall.Close:", err)
	}

	fmt.Println("socket", fd, "is closed.")
	os.Exit(0)
}
