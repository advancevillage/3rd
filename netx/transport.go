package netx

import "context"

type IMQ interface {
}

type mq struct {
	msgType MessageType
	msgId   uint16
	mCli    IMessage

	wc chan *message
	rc chan *message
}

func (m *mq) inqueue(ctx context.Context) {

}
