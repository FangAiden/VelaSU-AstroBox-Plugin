package plugin

type FileSortMode string

const (
	FileSortByName FileSortMode = "name"
	FileSortBySize FileSortMode = "size"
	FileSortByDate FileSortMode = "date"
)

type FileViewMode string

const (
	FileViewGrid FileViewMode = "grid"
	FileViewList FileViewMode = "list"
)

type FileEntry struct {
	Name      string
	Path      string
	Exists    bool
	IsDir     bool
	Size      int64
	MetaReady bool
	MetaErr   string
}

type ClipboardState struct {
	SourcePath  string
	SourceIsDir bool
	Mode        string // copy | move
}

type EditorState struct {
	Path       string
	Text       string
	IsBinary   bool
	HexPreview string
	Loaded     bool
}

type TransferJob struct {
	Kind      string // upload | download
	Path      string
	Progress  string
	Completed bool
	Error     string
}
