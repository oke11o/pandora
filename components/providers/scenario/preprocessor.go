package scenario

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type iterator interface {
	next(segment string) int
	rand(length int) int
}

func newNextIterator(seed int64) iterator {
	return &nextIterator{
		gs:  make(map[string]*atomic.Uint64),
		rnd: rand.New(rand.NewSource(seed)),
	}
}

type nextIterator struct {
	mx  sync.Mutex
	gs  map[string]*atomic.Uint64
	rnd *rand.Rand
}

func (n *nextIterator) rand(length int) int {
	return n.rnd.Intn(length)
}

func (n *nextIterator) next(segment string) int {
	a, ok := n.gs[segment]
	if !ok {
		n.mx.Lock()
		n.gs[segment] = &atomic.Uint64{}
		n.mx.Unlock()
		return 0
	}
	add := a.Add(1)
	return int(add)
}

type Preprocessor struct {
	Variables map[string]string
	iterator  iterator
}

func (p *Preprocessor) Process(templateVars map[string]any, sourceVars map[string]any) error {
	if templateVars == nil {
		return errors.New("templateVars must not be nil")
	}
	for k, v := range p.Variables {
		val, err := p.getValue(templateVars, v)
		if err != nil {
			var pathError *errSegmentNotFound
			if !errors.As(err, &pathError) {
				return fmt.Errorf("failed to get value for %s: %w", k, err)
			}
			val, err = p.getValue(sourceVars, v)
			if err != nil {
				return fmt.Errorf("failed to get value for %s: %w", k, err)
			}
		}
		err = p.setValue(templateVars, k, val)
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

type errSegmentNotFound struct {
	path    string
	segment string
}

func (e *errSegmentNotFound) Error() string {
	return fmt.Sprintf("segment %s not found in path %s", e.segment, e.path)
}

func (p *Preprocessor) getValue(reqMap map[string]any, path string) (any, error) {
	var curSegment strings.Builder
	segments := strings.Split(path, ".")

	currentData := reqMap
	for i, segment := range segments {
		segment = strings.TrimSpace(segment)
		curSegment.WriteByte('.')
		curSegment.WriteString(segment)
		if strings.Contains(segment, "[") && strings.HasSuffix(segment, "]") {
			openBraceIdx := strings.Index(segment, "[")
			indexStr := strings.ToLower(strings.TrimSpace(segment[openBraceIdx+1 : len(segment)-1]))

			segment = segment[:openBraceIdx]
			value, exists := currentData[segment]
			if !exists {
				return nil, &errSegmentNotFound{path: path, segment: segment}
			}

			mval, isMval := value.([]map[string]string)
			if isMval {
				index, err := p.calcIndex(indexStr, curSegment.String(), len(mval))
				if err != nil {
					return nil, fmt.Errorf("failed to calc index: %w", err)
				}
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
					index, err := p.calcIndex(indexStr, curSegment.String(), len(anySlice))
					if err != nil {
						return nil, fmt.Errorf("failed to calc index: %w", err)
					}
					if i != len(segments)-1 {
						return nil, fmt.Errorf("not last segment %s in path %s", segment, path)
					}
					return anySlice[index], nil
				}
				stringSlice, isStringSlice := value.([]string)
				if isStringSlice {
					index, err := p.calcIndex(indexStr, curSegment.String(), len(stringSlice))
					if err != nil {
						return nil, fmt.Errorf("failed to calc index: %w", err)
					}
					if i != len(segments)-1 {
						return nil, fmt.Errorf("not last segment %s in path %s", segment, path)
					}
					return stringSlice[index], nil
				}
				return nil, fmt.Errorf("invalid type of segment %s in path %s", segment, path)
			}

			index, err := p.calcIndex(indexStr, curSegment.String(), len(mapSlice))
			if err != nil {
				return nil, fmt.Errorf("failed to calc index: %w", err)
			}
			currentData = mapSlice[index]
		} else {
			value, exists := currentData[segment]
			if !exists {
				return nil, &errSegmentNotFound{path: path, segment: segment}
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

func (p *Preprocessor) calcIndex(indexStr string, segment string, length int) (int, error) {
	index, err := strconv.Atoi(indexStr)
	if err != nil && indexStr != "next" && indexStr != "rand" && indexStr != "last" {
		return 0, fmt.Errorf("invalid index: %s", indexStr)
	}
	if indexStr != "next" && indexStr != "rand" && indexStr != "last" {
		if index >= 0 && index < length {
			return index, nil
		}
		index %= length
		if index < 0 {
			index += length
		}
		return index, nil
	}

	if indexStr == "last" {
		return length - 1, nil
	}
	if indexStr == "rand" {
		return p.iterator.rand(length), nil
	}
	index = p.iterator.next(segment)
	if index >= length {
		index %= length
	}
	return index, nil
}
