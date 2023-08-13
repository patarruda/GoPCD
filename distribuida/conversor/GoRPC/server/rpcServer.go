package main

import (
	"fmt"
	"net"
	"net/rpc"
	"pcd/distribuida/conversor/base"
)

type ConversorService struct{}

func (s *ConversorService) Converter(req base.RequestConversor, reply *base.RequestConversor) error {
	*reply = base.ConversorMedidas{}.Invoke(req)
	return nil
}

func main() {
	conversorService := new(ConversorService)
	rpc.Register(conversorService)

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
