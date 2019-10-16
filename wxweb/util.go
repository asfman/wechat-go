package wxweb

import (
	"time"
	"math/rand"
	"reflect"
)

func GetRandomStringFromNum(length int) string {
	bytes := []byte("0123456789")
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func GetSyncKeyList(mp map[string]interface{}) (*SyncKeyList, error) {
	syncKey := mp["SyncKey"].(map[string]interface{})
	synks := make([]SyncKey, 0)
	list := syncKey["List"]
	is := list.([]interface{})
	for _, v := range is {
		vm := v.(map[string]interface{})
		sk := SyncKey{
			Key: int(vm["Key"].(float64)),
			Val: int(vm["Val"].(float64)),
		}
		synks = append(synks, sk)
	}
	return &SyncKeyList{
		Count: int(syncKey["Count"].(float64)),
		List:  synks,
	}, nil
}

func GetUserInfo(mp map[string]interface{}) *User {
	u := &User{}
	SetModel(u, mp)
	return u
}

func GetMemberList(mp map[string]interface{}) []*User {
	memberList := mp["MemberList"].([]interface{})
	retList := make([]*User, 0)
	for _, member := range memberList {
		user := member.(map[string]interface{})
		retList = append(retList, GetUserInfo(user))
	}
	return retList
}

func GetTextMessage(mp map[string]interface{}) *TextMessage {
	u := &TextMessage{}
	SetModel(u, mp)
	return u
}

func SetModel(u interface{}, mp map[string]interface{}) {
	t := reflect.TypeOf(u).Elem()
	fieldNum := t.NumField()
	fields := reflect.ValueOf(u).Elem()
	for i:= 0; i < fieldNum; i++ {
		fieldName := t.Field(i).Name
		v := mp[fieldName]
		if v == nil {
			continue
		}
		field := fields.FieldByName(fieldName)
		if vv, ok := v.(float64); ok {
			field.Set(reflect.ValueOf(int(vv)))
		} else {
			field.Set(reflect.ValueOf(v))
		}
	}
}

func GetNickName(memberList []*User, userName string) string {
	for _, member := range memberList {
		if member.UserName == userName {
			return member.NickName
		}
	}
	return userName
}

