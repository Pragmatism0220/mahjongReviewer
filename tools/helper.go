package tools

import (
	"github.com/Pragmatism0220/majsoul/message"
	"github.com/golang/protobuf/proto"
)

// 定义了配置文件的结构体，图省事写在helper这个文件下了。
type Config struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	LoginUUID    string `json:"loginUUID"`
	ReviewerPath string `json:"reviewerPath"`
}

func UnwrapData(rawData []byte) (methodName string, data []byte, err error) {
	wrapper := message.Wrapper{}
	if err = proto.Unmarshal(rawData, &wrapper); err != nil {
		return
	}
	return wrapper.GetName(), wrapper.GetData(), nil
}

// TODO: auto UnwrapMessage by methodName

func UnwrapMessage(rawData []byte, message proto.Message) error {
	methodName, data, err := UnwrapData(rawData)
	if err != nil {
		return err
	}
	// TODO: assert methodName when its not empty
	_ = methodName
	return proto.Unmarshal(data, message)
}
