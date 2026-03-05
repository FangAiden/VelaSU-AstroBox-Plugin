package plugin

import (
	dialog "astroboxplugin/bindings/astrobox_psys_host_dialog"
	"fmt"
	"strings"

	"github.com/bytecodealliance/wit-bindgen/wit_types"
)

const (
	dialogConfirmID = "confirm"
	dialogCancelID  = "cancel"
)

func confirmDialog(title string, content string) bool {
	ret := dialog.ShowDialog(
		dialog.DialogTypeAlert,
		dialog.DialogStyleSystem,
		dialog.DialogInfo{
			Title:   title,
			Content: content,
			Buttons: []dialog.DialogButton{
				{Id: dialogConfirmID, Primary: true, Content: "确认"},
				{Id: dialogCancelID, Primary: false, Content: "取消"},
			},
		},
	).Read()
	return ret.ClickedBtnId == dialogConfirmID
}

func alertDialog(title string, content string) {
	dialog.ShowDialog(
		dialog.DialogTypeAlert,
		dialog.DialogStyleSystem,
		dialog.DialogInfo{
			Title:   title,
			Content: content,
			Buttons: []dialog.DialogButton{
				{Id: dialogConfirmID, Primary: true, Content: "确定"},
			},
		},
	).Read()
}

func promptInputDialog(title string, content string, defaultValue string) (string, bool) {
	message := content
	if strings.TrimSpace(defaultValue) != "" {
		message = fmt.Sprintf("%s\n当前值: %s", content, defaultValue)
	}
	ret := dialog.ShowDialog(
		dialog.DialogTypeInput,
		dialog.DialogStyleWebsite,
		dialog.DialogInfo{
			Title:   title,
			Content: message,
			Buttons: []dialog.DialogButton{
				{Id: dialogConfirmID, Primary: true, Content: "确认"},
				{Id: dialogCancelID, Primary: false, Content: "取消"},
			},
		},
	).Read()
	if ret.ClickedBtnId != dialogConfirmID {
		return "", false
	}
	val := strings.TrimSpace(ret.InputResult)
	if val == "" {
		val = strings.TrimSpace(defaultValue)
	}
	return val, true
}

func pickLocalFile() (string, []byte, error) {
	ret := dialog.PickFile(
		dialog.PickConfig{Read: true, CopyTo: wit_types.None[string]()},
		dialog.FilterConfig{Multiple: false, Extensions: []string{}, DefaultDirectory: "", DefaultFileName: ""},
	).Read()
	if strings.TrimSpace(ret.Name) == "" {
		return "", nil, fmt.Errorf("未选择文件")
	}
	return ret.Name, ret.Data, nil
}

func startLocalSaveSession(defaultFileName string) (uint64, string, error) {
	ret := dialog.SaveFileStart(
		dialog.FilterConfig{
			Multiple:         false,
			Extensions:       []string{},
			DefaultDirectory: "",
			DefaultFileName:  strings.TrimSpace(defaultFileName),
		},
	).Read()
	if ret.IsErr() {
		return 0, "", fmt.Errorf("未选择保存位置")
	}
	session := ret.Ok()
	name := strings.TrimSpace(session.Name)
	if name == "" {
		name = strings.TrimSpace(defaultFileName)
	}
	return session.SessionId, name, nil
}

func writeLocalSaveChunk(sessionID uint64, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	ret := dialog.SaveFileWriteChunk(sessionID, data).Read()
	if ret.IsErr() {
		return fmt.Errorf("写入本地分块失败")
	}
	return nil
}

func finishLocalSaveSession(sessionID uint64) error {
	ret := dialog.SaveFileFinish(sessionID).Read()
	if ret.IsErr() {
		return fmt.Errorf("完成本地保存失败")
	}
	return nil
}

func abortLocalSaveSession(sessionID uint64) {
	dialog.SaveFileAbort(sessionID).Read()
}
