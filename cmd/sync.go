package main

import (
	"context"
	"log"
	"strings"

	"github.com/CloudAceEmma/wechat"
)

type PingMessage struct {
	FromUserName string
	ToNickName   string
}

type PeriodicSyncOption struct {
	Cancel context.CancelFunc
}

func periodicSync(options PeriodicSyncOption) {
	var errSlice []bool
	for {
		var err error
		if err = core.SyncPolling(); err == nil {
			if len(errSlice) >= 10 {
				options.Cancel()
			}
			errSlice = nil
			continue
		}
		if err == wechat.ErrAlreadyLoggedOut {
			if len(errSlice) >= 10 {
				options.Cancel()
			}
			errSlice = append(errSlice, true)
		} else {
			log.Println("sync error:", err.Error())
		}
	}
}

func onMsgRecv(data *wechat.SyncResponse) error {
	for _, message := range data.AddMsgList {
		if len(message.Content) == 0 {
			return nil
		}

		pingMsg := extractPingMessage(message.Content)
		userNick := core.User.NickName

		if pingMsg != nil &&
			strings.HasPrefix(pingMsg.ToNickName, userNick) {
			to := message.FromUserName
			msg := "What can I do for you?"
			if err := core.SendMsg(msg, to); err != nil {
				log.Println("Send message error:", err)
			}
		}
	}
	return nil
}

func extractPingMessage(message string) *PingMessage {
	if strings.HasPrefix(message, "@") {
		fromIdxEnd := strings.Index(message, ":")
		toIdxStart := strings.Index(message, "<br/>@")

		if fromIdxEnd == -1 || toIdxStart == -1 {
			return nil
		}

		padding := len("<br/>@")

		return &PingMessage{
			FromUserName: message[0:fromIdxEnd],
			ToNickName:   message[toIdxStart+padding:],
		}
	}
	return nil
}
