//go:build !gui

package platform

import (
	"github.com/thang-qt/Readn/src/server"
)

func Start(s *server.Server) {
	s.Start()
}
