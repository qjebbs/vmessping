package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
)

// MergeJSONs merge multiple json config files to Conifg
func MergeJSONs(paths []string) ([]byte, error) {
	files, err := pathsToFiles(paths)
	if err != nil {
		return nil, err
	}
	conf := make(map[string]interface{}, 0)
	for _, file := range files {
		c, err := jsonToMap(file)
		if err != nil {
			return nil, err
		}
		if err = mergeMaps(conf, c); err != nil {
			return nil, err
		}
	}
	sortSlicesInMap(conf)
	removePriorityKey(conf)
	data, err := json.Marshal(conf)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func isZero(v interface{}) bool {
	return getValue(reflect.ValueOf(v)).IsZero()
}

func getPriority(v interface{}) float64 {
	var m map[string]interface{}
	var ok bool
	if m, ok = v.(map[string]interface{}); !ok {
		return 0
	}
	if i, ok := m["priority"]; ok {
		if p, ok := i.(float64); ok {
			return p
		}
	}
	return 0
}
func sortSlicesInMap(target map[string]interface{}) {
	for key, value := range target {
		if slice, ok := value.([]interface{}); ok {
			sort.Slice(slice, func(i, j int) bool { return getPriority(slice[i]) < getPriority(slice[j]) })
			target[key] = slice
		} else if field, ok := value.(map[string]interface{}); ok {
			sortSlicesInMap(field)
		}
	}
}
func removePriorityKey(target map[string]interface{}) {
	for key, value := range target {
		if _, ok := value.(float64); key == "priority" && ok {
			delete(target, key)
		} else if slice, ok := value.([]interface{}); ok {
			for _, e := range slice {
				if el, ok := e.(map[string]interface{}); ok {
					removePriorityKey(el)
				}
			}
		} else if field, ok := value.(map[string]interface{}); ok {
			removePriorityKey(field)
		}
	}
}
func mergeMaps(target map[string]interface{}, source map[string]interface{}) error {
	for key, value := range source {
		// fmt.Printf("[%s] type: %s, kind: %s\n", key, getType(fieldTypeSrc.Type).Name(), getType(fieldTypeSrc.Type).Kind())
		if (value == nil) || isZero(value) {
			continue
		}
		if target[key] == nil || isZero(value) {
			target[key] = value
			continue
		}
		if slice, ok := value.([]interface{}); ok {
			if tslice, ok := target[key].([]interface{}); ok {
				target[key] = append(tslice, slice...)
			} else {
				return fmt.Errorf("value type of key (%s) mismatch, source is 'slice' but target not", key)
			}
		} else if field, ok := value.(map[string]interface{}); ok {
			if mapField, ok := target[key].(map[string]interface{}); ok {
				if err := mergeMaps(mapField, field); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("value type of key (%s) mismatch, source is 'map[string]interface{}' but target not", key)
			}
		}
	}
	return nil
}

func jsonToMap(f string) (map[string]interface{}, error) {
	c := make(map[string]interface{})
	r, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	err = decodeJSONConfig(r, &c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
