package models

type Superchat struct {
	Sender  string `json:"sender"`
	Amount  int    `json:"amount"`
	Message string `json:"message"`
}
