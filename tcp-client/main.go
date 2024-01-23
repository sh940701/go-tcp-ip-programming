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

	// socket file descriptor 생성
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		log.Fatalln("Error in syscall.Socket:", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	ch := make(chan os.Signal, 1)
	go handleSignal(fd, ch, &wg)

	// 소켓이 connection 을 보낼 주소값을 만들어준다.
	serverAddr := &syscall.SockaddrInet4{Port: 8080} // 8080 포트
	serverAddr.Addr = [4]byte{127, 0, 0, 1}          // localhost

	// 서버에 연결한다.
	err = syscall.Connect(fd, serverAddr)
	if err != nil {
		log.Fatalln("Error in syscall.Connect:", err)
	}

	fmt.Println("Client socket", fd, "is connected")

	for {
		// user input 을 받는다.
		fmt.Print("Enter message: ")

		// 이 때 golang 에서는 Scan, Scanf 를 통해서 \n 을 걸러내는 방식을 지원하지 않기 때문에 띄어쓰기를 만나면 input 을 종료해버린다.
		// 이를 해결하기 위해서 bufio.NewScanner 를 사용한다.
		scanner := bufio.NewScanner(os.Stdin)

		var message string

		if scanner.Scan() {
			message = scanner.Text()
		}
		if err != nil {
			log.Fatalln("Error in fmt.Scanf", err)
		}

		// server 에 input 을 message 로 보낸다.
		// 이 때 이 작업은, file 에 data 를 write 하는것과 같다.
		// 다만 write 하는 file 이 socket 이기 때문에, 네트워크 상으로 전송하는 action 을 취하게 되는 것이다.
		_, err = syscall.Write(fd, []byte(message))
		if err != nil {
			log.Fatalln("Error in syscall.Write:", err)
		}

		if message == "quit" {
			ch <- syscall.SIGINT
			wg.Wait()
			break
		}

		// 서버로부터의 응답을 받아 버퍼에 저장한다.
		//var buf []byte
		buf := make([]byte, 1024)
		n, err := syscall.Read(fd, buf)
		if err != nil {
			log.Fatalln("Error in syscall.Read:", err)
		}

		// 서버로부터 받은 응답을 출력한다.
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
