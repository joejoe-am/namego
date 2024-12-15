package rpc

import "github.com/joejoe-am/namego/configs"

const (
	RpcQueueTemplate                       = "rpc-%s"
	RpcReplyQueueTemplate                  = "rpc.reply-%s-%s"
	RpcReplyQueueTtl                       = 300000 // ms (5 min)
	EventHandlerBroadCaseQueueTemplate     = "evt-%s-%s--%s.%s-%s"
	EventHandlerSingletonCaseQueueTemplate = "evt-%s-%s"
	EventHandlerServicePoolQueueTemplate   = "evt-%s-%s--%s.%s"
)

var Cfg *configs.Configs

func init() {
	Cfg = configs.GetConfigs()
}
