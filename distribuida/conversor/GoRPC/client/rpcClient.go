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

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Uso: go run rpc_client.go <número_de_clientes>")
		return
	}

	numClients, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Erro ao converter o número de clientes:", err)
		return
	}

	numRequests := 10000 // Número de solicitações por cliente

	file, err := os.Create("desempenhoRPC.csv")
	if err != nil {
		fmt.Println("Erro ao criar o arquivo CSV:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	defer writer.Flush()

	wg.Add(numClients)
	for i := 0; i < numClients; i++ {
		go runClient(i, numRequests, writer)
	}

	wg.Wait()

	// Aguardar até que todos os clientes terminem
	time.Sleep(time.Duration(numClients) * time.Second)
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
		req := base.RequestConversor{
			Valor:    100,
			FromUnit: "C",
			ToUnit:   "F",
		}

		startTime := time.Now()
		var reply base.RequestConversor
		err = client.Call("ConversorService.Converter", req, &reply)
		if err != nil {
			fmt.Printf("Cliente %d: Erro na chamada %d: %s\n", clientID, i+1, err)
			continue
		}

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
