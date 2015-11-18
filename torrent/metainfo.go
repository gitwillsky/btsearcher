package torrent

// torrent information
type MetaInfo struct {
	InfoHash   string
	Announce   string
	CreateDate int64 `bencode:"creation date"`
	Comment    string `bencode:"comment"`
	Encoding   string `bencode:"encoding"`

	Info       InfoDict
}

// torrent info section
type InfoDict struct {
	Files        []FileDict // multiple files.
	Name         string
	Length       uint64 `bencode:"length"`
	Md5sum       string

	PieceLength  uint64 `bencode:"piece length"`
	Pieces       string `bencode:"pieces"`
	Publisher    string `bencode:"publisher"`
	PublisherURL string `bencode:"publisher-url"`
}

// torrent file information
type FileDict struct {
	Length uint64
	Path   []string
	Md5sum string `bencode:"md5sum"`
}

// directory information
type DirectoryInfo struct {
	Name  string
	Dirs  []*DirectoryInfo
	Files []*File
}

// File
type File struct {
	Path   string // file path
	Length int64  // file length
}