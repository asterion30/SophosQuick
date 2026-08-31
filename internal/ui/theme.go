package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// DarkSlateTheme provides an ultra-clean, modern Dark Slate aesthetic (macOS/Raycast/Linear style).
type DarkSlateTheme struct{}

var _ fyne.Theme = (*DarkSlateTheme)(nil)

// Palette definitions
var (
	ColorDarkSlate900 = color.NRGBA{R: 0x0F, G: 0x17, B: 0x2A, A: 0xFF} // #0F172A
	ColorDarkSlate800 = color.NRGBA{R: 0x1E, G: 0x29, B: 0x3B, A: 0xFF} // #1E293B
	ColorDarkSlate700 = color.NRGBA{R: 0x33, G: 0x41, B: 0x55, A: 0xFF} // #334155
	ColorDarkSlate600 = color.NRGBA{R: 0x47, G: 0x55, B: 0x69, A: 0xFF} // #475569
	ColorSlate400     = color.NRGBA{R: 0x94, G: 0xA3, B: 0xB8, A: 0xFF} // #94A3B8
	ColorSlate100     = color.NRGBA{R: 0xF1, G: 0xF5, B: 0xF9, A: 0xFF} // #F1F5F9
	ColorSlate50      = color.NRGBA{R: 0xF8, G: 0xFA, B: 0xFC, A: 0xFF} // #F8FAFC

	ColorIndigo500 = color.NRGBA{R: 0x63, G: 0x66, B: 0xF1, A: 0xFF} // #6366F1
	ColorIndigo600 = color.NRGBA{R: 0x4F, G: 0x46, B: 0xE5, A: 0xFF} // #4F46E5
	ColorCyan500   = color.NRGBA{R: 0x06, G: 0xB6, B: 0xD4, A: 0xFF} // #06B6D4
	ColorEmerald500 = color.NRGBA{R: 0x10, G: 0xB9, B: 0x81, A: 0xFF} // #10B981
	ColorRed500    = color.NRGBA{R: 0xEF, G: 0x44, B: 0x44, A: 0xFF} // #EF4444
	ColorAmber500  = color.NRGBA{R: 0xF5, G: 0x9E, B: 0x0B, A: 0xFF} // #F59E0B
)

func (t *DarkSlateTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return ColorDarkSlate900
	case theme.ColorNameButton:
		return ColorDarkSlate800
	case theme.ColorNameDisabledButton:
		return ColorDarkSlate700
	case theme.ColorNameDisabled:
		return ColorDarkSlate600
	case theme.ColorNameForeground:
		return ColorSlate50
	case theme.ColorNameHover:
		return ColorDarkSlate700
	case theme.ColorNameInputBackground:
		return ColorDarkSlate800
	case theme.ColorNamePlaceHolder:
		return ColorSlate400
	case theme.ColorNamePrimary:
		return ColorIndigo500
	case theme.ColorNameFocus:
		return ColorCyan500
	case theme.ColorNameSelection:
		return ColorIndigo600
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0x02, G: 0x06, B: 0x17, A: 0x99}
	case theme.ColorNameMenuBackground:
		return ColorDarkSlate800
	case theme.ColorNameOverlayBackground:
		return ColorDarkSlate800
	case theme.ColorNameError:
		return ColorRed500
	case theme.ColorNameSuccess:
		return ColorEmerald500
	case theme.ColorNameWarning:
		return ColorAmber500
	case theme.ColorNameHeaderBackground:
		return ColorDarkSlate900
	case theme.ColorNameSeparator:
		return ColorDarkSlate700
	default:
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

func (t *DarkSlateTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *DarkSlateTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *DarkSlateTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInlineIcon:
		return 16
	case theme.SizeNameScrollBar:
		return 8
	case theme.SizeNameScrollBarSmall:
		return 4
	case theme.SizeNameText:
		return 12.5
	case theme.SizeNameHeadingText:
		return 16
	case theme.SizeNameSubHeadingText:
		return 13
	case theme.SizeNameCaptionText:
		return 10.5
	case theme.SizeNameInputBorder:
		return 1.5
	default:
		return theme.DefaultTheme().Size(name)
	}
}
