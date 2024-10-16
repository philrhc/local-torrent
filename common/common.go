package common

import (
	"log/slog"
	"os"

	"github.com/anacrolix/envpprof"
)

func SetupTmpFolder() string {
	defer envpprof.Stop()
	tmpDir, err := os.MkdirTemp("", "torrent-zyre")
	AssertNil(err)
	slog.Info("made temp dir", slog.String("tmpDir", tmpDir))
	return tmpDir
}

func AssertNil(x any) {
	if x != nil {
		panic(x)
	}
}
