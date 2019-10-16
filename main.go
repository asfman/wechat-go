package main

import (
  "fmt"

  "github.com/asfman/wechat-go/wxweb"
)

func main() {
  session, err := wxweb.CreateSession()
  if err != nil {
    fmt.Println(err.Error())
    return
  }
  session.Login()
  session.WebNewLoginPage()
	session.WebWxInit()
	session.WebWxGetContact()
  session.SyncCheck()
}
