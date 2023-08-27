#!/bin/bash

# Verifica se o número de argumentos é válido
if [ $# -ne 1 ]; then
  echo "Uso: $0 <N>"
  exit 1
fi

# Lê o primeiro argumento (número de vezes a executar)
N=$1

# Comando a ser executado
command_to_run="go run client.go"

# Loop para executar o comando N vezes simultaneamente
i=1
while [ $i -le $N ]; do
  # Executa o comando em segundo plano (&) e redireciona a saída para o /dev/null
  $command_to_run &
  i=$((i + 1))
done

# Espera todas as execuções em segundo plano terminarem
wait

echo "Comando '$command_to_run' foi executado $N vezes simultaneamente."