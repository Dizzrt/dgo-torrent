package dlog_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/Dizzrt/dgo-torrent/dlog"
)

func TestLog(t *testing.T) {
	dlog.Init()
	dlog.L().Info(os.Getwd())
	dlog.L().Info(os.Args[0])
	dlog.L().Info(runtime.Caller(0))
	dlog.L().Sync()
}
