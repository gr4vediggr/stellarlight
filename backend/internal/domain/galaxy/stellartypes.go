package galaxy

import (
	"fmt"
	"image/color"
	"strings"
)

type StarType struct {
	ID           uint32  `json:"id"`           // Unique identifier for the star type
	Name         string  `json:"name"`         // Name of the star type
	Color        Color   `json:"color"`        // Color of the star type
	Description  string  `json:"description"`  // Description of the star type
	MinSize      float64 `json:"minSize"`      // Minimum size of the star
	MaxSize      float64 `json:"maxSize"`      // Maximum size of the star
	Chance       float64 `json:"chance"`       // Chance of this star type appearing
	PlanetChance float64 `json:"planetChance"` // Chance of planets orbiting this star type
	MinPlanets   int     `json:"minPlanets"`   // Minimum number of planets that can orbit this star type
	MaxPlanets   int     `json:"maxPlanets"`   // Maximum number of planets that can orbit this star type
}

func (st *StarType) GetChoiceWeight() float64 {
	return st.Chance
}

type PlanetType struct {
	ID          uint32  `json:"id"`          // Unique identifier for the planet type
	Name        string  `json:"name"`        // Name of the planet type
	Color       Color   `json:"color"`       // Color of the planet type
	Description string  `json:"description"` // Description of the planet type
	MinSize     float64 `json:"minSize"`     // Minimum size of the planet
	MaxSize     float64 `json:"maxSize"`     // Maximum size of the planet
	Chance      float64 `json:"chance"`      // Chance of this planet type appearing
	MoonChance  float64 `json:"moonChance"`  // Chance of moons orbiting this planet type
	MaxMoons    int     `json:"maxMoons"`    // Maximum number of moons that can orbit this planet type
}

func (pt *PlanetType) GetChoiceWeight() float64 {
	return pt.Chance
}

// Color type that marshals/unmarshals RGBA to/from hex string.
type Color struct {
	R, G, B, A uint8
}

// MarshalJSON encodes Color as "#RRGGBBAA"
func (c Color) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"#%02X%02X%02X%02X\"", c.R, c.G, c.B, c.A)), nil
}

// UnmarshalJSON decodes Color from "#RRGGBBAA" or "#RRGGBB"
func (c *Color) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")
	s = strings.TrimPrefix(s, "#")
	if len(s) == 6 {
		s += "FF"
	}
	if len(s) != 8 {
		return fmt.Errorf("invalid color hex: %s", s)
	}
	var r, g, b, a uint8
	_, err := fmt.Sscanf(s, "%02X%02X%02X%02X", &r, &g, &b, &a)
	if err != nil {
		return err
	}
	c.R, c.G, c.B, c.A = r, g, b, a
	return nil
}

// ToGGColor converts to gg.Color
func (c Color) ToGGColor() color.Color {
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}
