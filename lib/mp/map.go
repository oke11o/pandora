package mp

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type ErrSegmentNotFound struct {
	path    string
	segment string
}

func (e *ErrSegmentNotFound) Error() string {
	return fmt.Sprintf("segment %s not found in path %s", e.segment, e.path)
}

func MergeRecursive(dst, src map[string]any) error {
	for srcKey, srcVal := range src {
		dstVal, ok := dst[srcKey]
		if !ok {
			dst[srcKey] = srcVal
			continue
		}
		srcValueType := reflect.TypeOf(srcVal).Kind()
		dstValueType := reflect.TypeOf(dstVal).Kind()

		if srcValueType != dstValueType {
			return fmt.Errorf("value of field `%v` should be same type but got %T and %T", srcKey, srcVal, dstVal)
		}
		switch srcValueType {
		case reflect.Map:
			dstValM, okd := dstVal.(map[string]any)
			srcValM, oks := srcVal.(map[string]any)
			if okd && oks {
				if err := MergeRecursive(dstValM, srcValM); err != nil {
					return err
				}
			} else {
				dstValAM, okd := dstVal.(map[any]any)
				srcValAM, oks := srcVal.(map[any]any)
				if !okd || !oks {
					return fmt.Errorf("field `%s` should be support map[string]any or map[any]any; but is %T and %T", srcKey, dstVal, srcVal)
				}
				dstValM := make(map[string]any, len(dstValAM))
				for d1k, d1v := range dstValAM {
					dstValM[d1k.(string)] = d1v
				}
				srcValM := make(map[string]any, len(srcValAM))
				for d1k, d1v := range srcValAM {
					srcValM[d1k.(string)] = d1v
				}
				if err := MergeRecursive(dstValM, srcValM); err != nil {
					return err
				}
				dst[srcKey] = dstValM
			}
		case reflect.Slice:
			switch srcTypedVal := srcVal.(type) {
			case []any:
				dstTypedVal, ok := dstVal.([]any)
				var dstLen int
				if ok {
					dstLen = len(dstTypedVal)
				} else {
					return fmt.Errorf("field `%s` should be map[string]any but is `%T`", srcKey, dstVal)
				}
				srcLen := len(srcTypedVal)
				minLen := srcLen
				if minLen > dstLen {
					minLen = dstLen
				}
				for i := 0; i < minLen; i++ {
					d, ok := dstTypedVal[i].(map[string]any)
					if !ok {
						d1, ok := dstTypedVal[i].(map[any]any)
						if !ok {
							return fmt.Errorf("field `%s` should be []any where any=map[string]any or any=map[any]any but is `%T`", srcKey, dstVal)
						}
						d = make(map[string]any, len(d1))
						for d1k, d1v := range d1 {
							d1kk, ok := d1k.(string)
							_ = ok
							d[d1kk] = d1v
						}
					}
					s, ok := srcTypedVal[i].(map[string]any)
					if !ok {
						s1, ok := srcTypedVal[i].(map[any]any)
						if !ok {
							return fmt.Errorf("field `%s` should be []any where any=map[string]any or any=map[any]any but is `%T`", srcKey, srcVal)
						}
						s = make(map[string]any, len(s1))
						for s1k, s1v := range s1 {
							s1kk, ok := s1k.(string)
							_ = ok
							s[s1kk] = s1v
						}
					}
					err := MergeRecursive(d, s)
					if err != nil {
						return err
					}
					dstTypedVal[i] = d
				}
				if dstLen < srcLen {
					for i := dstLen; i < srcLen; i++ {
						dstTypedVal = append(dstTypedVal, srcTypedVal[i])
					}
					dst[srcKey] = dstTypedVal
				}
			case []map[string]any:
				dstTypedVal, ok := dstVal.([]map[string]any)
				var dstLen int
				if ok {
					dstLen = len(dstTypedVal)
				} else {
					return fmt.Errorf("field `%s` should be map[string]any but is `%T`", srcKey, dstVal)
				}
				srcLen := len(srcTypedVal)
				minLen := srcLen
				if minLen > dstLen {
					minLen = dstLen
				}
				for i := 0; i < minLen; i++ {
					err := MergeRecursive(dstTypedVal[i], srcTypedVal[i])
					if err != nil {
						return err
					}
				}
				if dstLen < srcLen {
					for i := dstLen; i < srcLen; i++ {
						dstTypedVal = append(dstTypedVal, srcTypedVal[i])
					}
					dst[srcKey] = dstTypedVal
				}
			default:
				dst[srcKey] = srcVal
			}
		default:
			dst[srcKey] = srcVal
		}
	}
	return nil
}

func GetMapValue(current map[string]any, path string, iter Iterator) (any, error) {
	var curSegment strings.Builder
	segments := strings.Split(path, ".")

	for i, segment := range segments {
		segment = strings.TrimSpace(segment)
		curSegment.WriteByte('.')
		curSegment.WriteString(segment)
		if strings.Contains(segment, "[") && strings.HasSuffix(segment, "]") {
			openBraceIdx := strings.Index(segment, "[")
			indexStr := strings.ToLower(strings.TrimSpace(segment[openBraceIdx+1 : len(segment)-1]))

			segment = segment[:openBraceIdx]
			pathVal, ok := current[segment]
			if !ok {
				return nil, &ErrSegmentNotFound{path: path, segment: segment}
			}
			sliceElement, err := extractFromSlice(pathVal, indexStr, curSegment.String(), iter)
			if err != nil {
				return nil, fmt.Errorf("cant extract value path=`%s`,segment=`%s`,err=%w", segment, path, err)
			}
			current, ok = sliceElement.(map[string]any)
			if !ok {
				if i != len(segments)-1 {
					return nil, fmt.Errorf("not last segment %s in path %s", segment, path)
				}
				return sliceElement, nil
			}
		} else {
			pathVal, ok := current[segment]
			if !ok {
				return nil, &ErrSegmentNotFound{path: path, segment: segment}
			}
			current, ok = pathVal.(map[string]any)
			if !ok {
				if i != len(segments)-1 {
					return nil, fmt.Errorf("not last segment %s in path %s", segment, path)
				}
				return pathVal, nil
			}
		}
	}

	return current, nil
}

func extractFromSlice(curValue any, indexStr string, curSegment string, iter Iterator) (result any, err error) {
	validTypes := []reflect.Type{
		reflect.TypeOf([]map[string]string{}),
		reflect.TypeOf([]map[string]any{}),
		reflect.TypeOf([]any{}),
		reflect.TypeOf([]string{}),
		reflect.TypeOf([]int{}),
		reflect.TypeOf([]int64{}),
		reflect.TypeOf([]float64{}),
	}

	var valueLen int
	var valueFound bool
	for _, valueType := range validTypes {
		if reflect.TypeOf(curValue) == valueType {
			valueLen = reflect.ValueOf(curValue).Len()
			valueFound = true
			break
		}
	}

	if !valueFound {
		return nil, fmt.Errorf("invalid type of value `%+v`, %T", curValue, curValue)
	}

	index, err := calcIndex(indexStr, curSegment, valueLen, iter)
	if err != nil {
		return nil, fmt.Errorf("failed to calc index for %T; err: %w", curValue, err)
	}

	switch v := curValue.(type) {
	case []map[string]string:
		currentData := make(map[string]any, len(v[index]))
		for k, val := range v[index] {
			currentData[k] = val
		}
		return currentData, nil
	case []map[string]any:
		return v[index], nil
	case []any:
		return v[index], nil
	case []string:
		return v[index], nil
	case []int:
		return v[index], nil
	case []int64:
		return v[index], nil
	case []float64:
		return v[index], nil
	}

	// This line should never be reached, as we've covered all valid types above
	return nil, fmt.Errorf("invalid type of value `%+v`, %T", curValue, curValue)
}

func calcIndex(indexStr string, segment string, length int, iter Iterator) (int, error) {
	index, err := strconv.Atoi(indexStr)
	if err != nil && indexStr != "next" && indexStr != "rand" && indexStr != "last" {
		return 0, fmt.Errorf("index should be integer or one of [next, rand, last], but got `%s`", indexStr)
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
		return iter.Rand(length), nil
	}
	index = iter.Next(segment)
	if index >= length {
		index %= length
	}
	return index, nil
}
