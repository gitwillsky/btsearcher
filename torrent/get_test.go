package torrent
import (
	"testing"
	"encoding/json"
	"time"
)


func Test_GetTorrentMetaInfo(t *testing.T) {
	b := &BitTorrent{
		Info_hash:"227ccb6ae4efdc440165c08c69b701bb327a1853",
	}

	m, err := b.GetTorrentMetaInfo()
	if err != nil {
		t.Error(err.Error())
	}
	dat, _ := json.MarshalIndent(m, "", "  ")
	t.Logf("download link =%s \n", b.DownloadLink)
	t.Logf("%v", string(dat))
	t.Log("create date " + time.Unix(m.CreateDate, 0).Format("2006-01-02"))
}