package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"pcd/distribuida/conversor/base"
	"strconv"
	"sync"
	"time"
)

var (
	fileLock sync.Mutex         // Lock para acesso concorrente ao arquivo
	csvFile  = "desempenho.csv" // arquivo CSV para armazenar os resultados de desempenho
)

// args: protocolo, idCliente, invocacoes, fromUnit, toUnit
func main() {
	//testes()

	aviso := " necessita desses argumentos: \n\"tcp\" ou \"udp\"\n id do cliente\n " +
		"quantidade de invocações\n unidade de medida (origem)\n unidede de medida (destino)"

	protocolo := os.Args[1]

	if len(os.Args) == 6 { // verifica se todos os argumentos foram passados
		idCliente := os.Args[2]
		invocacoes, _ := strconv.Atoi(os.Args[3])
		fromUnit := os.Args[4]
		toUnit := os.Args[5]

		switch protocolo {
		case "tcp":
			t1 := time.Now()
			ClientTCPConversor(invocacoes, fromUnit, toUnit)
			tTotal := time.Since(t1)
			writeToCSV(protocolo, idCliente, tTotal)
		case "udp":
			t1 := time.Now()
			ClientUDPConversor(invocacoes, fromUnit, toUnit)
			tTotal := time.Since(t1)
			writeToCSV(protocolo, idCliente, tTotal)
		default:
			fmt.Println(os.Args[0], aviso)
			os.Exit(0)
		}
		//fmt.Scanln()

	} else { //arugmentos inválidos
		fmt.Println(os.Args[0], aviso)
	}

}

func writeToCSV(protocolo, idCliente string, tTotal time.Duration) {
	// Converter duração para milliseconds
	millisecondsStr := fmt.Sprintf("%d", tTotal.Milliseconds())

	// imprimir no console
	fmt.Printf("%s,%s,%s\n", protocolo, idCliente, millisecondsStr)

	// Lock para acesso concorrente seguro ao arquivo
	fileLock.Lock()
	defer fileLock.Unlock()

	// Abrir arquivo CSV
	file, err := os.OpenFile(csvFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// 0666 = file permissions (rw-rw-rw-)
	if err != nil {
		fmt.Println("Erro ao abrir o arquivo:", csvFile, err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	// Write protocolo, idCliente, and tTotal to CSV file
	record := []string{protocolo, idCliente, millisecondsStr}
	err = writer.Write(record)
	if err != nil {
		fmt.Println("Erro ao escrever no arquivo:", csvFile, err)
		return
	}

	// Flush to disk (write to file)
	writer.Flush()
	if err := writer.Error(); err != nil {
		fmt.Println("Erro ao dar flush para o disco:", csvFile, err)
		return
	}
}

/*
func testes() {
	go main.ServerTCPConversor()
	go main.ServerUDPConversor()

	fmt.Scanln()

	//Teste 01 - TCP Conversão CelsiusToFahrenheit
	go ClientTCPConversor(10, "C", "F")

	fmt.Scanln()

	//Teste 02 - TCP Conversão não suportada
	go ClientTCPConversor(10, "k", "F")

	fmt.Scanln()

	//Teste 03 - UDP Conversão MetersToFeet
	go ClientUDPConversor(10, "M", "FT")

	fmt.Scanln()
}
*/

func GenerateRequest(fromoUnit string, toUnit string) base.RequestConversor {
	// gerar um float64 aleatório até 40.0
	rand.Seed(time.Now().UnixNano())
	n := rand.Float64() * 40.0
	//cria request
	return base.RequestConversor{Valor: n,
		FromUnit: fromoUnit,
		ToUnit:   toUnit} //CelsiusToFarhrenheit
}

func ClientTCPConversor(invocacoes int, fromUnit string, toUnit string) {
	// retorna endpoint do servidor TCP
	r, err := net.ResolveTCPAddr("tcp", base.HOST+base.PORTA_TCP)
	base.HandleError(err)

	// conecta ao servidor (porta local é alocada automaticamente)
	conn, err := net.DialTCP("tcp", nil, r)
	base.HandleError(err)

	// programa fechamento da conexão
	defer func(conn *net.TCPConn) {
		err := conn.Close()
		base.HandleError(err)
	}(conn)

	// cria enconder/decoder JSON
	jsonDecoder := json.NewDecoder(conn)
	jsonEncoder := json.NewEncoder(conn)

	// variáveis para receber requests e enviar respostas
	var msgFromServer base.RequestConversor
	var msgToServer base.RequestConversor

	// envia requests e recebe respostas
	for i := 0; i < invocacoes; i++ {
		//cria request
		msgToServer = GenerateRequest(fromUnit, toUnit)

		// serializa request com JSON (marshal) e envia para o servidor
		err = jsonEncoder.Encode(msgToServer)
		base.HandleError(err)

		// recebe resposta e decodifica com JSON (unmarshal)
		err = jsonDecoder.Decode(&msgFromServer)
		base.HandleError(err)

		//fmt.Println(msgFromServer)

		fmt.Println(fmt.Sprintf("Conversão %d: %.2f %s para %s => %.2f %s", i, msgToServer.Valor,
			msgToServer.FromUnit, msgToServer.ToUnit, msgFromServer.Valor, msgFromServer.ToUnit))
	}

}

func ClientUDPConversor(invocacoes int, fromUnit string, toUnit string) {
	// resolver endereço do servidor UDP
	r, err := net.ResolveUDPAddr("udp", base.HOST+base.PORTA_UDP)
	base.HandleError(err)

	// conecta ao servidor (porta local é alocada automaticamente) - UDP é sem conexão
	conn, err := net.DialUDP("udp", nil, r)
	base.HandleError(err)

	// programa fechamento da conexão
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		base.HandleError(err)
	}(conn)

	// cria enconder/decoder JSON
	jsonDecoder := json.NewDecoder(conn)
	jsonEncoder := json.NewEncoder(conn)

	// variáveis para receber requests e enviar respostas
	var msgFromServer base.RequestConversor
	var msgToServer base.RequestConversor

	// envia requests e recebe respostas
	for i := 0; i < invocacoes; i++ {
		//cria request
		msgToServer = GenerateRequest(fromUnit, toUnit)

		// serializa request com JSON (marshal) e envia para o servidor
		err = jsonEncoder.Encode(msgToServer)
		base.HandleError(err)

		// recebe resposta e decodifica com JSON (unmarshal)
		err = jsonDecoder.Decode(&msgFromServer)
		base.HandleError(err)

		fmt.Println(fmt.Sprintf("Conversão %d: %.2f %s para %s => %.2f %s", i, msgToServer.Valor,
			msgToServer.FromUnit, msgToServer.ToUnit, msgFromServer.Valor, msgFromServer.ToUnit))
	}
}
