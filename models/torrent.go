package models

import (
	"github.com/astaxie/beego/orm"
	"time"
	"errors"
)


type Torrent struct {
	Id             int64  `json:"id"`
	InfoHash       string `orm:"unique"  json:"infoHash"`
	Name           string `orm:"type(text)" json:"name"`
												   //	FileType       string    `orm:"size(30)"`
	LastDownloadAt time.Time `orm:"null;type(datetime)"  json:"lastDownDate"`
	LastDownloadIp string  `orm:"null;size(50)"`
	Hot            uint64    `json:"hot"`          // 下载热度
	FileLen        uint64    `json:"fileLength"`   // 文件大小
	FileCount      int       `json:"fileCount"`    // 文件数量
	Magnet         string    `json:"magnet"`       // 磁力链接
	Md5            string    `orm:"size(156);null"`
	Enable         bool      `json:"enable"`       // 审核标志
	FileList       []*File   `orm:"reverse(many)"` // multiple files.
	CreateDate     time.Time `orm:"type(datetime)"  json:"createDate"`
	CreateAt       time.Time `orm:"auto_now_add;type(datetime)" json:"-"`
}

type File struct {
	Id       int64    `json:"id"`
	Torrent  *Torrent `orm:"rel(fk);on_delete(do_nothing)"`
	Name     string  `orm:"type(text)"`
	Length   uint64
	Md5      string `orm:"size(256);null"`
	CreateAt time.Time `orm:"auto_now_add;type(datetime)" json:"-"`
}


// isHave
func (t *Torrent) IsExists() bool {
	o := orm.NewOrm()

	ip := t.LastDownloadIp
	downloadTime := t.LastDownloadAt

	if err := o.QueryTable("torrent").Filter("info_hash", t.InfoHash).One(t); err != nil {
		// 不存在
		return false
	}

	// 存在,更新最后下载IP 和时间和热度.
	go func(t *Torrent, o orm.Ormer) {
		if t.LastDownloadIp != ip {
			t.LastDownloadIp = ip
			t.LastDownloadAt = downloadTime
			t.Hot += 1
			o.Update(t, "last_download_ip", "last_download_at", "hot")
		}
	}(t, o)

	return true
}

// new torrent
func (t *Torrent) Create() error {
	o := orm.NewOrm()

	if t.IsExists() {
		return errors.New("Torrent already in database;")
	}

	// start transaction
	if err := o.Begin(); err != nil {
		return err
	}

	// get file length and file count.
	if t.FileList != nil {
		t.FileCount = len(t.FileList)
		t.FileLen = 0
		for _, file := range t.FileList {
			file.Torrent = t
			t.FileLen += file.Length
		}
	} else {
		t.FileCount = 1
	}

	// insert torrent.
	t.Hot = 1
	if _, err := o.Insert(t); err != nil {
		o.Rollback()
		return err
	}

	// insert torrent file.
	if t.FileList != nil {
		if _, err := o.InsertMulti(100, t.FileList); err != nil {
			o.Rollback()
			return err
		}
	}

	// commit transaction.
	if err := o.Commit(); err != nil {
		return err
	}
	return nil
}
