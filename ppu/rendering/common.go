package rendering

import (
	"image/color"

	"github.com/maxfierke/gogo-gb/ppu"
)

const (
	FB_WIDTH  = 160
	FB_HEIGHT = 144
)

type RenderedPixel struct {
	Layer     PixelLayer
	ColorID   ppu.ColorID
	PaletteID uint8
	Color     color.RGBA
}

type PixelLayer uint8

const (
	PIXEL_LAYER_BG  PixelLayer = iota // Background/window layer
	PIXEL_LAYER_BGP                   // Background/window layer w/ priority over objects
	PIXEL_LAYER_OBJ                   // Object layer
)

func isObjOverBackground(
	object *ppu.ObjectData,
	currentPixel RenderedPixel,
	objectPriorityMode ppu.ObjectPriorityMode,
	cgbMasterBgPriorityEnabled bool,
) bool {
	switch objectPriorityMode {
	case ppu.ObjectPriorityModeCGB:
		if currentPixel.ColorID == ppu.COLOR_ID_WHITE { // BG is color 0
			return true
		}

		// BG master priority isn't set
		if !cgbMasterBgPriorityEnabled {
			return true
		}

		// BG doesn't have priority (CGB) AND OBJ has priority over BG
		if currentPixel.Layer != PIXEL_LAYER_BGP && !object.Attributes.BGPriority {
			return true
		}
	case ppu.ObjectPriorityModeDMG:
		if currentPixel.ColorID == ppu.COLOR_ID_WHITE { // BG is color 0
			return true
		}

		// OBJ has priority over BG
		if !object.Attributes.BGPriority {
			return true
		}
	}

	return false
}
