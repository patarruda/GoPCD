package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"pcd/distribuida/conversor/base"
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const qos = 0

var (
	clientes    string  // quantidade total de clientes
	clientID    string  // id do cliente atual
	numRequests = 10000 // quantidade de requisições por cliente
	recebidos   = 0     // quantidade de respostas efetivas
	msgToServer base.RequestConversor
	//s           = NewSemaphore()			// testes

	cond             = sync.NewCond(&sync.Mutex{}) // variavel condicional para sincronização de requests e respostas
	responseReceived = false                       // indica se uma resposta foi recebida

	fileLock sync.Mutex // Lock para acesso concorrente ao arquivo .csv
	csvFile  = "desempenhoMQTT_TempoTotal.csv"
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
	perdaStr := fmt.Sprintf("%d", numRequests-recebidos)        // perda de requisições
	qosStr := fmt.Sprintf("%d", qos)
	record := []string{"mqtt", clientes, clientID, ttElapsedStr, perdaStr, qosStr, "var_cond"}
	writeToCSV(record)

	fmt.Println(record)
	fmt.Printf("Cliente %s: TEMPO TOTAL: %s\n", clientID, ttElapsedStr)
}

func Client(fromUnit string, toUnit string) {
	// configurar como um clienteMQTT do broker
	opts := MQTT.NewClientOptions()
	opts.AddBroker("mqtt://localhost:1883")
	opts.SetClientID("ClienteID" + clientID)
	opts.SetCleanSession(true) // limpar sessão deste cliente do broker quando desconectar

	// criar novo cliente do broker MQTT
	client := MQTT.NewClient(opts)

	// conectar ao broker
	token := client.Connect()
	token.Wait()
	base.HandleErrorMsg(token.Error(), "Erro ao conectar ao broker")
	//fmt.Println("Conectado ao broker...")

	// desconectar cliente do broker
	defer client.Disconnect(1) // 250 ms para desconectar

	// subscrever a um topico & usar um handler para receber as mensagens
	replyTopic := base.TOPIC_REPLIES + clientID
	requestTopic := base.TOPIC_REQUESTS + clientID
	token = client.Subscribe(replyTopic, qos, receiveHandler)
	base.HandleErrorMsg(token.Error(), "Erro ao subscrever ao tópico")

	//fmt.Printf(" Subscribed to %s\n Publishing to %s\n\n", replyTopic, requestTopic)

	// loop para publicar mensagens
	for i := 0; i < numRequests; i++ {
		// cria a mensagem
		msgToServer = base.GenerateRequest(fromUnit, toUnit)
		msg, err := json.Marshal(msgToServer)
		base.HandleErrorMsg(err, "Erro ao converter serializar mensagem")

		//publicar a mensagem
		token := client.Publish(requestTopic, qos, false, msg)
		token.Wait()
		base.HandleErrorMsg(token.Error(), "Erro ao publicar mensagem")

		fmt.Printf("Cliente %s - Solicita conversão nº %d: %.2f %s para %s\n", clientID, i+1, msgToServer.Valor, msgToServer.FromUnit, msgToServer.ToUnit)

		// Esperar pela resposta (variável condicional)
		waitForResponse()
		//s.waitForResponse()

		// PÉSSIMO DESEMPENHO!!! (BUSY WAINTING)
		//for !responseReceived { // enquanto não receber uma resposta espera
		//}
		//responseReceived = false // resetar a variável para a próxima requisição
		//time.Sleep(time.Millisecond)

	}

	// espera até que todas as respostas sejam recebidas
	//waitForAllResponses()

}

// função handler para receber mensagens do broker
var receiveHandler MQTT.MessageHandler = func(c MQTT.Client, m MQTT.Message) {
	//fmt.Println("\nrecebeu...")

	// desserializar mensagem recebida transformando em um objeto RequestConversor
	reply := base.RequestConversor{}
	err := json.Unmarshal(m.Payload(), &reply)
	base.HandleErrorMsg(err, "Erro ao desserializar mensagem")

	fmt.Printf("Cliente %s - Recebida conversão: %.2f %s para %s => %.2f %s\n", clientID, msgToServer.Valor, msgToServer.FromUnit, msgToServer.ToUnit, reply.Valor, reply.ToUnit)

	// sinalizar que uma resposta foi recebida (variável condicional)
	signalResponseReceived()
	//s.signalResponseReceived()

	recebidos++
	//signalIfAllResponsesReceived()

	//responseReceived = true
}

// Funções para sincronização de requests e respostas

func waitForResponse() {
	cond.L.Lock()         // Lock da variável condicional
	defer cond.L.Unlock() // Unlock da variável condicional ao final da função

	for !responseReceived { // enquanto não receber uma resposta espera
		cond.Wait() // espera até que a variável condicional seja sinalizada e libera o lock
	}

	// Após ser acordada, a thread readquire o lock e continua a execução (liberação do lock é feita no defer)
	responseReceived = false // resetar a variável para a próxima requisição
}

func signalResponseReceived() {
	cond.L.Lock()           // Lock da variável condicional
	responseReceived = true // atulaiza variavel para indicar que uma resposta foi recebida
	cond.L.Unlock()         // Unlock da variável condicional
	cond.Signal()           // acordar uma thread que esteja esperando na variável condicional
}

func waitForAllResponses() {
	cond.L.Lock()                  // Lock da variável condicional
	defer cond.L.Unlock()          // Unlock da variável condicional ao final da função
	for recebidos != numRequests { // enquanto não receber todas as respostas espera
		cond.Wait() // espera até que a variável condicional seja sinalizada e libera o lock
	}
}

func signalIfAllResponsesReceived() {
	if recebidos == numRequests {
		cond.L.Lock()   // Lock da variável condicional
		cond.Signal()   // acordar uma thread que esteja esperando na variável condicional
		cond.L.Unlock() // Unlock da variável condicional
	}
}

// Com Semáforo
/*
type ISemaphore interface {
	P()
	V()
}

type Semaphore struct {
	responseReceived bool
	cond             *sync.Cond
}

func NewSemaphore() *Semaphore {
	return &Semaphore{
		responseReceived: false,
		cond:             sync.NewCond(&sync.Mutex{}),
	}
}

func (s *Semaphore) waitForResponse() {
	s.cond.L.Lock()           // lock na variável condicional
	for !s.responseReceived { // (sem recursos disponíveis)
		fmt.Println("dormiu...")
		s.cond.Wait() // thread fica em espera
	}
	fmt.Println("acordou...")
	// após a thread ser acordada, segue a execução
	s.responseReceived = false
	fmt.Println("setou responseReceived false...")
	s.cond.L.Unlock() // unlock na variável condicional
}

func (s *Semaphore) signalResponseReceived() {
	s.cond.L.Lock() // lock na variável condicional
	s.responseReceived = true
	fmt.Println("setou responseReceived true...")
	s.cond.L.Unlock() // unlock na variável condicional
	s.cond.Signal()   // acorda uma thread em espera
	fmt.Println("sinalizou acordar...")
}*/

// Função para escrita no arquivo .csv
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
