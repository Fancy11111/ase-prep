package main

import (
	"ase-prep/communication"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"math"
	"os"
)

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type TestCase struct {
	Obstacle struct {
		Line   float64 `json:"line"`
		PointA Point   `json:"pointA"`
		PointB Point   `json:"pointB"`
	} `json:"obstacle"`
	Targets []Point `json:"targets"`
}

type Solution struct {
	AccessiblePoints []Point `json:"accessiblePoints"`
}

func (p Point) atan2() float64 {
	return math.Atan2(p.Y, p.X)
}

func solveTestcase(testCase TestCase) (Solution, error) {
	atanA := testCase.Obstacle.PointA.atan2()
	atanB := testCase.Obstacle.PointB.atan2()
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
	}, nil
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	client := communication.NewClient("http://localhost:3030", "test", "12019861")

	token, err := client.GetToken()
	if err != nil {
		log.Error().Err(err).Msg("Could not get token")
		os.Exit(-1)
	}

	log.Info().Str("token", token).Msg("Successfully got token")

	testcase := TestCase{}

	testcaseParams := communication.TestCaseParams{
		Stage:    "1",
		Token:    token,
		Testcase: "1",
	}

	_, err = client.GetTestCase(testcaseParams, &testcase)
	if err != nil {
		log.Error().Err(err).Msg("Could not get testcase")
		os.Exit(-1)
	}

	log.Info().
		Any("testcase", testcase).
		Msg("Successfully got testcase")

	solution, err := solveTestcase(testcase)
	if err != nil {
		log.Error().Err(err).Msg("Error while trying to solve testcase")
	}

	resp, err := client.SubmitSolution(testcaseParams, solution)
	if err != nil {
		log.Error().
			Err(err).
			Any("testcaseParams", testcaseParams).
			Msg("Could not submit testcases")
		os.Exit(-1)
	}

	log.Info().
		Any("solution", solution).
		Any("response", resp).
		Msg("Submitted solution")

}
