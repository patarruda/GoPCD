package base

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

const (
	HOST             = "localhost"
	PORTA_TCP        = ":1313"
	PORTA_UDP        = ":1314"
	PORTA_RPC        = ":1234"
	PORTA_RABBITMQ   = ":5672"
	PORTA_MQTT       = ":1883"
	USR_PSW_RABBITMQ = "guest:guest@"
	FILA_REQUESTS    = "requests_conversor"
	FILA_REPLIES     = "replies_conversor"
	TOPIC_REQUESTS   = "requests/client/"
	TOPIC_REPLIES    = "replies/client/"
)

type RequestConversor struct {
	Valor    float64
	FromUnit string
	ToUnit   string
}

func HandleError(err error) {
	if err != nil {
		fmt.Println("Erro: ", err.Error())
		os.Exit(1) // encerra o programa com status 1 (erro)
	}
}

func HandleErrorMsg(err error, msg string) {
	if err != nil {
		fmt.Println(msg, "\n  -> ", err.Error())
		os.Exit(1) // encerra o programa com status 1 (erro)
	}
}

func HandleNonFatal(err error, msg string) {
	if err != nil {
		fmt.Println(msg, "\n  -> ", err.Error())
	}
}

func GenerateRequest(fromoUnit string, toUnit string) RequestConversor {
	// gerar um float64 aleatório até 40.0
	rand.Seed(time.Now().UnixNano())
	n := rand.Float64() * 40.0
	//cria request
	return RequestConversor{
		Valor:    n,
		FromUnit: fromoUnit,
		ToUnit:   toUnit}
}
