package distribuida

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// ip e porta
	ip    = "localhost"
	porta = ":1313"
)

type medidas

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func fibServerTCP() {
	// definte endpoint do servidor TCP
	r, err := net.ResolveTCPAddr("tcp", ip+porta)
	handleError(err)

	// cria listener TCP
	l, err := net.ListenTCP("tcp", r)
	handleError(err)

	fmt.Println("FibServerTCP aguardando conexões - Porta" + porta + "...")

	// loop infinito para aceitar conexões
	for {
		// aceita conexão
		conn, err := l.Accept()
		handleError(err)

		// programa fechamento da conexão
		defer func(conn net.Conn) {
			err := conn.Close()
			handleError(err)
		}(conn)

		// cria goroutine para tratar a conexão
		go handleConnection(conn)

	}

}

func fibServerUDP() {
	//buffers para receber requests e enviar respostas
	req := make([]byte, 1024)

	// define endpoint do servidor UDP
	r, err := net.ResolveUDPAddr("udp", ip+porta)
	handleError(err)

	// prepara UDP para receber requests
	conn, err := net.ListenUDP("udp", r)
	handleError(err)

	// programa fechamento de conn
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		handleError(err)
	}(conn)

	fmt.Println("FibServerUDP aguardando requests - Porta" + porta + "...")

	// loop infinito para receber requests
	for {
		// recebe request
		n, addr, err := conn.ReadFromUDP(req)
		handleError(err)

		// processa request
		handleUDPRequest(req, conn, n, addr)

	}

}

func handleConnection(conn net.Conn) {
	// buffers para receber requests e enviar respostas
	req := make([]byte, 1024)
	res := make([]byte, 1024)

	// loop infinito para receber requests
	for {
		// recebe request
		n, err := conn.Read(req)
		handleError(err)

		// processa o request
		handleRequest(req, res, n)

	}
}

func handleRequest(req []byte, res []byte, n int) {
	// converte o request para string
	strReq := string(req[:n])

	// converte a string para inteiro
	num, err := strconv.Atoi(strings.TrimSpace(strReq))
	handleError(err)

	// calcula o fibonacci
	fib := fibonacci(num)

	// converte o fibonacci para string
	strFib := strconv.Itoa(fib)

	// converte o fibonacci para array de bytes
	res = []byte(strFib)

	// envia a resposta
	time.Sleep(3 * time.Second)
	fmt.Println("Enviando resposta...")
	_, err = conn.Write(res)
	handleError(err)

}

func handleError(err error) {
	if err != nil {
		fmt.Println("Erro: ", err.Error())
		os.Exit(1) // encerra o programa com status 1 (erro)
	}
}

func handleUDPRequest(req []byte, conn *net.UDPConn, n int, addr *net.UDPAddr) {

}
