package main

import (
	"btsearcher"
	"btsearcher/filter"
	"btsearcher/models"
	"bytes"
	"fmt"
	"github.com/marksamman/bencode"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/etcd/pkg/fileutil"
	"gopkg.in/alecthomas/kingpin.v2"
	"net"
	"os"
	"path"
	"strings"
	"time"
)

var (
	verbose         = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	addr            = kingpin.Flag("addr", "listen on given address (default all, ipv4 and ipv6)").Default("0.0.0.0").Short('a').IP()
	port            = kingpin.Flag("port", "listen udp port").Short('p').Default("6800").Uint16()
	maxFriends      = kingpin.Flag("max-friends", "max fiends to make with per second").Default("500").Uint()
	maxPeers        = kingpin.Flag("max-peers", "max peers to connect to download torrents").Default("500").Uint()
	downloadTimeout = kingpin.Flag("dl.timeout", "max time allowed for downloading torrents").Default("10s").Duration()
	downloadDir     = kingpin.Flag("dl.dir", "the directory to store the torrents").Default("torrents/").String()

	dbUser     = kingpin.Flag("db.user", "database user").Required().String()
	dbPassword = kingpin.Flag("db.password", "database password").Required().String()
	dbName     = kingpin.Flag("db.name", "database name").Required().String()
	dbAddr     = kingpin.Flag("db.addr", "database ip and port").Required().String()

	blackList = btsearcher.NewBlackList(5*time.Minute, 50000)
)

func run() error {
	tokens := make(chan struct{}, *maxPeers)

	dht, err := btsearcher.NewDHT(net.JoinHostPort(addr.String(), fmt.Sprint(*port)), int(*maxFriends))
	if err != nil {
		return err
	}

	log.Info("dht network start running...")
	dht.Run()

	for {
		select {
		case <-dht.Announcements().Wait():
			for {
				if ac := dht.Announcements().Get(); ac != nil {
					tokens <- struct{}{}
					go work(ac, tokens)
					continue
				}
				break
			}
		case <-dht.Die:
			return dht.ErrDie
		}
	}
}

func work(ac *btsearcher.Announcement, tokens chan struct{}) {
	defer func() {
		<-tokens
	}()

	t := &models.Torrent{
		LastDownloadIp: ac.From.IP.String(),
		LastDownloadAt: time.Now(),
		InfoHash:       strings.ToLower(ac.InfohashHex),
	}

	if t.IsExists() {
		return
	}

	peerAddr := ac.Peer.String()
	if blackList.Has(peerAddr) {
		return
	}

	wire := btsearcher.NewMetaWire(string(ac.Infohash), peerAddr, *downloadTimeout)
	defer wire.Free()

	meta, err := wire.Fetch()
	if err != nil {
		log.Infof("connect peer %s failed, add to blacklist", peerAddr)
		blackList.Add(peerAddr)
		return
	}

	go saveTorrentInfoDB(ac, meta)
	saveTorrentFile(ac.InfohashHex, meta)
}

func saveTorrentInfoDB(ac *btsearcher.Announcement, meta []byte) {
	torrent, err := btsearcher.ParseTorrent(ac.InfohashHex, meta)
	if err != nil {
		log.Errorf("parse torrent file failed, %s", err.Error())
		return
	}

	t := &models.Torrent{
		LastDownloadIp: ac.From.IP.String(),
		LastDownloadAt: time.Now(),
		InfoHash:       strings.ToLower(ac.InfohashHex),
	}

	t.Name = torrent.Name
	t.Enable = true
	if len(t.Name) < 2 || len([]byte(t.Name)) < 2 {
		log.Warnf("Torrent name length lt 2: %s", t.Name)
		return
	} else if filter.IsIllegalWords(t.Name) {
		// 非法资源
		t.Enable = false
	}

	t.FileLen = uint64(torrent.Length)
	t.Md5 = torrent.Md5
	t.Magnet = "magnet:?xt=urn:btih:" + torrent.InfohashHex
	t.CreateDate = time.Unix(torrent.CreateDate, 0)

	// file list
	for _, v := range torrent.Files {
		// 非法资源
		if filter.IsIllegalWords(v.Name) {
			t.Enable = false
		}

		file := &models.File{
			Name:   v.Name,
			Length: uint64(v.Length),
			Md5:    v.Md5,
		}

		t.FileList = append(t.FileList, file)
	}

	if err := t.Create(); err != nil {
		log.Errorf("save torrent information to db failed, %s", err.Error())
	}
}

func saveTorrentFile(infoHashHex string, meta []byte) {
	dir := path.Join(*downloadDir, time.Now().Format("2006-01-02"))
	filePath := path.Join(dir, infoHashHex+".torrent")

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		log.Infof("save torrent file canceled, file existed")
		return
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Errorf("create torrent dir failed, %s", err.Error())
		return
	}

	d, err := bencode.Decode(bytes.NewBuffer(meta))
	if err != nil {
		log.Errorf("decode torrent data failed, %s", err.Error())
		return
	}

	file, err := fileutil.TryLockFile(filePath, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Errorf("open torrent file failed, %s", err.Error())
		return
	}
	defer file.Close()

	if _, err := file.Write(bencode.Encode(map[string]interface{}{
		"info": d,
	})); err != nil {
		log.Errorf("write torrent file failed, %s", err.Error())
		return
	}
}

func main() {
	kingpin.Parse()

	if err := models.InitDBC(*dbUser, *dbPassword, *dbName, *dbAddr); err != nil {
		log.Fatalf("initial database failed, %s", err.Error())
	}

	if err := run(); err != nil {
		log.Fatalf("run failed, %s", err.Error())
	}
}
