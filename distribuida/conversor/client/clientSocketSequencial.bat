@echo off

REM verifica se os argumentos foram fornecidos
IF "%1"=="" (
    echo argumento nao fornecido: AMOSTRA.
    echo Uso: clientSocket.bat AMOSTRA CLIENTES PROTOCOLO
    exit /b
)

IF "%2"=="" (
    echo argumento nao fornecido: CLIENTES.
    echo Uso: clientSocket.bat AMOSTRA CLIENTES PROTOCOLO
    exit /b
)

IF "%3"=="" (
    echo argumento nao fornecido: PROTOCOLO.
    echo Uso: clientSocket.bat AMOSTRA CLIENTES PROTOCOLO
    echo PROTOCOLO: tcp ou udp
    exit /b
)

REM Set AMOSTRA, CLIENTES e PROTOCOLO
set AMOSTRA=%1
set CLIENTES=%2
set PROTOCOLO=%3

REM Compila o arquivo clientSocket.go
go build clientSocket.go

REM Execute o programa para a quantidade CLIENTES estabelecida em cada iteração
for /L %%a in (1,1,%AMOSTRA%) do (
    clientSocket.bat %CLIENTES% %PROTOCOLO%
    timeout /t 2 /nobreak > NUL
)