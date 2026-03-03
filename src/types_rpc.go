package plugin

import "encoding/json"

type RpcRequest struct {
	ID     string `json:"id"`
	Method string `json:"method"`
	Token  string `json:"token,omitempty"`
	Params any    `json:"params,omitempty"`
}

type RpcError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type RpcResponse struct {
	V       int             `json:"v,omitempty"`
	ID      string          `json:"id"`
	OK      bool            `json:"ok"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RpcError       `json:"error,omitempty"`
	Message string          `json:"message,omitempty"`
}

type HelloResult struct {
	Server        string `json:"server"`
	Protocol      int    `json:"protocol"`
	RemoteEnabled bool   `json:"remoteEnabled"`
	HasToken      bool   `json:"hasToken"`
	TS            int64  `json:"ts"`
}

type ShellExecResult struct {
	Cmd      string `json:"cmd"`
	Mode     string `json:"mode"`
	ExitCode *int   `json:"exitCode"`
	Output   string `json:"output"`
	PID      *int   `json:"pid"`
	JobID    string `json:"jobId"`
	Cwd      string `json:"cwd"`
}

type ShellCwdResult struct {
	Cwd string `json:"cwd"`
}

type FsStatResult struct {
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
	IsDir  bool   `json:"is_dir"`
	Size   int64  `json:"size"`
}

type FsReadResult struct {
	Path       string `json:"path"`
	Encoding   string `json:"encoding"`
	Offset     int    `json:"offset"`
	NextOffset int    `json:"next_offset"`
	Eof        bool   `json:"eof"`
	Size       int64  `json:"size"`
	Data       string `json:"data"`
}

type FsWriteResult struct {
	Path  string `json:"path"`
	Bytes int    `json:"bytes"`
	Mode  string `json:"mode"`
}
