package logging

import (
	"log/slog"
	"os"
)

var KubeLogger *slog.Logger
func SetLogger(isDebug bool){
	if isDebug {
		KubeLogger = slog.New(slog.NewTextHandler(os.Stdout,nil))
	} else {
		KubeLogger = slog.New(slog.DiscardHandler)
	}

}