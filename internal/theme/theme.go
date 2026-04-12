package theme

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	App            lipgloss.Style
	HeaderTitle    lipgloss.Style
	Panel          lipgloss.Style
	PanelFocused   lipgloss.Style
	PanelTitle     lipgloss.Style
	Text           lipgloss.Style
	Muted          lipgloss.Style
	Accent         lipgloss.Style
	Badge          lipgloss.Style
	BadgeMuted     lipgloss.Style
	Footer         lipgloss.Style
	Key            lipgloss.Style
	QuitConfirmKey lipgloss.Style
	Status         lipgloss.Style
	Success        lipgloss.Style
	Warning        lipgloss.Style
	RowFocused     lipgloss.Style
	Meta           lipgloss.Style
}

func New() Theme {
	bg := lipgloss.Color("#0B0F13")
	panelBg := lipgloss.Color("#121821")
	panelBgFocus := lipgloss.Color("#151E28")
	border := lipgloss.Color("#2A3642")
	focus := lipgloss.Color("#7EA9BC")
	text := lipgloss.Color("#E6EDF3")
	muted := lipgloss.Color("#8B9BA7")
	accent := lipgloss.Color("#C8D4DC")
	success := lipgloss.Color("#A0C487")
	warning := lipgloss.Color("#D89B57")
	surface := lipgloss.Color("#1B2530")
	surfaceHi := lipgloss.Color("#253240")
	rowFocus := lipgloss.Color("#1E2A36")
	panelBorder := lipgloss.Border{
		Top:         "▁",
		Bottom:      "▔",
		Left:        "▕",
		Right:       "▏",
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
	}
	return Theme{
		App: lipgloss.NewStyle().
			Foreground(text),

		HeaderTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(accent),

		Panel: lipgloss.NewStyle().
			Background(panelBg).
			Border(panelBorder).
			BorderForeground(border).
			Padding(1, 2),

		PanelFocused: lipgloss.NewStyle().
			Background(panelBgFocus).
			Border(panelBorder).
			BorderForeground(focus).
			Padding(1, 2),

		PanelTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(accent),

		Text: lipgloss.NewStyle().
			Foreground(text),

		Muted: lipgloss.NewStyle().
			Foreground(muted),

		Accent: lipgloss.NewStyle().
			Foreground(accent),

		Badge: lipgloss.NewStyle().
			Bold(true).
			Foreground(bg).
			Background(focus).
			Padding(0, 1),

		BadgeMuted: lipgloss.NewStyle().
			Foreground(text).
			Background(surface).
			Padding(0, 1),

		Footer: lipgloss.NewStyle().
			// Background(bg).
			Foreground(muted).
			Padding(0, 1),

		Key: lipgloss.NewStyle().
			Bold(true).
			Foreground(text).
			Background(surfaceHi).
			Padding(0, 1),

		QuitConfirmKey: lipgloss.NewStyle().
			Bold(true).
			Foreground(text).
			Background(surfaceHi).
			Padding(0, 0),

		Status: lipgloss.NewStyle().
			Foreground(muted),

		Success: lipgloss.NewStyle().
			Foreground(success),

		Warning: lipgloss.NewStyle().
			Foreground(warning),

		RowFocused: lipgloss.NewStyle().
			Foreground(text).
			Background(rowFocus).
			Bold(true).
			Padding(0, 1),

		Meta: lipgloss.NewStyle().
			Foreground(muted),
	}
}
