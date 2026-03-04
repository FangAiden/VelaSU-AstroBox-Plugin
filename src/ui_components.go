package plugin

import (
	ui "astroboxplugin/bindings/astrobox_psys_host_ui_v3"
	"github.com/bytecodealliance/wit-bindgen/wit_types"
)

func el(t ui.ElementType, content string) *ui.Element {
	if content == "" {
		return ui.MakeElement(t, wit_types.None[string]())
	}
	return ui.MakeElement(t, wit_types.Some(content))
}

func makeColumn() *ui.Element {
	return el(ui.ElementTypeDiv, "").Flex().FlexDirection(ui.FlexDirectionColumn)
}

func makeRow() *ui.Element {
	return el(ui.ElementTypeDiv, "").Flex().FlexDirection(ui.FlexDirectionRow)
}

func makeSpacer() *ui.Element {
	return el(ui.ElementTypeDiv, "").FlexGrow(1)
}

func makeTitle(value string) *ui.Element {
	return el(ui.ElementTypeSpan, value).
		Size(24).
		TextColor("#E8ECF8")
}

func makeSectionTitle(value string) *ui.Element {
	return el(ui.ElementTypeSpan, value).
		Size(16).
		TextColor("#D8DCF0")
}

func makeText(value string) *ui.Element {
	return el(ui.ElementTypeSpan, value).
		Size(13).
		TextColor("#D0D6E8")
}

func makeMutedText(value string) *ui.Element {
	return el(ui.ElementTypeSpan, value).
		Size(12).
		TextColor("#8F96AB")
}

func makeCodeBlock(value string) *ui.Element {
	return el(ui.ElementTypeCode, value).
		WidthFull().
		Size(12).
		TextColor("#C6F0E1").
		Bg("#0C131C").
		Border(1, "#233044").
		Radius(8).
		Padding(8)
}

func makeBadge(value string) *ui.Element {
	return el(ui.ElementTypeBadge, value).
		Bg("#2A334A").
		TextColor("#BFD0FF").
		Radius(10).
		PaddingLeft(8).
		PaddingRight(8).
		PaddingTop(2).
		PaddingBottom(2)
}

func makePrimaryButton(label string, eventID string) *ui.Element {
	return applyButtonMotion(
		el(ui.ElementTypeButton, label).
			Bg("#4A6CF7").
			TextColor("#FFFFFF").
			Border(1, "#5B7AF8").
			Radius(8).
			MinHeight(36).
			Padding(8).
			On(ui.EventClick, eventID),
	)
}

func makeDangerButton(label string, eventID string) *ui.Element {
	return applyButtonMotion(
		el(ui.ElementTypeButton, label).
			Bg("#B63A4B").
			TextColor("#FFFFFF").
			Border(1, "#C44D5E").
			Radius(8).
			MinHeight(36).
			Padding(8).
			On(ui.EventClick, eventID),
	)
}

func makeSecondaryButton(label string, eventID string) *ui.Element {
	return applyButtonMotion(
		el(ui.ElementTypeButton, label).
			Bg("#252D44").
			TextColor("#D4DAEE").
			Border(1, "#394360").
			Radius(8).
			MinHeight(36).
			Padding(8).
			On(ui.EventClick, eventID),
	)
}

func makeActionButton(label string, eventID string) *ui.Element {
	return applyButtonMotion(
		el(ui.ElementTypeButton, label).
			Bg("#1A1F32").
			TextColor("#D4DAEE").
			Border(1, "#323D5A").
			Radius(16).
			Padding(8).
			Size(12).
			On(ui.EventClick, eventID),
	)
}

func makeDangerActionButton(label string, eventID string) *ui.Element {
	return applyButtonMotion(
		el(ui.ElementTypeButton, label).
			Bg("#35161C").
			TextColor("#F8B4BF").
			Border(1, "#63343D").
			Radius(16).
			Padding(8).
			Size(12).
			On(ui.EventClick, eventID),
	)
}

func makeSingleLineInput(value string, eventID string) *ui.Element {
	return el(ui.ElementTypeInput, value).
		WithoutDefaultStyles().
		WidthFull().
		Bg("#11172A").
		TextColor("#E6ECFF").
		Border(1, "#313A56").
		Radius(8).
		Padding(8).
		On(ui.EventInput, eventID).
		On(ui.EventChange, eventID)
}

func makeInputArea(value string, eventID string) *ui.Element {
	return el(ui.ElementTypeTextarea, value).
		WithoutDefaultStyles().
		WidthFull().
		Bg("#11172A").
		TextColor("#E6ECFF").
		Border(1, "#313A56").
		Radius(8).
		Padding(8).
		On(ui.EventInput, eventID).
		On(ui.EventChange, eventID)
}

func makeSettingCard(icon, title, subtitle string, action *ui.Element) *ui.Element {
	row := el(ui.ElementTypeCard, "").
		WidthFull().
		Bg("#141A2B").
		Border(1, "#28314A").
		Radius(12).
		Padding(12).
		MarginBottom(10).
		Flex().AlignCenter().Gap(10)

	row = row.Child(makeSVGIcon(icon))

	content := makeColumn().FlexGrow(1)
	content = content.Child(makeText(title).Size(14).TextColor("#E8ECF8"))
	if subtitle != "" {
		content = content.Child(makeMutedText(subtitle).MarginTop(2))
	}
	row = row.Child(content)

	if action != nil {
		row = row.Child(action)
	}
	return row
}

func makePanel() *ui.Element {
	return el(ui.ElementTypeCard, "").
		WidthFull().
		Bg("#12182A").
		Border(1, "#2A344F").
		Radius(12).
		Padding(10)
}

func makeDashboardCard(iconStr, title, desc, eventID string) *ui.Element {
	content := makeColumn().
		AlignCenter().
		JustifyCenter().
		Gap(8).
		Child(makeSVGIcon(iconStr)).
		Child(makeText(title).Size(16).TextColor("#EEF2FF")).
		Child(makeMutedText(desc).Size(12))

	return applyButtonMotion(
		el(ui.ElementTypeButton, "").
			WithoutDefaultStyles().
			Width(180).
			MinHeight(150).
			Bg("#171E32").
			Border(1, "#2A3552").
			Radius(14).
			Padding(16).
			Margin(8).
			On(ui.EventClick, eventID).
			Child(content),
	)
}

func buildOptionPill(iconStr string, label string, eventID string) *ui.Element {
	content := makeRow().AlignCenter().Gap(4)
	if iconStr != "" {
		content = content.Child(makeSVGIcon(iconStr))
	}
	if label != "" {
		content = content.Child(makeText(label).Size(13).TextColor("#E8ECF8"))
	}

	return applyButtonMotion(
		el(ui.ElementTypeButton, "").
			WithoutDefaultStyles().
			Bg("#232D47").
			Border(1, "#3A4462").
			Radius(18).
			Padding(8).
			On(ui.EventClick, eventID).
			Child(content),
	)
}
