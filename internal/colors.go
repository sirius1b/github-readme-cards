package internal

type ColorTheme int

const (
	SolarizedLight ColorTheme = iota
	GithubLight
	RosePineDawn
	QuietLight
	Dracula
	GruvBox
	Nord
	OneDark
	Monokai
	TokyoNight
)

// ThemeColors holds color values for a theme
type ThemeColors struct {
	Primary    string
	Secondary  string
	Accent     string
	Background string
	Text       string
}

// ThemeColorMap maps ColorTheme constants to their ThemeColors
var (
	ThemeColorMap = map[ColorTheme]ThemeColors{
		SolarizedLight: {
			Primary:    "#268bd2",
			Secondary:  "#2aa198",
			Accent:     "#b58900",
			Background: "#fdf6e3",
			Text:       "#657b83",
		},
		GithubLight: {
			Primary:    "#0969da",
			Secondary:  "#6e7781",
			Accent:     "#d4a72c",
			Background: "#ffffff",
			Text:       "#24292f",
		},
		RosePineDawn: {
			Primary:    "#b4637a",
			Secondary:  "#ea9d34",
			Accent:     "#56949f",
			Background: "#faf4ed",
			Text:       "#575279",
		},
		QuietLight: {
			Primary:    "#6c6c6c",
			Secondary:  "#b3b3b3",
			Accent:     "#ffab70",
			Background: "#f5f5f5",
			Text:       "#333333",
		},
		Dracula: {
			Primary:    "#bd93f9",
			Secondary:  "#ff79c6",
			Accent:     "#50fa7b",
			Background: "#282a36",
			Text:       "#f8f8f2",
		},
		GruvBox: {
			Primary:    "#fabd2f",
			Secondary:  "#b8bb26",
			Accent:     "#fe8019",
			Background: "#282828",
			Text:       "#ebdbb2",
		},
		Nord: {
			Primary:    "#5e81ac",
			Secondary:  "#88c0d0",
			Accent:     "#a3be8c",
			Background: "#2e3440",
			Text:       "#d8dee9",
		},
		OneDark: {
			Primary:    "#61afef",
			Secondary:  "#c678dd",
			Accent:     "#e5c07b",
			Background: "#282c34",
			Text:       "#abb2bf",
		},
		Monokai: {
			Primary:    "#f92672",
			Secondary:  "#a6e22e",
			Accent:     "#fd971f",
			Background: "#272822",
			Text:       "#f8f8f2",
		},
		TokyoNight: {
			Primary:    "#7aa2f7",
			Secondary:  "#bb9af7",
			Accent:     "#7dcfff",
			Background: "#1a1b26",
			Text:       "#c0caf5",
		},
	}
)
