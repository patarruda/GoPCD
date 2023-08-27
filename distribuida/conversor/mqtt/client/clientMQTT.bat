@echo off

REM verifica se os argumentos foram fornecidos
IF "%1"=="" (
    echo argumento nao fornecido: CLIENTES.
    exit /b
)

REM Set CLIENTES e PROTOCOLO
set CLIENTES=%1

REM Compila o arquivo clientSocket.go
go build clientMQTT.go

REM Execute o programa para a quantidade CLIENTS estabelecida
for /L %%i in (1,1,%CLIENTES%) do (
    REM cliente RPC
    start cmd /k clientMQTT.exe %CLIENTES% %%i
)
