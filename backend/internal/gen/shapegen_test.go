package gen_test

import (
	"math"
	"testing"

	"github.com/gr4vediggr/stellarlight/internal/gen"
)

func TestGenerateSpiralgenPoints(t *testing.T) {
	points := gen.GenerateSpiralGalaxyPoints(1000, 2, 0.25, 0.06, 2, 0.01)
	if len(points) != 2000 {
		t.Errorf("Expected 2000 points, got %d", len(points))
	}
	for _, p := range points {
		if math.IsNaN(p.X) || math.IsNaN(p.Y) {
			t.Error("Generated point is NaN")
		}
	}

	// Optionally, save the points to an image for visual inspection
	img := gen.DrawPoints(points, 1024, 1024)

	if err := gen.SaveImage(img, "spiral_gen.png"); err != nil {
		t.Errorf("Failed to save image: %v", err)
	}
	triangulate := gen.Triangulate(points)

	tf := gen.CreateFromDelaunay(triangulate)

	remaining := tf.RemoveLongEdgesConnected(0.15)
	tf.Edges = remaining

	tf.DrawEdges("spiral_gen_edges.png")

	tf.ReduceHighDegreeEdges(5, 0.1, 20)
	tf.DrawEdges("spiral_gen_reduced_edges.png")
	t.Error("Test completed, check spiral_gen.png for visual output")

}

func TestGenerateSpiralgenPointsInterarm(t *testing.T) {
	points := gen.GenerateSpiralGalaxyPointsWithInterarm(2000, 2, 0.4, 0.06, 2, 0.01, 0.02)
	if len(points) != 2000 {
		t.Errorf("Expected 2000 points, got %d", len(points))
	}
	for _, p := range points {
		if math.IsNaN(p.X) || math.IsNaN(p.Y) {
			t.Error("Generated point is NaN")
		}
	}

	// Optionally, save the points to an image for visual inspection
	img := gen.DrawPoints(points, 1024, 1024)

	if err := gen.SaveImage(img, "spiral_gen.png"); err != nil {
		t.Errorf("Failed to save image: %v", err)
	}
	triangulate := gen.Triangulate(points)

	tf := gen.CreateFromDelaunay(triangulate)

	remaining := tf.RemoveLongEdgesConnected(0.2)
	tf.Edges = remaining

	tf.DrawEdges("spiral_gen_edges.png")

	tf.ReduceHighDegreeEdges(5, 0.25, 20)
	tf.DrawEdges("spiral_gen_reduced_edges.png")
	t.Error("Test completed, check spiral_gen.png for visual output")

}
