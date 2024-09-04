package communication

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

type client struct {
	scenario string
	mnr      string
	baseUrl  string
	errored  bool
}

func NewClient(baseUrl string, scenario string, mnr string) *client {
	return &client{mnr: mnr, scenario: scenario, baseUrl: fmt.Sprintf("%s/%s/assignment/%s", baseUrl, scenario, mnr)}
}

func (c *client) GetToken() (string, error) {
	requestURL := fmt.Sprintf("%s/token", c.baseUrl)
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		log.Error().Err(err).Msg("Could not create token request")
		return "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Could not execute token request")
		return "", err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error().Err(err).Msg("Could not read token response")
	}
	return string(body), nil
}

type TestCaseParams struct {
	Stage    string
	Testcase string
	Token    string
}

func (c *client) GetTestCase(params TestCaseParams, v any) (string, error) {
	requestURL := fmt.Sprintf("%s/stage/%s/testcase/%s?token=%s", c.baseUrl, params.Stage, params.Testcase, params.Token)
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		log.Error().Err(err).Msg("Could not create token request")
		return "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Could not execute token request")
		return "", err
	}

	defer res.Body.Close()
	/*body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error().Err(err).Msg("Could not read token response")
	}*/

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(v)
	if err != nil {
		log.Error().Err(err).Any("body", res.Body).Msg("Could not decode test case json")
	}

	return "", nil
}

type SolutionResult struct {
	Message        string `json:"message"`
	LinkToNextTask string `json:"linkToNextTask"`
}

func (c *client) SubmitSolution(params TestCaseParams, solution any) (SolutionResult, error) {

	requestURL := fmt.Sprintf("%s/stage/%s/testcase/%s?token=%s", c.baseUrl, params.Stage, params.Testcase, params.Token)

	reqBody, err := json.Marshal(solution)

	if err != nil {
		log.Error().Err(err).Any("solution", solution).Msg("Could not serialize solution")
	}

	//reqBody :=

	req, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(reqBody))
	if err != nil {
		log.Error().Err(err).Msg("Could not create token request")
		return SolutionResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Could not execute token request")
		return SolutionResult{}, err
	}

	defer res.Body.Close()
	result := SolutionResult{}

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&result)
	if err != nil {
		log.Error().Err(err).Any("body", res.Body).Msg("Could not decode submition response")
		return result, err
	}
	return result, nil
}

func (c *client) Finish() {

}
