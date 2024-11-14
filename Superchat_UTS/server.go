package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"Superchat_UTS/models"

	"github.com/gorilla/websocket"
)

var (
	users       = map[string]int{"Nathan": 1000} // Default user Nathan dengan saldo 1000
	superchats  = []models.Superchat{}
	mutex       = sync.Mutex{}
	upgrader    = websocket.Upgrader{}
	wsClients   = make(map[*websocket.Conn]bool) // Menyimpan klien WebSocket
	wsBroadcast = make(chan models.Superchat)    // Channel untuk menyiarkan donasi baru
)

func main() {
	go startUDPServer() // Server UDP untuk donasi
	go startTCPServer() // Server TCP untuk top-up dan tambah user

	http.HandleFunc("/api/donations", donationsHandler)
	http.HandleFunc("/ws", wsHandler) // WebSocket endpoint
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/streamer", streamerPageHandler)

	go handleWebSocketBroadcast() // Goroutine untuk broadcast WebSocket

	fmt.Println("Server HTTP berjalan di :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// startUDPServer menjalankan server UDP untuk menerima donasi
func startUDPServer() {
	udpAddr, err := net.ResolveUDPAddr("udp", ":8081")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Gagal menjalankan UDP server:", err)
		return
	}
	defer conn.Close()
	fmt.Println("Server UDP berjalan di :8081 untuk menerima donasi.")

	for {
		handleUDPConnection(conn)
	}
}

// handleUDPConnection memproses koneksi UDP untuk menerima donasi
func handleUDPConnection(conn *net.UDPConn) {
	buffer := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("Error membaca data UDP:", err)
		return
	}

	var chat models.Superchat
	err = json.Unmarshal(buffer[:n], &chat)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	mutex.Lock()
	superchats = append(superchats, chat)
	mutex.Unlock()

	fmt.Println("Donasi diterima:", chat)

	// Kirim data donasi ke semua klien WebSocket yang terhubung
	wsBroadcast <- chat
}

// startTCPServer menjalankan server TCP untuk menerima top-up dan tambah user
func startTCPServer() {
	listener, err := net.Listen("tcp", ":8082")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server TCP berjalan di :8082 untuk menerima top-up dan tambah user.")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Gagal menerima koneksi:", err)
			continue
		}
		go handleTopUpConnection(conn) // Menangani setiap koneksi di goroutine
	}
}

// handleTopUpConnection memproses permintaan TCP untuk top-up dan tambah user
func handleTopUpConnection(conn net.Conn) {
	defer conn.Close()

	var request map[string]interface{}
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&request); err != nil {
		fmt.Println("Error parsing request:", err)
		conn.Write([]byte("Invalid request"))
		return
	}

	action := request["action"].(string)
	switch action {
	case "view_balance":
		username := request["username"].(string)
		balance, exists := getUserBalance(username)
		if !exists {
			conn.Write([]byte("User tidak ditemukan"))
			return
		}
		conn.Write([]byte(fmt.Sprintf("%d", balance)))
	case "add_user":
		username := request["username"].(string)
		initialBalance := int(request["balance"].(float64))
		response := addUser(username, initialBalance)
		conn.Write([]byte(response))
	default:
		conn.Write([]byte("Aksi tidak dikenal"))
	}
}

// addUser menambahkan user baru ke sistem
func addUser(username string, initialBalance int) string {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := users[username]; exists {
		return "User sudah ada."
	}
	users[username] = initialBalance
	return fmt.Sprintf("User %s berhasil ditambahkan dengan saldo %d", username, initialBalance)
}

// getUserBalance mengembalikan saldo user
func getUserBalance(username string) (int, bool) {
	mutex.Lock()
	defer mutex.Unlock()

	balance, exists := users[username]
	return balance, exists
}

// donationsHandler mengirimkan daftar donasi sebagai JSON
func donationsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	mutex.Lock()
	defer mutex.Unlock()
	json.NewEncoder(w).Encode(superchats)
}

// wsHandler mengelola koneksi WebSocket
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error membuka koneksi WebSocket:", err)
		return
	}
	defer conn.Close()

	wsClients[conn] = true
	fmt.Println("Klien WebSocket terhubung.")

	for {
		if _, _, err := conn.NextReader(); err != nil {
			delete(wsClients, conn)
			fmt.Println("Klien WebSocket terputus.")
			break
		}
	}
}

// handleWebSocketBroadcast menyiarkan donasi baru ke semua klien WebSocket
func handleWebSocketBroadcast() {
	for {
		donation := <-wsBroadcast
		for client := range wsClients {
			err := client.WriteJSON(donation)
			if err != nil {
				fmt.Println("Error menulis ke WebSocket:", err)
				client.Close()
				delete(wsClients, client)
			}
		}
	}
}

// streamerPageHandler melayani halaman streamer
func streamerPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/streamer.html")
}
