@echo off

REM verifica se os argumentos foram fornecidos
IF "%1"=="" (
    echo argumento nao fornecido: CLIENTES.
    echo Uso: clientSocket.bat CLIENTES PROTOCOLO
    exit /b
)

IF "%2"=="" (
    echo argumento nao fornecido: PROTOCOLO.
    echo Usage: script.bat CLIENTES PROTOCOLO
    echo PROTOCOLO: tcp ou udp
    exit /b
)

REM Set CLIENTES e PROTOCOLO
set CLIENTES=%1
set PROTOCOLO=%2

REM Compila o arquivo clientSocket.go
go build clientSocket.go

REM Execute o programa para a quantidade CLIENTS estabelecida
for /L %%i in (1,1,%CLIENTES%) do (
    REM cliente TCP, 10000 invocações, CelsiusToFahrenheit
    start /B clientSocket.exe %PROTOCOLO% %CLIENTES% %%i 10000 C F
)