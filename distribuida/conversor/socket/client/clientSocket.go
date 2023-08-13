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

/*
	ORIENTAÇÕES PARA TESTES DE DESEMPENHO

	1. Métrica: Tempo Total de Execução
		- Tempo total de execução do cliente, desde a geração da primeira requisição até o recebimento da última resposta.
		- O tempo total de execução deve ser registrado no arquivo desempenhoTempoExecucao.csv.
		- Descomentar campos marcados com "TEMPO DE EXECUÇÃO" no código fonte.
		- Comentar campos marcados com "RTT" no código fonte.
	2. Métrica: RTT (Round-Trip Time)
		- Tempo de resposta de cada requisição.
		- O RTT deve ser registrado no arquivo desempenhoRTT.csv.
		- Descomentar campos marcados com "RTT" no código fonte.
		- Comentar campos marcados com "TEMPO DE EXECUÇÃO" no código fonte.
*/

var (
	fileLock sync.Mutex // Lock para acesso concorrente ao arquivo .csv

	//TEMPO DE EXECUÇÃO (Comentar/Descomentar para conforme testes desejados)
	//csvFile  = "desempenhoTempoExecucao.csv" // arquivo CSV para armazenar os resultados de desempenho
	//RTT (Comentar/Descomentar para conforme testes desejados)
	csvFile = "desempenhoRTT.csv" // arquivo CSV para armazenar os resultados de desempenho

	isClienteMedio bool   // indica se é o cliente médio, para registro do RTT
	protocolo      string // protocolo de comunicação
	clientes       string // quantidade de clientes
	idCliente      string // id do cliente
)

// args: protocolo, idCliente, invocacoes, fromUnit, toUnit
func main() {
	//testes()

	aviso := " necessita desses argumentos: \n\"tcp\" ou \"udp\"\n id do cliente\n " +
		"quantidade de invocações\n unidade de medida (origem)\n unidade de medida (destino)"

	protocolo = os.Args[1]

	if len(os.Args) == 7 { // verifica se todos os argumentos foram passados
		clientes = os.Args[2]
		idCliente = os.Args[3]
		invocacoes, _ := strconv.Atoi(os.Args[4])
		fromUnit := os.Args[5]
		toUnit := os.Args[6]

		// Para selecionar o cliente utilizado nos testes de desempenho RTT
		intClientes, _ := strconv.Atoi(clientes)
		intId, _ := strconv.Atoi(idCliente)
		isClienteMedio = intId == (intClientes/2)+1

		switch protocolo {
		case "tcp":
			//TEMPO DE EXECUÇÃO (Descomentar para testes por tempo de execução total)
			// para registrar tempo de execução total: t1, tTotal e writeToCsv
			//t1 := time.Now()
			ClientTCPConversor(invocacoes, fromUnit, toUnit)
			//tTotal := time.Since(t1)
			//writeToCSV(csvFile, tTotal)
		case "udp":
			//TEMPO DE EXECUÇÃO (Descomentar para testes por tempo de execução total)
			// para registrar tempo de execução total: t1, tTotal e writeToCsv
			//t1 := time.Now()
			ClientUDPConversor(invocacoes, fromUnit, toUnit)
			//tTotal := time.Since(t1)
			//writeToCSV(csvFile, tTotal)
		default:
			fmt.Println(os.Args[0], aviso)
			os.Exit(0)
		}
		//fmt.Scanln()

	} else { //arugmentos inválidos
		fmt.Println(os.Args[0], aviso)
	}

}

func writeToCSV(csvFile string, tTotal time.Duration) {
	// Converter duração
	//tTotalStr := fmt.Sprintf("%d", tTotal.Nanoseconds()/1000) // microsegundos
	tTotalStr := fmt.Sprintf("%d", tTotal.Nanoseconds()) // nanosegundos

	// imprimir no console
	fmt.Printf("%s,%s,%s,%s\n", protocolo, clientes, idCliente, tTotalStr)

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
	record := []string{protocolo, clientes, idCliente, tTotalStr}
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

	//RTT (Comentar/Descomentar para testar sem e com RTT)
	// variáveis para calcular RTT
	var rtt_inicio time.Time
	var rtt_total time.Duration

	// envia requests e recebe respostas
	for i := 0; i < invocacoes; i++ {
		//cria request
		msgToServer = GenerateRequest(fromUnit, toUnit)

		//RTT (Comentar/Descomentar para testar sem e com RTT)
		// calcula rtt
		rtt_inicio = time.Now()

		// serializa request com JSON (marshal) e envia para o servidor
		err = jsonEncoder.Encode(msgToServer)
		base.HandleError(err)

		// recebe resposta e decodifica com JSON (unmarshal)
		err = jsonDecoder.Decode(&msgFromServer)
		base.HandleError(err)

		//RTT (Comentar/Descomentar para testar sem e com RTT)
		// Verifica se é o cliente médio para registrar o RTT
		if isClienteMedio {
			rtt_total = time.Since(rtt_inicio)
			//fmt.Println("rtt_total", rtt_total.Nanoseconds())
			writeToCSV(csvFile, rtt_total)
		}

		//fmt.Println(msgFromServer)

		//RTT e TEMPO DE EXECUÇÃO (Comentar para não impactar no tempo aferido)
		//fmt.Printf("Conversão %d: %.2f %s para %s => %.2f %s\n", i, msgToServer.Valor,
		//	msgToServer.FromUnit, msgToServer.ToUnit, msgFromServer.Valor, msgFromServer.ToUnit)
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

	//RTT (Comentar/Descomentar para testar sem e com RTT)
	// variáveis para calcular RTT
	var rtt_inicio time.Time
	var rtt_total time.Duration

	// envia requests e recebe respostas
	for i := 0; i < invocacoes; i++ {
		//cria request
		msgToServer = GenerateRequest(fromUnit, toUnit)

		//RTT (Comentar/Descomentar para testar sem e com RTT)
		// calcula rtt
		rtt_inicio = time.Now()

		// serializa request com JSON (marshal) e envia para o servidor
		err = jsonEncoder.Encode(msgToServer)
		base.HandleError(err)

		// recebe resposta e decodifica com JSON (unmarshal)
		err = jsonDecoder.Decode(&msgFromServer)
		base.HandleError(err)

		//RTT (Comentar/Descomentar para testar sem e com RTT)
		// Verifica se é o cliente médio para registrar o RTT
		if isClienteMedio {
			rtt_total = time.Since(rtt_inicio)
			writeToCSV(csvFile, rtt_total)
		}

		//RTT e TEMPO DE EXECUÇÃO (Comentar para não impactar no tempo aferido)
		//fmt.Printf("Conversão %d: %.2f %s para %s => %.2f %s\n", i, msgToServer.Valor,
		//	msgToServer.FromUnit, msgToServer.ToUnit, msgFromServer.Valor, msgFromServer.ToUnit)
	}
}
