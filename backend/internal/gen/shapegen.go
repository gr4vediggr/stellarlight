package gen

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"
	"math/rand"
	"os"
	"sort"

	"github.com/fogleman/delaunay"

	"github.com/fogleman/gg"
)

type Point struct {
	X float64 // X coordinate
	Y float64 // Y coordinate
}

func GenerateSpiralPoints(numPoints int) []Point {
	points := make([]Point, numPoints)
	for i := 0; i < numPoints; i++ {
		angle := float64(i) * 2 * math.Pi / float64(numPoints)
		radius := float64(i) * 0.1 // Adjust the spiral density
		x := radius * math.Cos(angle)
		y := radius * math.Sin(angle)
		points[i] = Point{X: x, Y: y}
	}
	return points
}

func GenerateSpiralGalaxyPoints(numPoints, numArms int, armSpread, minRadius, twist, minDist float64) []Point {
	points := make([]Point, 0, numPoints)
	attempts := 0
	maxAttempts := numPoints * 200 // Prevent infinite loops
	for len(points) < numPoints && attempts < maxAttempts {
		i := len(points)
		attempts++
		arm := i % numArms

		frac := (minRadius + float64(i)/float64(numPoints)) / (1 + minRadius)
		biasedFrac := math.Pow(frac, 1)

		// Each arm makes about one full rotation

		armFactor := float64(arm) / float64(numArms) * 2 * math.Pi

		angle := biasedFrac*twist*math.Pi + armFactor
		// Bias radius towards center using sqrt, but start at minRadius
		radius := (minRadius + biasedFrac)
		// More jitter near the center for a core effect
		jitterFactor := (1.0 - biasedFrac) // high near center, low at edge
		jitterAngle := angle + (rand.Float64()-0.5)*2*math.Pi/float64(numArms)*(1-frac)
		jitterRadius := radius + (rand.Float64()-0.5)*armSpread*(jitterFactor+0.1)*radius
		x := jitterRadius * math.Cos(jitterAngle)
		y := jitterRadius * math.Sin(jitterAngle)
		tooClose := false
		for _, p := range points {
			dx := x - p.X
			dy := y - p.Y
			if dx*dx+dy*dy < minDist*minDist {
				tooClose = true
				break
			}
		}
		if !tooClose {
			points = append(points, Point{X: x, Y: y})
		}
	}
	return points
}

// GenerateSpiralGalaxyPointsPoisson generates spiral galaxy points using random sampling and a minimum distance constraint.
// numPoints: number of points to generate
// numArms: number of spiral arms
// armSpread: spread of the arms
// minRadius: minimum radius for the core
// minDist: minimum allowed distance between points
func GenerateSpiralGalaxyPointsPoisson(numPoints, numArms int, armSpread, minRadius, minDist, twist float64) []Point {
	points := make([]Point, 0, numPoints)
	attempts := 0
	maxAttempts := numPoints * 20 // Prevent infinite loops

	for len(points) < numPoints && attempts < maxAttempts {
		attempts++
		// Sample r with bias towards center
		r := minRadius + math.Pow(rand.Float64(), 1.5)*(1-minRadius)
		theta := rand.Float64() * 2 * math.Pi

		// Nudge strength depends on r: weaker near core, stronger at edge
		nudgeStrength := math.Min(1, (r-minRadius)/(1-minRadius)) // r ∈ [minRadius, 1], so nudge is weaker near core

		armIdx := int(theta / (2 * math.Pi / float64(numArms)))
		armBase := float64(armIdx) * 2 * math.Pi / float64(numArms)
		spiralTwist := twist * math.Pi
		targetTheta := armBase + spiralTwist*r
		theta = theta*(1-nudgeStrength) + targetTheta*nudgeStrength

		armAngle := float64(numArms) * theta
		armOffset := armSpread * math.Sin(armAngle)
		finalTheta := theta + armOffset

		x := r * math.Cos(finalTheta)
		y := r * math.Sin(finalTheta)

		tooClose := false
		for _, p := range points {
			dx := x - p.X
			dy := y - p.Y
			if dx*dx+dy*dy < minDist*minDist {
				tooClose = true
				break
			}
		}
		if !tooClose {
			points = append(points, Point{X: x, Y: y})
		}
	}
	return points
}

// GenerateSpiralGalaxyPointsWithInterarm generates spiral galaxy points with a small chance to place a point between arms.
// numPoints: number of points to generate
// numArms: number of spiral arms
// armSpread: spread of the arms
// minRadius: minimum radius for the core
// twist: spiral twist
// minDist: minimum allowed distance between points
// interarmChance: probability to place a point between arms (e.g. 0.01 for 1%)
func GenerateSpiralGalaxyPointsWithInterarm(numPoints, numArms int, armSpread, minRadius, twist, minDist, interarmChance float64) []Point {
	points := make([]Point, 0, numPoints)
	attempts := 0
	maxAttempts := numPoints * 200 // Prevent infinite loops
	for len(points) < numPoints && attempts < maxAttempts {
		i := len(points)
		attempts++
		var arm int
		var angle float64
		var radius float64

		frac := (minRadius + float64(i)/float64(numPoints)) / (1 + minRadius)
		biasedFrac := math.Pow(frac, 1.5)

		// Decide if this point is interarm
		if rand.Float64() < interarmChance && biasedFrac > 0.5 {
			// Pick a random angle between arms
			arm = rand.Intn(numArms)
			armOffset := float64(arm) / float64(numArms) * 2 * math.Pi
			// Offset by half an arm to be between arms
			angle = biasedFrac*twist*math.Pi + armOffset + math.Pi/float64(numArms)
			radius = minRadius + biasedFrac
		} else {
			arm = i % numArms
			armOffset := float64(arm) / float64(numArms) * 2 * math.Pi
			angle = biasedFrac*twist*math.Pi + armOffset
			radius = minRadius + biasedFrac
		}

		jitterFactor := (1.0 - biasedFrac)
		jitterAngle := angle + (rand.Float64()-0.5)*2*math.Pi/float64(numArms)*(jitterFactor)
		jitterRadius := radius + (rand.Float64()-0.5)*armSpread*(jitterFactor+0.1)*radius
		x := jitterRadius * math.Cos(jitterAngle)
		y := jitterRadius * math.Sin(jitterAngle)

		tooClose := false
		for _, p := range points {
			dx := x - p.X
			dy := y - p.Y
			if dx*dx+dy*dy < minDist*minDist {
				tooClose = true
				break
			}
		}
		if !tooClose {
			points = append(points, Point{X: x, Y: y})
		}
	}
	return points
}

// SmoothPointsSpatial averages each point with its spatial neighbors within a given radius.
// This produces a more natural smoothing effect.
func SmoothPointsSpatial(points []Point, radius float64, passes int) []Point {
	smoothed := make([]Point, len(points))
	copy(smoothed, points)
	for p := 0; p < passes; p++ {
		next := make([]Point, len(smoothed))
		for i, pt := range smoothed {
			sumX, sumY, count := pt.X, pt.Y, 1
			for j, other := range smoothed {
				if i == j {
					continue
				}
				dx := pt.X - other.X
				dy := pt.Y - other.Y
				if dx*dx+dy*dy <= radius*radius {
					sumX += other.X
					sumY += other.Y
					count++
				}
			}
			next[i].X = sumX / float64(count)
			next[i].Y = sumY / float64(count)
		}
		copy(smoothed, next)
	}
	return smoothed
}

func DrawPoints(points []Point, width, height int) image.Image {

	// Find minimum and maximum coordinates to scale points

	minX, minY := points[0].X, points[0].Y
	maxX, maxY := points[0].X, points[0].Y
	for _, p := range points {
		if p.X < minX {
			minX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill background with black
	draw.Draw(img, img.Bounds(), &image.Uniform{C: image.Black}, image.Point{}, draw.Src)
	for _, p := range points {

		x := int((p.X - minX) / (maxX - minX) * float64(width))
		y := int((p.Y - minY) / (maxY - minY) * float64(height))

		if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(x, y, image.White) // Draw point in white
		} else {
			fmt.Printf("Point (%f, %f) out of bounds\n", p.X, p.Y)
		}
	}
	return img
}

func SaveImage(img image.Image, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create image file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}
	return nil
}

func Triangulate(points []Point) *delaunay.Triangulation {

	dpoints := make([]delaunay.Point, len(points))
	for i, p := range points {
		dpoints[i].X = p.X
		dpoints[i].Y = p.Y
	}

	triangulation, err := delaunay.Triangulate(dpoints)
	if err != nil {
		fmt.Printf("Failed to triangulate points: %v\n", err)
	}

	return triangulation
}

type Edge struct {
	A, B int // Indices of the points in the triangulation
}

func (e Edge) Equal(other Edge) bool {
	return (e.A == other.A && e.B == other.B) || (e.A == other.B && e.B == other.A)
}

type Triangulation struct {
	Points []Point
	Edges  []Edge
	Adj    map[int]map[int]struct{} // adjacency list for fast lookup
}

// BuildAdjacencyList builds the adjacency list for the triangulation.
func (t *Triangulation) BuildAdjacencyList() {
	n := len(t.Points)
	adj := make(map[int]map[int]struct{}, n)
	for i := 0; i < n; i++ {
		adj[i] = make(map[int]struct{})
	}
	for _, e := range t.Edges {
		adj[e.A][e.B] = struct{}{}
		adj[e.B][e.A] = struct{}{}
	}
	t.Adj = adj
}

// IsConnected checks if the graph is fully connected using BFS.
func (t *Triangulation) IsConnected() bool {
	n := len(t.Points)
	if n == 0 || t.Adj == nil {
		return false
	}
	visited := make(map[int]bool, n)
	queue := []int{0}
	visited[0] = true
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		for neighbor := range t.Adj[node] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}
	return len(visited) == n
}

// RemoveEdge removes an edge from the triangulation and updates adjacency.
func (t *Triangulation) RemoveEdge(idx int) {
	e := t.Edges[idx]
	if t.Adj[e.A] != nil {
		delete(t.Adj[e.A], e.B)
	}
	if t.Adj[e.B] != nil {
		delete(t.Adj[e.B], e.A)
	}
	t.Edges = append(t.Edges[:idx], t.Edges[idx+1:]...)
	// Rebuild adjacency to ensure consistency after removal
	t.BuildAdjacencyList()
}

// tryRemoveEdge attempts to remove an edge and restores adjacency if not connected.
// No backup or deep copy needed, just remove and re-add adjacency if needed.
func (t *Triangulation) tryRemoveEdge(removeIdx int) bool {
	e := t.Edges[removeIdx]
	// Remove adjacency
	delete(t.Adj[e.A], e.B)
	delete(t.Adj[e.B], e.A)
	// Check connectivity
	if t.IsConnected() {
		// Remove edge from slice and rebuild adjacency
		t.Edges = append(t.Edges[:removeIdx], t.Edges[removeIdx+1:]...)
		t.BuildAdjacencyList()
		return true
	}
	// Restore adjacency if not connected
	t.Adj[e.A][e.B] = struct{}{}
	t.Adj[e.B][e.A] = struct{}{}
	return false
}

// CreateFromDelaunay creates a Triangulation from a Delaunay triangulation.
func CreateFromDelaunay(t *delaunay.Triangulation) *Triangulation {

	tri := &Triangulation{
		Points: make([]Point, len(t.Points)),
		Edges:  make([]Edge, 0, len(t.Triangles)/3),
	}

	for i, p := range t.Points {
		tri.Points[i] = Point{X: p.X, Y: p.Y}
	}

	edges := Edges(t)
	for _, edge := range edges {
		for _, existing := range tri.Edges {
			if (edge.A == existing.A && edge.B == existing.B) || (edge.A == existing.B && edge.B == existing.A) {
				continue
			}
		}
		tri.Edges = append(tri.Edges, edge)
	}

	return tri
}

// Edges extracts the edges from a Delaunay triangulation.
func Edges(t *delaunay.Triangulation) []Edge {
	edges := make([]Edge, 0, len(t.Triangles)/3)
	for i := 0; i < len(t.Triangles); i += 3 {
		a, b, c := t.Triangles[i], t.Triangles[i+1], t.Triangles[i+2]
		edges = append(edges, Edge{A: a, B: b}, Edge{A: b, B: c}, Edge{A: c, B: a})
	}
	return edges
}

// DrawEdges draws the edges of the triangulation to a PNG file.
func (t *Triangulation) DrawEdges(name string) {
	W, H := 2048, 2048 // Image dimensions

	// Compute point bounds for rendering
	min := t.Points[0]
	max := t.Points[0]
	for _, p := range t.Points {
		min.X = math.Min(min.X, p.X)
		min.Y = math.Min(min.Y, p.Y)
		max.X = math.Max(max.X, p.X)
		max.Y = math.Max(max.Y, p.Y)
	}

	size := Point{X: max.X - min.X, Y: max.Y - min.Y}
	center := Point{X: min.X + size.X/2, Y: min.Y + size.Y/2}
	scale := math.Min(float64(W)/size.X, float64(H)/size.Y) * 0.9

	dc := gg.NewContext(W, H)
	dc.SetRGB(0, 0, 0)
	dc.Clear()
	dc.SetRGB(0.7, 0.7, 0.7)

	dc.Translate(float64(W)/2, float64(H)/2)
	dc.Scale(scale, scale)
	dc.Translate(-center.X, -center.Y)

	for _, edge := range t.Edges {
		p1 := t.Points[edge.A]
		p2 := t.Points[edge.B]
		dc.DrawLine(p1.X, p1.Y, p2.X, p2.Y)
	}
	dc.Stroke()
	dc.SetRGB(1, 1, 1)

	for _, p := range t.Points {
		dc.DrawPoint(p.X, p.Y, 5)
	}
	dc.Fill()

	dc.SavePNG(name)
}

// RemoveLongEdgesConnected removes as many edges longer than maxLength as possible, always starting from the longest.
// It keeps removing edges as long as the graph remains connected.
func (t *Triangulation) RemoveLongEdgesConnected(maxLength float64) []Edge {
	t.BuildAdjacencyList()
	type edgeWithLength struct {
		Idx    int
		Length float64
	}
	// Repeat until no more removable long edges
	for {
		edgesWithLen := make([]edgeWithLength, 0, len(t.Edges))
		for idx, edge := range t.Edges {
			p1 := t.Points[edge.A]
			p2 := t.Points[edge.B]
			dx := p2.X - p1.X
			dy := p2.Y - p1.Y
			length := math.Sqrt(dx*dx + dy*dy)
			if length > maxLength {
				edgesWithLen = append(edgesWithLen, edgeWithLength{Idx: idx, Length: length})
			}
		}
		if len(edgesWithLen) == 0 {
			break
		}
		// Sort by length descending
		sort.Slice(edgesWithLen, func(i, j int) bool {
			return edgesWithLen[i].Length > edgesWithLen[j].Length
		})
		removedAny := false
		for _, ewl := range edgesWithLen {
			// Try to remove the longest edge
			if ewl.Idx < 0 || ewl.Idx >= len(t.Edges) {
				continue
			}
			canRemove := len(t.Adj[t.Edges[ewl.Idx].A]) > 2 && len(t.Adj[t.Edges[ewl.Idx].B]) > 2

			if canRemove && t.tryRemoveEdge(ewl.Idx) {
				removedAny = true
				break // After removal, restart the process
			}
		}
		if !removedAny {
			break // No more removable edges
		}
	}
	return t.Edges
}

// ReduceHighDegreeEdges removes edges if either endpoint exceeds maxDegree, or probabilistically.
// Edges are shuffled in-place and each edge is checked only once.
// Edges are removed immediately if removal keeps the graph connected.
func (t *Triangulation) ReduceHighDegreeEdges(maxDegree int, removeChance float64, nTries int) {
	fmt.Println("Reducing high degree edges with maxDegree:", maxDegree, "removeChance:", removeChance, "nTries:", nTries)
	if len(t.Edges) == 0 {
		return
	}

	idx := 0
	for idx < len(t.Edges) {

		e := t.Edges[idx]
		degreeA := len(t.Adj[e.A])
		degreeB := len(t.Adj[e.B])
		shouldRemove := degreeA > maxDegree || degreeB > maxDegree
		if (shouldRemove) && degreeA > 1 && degreeB > 1 {
			if t.tryRemoveEdge(idx) {
				// Edge removed, so do not increment idx (next edge shifts into idx)
				continue
			}
		}
		idx++
	}
	t.BuildAdjacencyList()

	for i, adj := range t.Adj {
		if len(adj) <= maxDegree {
			continue // No need to reduce this vertex
		}

		fmt.Println("Maxlen violation vertex with degree:", len(adj), "maxDegree:", maxDegree, "vertex:", i)

	}
}

type BaseGenerationConfig struct {
	NumPoints    int     // Number of points to generate
	MaxDistance  float64 // Maximum distance between points
	MaxDegree    int     // Maximum degree of any point in the triangulation
	RemoveChance float64 // Probability of removing an edge during reduction
}

func GenerateValidGalaxy(config BaseGenerationConfig, genFunc func() []Point) (*Triangulation, error) {
	attempts := 0
	maxAttempts := 1000 // Limit attempts to prevent infinite loops
	for attempts < maxAttempts {
		attempts++
		points := genFunc()
		if len(points) < config.NumPoints {
			continue // Regenerate if not enough points
		}

		tri := Triangulate(points)
		if tri == nil {
			continue
		}

		t := CreateFromDelaunay(tri)
		t.Edges = t.RemoveLongEdgesConnected(config.MaxDistance)
		t.BuildAdjacencyList()
		if !t.IsConnected() {
			continue // Regenerate if not connected
		}

		t.ReduceHighDegreeEdges(config.MaxDegree, config.RemoveChance, 10)

		return t, nil

	}
	return nil, fmt.Errorf("failed to generate valid galaxy")

}
