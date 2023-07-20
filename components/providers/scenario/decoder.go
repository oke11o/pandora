package scenario

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func decodeAmmo(cfg AmmoConfig) ([]*Ammo, error) {
	reqRegistry := map[string]Request{}

	for _, req := range cfg.Requests {
		reqRegistry[req.Name] = req
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

	div := gcdm(weights...)
	names := make(map[string]int)
	total := 0
	for _, sc := range input {
		cnt := int(sc.Weight / div)
		total += cnt
		names[sc.Name] = cnt
	}
	return names, total
}

func gcd(a, b int64) int64 {
	for a > 0 && b > 0 {
		if a >= b {
			a = a % b
		} else {
			b = b % a
		}
	}
	if a > b {
		return a
	}
	return b
}

func gcdm(weights ...int64) int64 {
	l := len(weights)
	if l < 2 {
		return 0
	}
	res := gcd(weights[l-2], weights[l-1])
	if l == 2 {
		return res
	}
	return gcd(gcdm(weights[:l-1]...), res)
}

func lcm(a, b int64) int64 {
	return (a * b) / gcd(a, b)
}

func lcmm(a ...int64) int64 {
	l := len(a)
	if l < 2 {
		return 0
	}
	res := lcm(a[l-2], a[l-1])
	if l == 2 {
		return res
	}
	return lcm(lcmm(a[:l-1]...), res)

}
