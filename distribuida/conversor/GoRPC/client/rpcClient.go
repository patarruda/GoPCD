package main

import (
	"encoding/csv"
	"fmt"
	"net/rpc"
	"os"
	"pcd/distribuida/conversor/base"
	"strconv"
	"sync"
	"time"
)

var wg sync.WaitGroup
var (
	clientes       string  // quantidade total de clientes
	clientID       int     // id do cliente atual
	numRequests    = 10000 // quantidade de requisições por cliente
	isClienteMedio bool    // indica se é o cliente médio, para registro do RTT

	fileLock sync.Mutex                         // Lock para acesso concorrente ao arquivo .csv
	csvFile  = "desempenhoRPC_clienteMedio.csv" // arquivo CSV para armazenar os resultados de desempenho
	//csvFile = "desempenhoRPC_TempoTotal.csv" // arquivo CSV para armazenar os resultados de desempenho
)

func main_() {
	if len(os.Args) != 2 {
		fmt.Println("Uso: go run rpc_client.go <número_de_clientes>")
		return
	}

	numClients, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Erro ao converter o número de clientes:", err)
		return
	}

	//numRequests := 10000 // Número de solicitações por cliente

	file, err := os.Create("desempenhoRPC.csv")
	if err != nil {
		fmt.Println("Erro ao criar o arquivo CSV:", err)
		return
	}
	defer file.Close()

	//writer = csv.NewWriter(file)

	//defer writer.Flush()

	//wg.Add(numClients)
	//for i := 0; i < numClients; i++ {
	//	go runClient(i, numRequests, writer)
	//}

	wg.Wait()

	// Aguardar até que todos os clientes terminem
	time.Sleep(time.Duration(numClients) * time.Second)
}

func main() {

	// Verificar se os argumentos foram passados corretamente
	if len(os.Args) != 3 {
		fmt.Println("Uso: go run rpc_client.go <clientes> <id_do_cliente>")
		return
	}

	// Setar as variáveis globais
	var err error
	var intClientes int
	clientes = os.Args[1]
	intClientes, err = strconv.Atoi(clientes)
	base.HandleErrorMsg(err, "Erro ao converter clientes")
	clientID, err = strconv.Atoi(os.Args[2])
	base.HandleErrorMsg(err, "Erro ao converter o id_do_cliente")
	isClienteMedio = clientID == (intClientes/2)+1 // Para identificar o cliente que deve ter o RTT registrado

	// TEMPO TOTAL - Inicia a contagem do tempo
	//tt_inicio := time.Now()

	// Executar o cliente RPC
	ClientRPC(numRequests, "C", "F")

	// TEMPO TOTAL - Finaliza a contagem do tempo
	/*tt_elapsed := time.Since(tt_inicio)
	tt_elapsedStr := fmt.Sprintf("%d", tt_elapsed.Nanoseconds()) // nanosegundos
	record := []string{
		"rpc",
		clientes,
		strconv.Itoa(clientID),
		tt_elapsedStr,
	}
	writeToCSV(record)
	fmt.Printf("Cliente %d: TEMPO TOTAL: %s\n", clientID, tt_elapsedStr)*/

}

func runClient(clientID, numRequests int, writer *csv.Writer) {

	defer wg.Done()

	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		fmt.Printf("Cliente %d: Erro ao conectar ao servidor: %s\n", clientID, err)
		return
	}
	defer client.Close()

	for i := 0; i < numRequests; i++ {
		//req := base.RequestConversor{
		//	Valor:    100,
		//	FromUnit: "C",
		//	ToUnit:   "F",
		//}
		req := base.GenerateRequest("C", "F")

		startTime := time.Now()
		var reply base.RequestConversor
		err = client.Call("ConversorService.Converter", req, &reply)
		if err != nil {
			fmt.Printf("Cliente %d: Erro na chamada %d: %s\n", clientID, i+1, err)
			continue
		}
		fmt.Printf("Conversão %d: %.2f %s para %s => %.2f %s\n", i, req.Valor,
			req.FromUnit, req.ToUnit, reply.Valor, reply.ToUnit)

		elapsedTime := time.Since(startTime)

		// Armazenar as informações no arquivo CSV
		record := []string{
			strconv.Itoa(clientID),
			strconv.Itoa(i + 1),
			elapsedTime.String(),
		}
		if err := writer.Write(record); err != nil {
			fmt.Println("Erro ao escrever no arquivo CSV:", err)
		}
		fmt.Printf("Cliente %d: Requisição %d - RTT: %v\n", clientID, i+1, elapsedTime)
	}
}

func ClientRPC(numRequests int, fromUnit string, toUnit string) {
	// Cria um cliente RPC e se conecta ao servidor (CalcularService)
	client, err := rpc.Dial("tcp", base.HOST+base.PORTA_RPC)
	base.HandleErrorMsg(err, "Cliente "+strconv.Itoa(clientID)+" Erro ao conectar ao servidor.")

	// programa o fechamento da conexão
	defer func(client *rpc.Client) {
		err := client.Close()
		base.HandleErrorMsg(err, "Cliente "+strconv.Itoa(clientID)+" Erro ao fechar a conexão com o servidor.")
	}(client)

	// Faz as requisições ao servidor
	for i := 0; i < numRequests; i++ {
		// Cria request
		req := base.GenerateRequest(fromUnit, toUnit)

		// Variável para armazenar a resposta do servidor
		var reply base.RequestConversor

		// RTT - Inicia a contagem do tempo
		startTime := time.Now()

		// Faz a chamada remota
		err = client.Call("ConversorService.Converter", req, &reply)
		if err != nil {
			fmt.Printf("Cliente %d: Erro na chamada %d: %s\n", clientID, i+1, err)
			continue
		}

		// RTT - Armazenar as informações no arquivo CSV {protocolo, clientes, id_do_cliente, requisicao, tempo}
		elapsedTime := time.Since(startTime)
		elapsedTimeStr := fmt.Sprintf("%d", elapsedTime.Nanoseconds()) // nanosegundos
		if isClienteMedio {
			record := []string{
				"rpc",
				clientes,
				strconv.Itoa(clientID),
				elapsedTimeStr,
			}
			writeToCSV(record)
			// Imprime o RTT
			fmt.Printf("Cliente %d: Requisição %d - RTT: %s\n", clientID, i+1, elapsedTimeStr)
		}

		// RTT e TEMPO TOTAL - Deixar comentado para testes de desempenho
		// Imprime a resposta do servidor
		//fmt.Printf("Cliente %d: Conversão %d: %.2f %s para %s => %.2f %s\n", clientID, i+1, req.Valor,
		//	req.FromUnit, req.ToUnit, reply.Valor, reply.ToUnit)

	}

}

func writeToCSV(record []string) {
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
	//record := []string{protocolo, clientes, idCliente, tTotalStr}
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
