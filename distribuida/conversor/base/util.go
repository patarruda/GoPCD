package base

import (
	"fmt"
	"os"
)

const (
	HOST      = "localhost"
	PORTA_TCP = ":1313"
	PORTA_UDP = ":1314"
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
