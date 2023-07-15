package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Car struct {
	ID           int
	CurrentSpeed int
	TireQuality  float64
	mu           sync.Mutex
}

type Client struct {
	Addr   *net.UDPAddr
	CarID  int
	mu     sync.Mutex
}

func main() {
	// Cria um endereço UDP
	address, err := net.ResolveUDPAddr("udp", ":8000")
	if err != nil {
		fmt.Println("Erro ao resolver endereço:", err)
		return
	}

	// Cria uma conexão UDP
	conn, err := net.ListenUDP("udp", address)
	if err != nil {
		fmt.Println("Erro ao ouvir:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Servidor aguardando conexões...")

	// Aguarda 10 segundos antes de iniciar a coleta de dados
	time.Sleep(10 * time.Second)

	// Cria um slice para armazenar as informações dos carros
	cars := make([]Car, 0)

	// Cria um map para armazenar as informações dos clientes
	clients := make(map[string]Client)
	clientsMutex := sync.Mutex{}

	// Loop principal do servidor
	for {
		// Buffer para armazenar os dados recebidos
		buffer := make([]byte, 1024)

		// Lê os dados do cliente
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Erro ao ler dados:", err)
			continue
		}

		// Extrai a velocidade do carro do pacote recebido
		velocidade, err := strconv.Atoi(strings.TrimSpace(string(buffer[:n])))
		if err != nil {
			fmt.Println("Erro ao converter velocidade:", err)
			continue
		}

		// Verifica se o cliente já está registrado
		client, exists := getClient(addr, clients, &clientsMutex)
		if !exists {
			// Verifica se o carro já está registrado
			carID := getCarID(addr, clients, &clientsMutex, cars)
			if carID == -1 {
				// Registra um novo carro
				car := Car{
					ID:           len(cars) + 1,
					CurrentSpeed: velocidade,
					TireQuality:  100.0,
				}
				cars = append(cars, car)
				carID = len(cars) - 1
			}

			// Registra um novo cliente
			client = Client{
				Addr:  addr,
				CarID: carID,
			}
			addClient(client, clients, &clientsMutex)

			// Envia a resposta inicial para o cliente
			response := fmt.Sprintf("%d:%.2f", client.CarID, cars[client.CarID].TireQuality)
			_, err = conn.WriteToUDP([]byte(response), addr)
			if err != nil {
				fmt.Println("Erro ao enviar resposta:", err)
				continue
			}

			fmt.Println("Novo cliente registrado (Carro ID:", client.CarID+1, ")", "Endereço:", addr)
		} else {
			// Atualiza a velocidade do carro existente
			updateCarSpeed(client.CarID, velocidade, cars)

			// Atualiza a porcentagem de desgaste do pneu
			tireQuality := updateTireQuality(client.CarID, cars)

			// Verifica se o pneu precisa ser trocado
			if tireQuality <= 20.0 {
				tireQuality = 100.0
				resetTireQuality(client.CarID, cars)
				fmt.Println("Carro (ID:", cars[client.CarID].ID, ") trocou os pneus no box.")
			}

			// Envia a resposta para o cliente
			response := fmt.Sprintf("%d:%.2f", client.CarID, tireQuality)
			_, err = conn.WriteToUDP([]byte(response), addr)
			if err != nil {
				fmt.Println("Erro ao enviar resposta:", err)
				continue
			}

			fmt.Println("Resposta enviada para o cliente (Carro ID:", client.CarID+1, ")", "Endereço:", addr)
		}
	}
}

// Função para obter o cliente com base no endereço do cliente
func getClient(addr *net.UDPAddr, clients map[string]Client, clientsMutex *sync.Mutex) (Client, bool) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	client, exists := clients[addr.String()]
	return client, exists
}

// Função para adicionar um novo cliente ao mapa de clientes
func addClient(client Client, clients map[string]Client, clientsMutex *sync.Mutex) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	clients[client.Addr.String()] = client
}

// Função para obter o ID de um carro com base no endereço do cliente
func getCarID(addr *net.UDPAddr, clients map[string]Client, clientsMutex *sync.Mutex, cars []Car) int {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for _, client := range clients {
		if client.Addr.IP.String() == addr.IP.String() && client.Addr.Port == addr.Port {
			return client.CarID
		}
	}
	return -1
}

// Função para atualizar a velocidade de um carro existente
func updateCarSpeed(carID int, speed int, cars []Car) {
	cars[carID].mu.Lock()
	cars[carID].CurrentSpeed = speed
	cars[carID].mu.Unlock()
}

// Função para atualizar a porcentagem de desgaste do pneu de um carro
func updateTireQuality(carID int, cars []Car) float64 {
	cars[carID].mu.Lock()
	car := &cars[carID]
	car.TireQuality -= float64(car.CurrentSpeed) * 0.01 // Exemplo de diminuição progressiva
	if car.TireQuality < 0 {
		car.TireQuality = 0
	}
	tireQuality := car.TireQuality
	cars[carID].mu.Unlock()
	return tireQuality
}

// Função para resetar a porcentagem de desgaste do pneu para 100% de um carro
func resetTireQuality(carID int, cars []Car) {
	cars[carID].mu.Lock()
	cars[carID].TireQuality = 100.0
	cars[carID].mu.Unlock()
}
