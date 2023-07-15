package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

func main() {
	// Endereço do servidor
	serverAddr, err := net.ResolveTCPAddr("tcp", "localhost:8000")
	if err != nil {
		fmt.Println("Erro ao resolver endereço do servidor:", err)
		return
	}

	// Cria uma conexão TCP para o servidor
	conn, err := net.DialTCP("tcp", nil, serverAddr)
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}
	defer conn.Close()

	// Gera um ID de carro aleatório
	rand.Seed(time.Now().UnixNano())
	carID := rand.Intn(100)

	// Envia o ID do carro para o servidor
	_, err = conn.Write([]byte(strconv.Itoa(carID) + "\n"))
	if err != nil {
		fmt.Println("Erro ao enviar o ID do carro para o servidor:", err)
		return
	}

	// Cria um leitor de buffer para a conexão
	reader := bufio.NewReader(conn)

	// Número de solicitações
	numRequests := 10000

	// Slice para armazenar os tempos de RTT
	rttTimes := make([]time.Duration, numRequests)

	// Loop para enviar as solicitações e coletar os tempos de RTT
	for i := 0; i < numRequests; i++ {
		// Gera um valor de velocidade aleatório
		velocidade := strconv.Itoa(rand.Intn(100))

		// Registra o tempo de envio
		sendTime := time.Now()

		// Envia a velocidade para o servidor
		_, err = conn.Write([]byte(velocidade + "\n"))
		if err != nil {
			fmt.Println("Erro ao enviar dados para o servidor:", err)
			return
		}

		// Lê a resposta do servidor
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Erro ao ler resposta do servidor:", err)
			return
		}

		// Registra o tempo de recebimento
		receiveTime := time.Now()

		// Calcula o tempo de RTT
		rtt := receiveTime.Sub(sendTime)

		// Armazena o tempo de RTT
		rttTimes[i] = rtt

		// Extrai a porcentagem de desgaste do pneu da resposta
		parts := strings.Split(strings.TrimSpace(response), ":")
		pneuID, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		pneuDesgaste, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)

		// Imprime a porcentagem de desgaste do pneu recebida
		fmt.Printf("Resposta do servidor (Pneu ID: %d) - Porcentagem de desgaste do pneu: %.2f%%\n", pneuID, pneuDesgaste)

		// Aguarda 1 segundo antes de enviar a próxima velocidade
		//time.Sleep(1 * time.Second)
	}

	// Calcula a média dos tempos de RTT
	var totalRTT time.Duration
	for _, rtt := range rttTimes {
		totalRTT += rtt
	}
	averageRTT := totalRTT / time.Duration(numRequests)

	// Imprime a média dos tempos de RTT
	fmt.Println("Tempo médio de RTT:", averageRTT)
}
