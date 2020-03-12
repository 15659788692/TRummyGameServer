package gate

import (

	"github.com/lonng/nano/component"
	"github.com/lonng/nano/examples/cluster/protocol"
	"github.com/lonng/nano/internal/log"
	"github.com/lonng/nano/session"
	"github.com/pingcap/errors"
)

type BindService struct {
	component.Base
	nextGateUid int64
}

func newBindService() *BindService {
	return &BindService{}
}

type (
	LoginRequest struct {
		Nickname string `json:"nickname"`
	}
	LoginResponse struct {
		Code int `json:"code"`
	}
)

func (bs *BindService) Login(s *session.Session, msg *LoginRequest) error {
	bs.nextGateUid++
	uid := bs.nextGateUid

	log.Println("Login Session:", s.RemoteAddr() )

	request := &protocol.NewUserRequest{
		Nickname: msg.Nickname,
		GateUid:  uid,
	}


log.Println( "TopicSerice.NewUser :" ,request.Nickname  )


	if err := s.RPC("TopicService.NewUser", request); err != nil {
		return errors.Trace(err)
	}

	//log.Println( "TopicSerice.NewUser : ****" ,request.Nickname  )


	return s.Response(&LoginResponse{})
}

func (bs *BindService) BindChatServer(s *session.Session, msg []byte) error {
	return errors.Errorf("not implement")
}
