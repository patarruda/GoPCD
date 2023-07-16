package main

import (
	//"conversor/base"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"pcd/distribuida/conversor/base"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "tcp":
			ServerTCPConversor()
		case "udp":
			ServerUDPConversor()
		default:
			fmt.Println(os.Args[0], " necessita de um desses argumentos: \"tcp\" ou \"udp\".")
			os.Exit(0)
		}
		fmt.Scanln()
	} else {
		fmt.Println(os.Args[0], " necessita de um desses argumentos: \"tcp\" ou \"udp\".")
	}
}

func ServerTCPConversor() {
	// definte endpoint do servidor TCP
	r, err := net.ResolveTCPAddr("tcp", base.HOST+base.PORTA_TCP)
	base.HandleError(err)

	// cria listener TCP
	l, err := net.ListenTCP("tcp", r)
	base.HandleError(err)

	fmt.Println("##### CONVERSOR DE MEDIDAS #####\n\nServerTCP aguardando conexões - Porta" +
		base.PORTA_TCP + "...")

	// loop infinito para aceitar conexões
	for {
		// aceita conexão (fechamento da conexão é feito pela goroutine)
		fmt.Println("Aguardando conexão...")
		conn, err := l.Accept()
		fmt.Println("Conexão estabelecida...")
		base.HandleError(err)

		// cria goroutine para tratar a conexão
		go handleTCPConnection(conn)
		fmt.Println("goroutine criada... handleTCPConnection")

	}

}

func ServerUDPConversor() {
	//buffers para receber requests e enviar respostas
	req := make([]byte, 1024)

	// define endpoint do servidor UDP
	r, err := net.ResolveUDPAddr("udp", base.HOST+base.PORTA_UDP)
	base.HandleError(err)

	// prepara UDP para receber requests
	conn, err := net.ListenUDP("udp", r)
	base.HandleError(err)

	// programa fechamento de conn
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		base.HandleError(err)
	}(conn)

	fmt.Println("##### CONVERSOR DE MEDIDAS #####\n\nServerUDP aguardando requests - Porta" +
		base.PORTA_UDP + "...")

	// loop infinito para receber requests de diversos clientes
	for {
		// recebe request
		n, addr, err := conn.ReadFromUDP(req)
		base.HandleError(err)

		// processa request
		handleUDPRequest(req, n, conn, addr)

	}

}

func handleTCPConnection(conn net.Conn) {
	// variáveis para receber requests e enviar respostas
	var msgFromClient base.RequestConversor
	var msgToClient base.RequestConversor

	// programa fechamento da conexão
	defer func(conn net.Conn) {
		err := conn.Close()
		base.HandleError(err)
		fmt.Println("Conexão encerrada. Aguardando próximo cliente...")
	}(conn)

	// cria enconder/decoder JSON
	jsonDecoder := json.NewDecoder(conn)
	jsonEncoder := json.NewEncoder(conn)

	// loop infinito para receber requests do cliente
	for {
		// recebe request e decodifica com JSON (unmarshal)
		err := jsonDecoder.Decode(&msgFromClient)
		if err != nil && err.Error() == "EOF" {
			fmt.Println("Cliente finalizou requests...")
			//conn.Close() // O fechamento da conexão será feito pelo "defer"
			break
		}

		//fmt.Println("msgFromClient: ", msgFromClient)

		// processa o request e cria a resposta para o cliente
		msgToClient = base.ConversorMedidas{}.Invoke(msgFromClient)
		//fmt.Println(msgToClient)
		//resultado := base.ConversorMedidas{}.Invoke(msgFromClient)
		//msgToClient = base.RequestConversor{Valor: resultado}

		// serializa resposta com JSON (marshal) e envia para o cliente
		err = jsonEncoder.Encode(msgToClient)
		base.HandleError(err)

		//fmt.Println("msgToClient: ", msgToClient)

	}
}

func handleUDPRequest(msgFromClient []byte, n int, conn *net.UDPConn, addr *net.UDPAddr) {
	var msgToClient []byte
	var request base.RequestConversor
	var reply base.RequestConversor

	// decodifica request com JSON (unmarshal)
	err := json.Unmarshal(msgFromClient[:n], &request)
	base.HandleError(err)

	// processa o request e cria a resposta para o cliente
	// resultado := base.ConversorMedidas{}.Invoke(request)
	// reply = base.RequestConversor{Valor: resultado}
	reply = base.ConversorMedidas{}.Invoke(request)

	// serializa resposta com JSON (marshal)
	msgToClient, err = json.Marshal(reply)
	base.HandleError(err)

	// envia resposta para o cliente
	_, err = conn.WriteToUDP(msgToClient, addr)
	base.HandleError(err)

}
