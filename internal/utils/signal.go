package utils

import (
	"os"
	"os/signal"
	"syscall"
)

type SignalExitFunc func()

func ProcessWaitForSignal(exitFunc SignalExitFunc) {
SIGNAL_WAIT:
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	for {
		select {
		case sigVal := <-signalCh:
			{
				switch sigVal {
				case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL:
					if exitFunc != nil {
						exitFunc()
					}
					close(signalCh)
					return
				case syscall.SIGHUP:
					close(signalCh)
					goto SIGNAL_WAIT
				default:
					close(signalCh)
					goto SIGNAL_WAIT
				}
			}
		}
	}
}
