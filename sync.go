package wechat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/CloudAceEmma/wechat/utils"
)

type SyncType = int

const (
	Normal         SyncType = 0
	MessageContact SyncType = 2
	ModProfile     SyncType = 4
	ModChatRoom    SyncType = 7
)

type SyncFunc = func(data *SyncResponse) error

func (core *Core) StatusNotify() error {
	params := url.Values{}
	params.Add("pass_ticket", core.SessionData.PassTicket)
	params.Add("lang", "zh_CN")

	u, err := url.ParseRequestURI(core.Config.Api.StatusNotify)
	if err != nil {
		return err
	}
	u.RawQuery = params.Encode()

	baseRequest, err := core.GetBaseRequest()
	if err != nil {
		return err
	}

	var code int
	var userName string
	if len(core.NotifyUserName) > 0 {
		code = 1
		userName = core.NotifyUserName
	} else {
		code = 3
		userName = core.User.UserName
	}
	core.NotifyUserName = ""

	data := StatusNotifyRequest{
		BaseRequest:  *baseRequest,
		Code:         code,
		FromUserName: core.User.UserName,
		ToUserName:   userName,
		ClientMsgId:  time.Now().UnixNano(),
	}

	marshalled, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(marshalled))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := core.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg := utils.GetErrorMsgInt(resp.StatusCode)
		return errors.New(errMsg)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result StatusNotifyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.BaseResponse.Ret != 0 {
		errMsg := utils.GetErrorMsgInt(result.BaseResponse.Ret)
		return errors.New(errMsg)
	}

	return nil
}

func (core *Core) SyncCheck() error {
	ts := time.Now().UnixNano() / int64(time.Millisecond)

	params := url.Values{}
	params.Add("r", fmt.Sprintf("%d", int64(ts)))
	params.Add("sid", core.SessionData.Sid)
	params.Add("uin", core.SessionData.Uin)
	params.Add("skey", core.SessionData.Skey)
	params.Add("deviceid", utils.GetDeviceID())
	params.Add("synckey", core.FormatedSyncKey)

	u, err := url.ParseRequestURI(core.Config.Api.SyncCheck)
	if err != nil {
		return err
	}

	u.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}

	resp, err := core.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if strings.Contains(string(body), "retcode:\"1101\"") {
		return ErrAlreadyLoggedOut
	}

	start := strings.Index(string(body), "selector:")
	start += len("selector:") + 1
	end := len(string(body)) - 2

	selectorStr := string(body)[start:end]
	selector, err := strconv.ParseInt(selectorStr, 10, 64)
	if err != nil {
		return err
	}
	core.SyncSelector = SyncType(selector)

	return nil
}

func (core *Core) Sync() (*SyncResponse, error) {
	ts := ^time.Now().UnixNano()

	params := url.Values{}
	params.Add("sid", core.SessionData.Sid)
	params.Add("skey", core.SessionData.Skey)
	params.Add("pass_ticket", core.SessionData.PassTicket)
	params.Add("lang", "zh_CN")

	u, err := url.ParseRequestURI(core.Config.Api.Sync)
	if err != nil {
		return nil, err
	}

	u.RawQuery = params.Encode()

	baseRequest, err := core.GetBaseRequest()
	if err != nil {
		return nil, err
	}

	data := SyncRequest{
		BaseRequest: *baseRequest,
		SyncKey:     core.SyncKey,
		RR:          int64(ts),
	}

	marshalled, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(marshalled))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := core.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg := utils.GetErrorMsgInt(resp.StatusCode)
		return nil, errors.New(errMsg)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result SyncResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.BaseResponse.Ret != 0 {
		errMsg := utils.GetErrorMsgInt(result.BaseResponse.Ret)
		return nil, errors.New(errMsg)
	}

	core.SyncKey = result.SyncCheckKey
	core.SetFormatedSyncKey(core.SyncKey)

	return &result, nil
}

func (core *Core) SetFormatedSyncKey(syncKey SyncKey) {
	syncKeyList := make([]string, len(syncKey.List))
	for i, item := range syncKey.List {
		syncKeyList[i] = strconv.Itoa(item.Key) + "_" +
			strconv.Itoa(item.Val)
	}
	core.FormatedSyncKey = strings.Join(syncKeyList, "|")
}

func (core *Core) SyncPolling() error {
	if err := core.SyncCheck(); err != nil {
		return err
	}

	if core.SyncSelector == Normal {
		return nil
	}

	var err error
	var data *SyncResponse
	if data, err = core.Sync(); err != nil {
		return err
	}

	core.LastSyncTime = time.Now().UnixNano()

	switch core.SyncSelector {
	case MessageContact:
		core.modDelContact(data) // This will not fail
		if core.SyncMsgFunc == nil || data.AddMsgCount == 0 {
			goto sync_contact
		}
		if err := core.SyncMsgFunc(data); err != nil {
			return err
		}

	sync_contact:
		if core.SyncContactFunc == nil {
			return nil
		}
		if err := core.SyncContactFunc(data); err != nil {
			return err
		}
	case ModProfile:
		fmt.Println("profile modified") // TODO Handle this
	case ModChatRoom:
		fmt.Println("chatroom modified") // TODO Handle this
	}

	return nil
}

func (core *Core) modDelContact(data *SyncResponse) {
	if data.ModContactCount > 0 {
		// Handle new contacts
		for _, contact := range data.ModContactList {
			log.Println("mod contact: " + contact.UserName)
			core.ContactMap[contact.UserName] = contact
		}
	}

	if data.DelContactCount > 0 {
		// Handle new contacts
		for _, contact := range data.DelContactList {
			log.Println("del contact: " + contact.UserName)
			delete(core.ContactMap, contact.UserName)
		}
	}
}
