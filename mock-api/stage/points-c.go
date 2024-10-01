package stage

import (
	"math"
	"sort"

	"github.com/rs/zerolog/log"
)

type StagePointsC struct {
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Obstacle struct {
	Line   float64 `json:"line"`
	PointA Point   `json:"pointA"`
	PointB Point   `json:"pointB"`
}

type TestCase struct {
	Obstacle Obstacle `json:"obstacle"`
	Targets  []Point  `json:"targets"`
}

type Solution struct {
	AccessiblePoints []Point `json:"accessiblePoints"`
}

func (p Point) Atan2() float64 {
	return math.Atan2(p.Y, p.X)
}

func NewStagePointC() Stage[TestCase, Solution] {
	return StagePointsC{}
}

func solveTestcase(testCase TestCase) Solution {
	atanA := testCase.Obstacle.PointA.Atan2()
	atanB := testCase.Obstacle.PointB.Atan2()
	log.Info().
		Float64("atan2-pointA", atanA).
		Float64("atan2-pointB", atanB).
		Msg("calculated ata2 for pointA")

	min, max := math.Min(atanA, atanB), math.Max(atanB, atanA)
	absLine := math.Abs(testCase.Obstacle.Line)

	accessiblePoints := make([]Point, 0)

	for _, el := range testCase.Targets {
		elAtan := math.Atan2(el.Y, el.X)
		if elAtan < min || elAtan > max || math.Abs(el.Y) < absLine {
			accessiblePoints = append(accessiblePoints, el)
		}
	}

	return Solution{
		AccessiblePoints: accessiblePoints,
	}
}

func (s StagePointsC) CreateTestcase(token string, nr int) TestCase {
	randGen := RandFromTokenAndTestcase(token, nr)
	nF := math.Pow(3, float64(nr))
	n := int(nF)
	targets := make([]Point, n)
	for i := 0; i < n; i++ {
		targets[i] = Point{
			X: randGen.Float64()*(nF+100.0) - (nF/2 + 50),
			Y: randGen.Float64()*(nF+100.0) - (nF/2 + 50),
		}
	}

	line := randGen.Float64()*(nF+75.0) - (nF/2 + 75.0/2)
	yA := randGen.Float64()*(nF+100.0) - (nF/2 + 50)
	yB := randGen.Float64()*(nF+100.0) - (nF/2 + 50)

	if math.Signbit(line) != math.Signbit(yA) {
		yA = yA * -1
	}

	if math.Signbit(line) != math.Signbit(yB) {
		yB = yB * -1
	}
	return TestCase{
		Obstacle: Obstacle{
			Line: line,
			PointA: Point{
				X: randGen.Float64()*(nF+100.0) - (nF/2 + 50),
				Y: yA,
			},
			PointB: Point{
				X: randGen.Float64()*(nF+100.0) - (nF/2 + 50),
				Y: yB,
			},
		},
		Targets: targets,
	}
}

func (s StagePointsC) GetSolution(token string, nr int) Solution {
	testcase := s.CreateTestcase(token, nr)
	return solveTestcase(testcase)
}

func (s StagePointsC) ValidateSolution(token string, nr int, solution Solution) bool {
	validSolution := s.GetSolution(token, nr)

	if len(solution.AccessiblePoints) != len(validSolution.AccessiblePoints) {
		return false
	}

	sort.Slice(solution.AccessiblePoints, func(i, j int) bool {
		a, b := solution.AccessiblePoints[i], solution.AccessiblePoints[j]
		return a.X < b.X || (a.X == b.X && a.Y <= b.Y)
	})

	sort.Slice(validSolution.AccessiblePoints, func(i, j int) bool {
		a, b := validSolution.AccessiblePoints[i], validSolution.AccessiblePoints[j]
		return a.X < b.X || (a.X == b.X && a.Y <= b.Y)
	})

	for i, el := range validSolution.AccessiblePoints {
		if el != solution.AccessiblePoints[i] {
		}
	}

	return false
}
