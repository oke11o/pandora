package scenario

import (
	"fmt"
	"strconv"
	"strings"
)

type Preprocessor struct {
	Name      string
	Variables map[string]string
}

func (p *Preprocessor) Process(reqMap map[string]any) error {
	for k, v := range p.Variables {
		val, err := p.getValue(reqMap, v)
		if err != nil {
			return fmt.Errorf("failed to get value for %s: %w", k, err)
		}
		err = p.setValue(reqMap, k, val)
		if err != nil {
			return fmt.Errorf("failed to set value for %s: %w", k, err)
		}
	}
	return nil
}

func (p *Preprocessor) setValue(reqMap map[string]any, k string, v any) error {
	target, exists := reqMap["preprocessor"]
	if !exists {
		reqMap["preprocessor"] = map[string]any{k: v}
		return nil
	}
	reqTarget, isMap := target.(map[string]any)
	if !isMap {
		return fmt.Errorf("preprocessor is not a map")
	}
	reqTarget[k] = v

	return nil
}

func (p *Preprocessor) getValue(reqMap map[string]any, path string) (any, error) {
	segments := strings.Split(path, ".")

	currentData := reqMap
	for i, segment := range segments {
		segment = strings.TrimSpace(segment)
		if strings.Contains(segment, "[") && strings.HasSuffix(segment, "]") {
			openBraceIdx := strings.Index(segment, "[")
			indexStr := segment[openBraceIdx+1 : len(segment)-1]
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, fmt.Errorf("invalid index: %s", indexStr)
			}

			segment = segment[:openBraceIdx]
			value, exists := currentData[segment]
			if !exists {
				return nil, fmt.Errorf("path not found: %s", path)
			}

			slice, isSlice := value.([]map[string]any)
			if !isSlice || index < 0 || index >= len(slice) {
				anySlice, isAnySlice := value.([]any)
				if isAnySlice || index < 0 || index >= len(anySlice) {
					if i != len(segments)-1 {
						return nil, fmt.Errorf("invalid index for slice: %s", segment)
					}
					return anySlice[index], nil
				}
				return nil, fmt.Errorf("invalid index for slice: %s", segment)
			}

			currentData = slice[index]
		} else {
			value, exists := currentData[segment]
			if !exists {
				return nil, fmt.Errorf("path not found: %s", path)
			}
			var ok bool
			currentData, ok = value.(map[string]any)
			if !ok {
				if i != len(segments)-1 {
					return nil, fmt.Errorf("invalid index for slice: %s", segment)
				}
				return value, nil
			}
		}
	}

	return currentData, nil
}
