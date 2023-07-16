@echo off

REM verifica se o argumento PROTOCOLO foi fornecido
IF "%1"=="" (
    echo argumento nao fornecido: PROTOCOLO.
    echo Uso: servidorSocket.bat PROTOCOLO
    echo PROTOCOLO: tcp ou udp
    exit /b
)

REM Set PROTOCOLO
set PROTOCOLO=%1

REM Compila o arquivo serverSocket.go
go build serverSocket.go

serverSocket.exe %PROTOCOLO%