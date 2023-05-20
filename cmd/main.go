package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/CloudAceEmma/wechat"
	"github.com/gin-gonic/gin"
)

var core *wechat.Core

func main() {
	var err error

	if core, err = wechat.New(wechat.CoreOption{
		SyncMsgFunc:     onMsgRecv,
		SyncContactFunc: nil,
	}); err != nil {
		log.Fatal(err)
	}

	if err = core.GetUUID(); err != nil {
		log.Fatal(err)
	}

	fmt.Println(core.QrCode)    // print qrcode
	fmt.Println(core.QrCodeUrl) // qrcode url

	if err = core.PreLogin(); err != nil {
		log.Fatal(err)
	}

	if err = core.Login(); err != nil {
		log.Fatal(err)
	}

	if err = core.Init(); err != nil {
		log.Fatal(err)
	}

	if err = core.StatusNotify(); err != nil {
		log.Fatal(err)
	}

	if err = core.GetContact(); err != nil {
		log.Fatal(err)
	}

	interruptContext, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	ctx, cancel := context.WithCancel(interruptContext)

	defer cancel()
	defer stop()

	go periodicSync(PeriodicSyncOption{ // as a goroutine
		Cancel: cancel,
	})

	router := gin.Default()
	initAllApiHanlders(router) // init all api handlers

	localIp, err := getOutboundIP()
	if err != nil {
		log.Println("Failed to get local ip")
		log.Fatal(err)
	}

	listenAddr := localIp.String() + ":8080"
	log.Println("got local ip:", localIp.String())
	srv := &http.Server{Addr: listenAddr, Handler: router}

	go func() {
		if err := srv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Println("listen error:", err)
			cancel()
		}
	}()

	select {
	case <-ctx.Done(): // When sync returned 1101
		if err = core.Logout(); err != nil {
			log.Println("logout error:", err.Error())
		}
		if err := srv.Shutdown(ctx); err != nil {
			log.Println("shutdown error:", err.Error())
		}
		log.Println("logged out:", core.User.NickName)
	}
}

// Get preferred outbound ip of this machine
func getOutboundIP() (*net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return &localAddr.IP, nil
}
