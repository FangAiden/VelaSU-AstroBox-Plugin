package plugin

import (
	ui "astroboxplugin/bindings/astrobox_psys_host_ui_v3"
)

const (
	pageAnimationDurationMs   = 260
	sectionAnimationDelayMs   = 70
	buttonAnimationDurationMs = 180
)

func applyPageMotion(route string, node *ui.Element) *ui.Element {
	preset := "fade-in-up"
	switch route {
	case RouteDashboard:
		preset = "fade-in"
	case RouteTerminal:
		preset = "slide-in-left"
	case RouteFileMgr:
		preset = "slide-in-right"
	case RouteSettings:
		preset = "fade-in-down"
	case RouteLogs:
		preset = "zoom-in"
	}
	return node.
		AnimationPreset(preset).
		AnimationDurationMs(pageAnimationDurationMs).
		AnimationFillMode("both").
		AnimationEasing("cubic-bezier(0.22, 1, 0.36, 1)").
		WillChange("opacity, transform").
		Transform("translateZ(0)").
		BackfaceVisibility("hidden")
}

func applySectionMotion(node *ui.Element, delayMs uint32) *ui.Element {
	return node.
		AnimationPreset("fade-in-up").
		AnimationDelayMs(delayMs).
		AnimationDurationMs(pageAnimationDurationMs).
		AnimationFillMode("both").
		WillChange("opacity, transform").
		Transform("translateZ(0)")
}

func applyButtonMotion(node *ui.Element) *ui.Element {
	return node.
		AnimationPreset("fade-in").
		AnimationDurationMs(buttonAnimationDurationMs).
		AnimationFillMode("both").
		Transition("transform 140ms ease, filter 160ms ease, background-color 180ms ease, border-color 180ms ease, box-shadow 200ms ease").
		WillChange("transform, filter, background-color, border-color").
		Transform("translateZ(0)")
}
