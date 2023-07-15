package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	"math/rand"
)

func main() {
	// Endereço do servidor
	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:8000")
	if err != nil {
		fmt.Println("Erro ao resolver endereço do servidor:", err)
		return
	}

	// Cria uma conexão UDP para o servidor
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}
	defer conn.Close()

	// Número de solicitações
	numRequests := 10000

	// Slice para armazenar os tempos de RTT
	rttTimes := make([]time.Duration, numRequests)

	// Loop para enviar as solicitações e coletar os tempos de RTT
	for i := 0; i < numRequests; i++ {
		// Gera um valor de velocidade aleatório
		velocidade := strconv.Itoa(rand.Intn(100))

		// Envia a velocidade para o servidor
		_, err := conn.Write([]byte(velocidade))
		if err != nil {
			fmt.Println("Erro ao enviar dados para o servidor:", err)
			return
		}

		// Registra o tempo de envio
		sendTime := time.Now()

		// Buffer para armazenar a resposta do servidor
		buffer := make([]byte, 1024)

		// Lê a resposta do servidor
		n, _, err := conn.ReadFromUDP(buffer)
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
		parts := strings.Split(strings.TrimSpace(string(buffer[:n])), ":")
		pneuID, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		pneuDesgaste, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)

		// Imprime a porcentagem de desgaste do pneu recebida
		fmt.Printf("Resposta do servidor (Pneu ID: %d) - Porcentagem de desgaste do pneu: %.2f%%\n", pneuID, pneuDesgaste)
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
