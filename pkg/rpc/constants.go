package rpc

const (
	RpcQueueTemplate      = "rpc-%s"
	RpcReplyQueueTemplate = "rpc.reply-%s-%s"
	RpcReplyQueueTtl      = 300000 // ms (5 min)
)
