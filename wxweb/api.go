package wxweb

import (
	"fmt"
  "regexp"
  "time"
  "strconv"
	"net/url"
	"net/http"
  "strings"
	"encoding/xml"
	"encoding/json"
	"io/ioutil"

  "github.com/spf13/viper"
)

type Api struct {
	client *Client
}

func initConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./wxweb/config")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found!", err)
		} else {
			fmt.Println(err)
		}
	}
}

func init() {
	initConfig()
}

func NewApi() *Api {
	userAgent := viper.GetString("web.user_agent")
  return &Api{
    client: NewClient(userAgent),
  }
}

// 第一步获取qrcode uuid
func (api *Api) GetUUID() (string, error) {
	km := url.Values{}
	km.Add("appid", viper.GetString("web.appid"))
	km.Add("fun", "new")
	km.Add("lang", "en_US")
	km.Add("_", strconv.FormatInt(time.Now().Unix(), 10))
  url := fmt.Sprintf("%s/jslogin?%s", viper.GetString("web.login_url"), km.Encode())
  fmt.Printf("url: %s\n", url)

	body, err := api.client.Get(url)
	if err != nil {
		return "", err
  }
  reg := regexp.MustCompile(`uuid\s*=\s*"([\w=-]+)"`)
  sub := reg.FindStringSubmatch(string(body))
	if len(sub) < 2 {
    return "", fmt.Errorf("jslogin response invalid")
	}
  fmt.Println(sub, "sub")
	return sub[1], nil
}

// 第二步获取redirectUri
func (api *Api) Login(uuid string) (string, error) {
	km := url.Values{}
	km.Add("tip", "0")
	km.Add("uuid", uuid)
	km.Add("r", strconv.FormatInt(time.Now().Unix(), 10))
	km.Add("_", strconv.FormatInt(time.Now().Unix(), 10))
	km.Add("loginicon", "true")
	url := fmt.Sprintf("%s/cgi-bin/mmwebwx-bin/login?%s", viper.GetString("web.login_url"), km.Encode())
	body, _ := api.client.Get(url)
	strb := string(body)
	if strings.Contains(strb, "window.code=200") &&
		strings.Contains(strb, "window.redirect_uri") {
		ss := strings.Split(strb, "\"")
		if len(ss) < 2 {
			return "", fmt.Errorf("parse redirect_uri fail, %s", strb)
		}
		return ss[1], nil
	}

	return "", fmt.Errorf("login response, %s", strb)
}

// 第三步根据redirectUrl返回xmlResponse, 包括pass_ticket等
// redirect_uri: https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxnewloginpage?ticket=AfJ6WDfJpst_nJRZIRcMXBhK@qrticket_0&uuid=4ejB1M5AkQ==&lang=en_US&scan=1569643013
func (api *Api) WebNewLoginPage(redirectUri string, xmlResponse *XmlResponse) ([]*http.Cookie, error) {
	km := url.Values{}
	km.Add("fun", "new")
	url := fmt.Sprintf("%s&%s", redirectUri, km.Encode())
	res, _ := api.client.fetchResponse("GET", url, nil, Header{})
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	fmt.Println(string(body), "[WebNewLoginPage]")
	if err := xml.Unmarshal(body, xmlResponse); err != nil {
		fmt.Println(fmt.Errorf("xml unmarshal err: %v", err))
		return nil, err
	}
	if xmlResponse.Ret != 0 {
		err := fmt.Errorf("xc.Ret != 0, %s", string(body))
		fmt.Println(err)
		return nil, err
	}
	return res.Cookies(), nil
}

// 第四部根据xmlResponse，返回用户信息
func (api *Api) WebWxInit(xmlResponse *XmlResponse) ([]byte, error) {
	km := url.Values{}
	km.Add("pass_ticket", xmlResponse.PassTicket)
	km.Add("skey", xmlResponse.Skey)
	km.Add("r", strconv.FormatInt(time.Now().Unix(), 10))
	url := viper.GetString("web.base_url") + "/cgi-bin/mmwebwx-bin/webwxinit?" + km.Encode()
	baseRequest := &BaseRequest{
		xmlResponse.Wxuin,
		xmlResponse.Wxsid,
		xmlResponse.Skey,
		"e" + GetRandomStringFromNum(15),
	}
	bytes, _ := json.Marshal(InitRequestBody{baseRequest})
	body, err := api.client.PostJsonBytes(url, bytes)
	// fmt.Println(string(body), "[WebWxInit]")
	return body, err
}

// 报告自己的状态BaseResponse.Ret为0表示成功
func (api *Api) WebWxStatusNotify(xmlResponse *XmlResponse, bot *User) (int, error) {
	km := url.Values{}
	km.Add("pass_ticket", xmlResponse.PassTicket)
	km.Add("lang", "en_US")
	url := viper.GetString("web.base_url") + "/cgi-bin/mmwebwx-bin/webwxstatusnotify?" + km.Encode()
	baseRequest := &BaseRequest{
		xmlResponse.Wxuin,
		xmlResponse.Wxsid,
		xmlResponse.Skey,
		"e" + GetRandomStringFromNum(15),
	}
	reqBody := NotifyRequestBody{
		BaseRequest: baseRequest,
		Code:         3,
		FromUserName: bot.UserName,
		ToUserName:   bot.UserName,
		ClientMsgId:  int(time.Now().Unix()),
	}

	bytes, _ := json.Marshal(reqBody)

	body, err := api.client.PostJsonBytes(url, bytes)
  reg := regexp.MustCompile(`"Ret":\s*(\d+)`)
  sub := reg.FindStringSubmatch(string(body))
  ret, _ := strconv.Atoi(sub[1])
	return ret, err
}

// 获取所有联系人, 测试发现url不加query也能正常返回，可能后台只验证Cookie
func (api *Api) WebWxGetContact(xmlResponse *XmlResponse) ([]byte, error) {
	km := url.Values{}
	km.Add("pass_ticket", xmlResponse.PassTicket)
	km.Add("lang", "en_US")
	url := viper.GetString("web.base_url") + "/cgi-bin/mmwebwx-bin/webwxgetcontact?" + km.Encode()
	baseRequest := &BaseRequest{
		xmlResponse.Wxuin,
		xmlResponse.Wxsid,
		xmlResponse.Skey,
		"e" + GetRandomStringFromNum(15),
	}
	bytes, _ := json.Marshal(InitRequestBody{baseRequest})
	body, err := api.client.PostJsonBytes(url, bytes)
	return body, err
}

func (api *Api) SyncCheck(xmlResponse *XmlResponse, skl *SyncKeyList) (retcode int, selector int, err error) {
	km := url.Values{}
	km.Add("r", strconv.FormatInt(time.Now().Unix()*1000, 10))
	km.Add("sid", xmlResponse.Wxsid)
	km.Add("uin", xmlResponse.Wxuin)
	km.Add("skey", xmlResponse.Skey)
	km.Add("deviceid", "e" + GetRandomStringFromNum(15))
	km.Add("synckey", skl.String())
	km.Add("_", strconv.FormatInt(time.Now().Unix()*1000, 10))
	url := viper.GetString("web.base_push_url") + "/cgi-bin/mmwebwx-bin/synccheck?" + km.Encode()
	body, _:= api.client.Get(url)
	strb := string(body)
	// 正常返回结果
	// window.synccheck={retcode:"0",selector:"0"}
	// 有消息返回结果
	// window.synccheck={retcode:"0",selector:"6"}
	// 发送消息返回结果
	// window.synccheck={retcode:"0",selector:"2"}
	// 朋友圈有动态
	// window.synccheck={retcode:"0",selector:"4"}
	reg := regexp.MustCompile("window.synccheck={retcode:\"(\\d+)\",selector:\"(\\d+)\"}")
	sub := reg.FindStringSubmatch(strb)
	retcode = 0
	selector = 0
	if len(sub) >= 2 {
		retcode, _ = strconv.Atoi(sub[1])
		selector, _ = strconv.Atoi(sub[2])
	}
	return retcode, selector, nil
}

func (api *Api) SyncMessage(xmlResponse *XmlResponse, skl *SyncKeyList) ([]byte, error) {
	km := url.Values{}
	km.Add("skey", xmlResponse.Skey)
	km.Add("sid", xmlResponse.Wxsid)
	km.Add("lang", "en_US")
	km.Add("pass_ticket", xmlResponse.PassTicket)
	url := viper.GetString("web.base_url") + "/cgi-bin/mmwebwx-bin/webwxsync?" + km.Encode()
	js := SyncRequestBody{
		BaseRequest: &BaseRequest{
			xmlResponse.Wxuin,
			xmlResponse.Wxsid,
			xmlResponse.Skey,
			"e" + GetRandomStringFromNum(15),
		},
		SyncKey: skl,
		rr:      ^int(time.Now().Unix()) + 1,
	}

	bytes, _ := json.Marshal(js)
	return api.client.PostJsonBytes(url, bytes)
}

