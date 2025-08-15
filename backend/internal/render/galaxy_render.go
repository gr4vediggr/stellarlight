package render

import (
	"math"

	"github.com/fogleman/gg"
	"github.com/gr4vediggr/stellarlight/internal/domain/galaxy"
)

func RenderGalaxyToImage(galaxy *galaxy.Galaxy, out string) {
	// compute point bounds for

	type Point struct {
		X, Y float64
	}

	points := make([]Point, 0, len(galaxy.StarSystems))

	for _, system := range galaxy.StarSystems {
		points = append(points, Point{X: system.LocationX, Y: system.LocationY})
	}

	min := points[0]
	max := points[0]
	for _, p := range points {
		min.X = math.Min(min.X, p.X)
		min.Y = math.Min(min.Y, p.Y)
		max.X = math.Max(max.X, p.X)
		max.Y = math.Max(max.Y, p.Y)
	}

	W, H := 2048, 2048 // Image dimensions

	size := Point{X: max.X - min.X, Y: max.Y - min.Y}
	center := Point{X: min.X + size.X/2, Y: min.Y + size.Y/2}
	scale := math.Min(float64(W)/size.X, float64(H)/size.Y) * 0.9

	// render points and edges
	dc := gg.NewContext(W, H)
	dc.SetHexColor("#301934")

	dc.Clear()

	dc.Translate(float64(W)/2, float64(H)/2)
	dc.Scale(scale, scale)
	dc.Translate(-center.X, -center.Y)

	dc.SetRGB(0.7, 0.7, 0.7)
	dc.SetLineWidth(1)
	for _, system := range galaxy.StarSystems {
		for _, connectedID := range system.ConnectedSystems {
			connectedSystem := galaxy.StarSystems[connectedID]
			dc.DrawLine(system.LocationX, system.LocationY, connectedSystem.LocationX, connectedSystem.LocationY)

		}
	}

	dc.Stroke()
	for _, system := range galaxy.StarSystems {
		dc.SetColor(system.Stars[0].Type.Color.ToGGColor())
		dc.DrawPoint(system.LocationX, system.LocationY, 3)
		dc.Fill()
	}
	dc.ClosePath()

	dc.SavePNG(out)

}
