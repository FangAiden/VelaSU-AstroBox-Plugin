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
		dialog.DialogStyleSystem,
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
