package torrent
import (
	"strings"
	"strconv"
	"crypto/sha1"
	"io"
	"fmt"
	"net/http"
	"errors"
	"time"
	"net"
	"io/ioutil"
	"github.com/zeebo/bencode"
)

// BitTorrent
type BitTorrent struct {
	Info_hash    string
	DownloadLink string
}

// new BitTorrent
func New(info_hash string) *BitTorrent {
	return &BitTorrent{
		Info_hash: info_hash,
	}
}

// get Bitcomet key with info_hash
func GetBitCometKey(info_hash string) string {
	var hashHex []byte

	hash := strings.ToLower(info_hash)
	halfLen := len(hash) / 2

	for i := 0; i < halfLen; i++ {
		val, _ := strconv.ParseInt(hash[i * 2:i * 2 + 2], 16, 0)
		hashHex = append(hashHex, byte(val))
	}

	bc := "bc" + string(hashHex) + "torrent"

	t := sha1.New()
	io.WriteString(t, bc)
	return fmt.Sprintf("%x", t.Sum(nil))
}

// 组合下载链接
func (b *BitTorrent) getDownloadLink() []string {
	var result []string

	uhash := strings.ToUpper(b.Info_hash)

	// use s2p.co
	link := fmt.Sprintf("http://s2p.co/get/%s.torrent", uhash)
	result = append(result, link)

	// use btdig
	link = fmt.Sprintf("http://btdig.com/%s.torrent", uhash)
	result = append(result, link)

	// use bt.box.n0808.com thunder torrent cache.
	link = fmt.Sprintf("http://bt.box.n0808.com/%s/%s/%s.torrent",
		uhash[0:2],
		uhash[len(uhash) - 2:],
		uhash,
	)
	result = append(result, link)


	return result
}

// get torrent meta info.
func (b *BitTorrent) GetTorrentMetaInfo() (*MetaInfo, error) {
	downloadLinks := b.getDownloadLink()

	for _, downloadLink := range downloadLinks {
		// new http request.
		req, err := http.NewRequest("GET", downloadLink, nil)
		if err != nil {
			continue
		}

		// set http header.
		req.Header.Add("User-Agent", "Mozilla/5.0")
		req.Header.Add("Accept", "*/*")
		req.Header.Add("Connection", "Keep-Alive")

		// set dead line time.
		client := &http.Client{
			Transport: &http.Transport{
				Dial: func(netw, addr string) (net.Conn, error) {
					// Read/write dead line.
					deadLine := time.Now().Add(time.Second * 20)
					// dial timeout.
					c, err := net.DialTimeout(netw, addr, time.Second * 3)
					if err != nil {
						return nil, err
					}

					c.SetDeadline(deadLine)
					return c, nil
				},
			},
		}

		// request
		response, err := client.Do(req)
		if err != nil {
			continue
		}
		defer response.Body.Close()

		// read torrent information
		torrentMeta, err := b.ReadTorrentInformation(response.Body)
		if err != nil {
			continue
		}


		b.DownloadLink = downloadLink

		return torrentMeta, nil
	}

	return nil, errors.New("Can not get torrent meta info!")
}


// Read torrent information
func (b *BitTorrent)ReadTorrentInformation(r io.Reader) (*MetaInfo, error) {
	// read file data.
	dat, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// decode dat to meta_info.
	m := &MetaInfo{}
	if err = bencode.DecodeBytes(dat, m); err != nil {
		return nil, err
	}

	m.InfoHash = b.Info_hash

	return m, nil
}