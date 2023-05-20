package wechat

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/CloudAceEmma/wechat/utils"
	"github.com/gabriel-vasile/mimetype"
)

type MessageType int

const (
	Text           MessageType = 1
	Image          MessageType = 3
	Attach         MessageType = 6
	Voice          MessageType = 34
	Video          MessageType = 43
	MicroVideo     MessageType = 62
	Emoticon       MessageType = 47
	App            MessageType = 49
	Voip           MessageType = 50
	VoipNotify     MessageType = 52
	VoipInvite     MessageType = 53
	Location       MessageType = 48
	StatusNotify   MessageType = 51
	SystemNotice   MessageType = 9999
	PossibleFriend MessageType = 40
	Verify         MessageType = 37
	ShareCard      MessageType = 42
	System         MessageType = 1e4
	Recalled       MessageType = 10002
)

type MediaMessage struct {
	Name      string
	FileBytes []byte
}

func (core *Core) SendMsg(msgAny interface{}, to string) error {
	params := url.Values{}
	params.Add("pass_ticket", core.SessionData.PassTicket)
	params.Add("lang", "zh_CN")

	clientMsgId := utils.GetClientMsgId()

	var uri string
	var messageReq MessageRequest

	msgText, validText := msgAny.(string)
	if validText {
		uri = core.Config.Api.SendMsg
		messageReq = MessageRequest{
			FromUserName: core.User.UserName,
			ToUserName:   to,
			Content:      &msgText,
			MediaId:      nil,
			Type:         Text,
			ClientMsgId:  clientMsgId,
			LocalID:      clientMsgId,
		}
	}

	msgMedia, validMedia := msgAny.(MediaMessage)
	if validMedia {
		var msgType MessageType
		mediaType, err := utils.DetectMediaType(msgMedia.FileBytes)
		if err != nil {
			return err
		}

		params.Add("fun", "async")
		params.Add("f", "json")
		resp, err := core.UploadMedia(&msgMedia)
		if err != nil {
			return err
		}

		var content string = ""
		if *mediaType == "pic" {
			uri = core.Config.Api.SendMsgImg
			msgType = Image
		} else if *mediaType == "video" {
			uri = core.Config.Api.SendVideoMsg
			msgType = Video
		} else if *mediaType == "doc" {
			uri = core.Config.Api.SendAppMsg
			msgType = Attach
			mtype := mimetype.Detect(msgMedia.FileBytes)
			appmsg := Appmsg{
				Appid:   "wxeb7ec651dd0aefa9",
				Sdkver:  "",
				Title:   msgMedia.Name,
				Des:     "",
				Action:  "",
				Type:    "6",
				Content: "",
				URL:     "",
				Lowurl:  "",
				Appattach: Appattach{
					Totallen: strconv.Itoa(len(msgMedia.FileBytes)),
					Attachid: resp.MediaID,
					Fileext:  mtype.Extension()[1:],
				},
				Extinfo: "",
			}
			marshalled, err := xml.Marshal(&appmsg)
			if err != nil {
				return err
			}
			content = string(marshalled)
		} else {
			return ErrInvalidMsgType
		}

		messageReq = MessageRequest{
			FromUserName: core.User.UserName,
			ToUserName:   to,
			Type:         msgType,
			ClientMsgId:  clientMsgId,
			LocalID:      clientMsgId,
		}

		if len(content) > 0 {
			messageReq.Content = &content
		} else {
			messageReq.MediaId = &resp.MediaID
		}
	}

	if !validText && !validMedia {
		return ErrInvalidMsgType
	}

	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return err
	}
	u.RawQuery = params.Encode()

	baseRequest, err := core.GetBaseRequest()
	if err != nil {
		return err
	}

	data := SendMsgRequest{
		BaseRequest: *baseRequest,
		Scene:       0,
		Message:     messageReq,
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(data)
	if err != nil {
		return err
	}

	reqBody := bytes.NewReader(buf.Bytes())
	req, err := http.NewRequest("POST", u.String(), reqBody)
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

	var result SendMsgResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.BaseResponse.Ret != 0 {
		errMsg := utils.GetErrorMsgInt(result.BaseResponse.Ret)
		return errors.New(errMsg)
	}

	return nil
}

func (core *Core) UploadMedia(msg *MediaMessage) (*UploadMediaResponse, error) {
	mediaType, err := utils.DetectMediaType(msg.FileBytes)
	if err != nil {
		return nil, err
	}

	baseRequest, err := core.GetBaseRequest()
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("f", "json")

	u, err := url.ParseRequestURI(core.Config.Api.UploadMedia)
	if err != nil {
		return nil, err
	}
	u.RawQuery = params.Encode()

	clientMsgId := utils.GetClientMsgId()

	data := UploadMediaRequest{
		BaseRequest:   *baseRequest,
		ClientMediaId: clientMsgId,
		TotalLen:      len(msg.FileBytes),
		StartPos:      0,
		DataLen:       len(msg.FileBytes),
		MediaType:     4,
		UploadType:    2,
		FromUserName:  core.User.UserName,
		ToUserName:    core.User.UserName,
	}

	marshalled, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	gmt := time.Now().UTC().Format(http.TimeFormat)

	formData := &bytes.Buffer{}
	writer := multipart.NewWriter(formData)

	// Add the form fields to the form.
	writer.WriteField("name", msg.Name)
	writer.WriteField("type", mimetype.Detect(msg.FileBytes).String())
	writer.WriteField("lastModifiedDate", gmt)
	writer.WriteField("size", fmt.Sprintf("%d", len(msg.FileBytes)))
	writer.WriteField("mediatype", *mediaType)
	writer.WriteField("uploadmediarequest", string(marshalled))
	writer.WriteField("webwx_data_ticket", core.SessionData.DataTicket)
	writer.WriteField("pass_ticket", core.SessionData.PassTicket)

	// Create a new form field for the file.
	part, err := writer.CreateFormFile("filename", msg.Name)
	if err != nil {
		return nil, err
	}

	part.Write(msg.FileBytes)

	// Close writer before use it in post request
	writer.Close()

	req, err := http.NewRequest("POST", u.String(), formData)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
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

	var result UploadMediaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.BaseResponse.Ret != 0 {
		errMsg := utils.GetErrorMsgInt(result.BaseResponse.Ret)
		return nil, errors.New(errMsg)
	}

	return &result, nil
}
