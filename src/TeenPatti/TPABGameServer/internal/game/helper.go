package game

import (
	"runtime"
	"strings"
	"github.com/lonng/nano/session"
	"TeenPatti/TPABGameServer/pkg/errutil"
	"TeenPatti/TPABGameServer/protocol"
)

const (

)

func verifyOptions(opts *protocol.DeskOptions) bool {


	if opts == nil {
		return false
	}

/*
	if opts.MaxRound != 1 && opts.MaxRound != 4 && opts.MaxRound != 8 && opts.MaxRound != 16 {
		return false
	}
*/


	return true
}

func requireCardCount(round int) int {

	return 0
}

func playerWithSession(s *session.Session) (*Player, error) {

	p, ok := s.Value(kCurPlayer).(*Player)
	if !ok {
		return nil, errutil.ErrPlayerNotFound
	}
	return p, nil
}


func stack() string {
	buf := make([]byte, 10000)
	n := runtime.Stack(buf, false)
	buf = buf[:n]

	s := string(buf)

	// skip nano frames lines
	const skip = 7
	count := 0
	index := strings.IndexFunc(s, func(c rune) bool {
		if c != '\n' {
			return false
		}
		count++
		return count == skip
	})
	return s[index+1:]
}
