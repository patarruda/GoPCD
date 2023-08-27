package main

import (
	"encoding/json"
	"fmt"
	"pcd/distribuida/conversor/base"
	//"strconv"

	"github.com/streadway/amqp"
)

func main() {

	// cria conexão com o broker
	conn, err := amqp.Dial("amqp://" + base.USR_PSW_RABBITMQ + base.HOST + base.PORTA_RABBITMQ)
	base.HandleErrorMsg(err, "Não foi possível se conectar ao broker")
	defer conn.Close()

	// cria um canal
	ch, err := conn.Channel()
	base.HandleErrorMsg(err, "Não foi possível estabelecer um canal de comunicação com o broker")
	defer ch.Close()

	// declara a fila no broker para receber mensagens dos clientes
	q, err := ch.QueueDeclare(
		base.FILA_REQUESTS, // Nome da fila
		false,              // Not Durable (queue mão sobrevive após servidor ser reiniciado)
		false,              // Not Auto-deleted (queue não é deletada quando não está em uso)
		false,              // Not Exclusive (a queue pode ser acessada por outros canais)
		false,              // Sem No-wait (cliente espera resposta do servidor)
		nil,                // Arguments (optional additional arguments)
	)
	base.HandleErrorMsg(err, "Não foi possível criar a fila no broker")

	// Consumer - escuta a fila no aguardo de mensagens dos clientes
	msgs, err := ch.Consume(
		q.Name,      // Name da fila de onde serão consumidas mensagens
		"conversor", // nome do Consumer (usado para rastrear consumers no RabbitMQ UI)
		true,        // Auto-acknowledgment (messages are automatically acknowledged)
		false,       // Not Exclusive (outros consumers também podem acessar a fila)
		false,       // Sem No-local (pode publicar mensagens da mesma conexão que está consumindo)
		false,       // Sem No-wait (cliente espera resposta do servidor)
		nil)         // Arguments (optional additional arguments)
	base.HandleErrorMsg(err, "Falha ao registrar o consumidor no broker")

	fmt.Printf("\n\n##### CONVERSOR DE MEDIDAS #####\n\n Servidor pronto: escutando fila \"%s\" (RabbitMQ)...\n\n", base.FILA_REQUESTS)

	// loop infinito para receber requests (amqp.Delivery) através do canal msgs
	//numThread := 1
	for delivery := range msgs {
		//threadStr := strconv.Itoa(numThread)
		//numThread++
		//fmt.Printf("Servidor - Thread nº %s - CorrelationID: %s\n", threadStr, delivery.CorrelationId)
		handleRequest(ch, delivery)
		//fmt.Printf("Servidor - Thread nº %s - CorrelationID: %s - handleRequest finalizado\n", threadStr, delivery.CorrelationId)

	}

}

func handleRequest(ch *amqp.Channel, delivery amqp.Delivery) {
	// recebe request e desserializa
	msgFromClient := base.RequestConversor{}
	err := json.Unmarshal(delivery.Body, &msgFromClient)
	base.HandleErrorMsg(err, "Falha ao desserializar a mensagem")

	//fmt.Printf("Request recebido...\n  UserId: %s / AppId: %s / MessageId: %s, CorrelationId: %s\n",
	//	delivery.UserId, delivery.AppId, delivery.MessageId, delivery.CorrelationId)

	// processa request
	msgToClient := base.ConversorMedidas{}.Invoke(msgFromClient)

	// prepara resposta
	replyMsgBytes, err := json.Marshal(msgToClient)
	base.HandleErrorMsg(err, "Falha ao serializar mensagem")

	// publica resposta
	err = ch.Publish(
		"",               // Tipo de Exchange (string vazia = default)
		delivery.ReplyTo, // Routing key (fila de destino é a registrada no campo ReplyTo da mensagem recebida)
		false,            // Mandatory (não é obrigatório)
		false,            // Immediate (não é necessário)
		amqp.Publishing{
			ContentType:   "text/plain",           // Tipo de conteúdo da mensagem
			CorrelationId: delivery.CorrelationId, // ID de correlação da resposta
			Body:          replyMsgBytes,          // Conteúdo da mensagem de resposta
		},
	)
	base.HandleErrorMsg(err, "Falha ao enviar a mensagem para o broker")
}
