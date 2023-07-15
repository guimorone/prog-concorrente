package main

import (
	"bufio"
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
	Conn   net.Conn
	CarID  int
	mu     sync.Mutex
}

func main() {
	// Cria um endereço TCP
	address, err := net.ResolveTCPAddr("tcp", ":8000")
	if err != nil {
		fmt.Println("Erro ao resolver endereço:", err)
		return
	}

	// Cria uma escuta TCP
	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		fmt.Println("Erro ao ouvir:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Servidor aguardando conexões...")

	// Aguarda 10 segundos antes de iniciar a coleta de dados
	time.Sleep(10 * time.Second)

	// Cria um mapa para armazenar as informações dos carros
	cars := make(map[int]*Car)

	// Cria um map para armazenar as informações dos clientes
	clients := make(map[string]Client)
	clientsMutex := sync.Mutex{}

	// Loop principal do servidor
	for {
		// Aceita a conexão do cliente
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Erro ao aceitar conexão:", err)
			continue
		}

		// Cria uma goroutine para lidar com o cliente
		go handleClient(conn, clients, &clientsMutex, &cars)
	}
}

// Função para lidar com um cliente
func handleClient(conn net.Conn, clients map[string]Client, clientsMutex *sync.Mutex, cars *map[int]*Car) {
	defer conn.Close()

	// Cria um leitor de buffer para a conexão
	reader := bufio.NewReader(conn)

	// Lê o ID do carro enviado pelo cliente
	carIDStr, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Erro ao ler ID do carro:", err)
		return
	}

	carID, err := strconv.Atoi(strings.TrimSpace(carIDStr))
	if err != nil {
		fmt.Println("Erro ao converter ID do carro:", err)
		return
	}

	// Verifica se o cliente já está registrado
	client, exists := getClient(conn.RemoteAddr().String(), clients, clientsMutex)
	if !exists {
		// Registra um novo cliente
		client = Client{
			Conn:  conn,
			CarID: carID,
		}
		addClient(client, clients, clientsMutex)
		addCar(carID, cars)

		// Envia a resposta inicial para o cliente
		response := fmt.Sprintf("%d:%.2f\n", client.CarID, (*cars)[client.CarID].TireQuality)
		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Erro ao enviar resposta:", err)
			return
		}

		fmt.Println("Novo cliente registrado (Carro ID:", client.CarID, ")", "Endereço:", conn.RemoteAddr())
	}

	// Loop para ler a velocidade enviada pelo cliente e enviar a resposta
	for {
		// Lê a velocidade enviada pelo cliente
		speedStr, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Erro ao ler velocidade do carro:", err)
			return
		}

		speed, err := strconv.Atoi(strings.TrimSpace(speedStr))
		if err != nil {
			fmt.Println("Erro ao converter velocidade do carro:", err)
			return
		}

		// Atualiza a velocidade do carro
		updateCarSpeed(client.CarID, speed, cars)

		// Atualiza a porcentagem de desgaste do pneu
		tireQuality := updateTireQuality(client.CarID, cars)

		// Verifica se o pneu precisa ser trocado
		if tireQuality <= 20.0 {
			tireQuality = 100.0
			resetTireQuality(client.CarID, cars)
			fmt.Println("Carro (ID:", client.CarID, ") trocou os pneus no box.")
		}

		// Envia a resposta para o cliente
		response := fmt.Sprintf("%d:%.2f\n", client.CarID, tireQuality)
		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Erro ao enviar resposta:", err)
			return
		}

		fmt.Println("Resposta enviada para o cliente (Carro ID:", client.CarID, ")", "Endereço:", conn.RemoteAddr())

		// Aguarda 1 segundo antes de ler a próxima velocidade
		//time.Sleep(1 * time.Second)
	}
}

// Função para obter o cliente com base no endereço do cliente
func getClient(addr string, clients map[string]Client, clientsMutex *sync.Mutex) (Client, bool) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	client, exists := clients[addr]
	return client, exists
}

// Função para adicionar um novo cliente ao mapa de clientes
func addClient(client Client, clients map[string]Client, clientsMutex *sync.Mutex) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	clients[client.Conn.RemoteAddr().String()] = client
}

// Função para adicionar um novo carro ao mapa de carros
func addCar(carID int, cars *map[int]*Car) {
	(*cars)[carID] = &Car{
		ID:           carID,
		CurrentSpeed: 0,
		TireQuality:  100.0,
	}
}

// Função para atualizar a velocidade de um carro existente
func updateCarSpeed(carID int, speed int, cars *map[int]*Car) {
	(*cars)[carID].mu.Lock()
	defer (*cars)[carID].mu.Unlock()

	(*cars)[carID].CurrentSpeed = speed
}

// Função para atualizar a porcentagem de desgaste do pneu de um carro
func updateTireQuality(carID int, cars *map[int]*Car) float64 {
	(*cars)[carID].mu.Lock()
	defer (*cars)[carID].mu.Unlock()

	car := (*cars)[carID]
	car.TireQuality -= float64(car.CurrentSpeed) * 0.01 // Exemplo de diminuição progressiva
	if car.TireQuality < 0 {
		car.TireQuality = 0
	}
	tireQuality := car.TireQuality
	return tireQuality
}

// Função para resetar a porcentagem de desgaste do pneu para 100
func resetTireQuality(carID int, cars *map[int]*Car) {
	(*cars)[carID].mu.Lock()
	defer (*cars)[carID].mu.Unlock()

	(*cars)[carID].TireQuality = 100.0
}
