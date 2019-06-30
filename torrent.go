package btsearcher

import (
	"bytes"
	"fmt"
	"github.com/marksamman/bencode"
	"strings"
)

type torrentFile struct {
	Name   string
	Length int64
	Md5    string
}

func (t *torrentFile) String() string {
	return fmt.Sprintf("name: %s\n, size: %d\n", t.Name, t.Length)
}

type torrent struct {
	InfohashHex string
	Name        string
	Length      int64
	CreateDate  int64
	Md5         string
	Files       []*torrentFile
}

func (t *torrent) String() string {
	return fmt.Sprintf(
		"link: %s\nname: %s\nsize: %d\nfile: %d\n",
		fmt.Sprintf("magnet:?xt=urn:btih:%s", t.InfohashHex),
		t.Name,
		t.Length,
		len(t.Files),
	)
}

func ParseTorrent(infohashHex string, meta []byte) (*torrent, error) {
	dict, err := bencode.Decode(bytes.NewBuffer(meta))
	if err != nil {
		return nil, err
	}

	t := &torrent{InfohashHex: infohashHex}
	if name, ok := dict["name.utf-8"].(string); ok {
		t.Name = name
	} else if name, ok := dict["name"].(string); ok {
		t.Name = name
	}
	if length, ok := dict["length"].(int64); ok {
		t.Length = length
	}
	if createDate, ok := dict["creation date"].(int64); ok {
		t.CreateDate = createDate
	}
	if md5Str, ok := dict["md5sum"].(string); ok {
		t.Md5 = md5Str
	}

	var totalSize int64
	var extractFiles = func(file map[string]interface{}) {
		var filename string
		var filelength int64
		var md5Str string
		if inter, ok := file["path.utf-8"].([]interface{}); ok {
			name := make([]string, len(inter))
			for i, v := range inter {
				name[i] = fmt.Sprint(v)
			}
			filename = strings.Join(name, "/")
		} else if inter, ok := file["path"].([]interface{}); ok {
			name := make([]string, len(inter))
			for i, v := range inter {
				name[i] = fmt.Sprint(v)
			}
			filename = strings.Join(name, "/")
		}
		if length, ok := file["length"].(int64); ok {
			filelength = length
			totalSize += filelength
		}
		if md5sum, ok := file["md5sum"].(string); ok {
			md5Str = md5sum
		}
		t.Files = append(t.Files, &torrentFile{Name: filename, Length: filelength, Md5: md5Str})
	}

	if files, ok := dict["files"].([]interface{}); ok {
		for _, file := range files {
			if f, ok := file.(map[string]interface{}); ok {
				extractFiles(f)
			}
		}
	}

	if t.Length == 0 {
		t.Length = totalSize
	}
	if len(t.Files) == 0 {
		t.Files = append(t.Files, &torrentFile{Name: t.Name, Length: t.Length})
	}

	return t, nil
}
