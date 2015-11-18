package main

import (
	"github.com/gitwillsky/btsearcher/dht"
	"github.com/gitwillsky/btsearcher/torrent"
	"github.com/gitwillsky/btsearcher/models"
	"time"
	"strings"
	"errors"
	"github.com/gitwillsky/btsearcher/filter"
	"github.com/astaxie/beego"
)

// insert metainfo to database.
func Insert(infoHash string, ip string) error {
	t := &models.Torrent{
		LastDownloadIp: ip,
		LastDownloadAt: time.Now(),
		InfoHash: strings.ToLower(infoHash),
	}

	if t.IsExists() {
		return errors.New("Torrent already in database;")
	}

	// get torrent file information.
	torrentFile := torrent.New(infoHash)
	m, err := torrentFile.GetTorrentMetaInfo()
	if err != nil {
		return err
	}

	t.Name = m.Info.Name
	t.Enable = true
	if len(t.Name) < 2 || len([]byte(t.Name)) < 2 {
		return errors.New("Torrent info name length lt 2!")
	} else if filter.IsIllegalWords(t.Name) {
		// 非法资源
		t.Enable = false
	}

	t.FileLen = m.Info.Length
	t.Md5 = m.Info.Md5sum
	t.Magnet = "magnet:?xt=urn:btih:" + t.InfoHash
	t.CreateDate = time.Unix(m.CreateDate, 0)

	// file list
	for _, v := range m.Info.Files {
		// 非法资源
		if filter.IsIllegalWords(v.Path[0]) {
			t.Enable = false
		}

		file := &models.File{
			Name:v.Path[0],
			Length:v.Length,
			Md5:v.Md5sum,
		}

		t.FileList = append(t.FileList, file)
	}

	return t.Create()
}

func excute(output chan string, port int) {
	myDHTNode := dht.NewDHTNode(output, port)
	myDHTNode.Run()
}

func main() {
	output := make(chan string)

	portA, _ := beego.AppConfig.Int("UDPPort")


	excute(output, portA)

	for {
		select {
		case dat := <-output:
			s := strings.Split(dat, ":")

			infohash := s[0]
			ip := s[1]

			go func(info string, ip string) {
				Insert(info, ip)
				/*if err := Insert(info, ip); err != nil {
					fmt.Printf("%s insert failed: %s \n", info, err.Error())
				}else {
					fmt.Printf("%s insert success! \n", info)
				}
*/
			}(infohash, ip)
		}
	}
}
