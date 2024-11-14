package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
)

func main() {
	RunUser2()
}

func RunUser2() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Masukkan username untuk top-up: ")
	scanner.Scan()
	username := scanner.Text()

	for {
		fmt.Println("\n=== Menu Top Up Saldo ===")
		fmt.Println("1. Top Up Saldo")
		fmt.Println("2. Keluar")
		fmt.Print("Pilih opsi: ")

		scanner.Scan()
		option := scanner.Text()

		switch option {
		case "1":
			topup(scanner, username)
		case "2":
			fmt.Println("Terima kasih!")
			return
		default:
			fmt.Println("Opsi tidak valid. Silakan coba lagi.")
		}
	}
}

func topup(scanner *bufio.Scanner, username string) {
	fmt.Print("Masukkan jumlah top-up: ")
	scanner.Scan()
	amount, err := strconv.Atoi(scanner.Text())
	if err != nil || amount <= 0 {
		fmt.Println("Jumlah top-up tidak valid.")
		return
	}

	conn, err := net.Dial("tcp", "localhost:8082")
	if err != nil {
		fmt.Println("Gagal terhubung ke server:", err)
		return
	}
	defer conn.Close()

	request := struct {
		Username string `json:"username"`
		Amount   int    `json:"amount"`
		Action   string `json:"action"`
	}{
		Username: username,
		Amount:   amount,
		Action:   "top_up",
	}
	data, _ := json.Marshal(request)
	conn.Write(data)

	buffer := make([]byte, 1024)
	n, _ := conn.Read(buffer)
	fmt.Println("Respons dari server:", string(buffer[:n]))
}
