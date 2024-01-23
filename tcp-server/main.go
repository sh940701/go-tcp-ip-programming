package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

// POSIX 란?
// 운영체제나 컴파일러에 종속되지 않고, 어떤 환경에서든 동일한 기능(시스템 콜 실행)을 수행할 수 있도록 만들어진
// 인터페이스와 코드의 집합(라이브러리)
func main() {
	// file descriptor 는 process 내부에서 사용하는 파일 식별자
	// 예를 들어 A 라는 프로세스와 B 라는 프로세스가 C 라는 파일에 접근하려고 할 때
	// 같은 C 라는 파일에 대해서 A 에서의 file descriptor 와 B 에서의 file descriptor 는 다르다.
	// 파일 입출력과 관련된 시스템 콜은, 첫번째 인자로 무조건 file descriptor 를 받는다.
	fmt.Println("Multi Thread Socket Server")

	// file descriptor (Socket)를 만드는 과정
	// AF_INET 은 IPv4 를, SOCK_STREAM 은 TCP 통신을 의미한다.
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		log.Fatalln("Error in syscall.Socket")
	}

	defer func() {
		// file descriptor 를 닫는다.
		// defer 로 지연평가를 사용해서 모든 함수가 끝난 후, 소켓을 닫는다.
		err = syscall.Close(fd)
		if err != nil {
			log.Fatalln("Error in syscall.Close:", err)
		}
		fmt.Println("socket", fd, "is closed.")
	}()

	ch := make(chan os.Signal, 1)
	go handleSignal(fd, ch)

	// 소켓 수준에서의 address(식별자 = port number) 를 지정해준다.
	sockAddr := &syscall.SockaddrInet4{Port: 8080}

	// 모든 IPv4 인터페이스로부터의 접근을 허용한다.
	//copy(sockAddr.Addr[:], []byte{0,0,0,0})
	sockAddr.Addr = [4]byte{0, 0, 0, 0}

	// socket 에 주소를 할당해준다.
	err = syscall.Bind(fd, sockAddr)

	// socket 을 listen 상태로 만들어준다.
	// 이 때 두번째 인자는, 한 순간에 접속할 수 있는 client 의 숫자이다.(buffer)
	// 이 말은 소켓이 10개까지만 열린다는 것이 아니라, 한 순간에 100 개의 client 가 접근을 요청하면
	// 10개까지만 수용을 한다는것이다.
	// 그러나 100개의 client 가 순차적으로 접근을 요청하면 이는 활성화 될 수 있다.
	err = syscall.Listen(fd, 10)
	if err != nil {
		log.Fatalln("Error in syscall.Listen:", err)
	}

	fmt.Println("Listening on", "localhost:8080")
	fmt.Println("Server Socket:", fd)

	for {
		// syscall.Accept 함수는 블로킹 함수로, client 로부터의 요청이 있을 때 까지 for 문은 진행되지 않고
		// 멈춰있게 된다.
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

	// incoming data 를 담기 위한 buffer 를 만들어준다.
	buf := make([]byte, 1024)
	zeroed := make([]byte, 1024)

	for {
		// client 로 부터 데이터를 읽어온다.
		// syscall.Read 함수도 블로킹함수로, client 로 부터 데이터가 들어올때까지 for문이 실행되지 않고 기다리게 된다.
		n, err := syscall.Read(fd, buf)
		if err != nil {
			log.Fatalln("Error in syscall.Read:", err)
		}

		if n == 0 || string(buf[:n]) == "quit" {
			return
		}

		clientIP := fmt.Sprintf("%d.%d.%d.%d", sockAddr.Addr[0], sockAddr.Addr[1], sockAddr.Addr[2], sockAddr.Addr[3])

		fmt.Printf("Received: %s\nFrom %s - Socket: %d\n---\n\n", string(buf[:n]), clientIP, fd)

		if checkContainsOperator(string(buf)) == true {
			result, err := calculate(string(buf[:n]))
			if err != nil {
				log.Fatalln("Error in calculate:", err)
			}

			data := []byte(result)

			_, err = syscall.Write(fd, data)
			if err != nil {
				log.Fatalln("Error in syscall.Write:", err)
			}
		} else {
			_, err = syscall.Write(fd, buf[:n])
			if err != nil {
				log.Fatalln("Error in syscall.Write:", err)
			}
		}

		// 기존에는 buf 를 그대로 두고 들어온 데이터의 길이까지만 읽어서 반환을 했었다.
		// 그러나 이번 구현에서는 계산 식인 경우가 있기 때문에, buffer 를 초기화 해줘야 한다.
		// 이 때 buf = buf[:0] 의 방식으로 초기화를 해주면, 버퍼의 길이 자체가 0이 되어버려서
		// block 되지 않고 for 문을 계속 반복한다.
		// 따라서 같은 크기의 빈 버퍼를 만들어두고, 이를 계속 복사하는 방식으로 초기화를 해 준다.
		copy(buf, zeroed)
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

func calculate(message string) (string, error) {
	messageSplit := strings.Split(message, " ")
	operator := messageSplit[1]

	var result string

	left, err := strconv.ParseFloat(messageSplit[0], 64)
	if err != nil {
		return result, err
	}

	right, err := strconv.ParseFloat(messageSplit[2], 64)
	if err != nil {
		return result, err
	}

	switch operator {
	case "+":
		result = fmt.Sprintf("%f", left+right)
		return result, nil
	case "-":
		result = fmt.Sprintf("%f", left-right)
		return result, nil
	case "*":
		result = fmt.Sprintf("%f", left*right)
		return result, nil
	case "/":
		if right == 0 {
			return result, errors.New("division by zero is not allowed")
		}
		result = fmt.Sprintf("%f", left/right)
		return result, nil
	}

	return result, errors.New("unknown Error")
}

func checkContainsOperator(message string) bool {
	operators := []string{"+", "-", "*", "/"}
	for _, op := range operators {
		if strings.Contains(message, op) {
			return true
		}
	}

	return false
}
