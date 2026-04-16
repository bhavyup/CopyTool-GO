package theme

import "github.com/charmbracelet/lipgloss"

// Palette is the raw color vocabulary for the app.
type Palette struct {
	Base    lipgloss.Color
	Surface lipgloss.Color
	Overlay lipgloss.Color
	Raised  lipgloss.Color

	BorderSubtle lipgloss.Color
	BorderFocus  lipgloss.Color
	BorderStrong lipgloss.Color

	TextPrimary   lipgloss.Color
	TextSecondary lipgloss.Color
	TextMuted     lipgloss.Color
	TextInverse   lipgloss.Color

	Accent  lipgloss.Color
	Success lipgloss.Color
	Warning lipgloss.Color
	Danger  lipgloss.Color

	SelFull    lipgloss.Color
	SelPartial lipgloss.Color
	SelNone    lipgloss.Color

	DirColor  lipgloss.Color
	FileColor lipgloss.Color
	ExtColor  lipgloss.Color

	LineNumColor lipgloss.Color
	CodeBg       lipgloss.Color
	Transparent  lipgloss.Color
}

// NightPalette is a GitHub-dark-inspired palette tuned for terminal contrast.
var NightPalette = Palette{
	Base:    lipgloss.Color("#0D1117"),
	Surface: lipgloss.Color("#161B22"),
	Overlay: lipgloss.Color("#1C2128"),
	Raised:  lipgloss.Color("#21262D"),

	BorderSubtle: lipgloss.Color("#30363D"),
	BorderFocus:  lipgloss.Color("#388BFD"),
	BorderStrong: lipgloss.Color("#484F58"),

	TextPrimary:   lipgloss.Color("#E6EDF3"),
	TextSecondary: lipgloss.Color("#8B949E"),
	TextMuted:     lipgloss.Color("#484F58"),
	TextInverse:   lipgloss.Color("#0D1117"),

	Accent:  lipgloss.Color("#388BFD"),
	Success: lipgloss.Color("#3FB950"),
	Warning: lipgloss.Color("#D29922"),
	Danger:  lipgloss.Color("#F85149"),

	SelFull:    lipgloss.Color("#F0A050"),
	SelPartial: lipgloss.Color("#7C6F52"),
	SelNone:    lipgloss.Color("#30363D"),

	DirColor:  lipgloss.Color("#79C0FF"),
	FileColor: lipgloss.Color("#E6EDF3"),
	ExtColor:  lipgloss.Color("#484F58"),

	LineNumColor: lipgloss.Color("#3D444D"),
	CodeBg:       lipgloss.Color("#161B22"),
	Transparent:  lipgloss.Color(""),
}

type Theme struct {
	// Layout
	App lipgloss.Style

	// Panels
	Panel        lipgloss.Style
	PanelFocused lipgloss.Style

	// Header
	HeaderBar     lipgloss.Style
	HeaderBrand   lipgloss.Style
	HeaderProject lipgloss.Style
	HeaderStat    lipgloss.Style
	HeaderBadge   lipgloss.Style
	HeaderDirty   lipgloss.Style

	// Tree / Navigator
	TreeRowNormal  lipgloss.Style
	TreeRowFocused lipgloss.Style
	TreeRowDir     lipgloss.Style
	TreeRowFile    lipgloss.Style
	TreeRowExt     lipgloss.Style
	TreeIndent     lipgloss.Style
	TreeSelFull    lipgloss.Style
	TreeSelPartial lipgloss.Style
	TreeSelNone    lipgloss.Style
	TreeDirOpen    lipgloss.Style
	TreeDirClosed  lipgloss.Style
	TreeDirEmpty   lipgloss.Style

	// Inspect
	InspectTitle    lipgloss.Style
	InspectSubtitle lipgloss.Style
	InspectSection  lipgloss.Style
	InspectKey      lipgloss.Style
	InspectVal      lipgloss.Style
	InspectMuted    lipgloss.Style
	InspectWarning  lipgloss.Style
	InspectLineNum  lipgloss.Style
	InspectCode     lipgloss.Style

	// Drawers / overlays
	DrawerPanel   lipgloss.Style
	DrawerTitle   lipgloss.Style
	DrawerSection lipgloss.Style
	DrawerInput   lipgloss.Style
	DrawerHint    lipgloss.Style
	DrawerBtn     lipgloss.Style

	// Status / footer
	StatusBar     lipgloss.Style
	StatusMsg     lipgloss.Style
	StatusSuccess lipgloss.Style
	StatusWarn    lipgloss.Style
	KeyChip       lipgloss.Style
	KeyChipAccent lipgloss.Style
	KeyLabel      lipgloss.Style

	// Palette / search
	PaletteInput   lipgloss.Style
	PaletteMatch   lipgloss.Style
	PaletteDir     lipgloss.Style
	PaletteFocused lipgloss.Style

	// Skeleton
	SkeletonDim lipgloss.Style
	SkeletonMid lipgloss.Style
	SkeletonHi  lipgloss.Style

	// Generic
	Text    lipgloss.Style
	Muted   lipgloss.Style
	Accent  lipgloss.Style
	Success lipgloss.Style
	Warning lipgloss.Style
	Danger  lipgloss.Style

	TextWS lipgloss.Style // for rendering whitespace (e.g. in file names) with a visible marker

	// Raw palette access
	P Palette

	// Deprecated compatibility aliases (to keep existing call-sites stable while
	// files are migrated gradually).
	HeaderTitle    lipgloss.Style
	PanelTitle     lipgloss.Style
	Badge          lipgloss.Style
	BadgeMuted     lipgloss.Style
	Footer         lipgloss.Style
	Key            lipgloss.Style
	QuitConfirmKey lipgloss.Style
	Status         lipgloss.Style
	RowFocused     lipgloss.Style
	Meta           lipgloss.Style
}

func New() Theme {
	p := NightPalette

	border := lipgloss.Border{
		Top:         "▁",
		Bottom:      "▔",
		Left:        "▕",
		Right:       "▏",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
	}

	keyChipBase := lipgloss.NewStyle().
		Background(p.Raised).
		Foreground(p.TextPrimary).
		Padding(0, 0)

	th := Theme{
		P: p,

		App: lipgloss.NewStyle().
			Foreground(p.TextPrimary),

		Panel: lipgloss.NewStyle().
			Background(p.Surface).
			Border(border).
			BorderForeground(p.BorderSubtle).
			Padding(1, 2),

		PanelFocused: lipgloss.NewStyle().
			Background(p.Surface).
			Border(border).
			BorderForeground(p.BorderFocus).
			Padding(1, 2),

		HeaderBar: lipgloss.NewStyle().
			Background(p.Surface).
			BorderBottom(true).
			Border(border).
			// BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(p.BorderSubtle).
			Padding(1, 2),

		HeaderBrand: lipgloss.NewStyle().
			Background(p.Surface).
			Foreground(p.TextPrimary).
			Bold(true),

		HeaderProject: lipgloss.NewStyle().
			Background(p.Accent).
			Foreground(p.TextInverse).
			Bold(true).
			Padding(0, 1),

		HeaderStat: lipgloss.NewStyle().
			Background(p.Surface).
			Foreground(p.TextMuted),

		HeaderBadge: lipgloss.NewStyle().
			Background(p.Raised).
			Foreground(p.TextSecondary).
			Padding(0, 1),

		HeaderDirty: lipgloss.NewStyle().
			Background(p.Warning).
			Foreground(p.TextInverse).
			Bold(true).
			Padding(0, 1),

		TreeRowNormal: lipgloss.NewStyle().
			Background(p.Surface).
			Foreground(p.TextPrimary),

		TreeRowFocused: lipgloss.NewStyle().
			Background(p.Raised).
			Foreground(p.TextPrimary).
			Bold(true),

		TreeRowDir: lipgloss.NewStyle().
			Background(p.Surface).
			Foreground(p.DirColor),

		TreeRowFile: lipgloss.NewStyle().
			Background(p.Surface).
			Foreground(p.FileColor),

		TreeRowExt: lipgloss.NewStyle().
			Background(p.Surface).
			Foreground(p.ExtColor),

		TreeIndent: lipgloss.NewStyle().
			Foreground(p.TextMuted).
			Background(p.Surface),

		TreeSelFull: lipgloss.NewStyle().
			Background(p.Surface).
			Foreground(p.SelFull).
			Bold(true),

		TreeSelPartial: lipgloss.NewStyle().
			Background(p.Surface).
			Foreground(p.SelPartial),

		TreeSelNone: lipgloss.NewStyle().
			Background(p.Surface).
			Foreground(p.SelNone),

		TreeDirOpen: lipgloss.NewStyle().
			Background(p.Surface).
			Foreground(p.TextSecondary),

		TreeDirClosed: lipgloss.NewStyle().
			Background(p.Surface).
			Foreground(p.TextSecondary),

		TreeDirEmpty: lipgloss.NewStyle().
			Foreground(p.TextMuted).
			Background(p.Surface),

		InspectTitle: lipgloss.NewStyle().
			Foreground(p.TextPrimary).
			Bold(true),

		InspectSubtitle: lipgloss.NewStyle().
			Foreground(p.TextSecondary),

		InspectSection: lipgloss.NewStyle().
			Foreground(p.Accent),

		InspectKey: lipgloss.NewStyle().
			Foreground(p.TextSecondary),

		InspectVal: lipgloss.NewStyle().
			Foreground(p.TextPrimary),

		InspectMuted: lipgloss.NewStyle().
			Foreground(p.TextMuted),

		InspectWarning: lipgloss.NewStyle().
			Foreground(p.Warning),

		InspectLineNum: lipgloss.NewStyle().
			Foreground(p.LineNumColor),

		InspectCode: lipgloss.NewStyle().
			Foreground(p.TextPrimary),

		DrawerPanel: lipgloss.NewStyle().
			Background(p.Overlay).
			Border(border).
			BorderForeground(p.BorderFocus).
			Padding(1, 3),

		DrawerTitle: lipgloss.NewStyle().
			Foreground(p.TextPrimary).
			Bold(true),

		DrawerSection: lipgloss.NewStyle().
			Foreground(p.Accent),

		DrawerInput: lipgloss.NewStyle().
			Background(p.Raised).
			Foreground(p.TextPrimary).
			Padding(0, 1),

		DrawerHint: lipgloss.NewStyle().
			Foreground(p.TextMuted),

		DrawerBtn: lipgloss.NewStyle().
			Background(p.Raised).
			Foreground(p.TextSecondary).
			Padding(0, 1),

		StatusBar: lipgloss.NewStyle().
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(p.BorderSubtle).
			Padding(0, 2),

		StatusMsg: lipgloss.NewStyle().
			Background(p.Surface).
			Foreground(p.TextSecondary),

		StatusSuccess: lipgloss.NewStyle().
			Background(p.Surface).
			Foreground(p.Success),

		StatusWarn: lipgloss.NewStyle().
			Background(p.Surface).
			Foreground(p.Warning),

		KeyChip: keyChipBase,

		KeyChipAccent: keyChipBase.Copy().
			Background(p.Accent).
			Foreground(p.TextInverse).
			Bold(true),

		KeyLabel: lipgloss.NewStyle().
			Background(p.Surface).
			Foreground(p.TextMuted),

		PaletteInput: lipgloss.NewStyle().
			Foreground(p.TextPrimary),

		PaletteMatch: lipgloss.NewStyle().
			Foreground(p.TextSecondary),

		PaletteDir: lipgloss.NewStyle().
			Foreground(p.DirColor),

		PaletteFocused: lipgloss.NewStyle().
			Background(p.Raised).
			Foreground(p.TextPrimary).
			Bold(true),

		SkeletonDim: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#21262D")),

		SkeletonMid: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#30363D")),

		SkeletonHi: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#484F58")),

		Text: lipgloss.NewStyle().
			Foreground(p.TextPrimary),

		Muted: lipgloss.NewStyle().
			Foreground(p.TextMuted),

		Accent: lipgloss.NewStyle().
			Foreground(p.Accent),

		Success: lipgloss.NewStyle().
			Foreground(p.Success),

		Warning: lipgloss.NewStyle().
			Foreground(p.Warning),

		Danger: lipgloss.NewStyle().
			Foreground(p.Danger),

		TextWS: lipgloss.NewStyle().
			Background(p.Surface).
			Foreground(p.TextMuted),
	}

	// Compatibility aliases
	th.HeaderTitle = th.HeaderBrand
	th.PanelTitle = th.InspectSection.Copy().Bold(true)
	th.Badge = th.HeaderProject
	th.BadgeMuted = th.HeaderBadge
	th.Footer = th.StatusBar.Copy().BorderTop(false).Padding(0, 1)
	th.Key = th.KeyChip
	th.QuitConfirmKey = th.KeyChip
	th.Status = th.StatusMsg
	th.RowFocused = th.TreeRowFocused.Copy().Padding(0, 1)
	th.Meta = th.InspectLineNum

	return th
}

// RuledSection renders a `──── LABEL ────` rule in accent color.
func RuledSection(th Theme, label string, width int) string {
	if width <= 0 {
		return ""
	}

	labelWidth := len([]rune(label))
	dashTotal := width - labelWidth - 2
	if dashTotal < 4 {
		return th.InspectSection.Render(label)
	}

	leftDashes := 2
	rightDashes := dashTotal - leftDashes
	if rightDashes < 0 {
		rightDashes = 0
	}

	rule := repeatStr("─", leftDashes) + " " + label + " " + repeatStr("─", rightDashes)
	return th.InspectSection.Render(rule)
}

// SelectionGlyph returns the colorized selection marker glyph.
func SelectionGlyph(th Theme, selected, partial bool, focused bool) string {
	bg := th.P.Raised
	TreeSelNone := th.TreeSelNone
	TreeSelPartial := th.TreeSelPartial
	TreeSelFull := th.TreeSelFull

	if focused {
		TreeSelNone = TreeSelNone.Copy().Background(bg)
		TreeSelPartial = TreeSelPartial.Copy().Background(bg)
		TreeSelFull = TreeSelFull.Copy().Background(bg)
	}

	switch {
	case partial:
		return TreeSelPartial.Render("◐")
	case selected:
		return TreeSelFull.Render("●")
	default:
		return TreeSelNone.Render("○")
	}
}

// DirArrow returns the colorized directory arrow marker.
func DirArrow(th Theme, expanded bool, hasChildren bool, focused bool) string {
	bg := th.P.Raised
	TreeDirEmpty := th.TreeDirEmpty
	TreeDirOpen := th.TreeDirOpen
	TreeDirClosed := th.TreeDirClosed

	if focused {
		TreeDirEmpty = TreeDirEmpty.Copy().Background(bg)
		TreeDirOpen = TreeDirOpen.Copy().Background(bg)
		TreeDirClosed = TreeDirClosed.Copy().Background(bg)
	}

	if !hasChildren {
		return TreeDirEmpty.Render("╴")
	}
	if expanded {
		return TreeDirOpen.Render("▼") //edit it to bigger triangle
	}
	return TreeDirClosed.Render("▶")
}

// KeyHint renders `[key] label` with standard key chip styling.
func KeyHint(th Theme, key, label string) string {
	return th.KeyChip.Render(key) + th.KeyLabel.Render(label)
}

// KeyHintAccent renders a primary-action key hint.
func KeyHintAccent(th Theme, key, label string) string {
	return th.KeyChipAccent.Render(key) + th.KeyLabel.Render(label)
}

func repeatStr(s string, n int) string {
	if n <= 0 {
		return ""
	}
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
