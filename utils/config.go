package utils

import (
	"regexp"
	"strings"
)

type Api struct {
	JsLogin         string
	Login           string
	SyncCheck       string
	DownloadMedia   string
	UploadMedia     string
	Preview         string
	Init            string
	GetContact      string
	Sync            string
	BatchGetContact string
	GetIcon         string
	SendMsg         string
	SendMsgImg      string
	SendVideoMsg    string
	SendEmoticon    string
	SendAppMsg      string
	GetHeadMsg      string
	GetMsgImg       string
	GetMedia        string
	GetVideo        string
	Logout          string
	GetVoice        string
	UpdateChatRoom  string
	CreateChatRoom  string
	StatusNotify    string
	CheckUrl        string
	VerifyUser      string
	SendFeedback    string
	StatusReport    string
	SearchContact   string
	OpLog           string
	CheckUpload     string
	RevokeMsg       string
}

type OpLogCmdId struct {
	TopContact     int
	ModRemarkNanme int
}

type DefaultContact struct {
	FileHelper      string `default:"filehelper"`
	NewsApp         string `default:"newsapp"`
	RecommandHelper string `default:"fmessage"`
}

type ContactFlag struct {
	Contact             bool
	ChatContact         bool
	ChatRoomContact     bool
	BlackListContact    bool
	DomainContact       bool
	HideContact         bool
	FavouriteContact    bool
	ThirdAppContact     bool
	SnsBlackListContact bool
	NotifyCloseContact  bool
	TopContact          bool
}

type Config struct {
	Lang           string `default:"zh-CN"`
	EmoticonReg    string `default:"img\\sclass="(qq)?emoji (qq)?emoji([\\da-f]*?)"\\s(text="[^<>(\\s]*")?\\s?src="[^<>(\\s]*"\\s*"`
	ResourcePath   string `default:"/zh_CN/htmledition/v2/"`
	OpLogCmdId     OpLogCmdId
	DefaultContact DefaultContact
	ContactFlag    ContactFlag
	Multimedia     struct {
		UserAttribute struct {
			Biz         bool
			Famous      bool
			BizBig      bool
			BizBrand    bool
			BizVerified bool
		}
		Data struct {
			Text            bool
			Html            bool
			Image           bool
			PrivateMsgText  bool
			PrivateMsgHtml  bool
			PrivateMsgImage bool
		}
	}
	SyncCheckRetSuccess int
	SyncCheckRetLogout  int
	Origin              string
	BaseUrl             string
	Api                 Api
}

type ConfigOption struct {
	Host string
}

func NewConfig(option ConfigOption) (*Config, error) {
	getHost := func() string {
		if len(option.Host) > 0 {
			return option.Host
		}
		return "wx.qq.com"
	}
	host := getHost()
	origin := "https://" + host
	loginUrl := "login.wx.qq.com"
	fileUrl := "file.wx.qq.com"
	pushUrl := "webpush.weixin.qq.com"

	found, err := regexp.MatchString("\\w+(\\.qq\\.com|\\.wechat\\.com)", host)
	if err != nil {
		return nil, err
	}

	if found {
		var prefix string
		var suffix string

		if strings.HasSuffix(host, ".qq.com") {
			if strings.HasPrefix(host, "wx.") {
				prefix = "wx."
			} else if strings.HasPrefix(host, "wx2.") {
				prefix = "wx2."
			} else if strings.HasPrefix(host, "wx8.") {
				prefix = "wx8."
			} else {
				prefix = "wx."
			}
			suffix = "qq.com"
		} else {
			if strings.HasPrefix(host, "web.") {
				prefix = "web."
			} else if strings.HasPrefix(host, "web2") {
				prefix = "web2."
			} else {
				prefix = "web."
			}
			suffix = "wechat.com"
		}
		loginUrl = "login." + prefix + suffix
		fileUrl = "file." + prefix + suffix
		pushUrl = "webpush." + prefix + suffix
	}

	conf := Config{
		SyncCheckRetSuccess: 0,
		SyncCheckRetLogout:  1101,
		Origin:              origin,
		BaseUrl:             origin + "/cgi-bin/mmwebwx-bin",
		Api: Api{
			JsLogin:         "https://" + loginUrl + "/jslogin?appid=wx782c26e4c19acffb&fun=new&lang=zh-CN&redirect_uri=https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxnewloginpage?mod=desktop",
			Login:           "https://" + loginUrl + "/cgi-bin/mmwebwx-bin/login",
			SyncCheck:       "https://" + pushUrl + "/cgi-bin/mmwebwx-bin/synccheck",
			DownloadMedia:   "https://" + fileUrl + "/cgi-bin/mmwebwx-bin/webwxgetmedia",
			UploadMedia:     "https://" + fileUrl + "/cgi-bin/mmwebwx-bin/webwxuploadmedia",
			Preview:         origin + "/cgi-bin/mmwebwx-bin/webwx" + "preview",
			Init:            origin + "/cgi-bin/mmwebwx-bin/webwx" + "init",
			GetContact:      origin + "/cgi-bin/mmwebwx-bin/webwx" + "getcontact",
			Sync:            origin + "/cgi-bin/mmwebwx-bin/webwx" + "sync",
			BatchGetContact: origin + "/cgi-bin/mmwebwx-bin/webwx" + "batchgetcontact",
			GetIcon:         origin + "/cgi-bin/mmwebwx-bin/webwx" + "geticon",
			SendMsg:         origin + "/cgi-bin/mmwebwx-bin/webwx" + "sendmsg",
			SendMsgImg:      origin + "/cgi-bin/mmwebwx-bin/webwx" + "sendmsgimg",
			SendVideoMsg:    origin + "/cgi-bin/mmwebwx-bin/webwx" + "sendvideomsg",
			SendEmoticon:    origin + "/cgi-bin/mmwebwx-bin/webwx" + "sendemoticon",
			SendAppMsg:      origin + "/cgi-bin/mmwebwx-bin/webwx" + "sendappmsg",
			GetHeadMsg:      origin + "/cgi-bin/mmwebwx-bin/webwx" + "getheadimg",
			GetMsgImg:       origin + "/cgi-bin/mmwebwx-bin/webwx" + "getmsgimg",
			GetMedia:        origin + "/cgi-bin/mmwebwx-bin/webwx" + "getmedia",
			GetVideo:        origin + "/cgi-bin/mmwebwx-bin/webwx" + "getvideo",
			Logout:          origin + "/cgi-bin/mmwebwx-bin/webwx" + "logout",
			GetVoice:        origin + "/cgi-bin/mmwebwx-bin/webwx" + "getvoice",
			UpdateChatRoom:  origin + "/cgi-bin/mmwebwx-bin/webwx" + "updatechatroom",
			CreateChatRoom:  origin + "/cgi-bin/mmwebwx-bin/webwx" + "createchatroom",
			StatusNotify:    origin + "/cgi-bin/mmwebwx-bin/webwx" + "statusnotify",
			CheckUrl:        origin + "/cgi-bin/mmwebwx-bin/webwx" + "checkurl",
			VerifyUser:      origin + "/cgi-bin/mmwebwx-bin/webwx" + "verifyuser",
			SendFeedback:    origin + "/cgi-bin/mmwebwx-bin/webwx" + "sendfeedback",
			StatusReport:    origin + "/cgi-bin/mmwebwx-bin/webwx" + "statreport",
			SearchContact:   origin + "/cgi-bin/mmwebwx-bin/webwx" + "seatchcontact",
			OpLog:           origin + "/cgi-bin/mmwebwx-bin/webwx" + "oplog",
			CheckUpload:     origin + "/cgi-bin/mmwebwx-bin/webwx" + "checkupload",
			RevokeMsg:       origin + "/cgi-bin/mmwebwx-bin/webwx" + "revokemsg",
		},
	}
	return &conf, nil
}
