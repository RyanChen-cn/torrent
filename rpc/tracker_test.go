package rpc

import (
	"fmt"
	"log"
	"math/rand"
	"testing"
	"torrent/entity"
)

func TestGetPeers(t *testing.T) {
	// 模拟一个 TorrentFile
	file, err := entity.UnmarshalTorrentFile("../testdata/ubuntu-24.04-live-server-amd64.iso.torrent")
	if err != nil {
		log.Fatalf("failed to unmarshal torrent file: %v", err)
	}
	var peerID [20]byte
	_, err = rand.Read(peerID[:])
	if err != nil {
	}

	request, err := buildTrackerRequest(&file, string(peerID[:]), 6881, 0, 0, 100, "start")
	fmt.Println(err)
	// 构建一个 TrackerRequest

	// 测试 GetPeers 方法
	trackerResponse, err := GetPeers(file, request)
	fmt.Println(trackerResponse, err)
	//81af07491915415dad45f87c0c2ae52fae92c06b
	//compact=1&downloaded=0&event=start&info_hash=%C0%C5%D6.%04%DA%DE%16m%80G%96%EBN%F0o%BF%0AC%5C&left=0&numwant=50&peer_id=1233523412&port=6881&uploaded=0
	//compact=1&downloaded=0&event=start&info_hash=%C0%C5%D6.%04%DA%DE%16m%80G%96%EBN%F0o%BF%0AC%5C&left=0&numwant=50&peer_id=1233523412&port=6881&uploaded=0
	// 其他断言...
}
