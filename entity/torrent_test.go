package entity

import (
	"log"
	"testing"
)

func TestUnmarshalTorrentFile(t *testing.T) {
	// 读取 .torrent 文件
	file, err := UnmarshalTorrentFile("../testdata/ubuntu-24.04-live-server-amd64.iso.torrent")
	if err != nil {
		log.Fatalf("failed to unmarshal torrent file: %v", err)
	}

	t.Logf("Torrent File: %+v\n", file)
}
