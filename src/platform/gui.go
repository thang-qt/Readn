//go:build (darwin || windows) && gui

package platform

import (
	"github.com/thang-qt/Readn/src/server"
	"github.com/thang-qt/Readn/src/systray"
)

func Start(s *server.Server) {
	systrayOnReady := func() {
		systray.SetIcon(Icon)
		systray.SetTooltip("yarr")

		menuOpen := systray.AddMenuItem("Open", "")
		systray.AddSeparator()
		menuQuit := systray.AddMenuItem("Quit", "")

		go func() {
			for {
				select {
				case <-menuOpen.ClickedCh:
					Open(s.GetAddr())
				case <-menuQuit.ClickedCh:
					systray.Quit()
				}
			}
		}()

		s.Start()
	}
	systray.Run(systrayOnReady, nil)
}
