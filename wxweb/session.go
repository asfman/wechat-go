package wxweb

import (
  "fmt"
  "os"
  "time"
	"encoding/json"

	"github.com/mdp/qrterminal"
)

type Session struct {
	QrcodeUUID    string
	RedirectUri   string
	XmlResponse	  *XmlResponse
	SyncKeyList    *SyncKeyList
	MemberList    []*User
	Bot           *User
	Api           *Api
}

func CreateSession() (*Session, error) {
  api := NewApi()
	uuid, err := api.GetUUID()
  fmt.Printf("uuid: %s\n", uuid)
	if err != nil {
		return nil, err
	}
  qrterminal.Generate("https://login.weixin.qq.com/l/"+uuid, qrterminal.L, os.Stdout)
  return &Session{
		Api:					api,
		QrcodeUUID:		uuid,
		XmlResponse:	&XmlResponse{},
  }, nil
}

func (session *Session) Login() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		redirectUri, err := session.Api.Login(session.QrcodeUUID)
		if err != nil {
			fmt.Printf(err.Error())
		} else {
      session.RedirectUri = redirectUri
      fmt.Println(redirectUri)
      return
		}
	}
}

func (session *Session) WebNewLoginPage() {
  redirectUri :=  session.RedirectUri
  if redirectUri == "" {
    fmt.Println("redirect_uri empty!")
    return
  }
  session.Api.WebNewLoginPage(redirectUri, session.XmlResponse)
}

func (session *Session) WebWxInit() {
	if session.XmlResponse.PassTicket == "" {
		fmt.Println("pass ticket empty!")
		return
	}
	b, _ := session.Api.WebWxInit(session.XmlResponse)
	var jm map[string]interface{}
	if err := json.Unmarshal(b, &jm); err != nil {
		fmt.Println(err)
		return
	}
	syncKeyList, _ :=  GetSyncKeyList(jm)
	session.SyncKeyList = syncKeyList
	user := GetUserInfo(jm["User"].(map[string]interface{}))
	session.Bot = user
	fmt.Println(syncKeyList, "[syncKeyList]")
	fmt.Println(user, "[user]")
	notifyRet, _ := session.Api.WebWxStatusNotify(session.XmlResponse, session.Bot)
  fmt.Println(notifyRet, "[response ret]")
  if notifyRet !=0 {
    fmt.Println("WebWxStatusNotify fail, BaseResponse.Ret != 0")
    return
  }
}

func (session *Session) WebWxGetContact() {
	b, _ := session.Api.WebWxGetContact(session.XmlResponse)
	var jm map[string]interface{}
	if err := json.Unmarshal(b, &jm); err != nil {
		fmt.Println(err)
		return
	}
	session.MemberList = GetMemberList(jm)
}

func (session *Session) SyncCheck() {
	for {
		retcode, selector, _ := session.Api.SyncCheck(session.XmlResponse, session.SyncKeyList)
		fmt.Println(retcode, selector, "[synccheck]")
		switch retcode {
			case 0:
				if selector !=0 {
					b, _ := session.Api.SyncMessage(session.XmlResponse, session.SyncKeyList)
					var jm map[string]interface{}
					if err := json.Unmarshal(b, &jm); err != nil {
						fmt.Println(err)
						return
					}
					syncKeyList, _ :=  GetSyncKeyList(jm)
					session.SyncKeyList = syncKeyList
					fmt.Println(session.SyncKeyList)
					msgList := jm["AddMsgList"].([]interface{})
					for _, msg := range msgList {
						msg := msg.(map[string]interface{})
						msgType := int(msg["MsgType"].(float64))
						fmt.Println(msgType, "[msgType]")
						switch msgType {
							case 1: //text message
								textMsg := GetTextMessage(msg)
								fromUserName := textMsg.FromUserName
								toUserName := textMsg.ToUserName
								from := GetNickName(session.MemberList, fromUserName)
								to := GetNickName(session.MemberList, toUserName)
								content := textMsg.Content
								fmt.Printf(fmt.Sprintf("from: %s, to: %s, content: %s\n", from, to, content))
						}
					}
				}
			case 1102,1101:
				fmt.Println("1101 1102 logout")
				return
			default:
				fmt.Printf("retcode: %v\n", retcode)
		}
	}
}

