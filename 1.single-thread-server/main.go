package main

import "fmt"

// POSIX 란?
// 운영체제나 컴파일러에 종속되지 않고, 어떤 환경에서든 동일한 기능(시스템 콜 실행)을 수행할 수 있도록 만들어진
// 인터페이스와 코드의 집합(라이브러리)
func main() {
	// file descriptor 는 process 내부에서 사용하는 파일 식별자
	// 예를 들어 A 라는 프로세스와 B 라는 프로세스가 C 라는 파일에 접근하려고 할 때
	// 같은 C 라는 파일에 대해서 A 에서의 file descriptor 와 B 에서의 file descriptor 는 다르다.
	// 파일 입출력과 관련된 시스템 콜은, 첫번째 인자로 무조건 file descriptor 를 받는다.
	fmt.Println("Single Thread Socket Server")
}
