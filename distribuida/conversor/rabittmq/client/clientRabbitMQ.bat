@echo off

REM verifica se os argumentos foram fornecidos
IF "%1"=="" (
    echo argumento nao fornecido: CLIENTES.
    echo Uso: clientSocket.bat CLIENTES
    exit /b
)

REM Set CLIENTES e PROTOCOLO
set CLIENTES=%1

REM Compila o arquivo clientSocket.go
go build clientRabbitMQ.go

REM Execute o programa para a quantidade CLIENTS estabelecida
for /L %%i in (1,1,%CLIENTES%) do (
    REM cliente RPC
    start cmd /k clientRabbitMQ.exe %CLIENTES% %%i
)