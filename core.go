package wechat

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/CloudAceEmma/wechat/utils"
	"github.com/skip2/go-qrcode"

	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type SessionData struct {
	UUID       string
	Skey       string
	Sid        string
	Uin        string
	PassTicket string
	DataTicket string
}

type Core struct {
	Config          utils.Config
	SessionData     SessionData
	User            User
	Avatar          string
	RedirectUri     string
	QrCodeUrl       string
	QrCode          string
	NotifyUserName  string
	ContactMap      map[string]Contact
	LastSyncTime    int64
	SyncKey         SyncKey
	SyncSelector    SyncType
	FormatedSyncKey string
	ContactSeq      int
	Client          *http.Client
	SyncMsgFunc     SyncFunc
	SyncContactFunc SyncFunc
}

type CoreOption struct {
	SyncMsgFunc     SyncFunc
	SyncContactFunc SyncFunc
}

func New(options CoreOption) (*Core, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	core := Core{
		SyncMsgFunc:     options.SyncMsgFunc,
		SyncContactFunc: options.SyncContactFunc,
		Client: &http.Client{
			CheckRedirect: nil,
			Jar:           jar,
		},
	}

	config, err := utils.NewConfig(utils.ConfigOption{})
	if err != nil {
		return nil, err
	}

	core.Config = *config
	return &core, nil
}

func (core *Core) GetUUID() error {
	resp, err := core.Client.Post(core.Config.Api.JsLogin, "", http.NoBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	start := strings.Index(string(body), "window.QRLogin.uuid = ")
	start += len("window.QRLogin.uuid = ") + 1
	end := len(string(body)) - 2

	uuid := string(body)[start:end]

	if resp.StatusCode != http.StatusOK {
		errMsg := utils.GetErrorMsgInt(resp.StatusCode)
		return errors.New(errMsg)
	}

	core.QrCodeUrl = "https://login.weixin.qq.com/qrcode/" + uuid
	qrCodeContent := "https://login.weixin.qq.com/l/" + uuid

	qrCode, err := qrcode.New(qrCodeContent, qrcode.Medium)
	if err != nil {
		return err
	}

	core.QrCode = qrCode.ToSmallString(false)
	core.SessionData.UUID = uuid
	return nil
}

func (core *Core) PreLogin() error {
	ts := ^time.Now().UnixNano()

	params := url.Values{}
	params.Add("tip", "0")
	params.Add("uuid", core.SessionData.UUID)
	params.Add("loginicon", "true")
	params.Add("r", fmt.Sprintf("%d", int64(ts)))

	u, err := url.ParseRequestURI(core.Config.Api.Login)
	if err != nil {
		return err
	}
	u.RawQuery = params.Encode()

	resp, err := core.Client.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	httpStatusSuccess := strings.Contains(string(body), "window.redirect_uri")
	httpStatusCreated := strings.Contains(string(body), "window.userAvatar")

	if !httpStatusCreated && !httpStatusSuccess {
		start := strings.Index(string(body), "window.code=")
		start += len("window.code=") + 1
		end := len(string(body)) - 2

		codeStr := string(body)[start:end]

		errMsg := utils.GetErrorMsgStr(codeStr)
		return errors.New(errMsg)
	}

	if httpStatusSuccess {
		start := strings.Index(string(body), "redirect_uri=")
		start += len("redirect_uri=") + 1
		end := len(string(body)) - 2
		redirectUri := string(body)[start:end]

		u, err := url.Parse(redirectUri)
		if err != nil {
			return err
		}

		config, err := utils.NewConfig(utils.ConfigOption{Host: u.Hostname()})
		if err != nil {
			return err
		}

		core.Config = *config
		core.RedirectUri = redirectUri
	}

	if httpStatusCreated {
		start := strings.Index(string(body), "userAvatar = ")
		start += len("userAvatar = ") + 1
		end := len(string(body)) - 2

		core.Avatar = string(body)[start:end]

		if err := core.PreLogin(); err != nil {
			return err
		}
	}

	return nil
}

func (core *Core) Login() error {
	core.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, err := http.NewRequest("GET", core.RedirectUri, nil)
	if err != nil {
		return err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36")
	req.Header.Add("client-version", "2.0.0")
	req.Header.Add("referer", "https://wx.qq.com/?&lang=zh_CN&target=t")
	req.Header.Add("extspam", "Go8FCIkFEokFCggwMDAwMDAwMRAGGvAESySibk50w5Wb3uTl2c2h64jVVrV7gNs06GFlWplHQbY/5FfiO++1yH4ykCyNPWKXmco+wfQzK5R98D3so7rJ5LmGFvBLjGceleySrc3SOf2Pc1gVehzJgODeS0lDL3/I/0S2SSE98YgKleq6Uqx6ndTy9yaL9qFxJL7eiA/R3SEfTaW1SBoSITIu+EEkXff+Pv8NHOk7N57rcGk1w0ZzRrQDkXTOXFN2iHYIzAAZPIOY45Lsh+A4slpgnDiaOvRtlQYCt97nmPLuTipOJ8Qc5pM7ZsOsAPPrCQL7nK0I7aPrFDF0q4ziUUKettzW8MrAaiVfmbD1/VkmLNVqqZVvBCtRblXb5FHmtS8FxnqCzYP4WFvz3T0TcrOqwLX1M/DQvcHaGGw0B0y4bZMs7lVScGBFxMj3vbFi2SRKbKhaitxHfYHAOAa0X7/MSS0RNAjdwoyGHeOepXOKY+h3iHeqCvgOH6LOifdHf/1aaZNwSkGotYnYScW8Yx63LnSwba7+hESrtPa/huRmB9KWvMCKbDThL/nne14hnL277EDCSocPu3rOSYjuB9gKSOdVmWsj9Dxb/iZIe+S6AiG29Esm+/eUacSba0k8wn5HhHg9d4tIcixrxveflc8vi2/wNQGVFNsGO6tB5WF0xf/plngOvQ1/ivGV/C1Qpdhzznh0ExAVJ6dwzNg7qIEBaw+BzTJTUuRcPk92Sn6QDn2Pu3mpONaEumacjW4w6ipPnPw+g2TfywJjeEcpSZaP4Q3YV5HG8D6UjWA4GSkBKculWpdCMadx0usMomsSS/74QgpYqcPkmamB4nVv1JxczYITIqItIKjD35IGKAUwAA==")
	resp, err := core.Client.Do(req)
	if err != nil {
		return err
	}
	core.Client.CheckRedirect = nil
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMovedPermanently {
		errMsg := utils.GetErrorMsgInt(resp.StatusCode)
		return errors.New(errMsg)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result LoginResponse
	if err := xml.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.Ret != "0" {
		return errors.New(result.Message)
	}

	core.SessionData.Skey = *result.Skey
	core.SessionData.Sid = *result.Wxsid
	core.SessionData.Uin = *result.Wxuin
	core.SessionData.PassTicket = *result.PassTicket

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "webwx_data_ticket" {
			core.SessionData.DataTicket = cookie.Value
		}
		if cookie.Name == "wxuin" {
			core.SessionData.Uin = cookie.Value
		}
		if cookie.Name == "wxsid" {
			core.SessionData.Sid = cookie.Value
		}
		if cookie.Name == "pass_ticket" {
			core.SessionData.PassTicket = cookie.Value
		}
	}

	return nil
}

func (core *Core) Init() error {
	ts := time.Now().UnixNano()
	r := ts / -1579

	params := url.Values{}
	params.Add("pass_ticket", core.SessionData.PassTicket)
	params.Add("r", fmt.Sprintf("%d", int64(r)))

	u, err := url.ParseRequestURI(core.Config.Api.Init)
	if err != nil {
		return err
	}
	u.RawQuery = params.Encode()

	baseRequest, err := core.GetBaseRequest()
	if err != nil {
		return err
	}

	data := InitRequest{
		BaseRequest: *baseRequest,
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

	var result InitResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.BaseResponse.Ret == core.Config.SyncCheckRetLogout {
		return ErrAlreadyLoggedOut
	}

	if result.BaseResponse.Ret != 0 {
		errMsg := utils.GetErrorMsgInt(result.BaseResponse.Ret)
		return errors.New(errMsg)
	}

	if len(result.SKey) > 0 {
		core.SessionData.Skey = result.SKey
	}

	core.SyncKey = result.SyncKey
	core.SetFormatedSyncKey(result.SyncKey)

	core.User = result.User
	core.ContactMap = make(map[string]Contact)

	for _, contact := range result.ContactList {
		core.ContactMap[contact.UserName] = contact
	}

	log.Println("logged in:", core.User.NickName)

	return nil
}

func (core *Core) GetBaseRequest() (*BaseRequest, error) {
	uin, err := strconv.ParseInt(core.SessionData.Uin, 10, 64)
	if err != nil {
		return nil, err
	}

	return &BaseRequest{
		Uin:      uin,
		Sid:      core.SessionData.Sid,
		Skey:     core.SessionData.Skey,
		DeviceID: utils.GetDeviceID(),
	}, nil
}

func (core *Core) Logout() error {
	params := url.Values{}
	params.Add("redirect", "1")
	params.Add("skey", core.SessionData.Skey)
	params.Add("type", "0")
	params.Add("lang", "zh_CN")

	u, err := url.ParseRequestURI(core.Config.Api.Logout)
	if err != nil {
		return err
	}
	u.RawQuery = params.Encode()

	req, err := http.NewRequest("POST", u.String(), nil)
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

	return nil
}
