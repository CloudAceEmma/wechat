package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/CloudAceEmma/wechat"
	"github.com/gin-gonic/gin"
)

func initAllApiHanlders(engine *gin.Engine) {
	engine.GET("/demo", demoHandler)
	engine.POST("/sendmsg", sendMsgHandler)
	engine.POST("/sendfile", sendFileHandler)
}

func demoHandler(c *gin.Context) {
	to := "filehelper"
	msg := "message sent by wechat bot"
	if err := core.SendMsg(msg, to); err != nil {
		log.Println("Send message error:", err)
	}

	jpgBytes, err := ioutil.ReadFile("media/gopher.jpg")
	if err != nil {
		log.Println("Read file error:", err)
	}

	msgJpg := wechat.MediaMessage{
		Name:      "gopher.jpg",
		FileBytes: jpgBytes,
	}
	if err := core.SendMsg(msgJpg, to); err != nil {
		log.Println("Send message error:", err)
	}

	mp4Bytes, err := ioutil.ReadFile("media/gopher.mp4")
	if err != nil {
		log.Println("Read file error:", err)
	}

	msgMp4 := wechat.MediaMessage{
		Name:      "gopher.mp4",
		FileBytes: mp4Bytes,
	}
	if err := core.SendMsg(msgMp4, to); err != nil {
		log.Println("Send message error:", err)
	}

	txtBytes, err := ioutil.ReadFile("media/hello.txt")
	if err != nil {
		log.Println("Read file error:", err)
	}

	msgTxt := wechat.MediaMessage{
		Name:      "hello.txt",
		FileBytes: txtBytes,
	}
	if err := core.SendMsg(msgTxt, to); err != nil {
		log.Println("Send message error:", err)
	}

	c.IndentedJSON(http.StatusOK, Message{Msg: "success"})
}

func sendMsgHandler(c *gin.Context) {
	sendMsgReq := SendMessageRequest{}

	if err := c.ShouldBind(&sendMsgReq); err != nil {
		c.IndentedJSON(http.StatusBadRequest, Message{
			Msg: err.Error(),
		})
		return
	}

	if len(sendMsgReq.TextMsg) == 0 || len(sendMsgReq.NickName) == 0 {
		c.IndentedJSON(http.StatusBadRequest, Message{
			Msg: "bad request",
		})
		return
	}

	var to string
	for _, contact := range core.ContactMap {
		if strings.Contains(contact.NickName, sendMsgReq.NickName) {
			to = contact.UserName
			break
		}
	}

	if len(to) == 0 {
		c.IndentedJSON(http.StatusNotFound, Message{
			Msg: "contact not found: " + sendMsgReq.NickName,
		})
		return
	}

	if len(sendMsgReq.TextMsg) > 0 {
		if err := core.SendMsg(sendMsgReq.TextMsg, to); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, Message{
				Msg: err.Error(),
			})
			return
		}
	}

	c.IndentedJSON(http.StatusOK, Message{Msg: "success"})
}

func sendFileHandler(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, Message{
			Msg: err.Error(),
		})
		return
	}

	nickName := c.Request.PostFormValue("NickName")

	buf := new(bytes.Buffer)
	buf.ReadFrom(file)

	var to string
	for _, contact := range core.ContactMap {
		if strings.Contains(contact.NickName, nickName) {
			to = contact.UserName
			break
		}
	}

	if len(to) == 0 {
		c.IndentedJSON(http.StatusNotFound, Message{
			Msg: "contact not found: " + nickName,
		})
		return
	}

	if len(buf.Bytes()) > 0 {
		if err := core.SendMsg(wechat.MediaMessage{
			Name:      header.Filename,
			FileBytes: buf.Bytes(),
		}, to); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, Message{
				Msg: err.Error(),
			})
			return
		}
	}
	c.IndentedJSON(http.StatusOK, Message{Msg: "success"})
}
