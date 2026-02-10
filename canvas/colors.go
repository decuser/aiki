package canvas

import "image/color"

// CGA 16 colors
var Colors = map[string]color.RGBA{
	"black":          {0, 0, 0, 255},
	"blue":           {0, 0, 170, 255},
	"green":          {0, 170, 0, 255},
	"cyan":           {0, 170, 170, 255},
	"red":            {170, 0, 0, 255},
	"magenta":        {170, 0, 170, 255},
	"brown":          {170, 85, 0, 255},
	"white":          {170, 170, 170, 255},
	"gray":           {85, 85, 85, 255},
	"bright_blue":    {85, 85, 255, 255},
	"bright_green":   {85, 255, 85, 255},
	"bright_cyan":    {85, 255, 255, 255},
	"bright_red":     {255, 85, 85, 255},
	"bright_magenta": {255, 85, 255, 255},
	"yellow":         {255, 255, 85, 255},
	"bright_white":   {255, 255, 255, 255},
}

var DefaultBG = Colors["black"]
var DefaultFG = Colors["bright_white"]
