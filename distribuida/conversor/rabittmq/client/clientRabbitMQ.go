package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	//"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"pcd/distribuida/conversor/base"

	"github.com/streadway/amqp"
)

var ()

var (
	clientes    string  // quantidade total de clientes
	clientID    string  // id do cliente atual
	numRequests = 10000 // quantidade de requisições por cliente

	fileLock sync.Mutex // Lock para acesso concorrente ao arquivo .csv
	csvFile  = "desempenhoRabbitMQ_TempoTotal.csv"
)

func main() {
	// Verificar se os argumentos foram passados corretamente
	if len(os.Args) != 3 {
		fmt.Println("Parâmetros exigidos: <clientes> <id_do_cliente>")
		return
	}

	// Setar as variáveis globais
	clientes = os.Args[1]
	clientID = os.Args[2]

	// TEMPO TOTAL - Inicia a contagem do tempo
	fmt.Printf("Cliente %s: iniciou..\n", clientID)
	ttInicio := time.Now()

	// Executar o cliente
	Client("C", "F")

	// TEMPO TOTAL - Finaliza a contagem do tempo
	ttElapsed := time.Since(ttInicio)
	ttElapsedStr := fmt.Sprintf("%d", ttElapsed.Milliseconds()) // ms
	record := []string{"rabbitmq", clientes, clientID, ttElapsedStr}
	writeToCSV(record)
	fmt.Println(record)
	fmt.Printf("Cliente %s: TEMPO TOTAL: %s\n", clientID, ttElapsedStr)
}

func Client(fromUnit string, toUnit string) {
	// cria conexão com o broker
	conn, err := amqp.Dial("amqp://" + base.USR_PSW_RABBITMQ + base.HOST + base.PORTA_RABBITMQ)
	base.HandleErrorMsg(err, "Não foi possível se conectar ao broker")
	defer conn.Close()

	// cria o canal
	ch, err := conn.Channel()
	base.HandleErrorMsg(err, "Não foi possível estabelecer um canal de comunicação com o broker")
	defer ch.Close()

	// declara a fila no broker para receber as respostas
	q, err := ch.QueueDeclare(
		"", // fila anônima (cada cliente tem sua própria fila)
		//base.FILA_REPLIES, // Nome da fila
		false, // Not Durable (queue mão sobrevive após servidor ser reiniciado)
		false, // Not Auto-deleted (queue não é deletada quando não está em uso)
		true,  // Exclusive (a queue não pode ser acessada por outros canais)
		false, // Sem No-wait (cliente espera resposta do servidor)
		nil,   // Arguments (optional additional arguments)
	)
	base.HandleErrorMsg(err, "Não foi possível criar a fila no broker")

	// Consumer - escuta a fila de respostas no aguardo de mensagens do servidor
	msgs, err := ch.Consume(
		q.Name,                // Name da fila de onde serão consumidas mensagens
		"clienteID_"+clientID, // nome do cliente (usado para rastrear consumers no RabbitMQ UI)
		true,                  // Auto-acknowledgment (messages are automatically acknowledged)
		false,                 // Not Exclusive (outros consumers também podem acessar a fila)
		false,                 // Sem No-local (pode publicar mensagens da mesma conexão que está consumindo)
		false,                 // Sem No-wait (cliente espera resposta do servidor)
		nil)                   // Arguments (optional additional arguments)
	base.HandleErrorMsg(err, "Falha ao registrar o cliente no broker")

	// loop para enviar requisições ao servidor e receber respostas
	for i := 0; i < numRequests; i++ {
		// prepara mensagem
		msgRequest := base.GenerateRequest("C", "F")
		msgRequestBytes, err := json.Marshal(msgRequest)
		base.HandleErrorMsg(err, "Falha ao serializar a mensagem")

		correlationID := "ID_" + clientID + "--REQ_" + strconv.Itoa(i)

		//fmt.Println("cliente - publicando mensagem...")
		// publica mensagem no broker
		err = ch.Publish(
			"",                 // Tipo de Exchange (string vazia = default)
			base.FILA_REQUESTS, // Routing key (fila de destino)
			false,
			false,
			amqp.Publishing{
				ContentType:   "text/plain",  // Tipo de conteúdo da mensagem
				CorrelationId: correlationID, // ID de correlação para rastreamento da resposta
				ReplyTo:       q.Name,        // fila de destino para a resposta do servidor
				Body:          msgRequestBytes,
			},
		)
		base.HandleErrorMsg(err, "Falha ao publicar a mensagem")
		//fmt.Println("cliente - publicação realizada...")

		// aguarda resposta do broker
		delivery := <-msgs
		//fmt.Println("cliente - recebe delivery...")

		// deserializada a mensagem
		msgResponse := base.RequestConversor{}
		err = json.Unmarshal(delivery.Body, &msgResponse)
		//base.HandleNonFatal(err, "Erro na deserialização da resposta")

		// Verifica se a resposta recebida é a esperada e imprime
		if correlationID != delivery.CorrelationId {
			fmt.Printf("Cliente %s: Conversão %d: Erro na resposta do servidor\n", clientID, i+1)
		} else {
			fmt.Printf("Cliente %s: Conversão %d: %.2f %s para %s => %.2f %s\n", clientID, i+1, msgRequest.Valor,
				msgRequest.FromUnit, msgRequest.ToUnit, msgResponse.Valor, msgResponse.ToUnit)
		}
	}
}

// writeToCSV escreve os dados de desempenho no arquivo CSV
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
