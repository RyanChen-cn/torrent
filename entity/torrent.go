package entity

import (
	"crypto/sha1"
	"github.com/RyanChen-cn/torrent/bencode"
	"os"
	"time"
)

// TorrentFile and related structs
type TorrentFile struct {
	Announce     string     `bencode:"announce"`
	AnnounceList [][]string `bencode:"announce-list"`
	CreationDate time.Time  `bencode:"creation date,omitempty"`
	Comment      string     `bencode:"comment,omitempty"`
	CreatedBy    string     `bencode:"created by,omitempty"`
	Encoding     string     `bencode:"encoding,omitempty"`
	Info         InfoDict   `bencode:"info"`
}

type InfoDict struct {
	//Files       []FileDict `bencode:"files,omitempty"`
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type FileDict struct {
	Length int      `bencode:"length"`
	Path   []string `bencode:"path"`
}

// Special unmarshal function for TorrentFile
func UnmarshalTorrentFile(name string) (TorrentFile, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return TorrentFile{}, err
	}
	var torrentFile TorrentFile
	serializer := bencode.NewBencodeSerializer()
	err = serializer.Unmarshal(data, &torrentFile)
	if err != nil {
		return TorrentFile{}, err
	}
	return torrentFile, nil
}

func GenerateInfoHash(info InfoDict) ([20]byte, error) {
	serializer := bencode.NewBencodeSerializer()
	encodedInfo, err := serializer.Marshal(info)
	if err != nil {
		return [20]byte{}, err
	}

	hash := sha1.Sum(encodedInfo)
	return hash, nil
}
