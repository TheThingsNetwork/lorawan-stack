// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package filter

import "fmt"

// Build builds a filter based of the input, which should be a map[string]interface{} or a []map[string]interface{}
// that describes the filter, usually read from config.
func Build(in interface{}) (Filter, error) {
	switch v := in.(type) {
	case map[string]interface{}:
		// Check which clause

		// MATCH clause
		if m, ok := v["match"]; ok {
			return match(m)
		}

		// AND clause
		if clauses, ok := v["and"]; ok {
			c, err := gather(clauses)
			if err != nil {
				return nil, err
			}

			return And(c...), nil
		}

		// OR clause
		if clauses, ok := v["or"]; ok {
			c, err := gather(clauses)
			if err != nil {
				return nil, err
			}

			return Or(c...), nil
		}

		return nil, fmt.Errorf("Unknown clause, expected `match`, `or` or `and` (got %v)", v)

	case []map[string]interface{}:
		// list of clauses is AND
		return Build(map[string]interface{}{
			"and": v,
		})

	case []interface{}:
		// list of clauses is AND
		return Build(map[string]interface{}{
			"and": v,
		})

	case map[interface{}]interface{}:
		return Build(stringMap(v))

	default:
		return nil, fmt.Errorf("Cannot parse filter description of type %T", in)
	}
}

func gather(in interface{}) ([]Filter, error) {
	list, ok := in.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Expected list of clauses, got %T", in)
	}

	filters := make([]Filter, 0, len(list))
	for _, item := range list {
		desc, err := Build(item)
		if err != nil {
			return nil, err
		}

		filters = append(filters, desc)
	}

	return filters, nil
}

func stringMap(in map[interface{}]interface{}) map[string]interface{} {
	res := make(map[string]interface{}, len(in))
	for key, val := range in {
		var skey string
		if str, ok := key.(string); ok {
			skey = str
		} else if s, ok := key.(fmt.Stringer); ok {
			skey = s.String()
		} else {
			skey = fmt.Sprintf("%v", key)
		}

		res[skey] = val
	}

	return res
}

func match(in interface{}) (Filter, error) {
	var field string
	var value string
	var ok bool

	switch m := in.(type) {
	case map[interface{}]interface{}:
		return match(stringMap(m))

	case map[string]interface{}:
		field, ok = m["field"].(string)
		if !ok {
			return nil, fmt.Errorf("match `field` should be a string, got %T", m["field"])
		}

		value, ok = m["value"].(string)
		if !ok {
			return nil, fmt.Errorf("match `value` should be a string, got %T", m["value"])
		}

	default:
		return nil, fmt.Errorf("match should be a map with keys `field` and `value`, got %v", m)
	}

	return FieldString(field, value), nil
}
