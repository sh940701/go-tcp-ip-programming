# http 서버 구현

## 개요

본 구현에서는 http 1.1 기반의 서버를 구현합니다.

## 기능

- 클라이언트가 `GET /` 요청을 하면 `index.html` 을 클라이언트에게 전송합니다.

## 제약사항

- 표준 라이브러리만을 사용하여 구현합니다.

- 멀티 스레드 구현을 위하여 [Gorutine](https://go.dev/tour/concurrency/1) 문법을 사용합니다.

- `net`, `net/http` 패키지 대신, `syscall` 패키지를 이용하여 직접
  [POSIX API](https://docs.oracle.com/cd/E19048-01/chorus5/806-6897/auto1/index.html)를 호출합니다.

- `GET /` 요청을 이외에는 `<h1>404 NOT FOUND</h1>` 을 반환합니다.

- curl, postman, 웹 브라우저 등 다양한 웹 에이전트를 사용하여 테스트합니다.
