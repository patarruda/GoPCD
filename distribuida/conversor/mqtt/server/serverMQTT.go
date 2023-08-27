package main

import (
	"encoding/json"
	"fmt"
	"os"
	"pcd/distribuida/conversor/base"
	"strconv"
	//"time"
	//"sync"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const qos = 0 // nível de serviço QoS (Quality of Service) 1 = "Entrega garantida"

var (
	clientes int // quantidade total de clientes esperada (passada como argumento)
)

func main() {
	// Verificar se os argumentos foram passados corretamente
	if len(os.Args) != 2 {
		fmt.Println("Parâmetros exigidos: <clientes_esperados>")
		return
	}

	// Setar as variáveis globais
	var err error
	clientes, err = strconv.Atoi(os.Args[1])
	base.HandleErrorMsg(err, "Erro ao converter clientes")

	// Iniciar o servidor MQTT
	Server()

}

func Server() {
	// configurar servidor como um cliente do Broker MQTT
	opts := MQTT.NewClientOptions()
	opts.AddBroker("mqtt://" + base.HOST + base.PORTA_MQTT)
	opts.SetClientID("conversor_server")
	opts.SetCleanSession(true) // limpar sessão deste cliente do broker quando desconectar

	// criar novo cliente do Broker MQTT
	client := MQTT.NewClient(opts)

	// conectar ao broker
	token := client.Connect()
	token.Wait()
	base.HandleErrorMsg(token.Error(), "Erro ao conectar ao broker")

	// programar desconexão do broker
	defer client.Disconnect(1)

	// Subscrever para tópicos (1 "request" e 1 "reply" por cliente)
	for i := 1; i <= clientes; i++ {
		// criar tópicos
		requestTopic := base.TOPIC_REQUESTS + strconv.Itoa(i)
		replyTopic := base.TOPIC_REPLIES + strconv.Itoa(i)

		// subscrever ao tópico e usar um handler para receber as mensagens
		//wg := sync.WaitGroup{}
		//wg.Add(1)
		//go func() {
		token = client.Subscribe(requestTopic, qos, createReceiveHandler(replyTopic))
		token.Wait()
		base.HandleErrorMsg(token.Error(), "Erro ao subscrever ao tópico...")

		fmt.Printf("\nTópicos para o cliente %d\n Subscribed to %s\n Publishing to %s\n", i, requestTopic, replyTopic)
		//wg.Wait()
		//}()
	}

	fmt.Println("\n\n##### CONVERSOR DE MEDIDAS #####\n\n Servidor pronto (MQTT - Mosquitto)...\n Para sair, pressione <ENTER>...\n\n")

	fmt.Scanln()
}

func createReceiveHandler(replyTopic string) MQTT.MessageHandler {
	// retorna uma função handler para tratar mensagens recebidas
	return func(c MQTT.Client, m MQTT.Message) {
		// desserializar mensagem recebida transformando em um objeto RequestConversor
		msgFromClient := base.RequestConversor{}
		err := json.Unmarshal(m.Payload(), &msgFromClient)
		base.HandleErrorMsg(err, "Erro ao desserializar mensagem")

		// invocar a operação do conversor de medidas
		resultado := base.ConversorMedidas{}.Invoke(msgFromClient)

		// serializar o resultado da operação
		msgToClient, err := json.Marshal(resultado)
		base.HandleErrorMsg(err, "Erro ao serializar mensagem")

		// publicar o resultado no tópico de resposta
		token := c.Publish(replyTopic, qos, false, msgToClient)
		//fmt.Println("passou do publish..")
		token.Wait()
		//if token.WaitTimeout(10 * time.Millisecond) {
		//	fmt.Println("atingiu timeout...")
		//	token = c.Publish(replyTopic, 0, false, msgToClient)
		//	token.Wait()
		//}
		//fmt.Println("passou do token.Wait()..")
		base.HandleErrorMsg(token.Error(), "Erro ao publicar mensagem")

		//fmt.Printf("Tópico %s - Conversão: %.2f %s para %s => %.2f %s\n", replyTopic, msgFromClient.Valor,
		//	msgFromClient.FromUnit, msgFromClient.ToUnit, resultado.Valor, resultado.ToUnit)

	}
}
