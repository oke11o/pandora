package scenario

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/yandex/pandora/lib/math"
)

func decodeAmmo(cfg AmmoConfig) ([]*Ammo, error) {
	reqRegistry := make(map[string]Request, len(cfg.Requests))
	paramRegistry := make(map[string][]string, len(cfg.Requests))

	for _, req := range cfg.Requests {
		reqRegistry[req.Name] = req
		paramRegistry[req.Name] = extractParams(req)
	}

	scenarioRegistry := map[string]Scenario{}
	for _, sc := range cfg.Scenarios {
		scenarioRegistry[sc.Name] = sc
	}

	names, size := spreadNames(cfg.Scenarios)
	result := make([]*Ammo, 0, size)
	for _, sc := range cfg.Scenarios {
		a, err := convertScenarioToAmmo(sc, reqRegistry)
		if err != nil {
			return nil, fmt.Errorf("failed to convert scenario %s: %w", sc.Name, err)
		}
		for i := 0; i < names[sc.Name]; i++ {
			result = append(result, a)
		}
	}

	return result, nil
}

var extractParamsRegex = regexp.MustCompile("{{.+?}}")

/*
*
Preprocessors  []Preprocessor    `yaml:"preprocessors"`
Postprocessors []Postprocessor   `yaml:"postprocessors"`
outputParams   []string
*/
func extractParams(req Request) []string {
	resUri := extractParamsRegex.FindAllString(req.Uri, -1)
	var resBody []string
	if req.Body != nil {
		resBody = extractParamsRegex.FindAllString(*req.Body, -1)
	}
	var headerRes []string
	for key, val := range req.Headers {
		ks := extractParamsRegex.FindAllString(key, -1)
		vs := extractParamsRegex.FindAllString(val, -1)
		headerRes = append(headerRes, ks...)
		headerRes = append(headerRes, vs...)
	}
	result := make([]string, 0, len(resUri)+len(resBody))
	result = append(result, resUri...)
	if len(resBody) > 0 {
		result = append(result, resBody...)
	}
	if len(headerRes) > 0 {
		result = append(result, headerRes...)
	}
	for i := range result {
		result[i] = strings.TrimSpace(result[i][2 : len(result[i])-2])
	}
	// TODO: remove duplicates
	return result
}

func convertScenarioToAmmo(sc Scenario, reqs map[string]Request) (*Ammo, error) {
	result := &Ammo{name: sc.Name}
	for _, sh := range sc.Shoot {
		name, cnt, err := parseShootName(sh)
		if err != nil {
			return nil, fmt.Errorf("failed to parse shoot %s: %w", sh, err)
		}
		if req, ok := reqs[name]; ok {
			for i := 0; i < cnt; i++ {
				result.Requests = append(result.Requests, req)
			}
		}
	}

	return result, nil
}

func parseShootName(shoot string) (string, int, error) {
	name, args, err := parseStringFunc(shoot)
	if err != nil {
		return "", 0, err
	}
	cnt := 1
	if len(args) > 0 && args[0] != "" {
		cnt, err = strconv.Atoi(args[0])
		if err != nil {
			return "", 0, fmt.Errorf("failed to parse count: %w", err)
		}
	}
	return name, cnt, nil
}

func parseStringFunc(shoot string) (string, []string, error) {
	openIdx := strings.IndexRune(shoot, '(')
	if openIdx == -1 {
		return shoot, nil, nil
	}
	name := strings.TrimSpace(shoot[:openIdx])

	arg := strings.TrimSpace(shoot[openIdx+1:])
	closeIdx := strings.IndexRune(arg, ')')
	if closeIdx != len(arg)-1 {
		return name, nil, errors.New("invalid close bracket position")
	}
	arg = strings.TrimSpace(arg[:closeIdx])
	args := strings.Split(arg, ",")
	return name, args, nil
}

func spreadNames(input []Scenario) (map[string]int, int) {
	if len(input) == 0 {
		return nil, 0
	}
	if len(input) == 1 {
		return map[string]int{input[0].Name: 1}, 1
	}

	scenarioRegistry := map[string]Scenario{}
	weights := make([]int64, len(input))
	for i, sc := range input {
		scenarioRegistry[sc.Name] = sc
		weights[i] = sc.Weight
	}

	div := math.GCDM(weights...)
	names := make(map[string]int)
	total := 0
	for _, sc := range input {
		cnt := int(sc.Weight / div)
		total += cnt
		names[sc.Name] = cnt
	}
	return names, total
}
