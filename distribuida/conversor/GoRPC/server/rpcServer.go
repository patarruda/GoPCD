package main

import (
	"fmt"
	"net"
	"net/rpc"
	"pcd/distribuida/conversor/base"
)

// ConversorService é o objeto (struct) que será registrado para ser chamado remotamente via RPC
type ConversorService struct{}

// Converter é o método que será chamado remotamente via RPC
func (s *ConversorService) Converter(req base.RequestConversor, reply *base.RequestConversor) error {
	*reply = base.ConversorMedidas{}.Invoke(req)
	return nil
}

func main() {
	conversorService := new(ConversorService) // cria instância de ConversorService
	rpc.Register(conversorService)            // registra ConversorService para ser chamada remotamente via RPC

	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		fmt.Println("Erro ao iniciar o servidor:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Servidor de conversão de medidas ouvindo na porta 1234...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Erro na conexão:", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
