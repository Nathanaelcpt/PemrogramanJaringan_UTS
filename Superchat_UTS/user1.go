package main

import (
	"Superchat_UTS/models"
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
)

func main() {
	RunUser1()
}

func RunUser1() {
	scanner := bufio.NewScanner(os.Stdin)
	username := "Nathan" // Default user

	for {
		fmt.Println("\n=== Menu Superchat ===")
		fmt.Println("1. Kirim Superchat")
		fmt.Println("2. Lihat Saldo")
		fmt.Println("3. Tambah User Baru")
		fmt.Println("4. Keluar")
		fmt.Print("Pilih opsi: ")

		scanner.Scan()
		option := scanner.Text()

		switch option {
		case "1":
			sendSuperchat(scanner, username)
		case "2":
			viewBalance(username)
		case "3":
			addNewUser(scanner)
		case "4":
			fmt.Println("Terima kasih!")
			return
		default:
			fmt.Println("Opsi tidak valid. Silakan coba lagi.")
		}
	}
}

func sendSuperchat(scanner *bufio.Scanner, username string) {
	fmt.Print("Masukkan jumlah donasi: ")
	scanner.Scan()
	amount, err := strconv.Atoi(scanner.Text())
	if err != nil || amount <= 0 {
		fmt.Println("Jumlah donasi tidak valid.")
		return
	}

	balance := viewBalance(username)
	if balance < amount {
		fmt.Println("Saldo tidak mencukupi untuk donasi.")
		return
	}

	fmt.Print("Masukkan pesan: ")
	scanner.Scan()
	message := scanner.Text()

	conn, err := net.Dial("udp", "localhost:8081")
	if err != nil {
		fmt.Println("Gagal terhubung ke server UDP:", err)
		return
	}
	defer conn.Close()

	chat := models.Superchat{
		Sender:  username,
		Amount:  amount,
		Message: message,
	}
	data, err := json.Marshal(chat)
	if err != nil {
		fmt.Println("Error serializing data:", err)
		return
	}

	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("Error mengirim data:", err)
		return
	}

	fmt.Println("Superchat berhasil dikirim melalui UDP!")
}

func viewBalance(username string) int {
	conn, err := net.Dial("tcp", "localhost:8082")
	if err != nil {
		fmt.Println("Gagal terhubung ke server:", err)
		return 0
	}
	defer conn.Close()

	request := struct {
		Username string `json:"username"`
		Action   string `json:"action"`
	}{
		Username: username,
		Action:   "view_balance",
	}
	data, _ := json.Marshal(request)
	conn.Write(data)

	buffer := make([]byte, 1024)
	n, _ := conn.Read(buffer)
	balance, _ := strconv.Atoi(string(buffer[:n]))
	fmt.Printf("Saldo Anda saat ini adalah: %d\n", balance)
	return balance
}

func addNewUser(scanner *bufio.Scanner) {
	fmt.Print("Masukkan nama user baru: ")
	scanner.Scan()
	newUsername := scanner.Text()

	fmt.Print("Masukkan saldo awal: ")
	scanner.Scan()
	initialBalance, err := strconv.Atoi(scanner.Text())
	if err != nil || initialBalance < 0 {
		fmt.Println("Saldo tidak valid.")
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
		Balance  int    `json:"balance"`
		Action   string `json:"action"`
	}{
		Username: newUsername,
		Balance:  initialBalance,
		Action:   "add_user",
	}
	data, _ := json.Marshal(request)
	conn.Write(data)

	buffer := make([]byte, 1024)
	n, _ := conn.Read(buffer)
	fmt.Println(string(buffer[:n]))
}
