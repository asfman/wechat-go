package wxweb

import (
	"encoding/xml"
	"strconv"
	"strings"
)

type XmlResponse struct {
	XMLName     xml.Name `xml:"error"`
	Ret         int      `xml:"ret"`
	Message     string   `xml:"message"`
	Skey        string   `xml:"skey"`
	Wxsid       string   `xml:"wxsid"`
	Wxuin       string   `xml:"wxuin"`
	PassTicket  string   `xml:"pass_ticket"`
	IsGrayscale int      `xml:"isgrayscale"`
}

type BaseRequest struct {
	Uin      string
	Sid      string
	Skey     string
	DeviceID string
}

type SyncKey struct {
	Key int
	Val int
}

type SyncKeyList struct {
	Count int
	List  []SyncKey
}

func (s *SyncKeyList) String() string {
	strs := make([]string, 0)
	for _, v := range s.List {
		strs = append(strs, strconv.Itoa(v.Key)+"_"+strconv.Itoa(v.Val))
	}
	return strings.Join(strs, "|")
}

type User struct {
	Uin      int
	UserName string
	NickName string
}

type InitRequestBody struct {
	BaseRequest *BaseRequest
}

type NotifyRequestBody struct {
	BaseRequest        *BaseRequest
	FromUserName       string
	ToUserName         string
	ClientMsgId        int
	Code							 int
}

type SyncRequestBody struct {
	BaseRequest *BaseRequest
  SyncKey     *SyncKeyList
  rr          int
}

type TextMessage struct {
	MsgType      int
	Content      string
	FromUserName string
	ToUserName   string
	ClientMsgId  int
}

