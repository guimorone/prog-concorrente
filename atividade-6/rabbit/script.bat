@echo off

REM Verifica se o número de argumentos é válido
if "%~1"=="" (
  echo Uso: %0 ^<N^>
  exit /b 1
)

REM Lê o primeiro argumento (número de vezes a executar)
set N=%1

REM Comando a ser executado
set "command_to_run=go run client.go"

REM Loop para executar o comando N vezes simultaneamente
setlocal enabledelayedexpansion
set "i=1"
:loop
if !i! gtr %N% goto :end
  REM Executa o comando em segundo plano (^) e redireciona a saída para NUL
  start /b %command_to_run%
  set /a "i+=1"
  goto :loop
:end
endlocal

REM Espera todas as execuções em segundo plano terminarem
:wait
for /l %%x in (1, 1, %N%) do (
  tasklist /fi "IMAGENAME eq go.exe" | find /i "go.exe" >nul
  if errorlevel 1 (
    goto :continue
  ) else (
    timeout /t 1 /nobreak >nul
    goto :wait
  )
)
:continue

echo Comando '%command_to_run%' foi executado %N% vezes simultaneamente.
