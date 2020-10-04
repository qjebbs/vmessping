package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
)

type routerConfigRaw struct {
	Settings       json.RawMessage   `json:"settings,omitempty"` // Deprecated
	RuleList       []json.RawMessage `json:"rules,omitempty"`
	DomainStrategy json.RawMessage   `json:"domainStrategy,omitempty"`
	Balancers      []json.RawMessage `json:"balancers,omitempty"`
}
type configRaw struct {
	Port            json.RawMessage   `json:"port,omitempty"` // Port of this Point server. Deprecated.
	LogConfig       json.RawMessage   `json:"log,omitempty"`
	RouterConfig    *routerConfigRaw  `json:"routing,omitempty"`
	DNSConfig       json.RawMessage   `json:"dns,omitempty"`
	InboundConfigs  []json.RawMessage `json:"inbounds,omitempty"`
	OutboundConfigs []json.RawMessage `json:"outbounds,omitempty"`
	InboundConfig   json.RawMessage   `json:"inbound,omitempty"`        // Deprecated.
	OutboundConfig  json.RawMessage   `json:"outbound,omitempty"`       // Deprecated.
	InboundDetours  []json.RawMessage `json:"inboundDetour,omitempty"`  // Deprecated.
	OutboundDetours []json.RawMessage `json:"outboundDetour,omitempty"` // Deprecated.
	Transport       json.RawMessage   `json:"transport,omitempty"`
	Policy          json.RawMessage   `json:"policy,omitempty"`
	API             json.RawMessage   `json:"api,omitempty"`
	Stats           json.RawMessage   `json:"stats,omitempty"`
	Reverse         json.RawMessage   `json:"reverse,omitempty"`
}
type ruleWithPriority struct {
	Priority int `json:"priority"`
	Rule     *json.RawMessage
}
type outboundWithPriority struct {
	Priority int
	Outbound *json.RawMessage
}

// MergeJSONsWithReflect merge multiple json config files to Conifg
// This keeps json orders, but only sorts routing rules and outbounds
func MergeJSONsWithReflect(paths []string) ([]byte, error) {
	files, err := pathsToFiles(paths)
	if err != nil {
		return nil, err
	}
	conf := &configRaw{}
	for _, file := range files {
		c, err := jsonToConfigRaw(file)
		if err != nil {
			return nil, err
		}
		err = mergeStructs(conf, c)
		if err != nil {
			return nil, err
		}
	}
	err = sortRulesByPriority(conf)
	if err != nil {
		return nil, err
	}
	err = sortOutboundsByPriority(conf)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(conf)
	if err != nil {
		return nil, err
	}
	return data, nil
}
func mergeStructs(target interface{}, source interface{}) error {
	return mergeValues(reflect.ValueOf(target), reflect.ValueOf(source))
}

func mergeValues(target reflect.Value, source reflect.Value) error {
	typeSrc := source.Type()
	typeTgt := target.Type()
	if typeSrc.Name() != typeTgt.Name() {
		return errors.New("type mismatch")
	}
	for i := 0; i < getValue(source).NumField(); i++ {
		fieldTypeSrc := getType(typeSrc).Field(i)
		fieldValueSrc := getValue(source).Field(i)
		// fmt.Printf("[%s] type: %s, kind: %s\n", fieldTypeSrc.Name, getType(fieldTypeSrc.Type).Name(), getType(fieldTypeSrc.Type).Kind())
		if (fieldValueSrc.Kind() == reflect.Ptr && fieldValueSrc.IsNil()) || getValue(fieldValueSrc).IsZero() {
			// fmt.Println(getValue(fieldValueSrc).Interface())
			continue
		}
		fieldValueTgt := getValue(target).FieldByName(fieldTypeSrc.Name)
		if !fieldValueTgt.CanSet() {
			fmt.Printf("%s: cannot set", fieldTypeSrc.Name)
			continue
		}
		if (fieldValueTgt.Kind() == reflect.Ptr && fieldValueTgt.IsNil()) || getValue(fieldValueTgt).IsZero() {
			fieldValueTgt.Set(fieldValueSrc)
			continue
		}
		if getValue(fieldValueSrc).Kind() == reflect.Slice {
			for i := 0; i < fieldValueSrc.Len(); i++ {
				el := fieldValueSrc.Index(i)
				fieldValueTgt.Set(reflect.Append(fieldValueTgt, el))
			}
		} else if getType(fieldTypeSrc.Type).Kind() == reflect.Struct {
			err := mergeValues(fieldValueTgt, fieldValueSrc)
			if err != nil {
				return err
			}
			continue
		}
	}
	return nil
}

func getValue(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		return v.Elem()
	}
	return v
}

func getType(v reflect.Type) reflect.Type {
	if v.Kind() == reflect.Ptr {
		return v.Elem()
	}
	return v
}

func sortRulesByPriority(c *configRaw) error {
	rules := make([]*ruleWithPriority, 0)
	if c.RouterConfig == nil || len(c.RouterConfig.RuleList) == 0 {
		return nil
	}
	for i := 0; i < len(c.RouterConfig.RuleList); i++ {
		raw := &c.RouterConfig.RuleList[i]
		r := &ruleWithPriority{Rule: raw}
		err := json.Unmarshal(*raw, r)
		if err != nil {
			return err
		}
		rules = append(rules, r)
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].Priority < rules[j].Priority })
	c.RouterConfig.RuleList = make([]json.RawMessage, 0)
	for _, r := range rules {
		c.RouterConfig.RuleList = append(c.RouterConfig.RuleList, *r.Rule)
	}
	return nil
}

func sortOutboundsByPriority(c *configRaw) error {
	outbounds := make([]*outboundWithPriority, 0)
	if len(c.OutboundConfigs) == 0 {
		return nil
	}
	for i := 0; i < len(c.OutboundConfigs); i++ {
		out := &c.OutboundConfigs[i]
		o := &outboundWithPriority{}
		err := json.Unmarshal(*out, o)
		if err != nil {
			return err
		}
		o.Outbound = out
		outbounds = append(outbounds, o)
	}
	sort.Slice(outbounds, func(i, j int) bool { return outbounds[i].Priority < outbounds[j].Priority })
	c.OutboundConfigs = make([]json.RawMessage, 0)
	for _, r := range outbounds {
		c.OutboundConfigs = append(c.OutboundConfigs, *r.Outbound)
	}
	return nil
}

func jsonToConfigRaw(f string) (*configRaw, error) {
	c := &configRaw{}
	r, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	err = decodeJSONConfig(r, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
