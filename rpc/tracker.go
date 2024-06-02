package rpc

import (
	"fmt"
	"github.com/RyanChen-cn/torrent/bencode"
	"github.com/RyanChen-cn/torrent/entity"
	"io/ioutil"
	"net/http"
	"net/url"
)

// TrackerRequest represents the fields sent in a request to the tracker.
type TrackerRequest struct {
	InfoHash   [20]byte `url:"info_hash"`
	PeerID     string   `url:"peer_id"`
	IP         string   `url:"ip,omitempty"`
	Port       int      `url:"port"`
	Uploaded   int      `url:"uploaded"`
	Downloaded int      `url:"downloaded"`
	Left       int      `url:"left"`
	Event      string   `url:"event,omitempty"`
	Compact    int      `url:"compact,omitempty"`
	NoPeerID   int      `url:"no_peer_id,omitempty"`
	NumWant    int      `url:"numwant,omitempty"`
	Key        string   `url:"key,omitempty"`
	TrackerID  string   `url:"trackerid,omitempty"`
}

// TrackerResponse represents the fields received in a response from the tracker.
type TrackerResponse struct {
	FailureReason  string      `bencode:"failure reason,omitempty"`
	WarningMessage string      `bencode:"warning message,omitempty"`
	Interval       int         `bencode:"interval"`
	MinInterval    int         `bencode:"min interval,omitempty"`
	TrackerID      string      `bencode:"tracker id,omitempty"`
	Complete       int         `bencode:"complete"`
	Incomplete     int         `bencode:"incomplete"`
	Peers          interface{} `bencode:"peers"` // Can be string (compact format) or []Peer
}

// Peer represents an individual peer in the tracker response.
type Peer struct {
	PeerID string `bencode:"peer id"`
	IP     string `bencode:"ip"`
	Port   int    `bencode:"port"`
}

// 构建 TrackerRequest 从 TorrentFile
func buildTrackerRequest(file *entity.TorrentFile, peerID string, port int, uploaded, downloaded, left int, event string) (*TrackerRequest, error) {
	infoHash, err := entity.GenerateInfoHash(file.Info)
	if err != nil {
		return nil, err
	}
	return &TrackerRequest{
		InfoHash:   infoHash,
		PeerID:     peerID,
		Port:       port,
		Uploaded:   uploaded,
		Downloaded: downloaded,
		Left:       file.Info.Length,
		Event:      event,
		Compact:    1,  // 一般请求 compact peer 列表
		NumWant:    50, // 一般请求 50 个 peers
	}, nil
}

// 发送 HTTP GET 请求到 tracker 服务器，并解析返回的 TrackerResponse
func GetPeers(torrent entity.TorrentFile, req *TrackerRequest) (*TrackerResponse, error) {
	trackerURLs := append([]string{torrent.Announce}, flattenAnnounceList(torrent.AnnounceList)...)

	for _, trackerURL := range trackerURLs {
		// 构建请求 URL
		reqURL, err := url.Parse(trackerURL)
		if err != nil {
			continue // 无法解析 URL，跳过
		}

		// 添加请求参数
		query := reqURL.Query()
		query.Set("info_hash", string(req.InfoHash[:]))
		query.Set("peer_id", req.PeerID)
		query.Set("port", fmt.Sprintf("%d", req.Port))
		query.Set("uploaded", fmt.Sprintf("%d", req.Uploaded))
		query.Set("downloaded", fmt.Sprintf("%d", req.Downloaded))
		query.Set("left", fmt.Sprintf("%d", req.Left))
		if req.Event != "" {
			query.Set("event", req.Event)
		}
		if req.Compact > 0 {
			query.Set("compact", fmt.Sprintf("%d", req.Compact))
		}
		if req.NoPeerID > 0 {
			query.Set("no_peer_id", fmt.Sprintf("%d", req.NoPeerID))
		}
		if req.NumWant > 0 {
			query.Set("numwant", fmt.Sprintf("%d", req.NumWant))
		}
		if req.Key != "" {
			query.Set("key", req.Key)
		}
		if req.TrackerID != "" {
			query.Set("trackerid", req.TrackerID)
		}
		reqURL.RawQuery = query.Encode()

		// 发送请求
		resp, err := http.Get(reqURL.String())
		if err != nil {
			continue // 请求失败，跳过
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue // 非200响应，跳过
		}

		// 读取响应数据
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue // 读取响应失败，跳过
		}

		// 解析 TrackerResponse
		var trackerResponse TrackerResponse
		serializer := bencode.NewBencodeSerializer()
		err = serializer.Unmarshal(body, &trackerResponse)
		if err != nil {
			continue // 解析失败，跳过
		}

		return &trackerResponse, nil // 成功解析
	}

	return nil, fmt.Errorf("unable to get peers from tracker")
}

// 辅助函数：展开 announce-list
func flattenAnnounceList(announceList [][]string) []string {
	var result []string
	for _, tier := range announceList {
		result = append(result, tier...)
	}
	return result
}
