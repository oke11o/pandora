package scenario

import (
	"fmt"
	"strconv"
	"strings"
)

type Preprocessor struct {
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

			mval, isMval := value.([]map[string]string)
			if isMval && index >= 0 && index < len(mval) {
				vval := mval[index]
				currentData = make(map[string]any, len(vval))
				for k, v := range vval {
					currentData[k] = v
				}
				continue
			}

			mapSlice, isMapSlice := value.([]map[string]any)
			if !isMapSlice {
				anySlice, isAnySlice := value.([]any)
				if isAnySlice {
					if index < 0 || index >= len(anySlice) {
						return nil, fmt.Errorf("invalid index %d for segment %s in path  %s", index, segment, path)
					}
					if i != len(segments)-1 {
						return nil, fmt.Errorf("not last segment %s in path %s", segment, path)
					}
					return anySlice[index], nil
				}
				return nil, fmt.Errorf("invalid type of segment %s in path %s", segment, path)
			}
			if index < 0 || index >= len(mapSlice) {
				return nil, fmt.Errorf("invalid path : %s", path)
			}

			currentData = mapSlice[index]
		} else {
			value, exists := currentData[segment]
			if !exists {
				return nil, fmt.Errorf("segment %s not found in path %s", segment, path)
			}
			var ok bool
			currentData, ok = value.(map[string]any)
			if !ok {
				if i != len(segments)-1 {
					return nil, fmt.Errorf("not last segment %s in path %s", segment, path)
				}
				return value, nil
			}
		}
	}

	return currentData, nil
}
