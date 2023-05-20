package main

type SendMessageRequest struct {
	TextMsg  string `json:"TextMsg"`
	NickName string `json:"NickName"`
}

type Message struct {
	Msg string `json:"message"`
}
