package hcl2

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// Unmarshal parses the HCL2 configuration data from the given byte slice and
// unmarshals it into the provided struct.
// shortage:
// 1. time.Duration is not supported. Using 'github.com/cocktail828/go-cty@v0.0.1' instead
func Unmarshal(data []byte, v any) error {
	file, diags := hclsyntax.ParseConfig(data, "example.hcl", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return diags
	}

	diags = gohcl.DecodeBody(file.Body, nil, v)
	for _, diag := range diags {
		if diag.Summary == "Unsupported argument" {
			// Allow unknow fields
			diag.Severity = hcl.DiagWarning
		}
	}

	if diags.HasErrors() {
		sort.Slice(diags, func(i, j int) bool {
			return diags[i].Severity < diags[j].Severity
		})
		return diags
	}
	return nil
}

type hclTag struct {
	Name    string
	IsBlock bool
	IsLabel bool
	IsSlice bool
}

func Marshal(v any) ([]byte, error) {
	f := hclwrite.NewEmptyFile()
	processStructFields(f.Body(), v)
	return f.Bytes(), nil
}

func extractLabels(obj any) []string {
	val := reflect.ValueOf(obj)
	typ := val.Type()

	if typ.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = val.Type()
	}

	if typ.Kind() != reflect.Struct {
		return nil
	}

	var labels []string
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tagName := field.Tag.Get("hcl")
		if tagName == "-" {
			continue
		}

		parts := strings.Split(tagName, ",")
		if len(parts) > 1 && parts[1] == "label" && field.IsExported() {
			fieldVal := val.Field(i)
			if !fieldVal.IsZero() {
				labels = append(labels, fmt.Sprintf("%v", fieldVal.Interface()))
			}
		}
	}
	return labels
}

func processStructFields(body *hclwrite.Body, obj any) {
	val := reflect.ValueOf(obj)
	typ := val.Type()

	if typ.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = val.Type()
	}

	if typ.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)
		if !field.IsExported() {
			continue
		}

		tagName := field.Tag.Get("hcl")
		if tagName == "-" {
			continue
		}

		hclTag := hclTag{
			Name:    field.Name,
			IsSlice: fieldVal.Kind() == reflect.Slice || fieldVal.Kind() == reflect.Array,
		}
		if tagName != "" {
			parts := strings.Split(tagName, ",")
			if len(parts) > 0 {
				hclTag.Name = parts[0]
			}

			if len(parts) > 1 {
				switch parts[1] {
				case "block":
					hclTag.IsBlock = true
				case "label":
					hclTag.IsLabel = true
				}
			}
		}

		if hclTag.IsLabel {
			continue
		}
		handleField(body, hclTag.Name, fieldVal, hclTag)
	}
}

func handleField(body *hclwrite.Body, fieldName string, fieldVal reflect.Value, hclTag hclTag) {
	isStruct := fieldVal.Kind() == reflect.Struct ||
		(fieldVal.Kind() == reflect.Ptr && fieldVal.Elem().Kind() == reflect.Struct)

	shouldBeBlock := hclTag.IsBlock || (isStruct && fieldName != "")
	if shouldBeBlock && !hclTag.IsSlice {
		var labels []string
		if fieldVal.CanInterface() {
			labels = extractLabels(fieldVal.Interface())
		}

		block := body.AppendNewBlock(fieldName, labels)
		blockBody := block.Body()
		processStructFields(blockBody, fieldVal.Interface())
		return
	}

	switch fieldVal.Kind() {
	case reflect.String:
		body.SetAttributeValue(fieldName, cty.StringVal(fieldVal.String()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		body.SetAttributeValue(fieldName, cty.NumberIntVal(fieldVal.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		body.SetAttributeValue(fieldName, cty.NumberIntVal(int64(fieldVal.Uint())))
	case reflect.Float32, reflect.Float64:
		body.SetAttributeValue(fieldName, cty.NumberFloatVal(fieldVal.Float()))
	case reflect.Bool:
		body.SetAttributeValue(fieldName, cty.BoolVal(fieldVal.Bool()))
	case reflect.Slice, reflect.Array:
		handleSlice(body, fieldName, fieldVal)
	case reflect.Map:
		handleMap(body, fieldName, fieldVal)
	case reflect.Struct:
		processStructFields(body, fieldVal.Interface())
	case reflect.Ptr:
		if !fieldVal.IsNil() {
			handleField(body, fieldName, fieldVal.Elem(), hclTag)
		}
	}
}

func handleSlice(body *hclwrite.Body, fieldName string, fieldVal reflect.Value) {
	if fieldVal.Len() == 0 {
		body.SetAttributeValue(fieldName, cty.ListValEmpty(cty.String))
		return
	}

	elemType := fieldVal.Type().Elem()
	isStructElem := elemType.Kind() == reflect.Struct ||
		(elemType.Kind() == reflect.Ptr && elemType.Elem().Kind() == reflect.Struct)

	if isStructElem {
		for i := 0; i < fieldVal.Len(); i++ {
			elem := fieldVal.Index(i)
			if elem.Kind() == reflect.Ptr && elem.IsNil() {
				continue
			}

			var labels []string
			if elem.CanInterface() {
				labels = extractLabels(elem.Interface())
			}

			block := body.AppendNewBlock(fieldName, labels)
			blockBody := block.Body()
			processStructFields(blockBody, elem.Interface())
		}
		return
	}

	elemKind := elemType.Kind()
	vals := make([]cty.Value, 0, fieldVal.Len())
	for i := 0; i < fieldVal.Len(); i++ {
		elem := fieldVal.Index(i)
		switch elemKind {
		case reflect.String:
			vals = append(vals, cty.StringVal(elem.String()))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			vals = append(vals, cty.NumberIntVal(elem.Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			vals = append(vals, cty.NumberIntVal(int64(elem.Uint())))
		case reflect.Float32, reflect.Float64:
			vals = append(vals, cty.NumberFloatVal(elem.Float()))
		case reflect.Bool:
			vals = append(vals, cty.BoolVal(elem.Bool()))
		default:
			vals = append(vals, cty.StringVal(fmt.Sprintf("%v", elem.Interface())))
		}
	}

	body.SetAttributeValue(fieldName, cty.ListVal(vals))
}

func handleMap(body *hclwrite.Body, fieldName string, fieldVal reflect.Value) {
	if fieldVal.IsNil() || fieldVal.Len() == 0 {
		body.SetAttributeValue(fieldName, cty.MapValEmpty(cty.String))
		return
	}

	valType := fieldVal.Type().Elem()
	isStructVal := valType.Kind() == reflect.Struct ||
		(valType.Kind() == reflect.Ptr && valType.Elem().Kind() == reflect.Struct)

	if isStructVal {
		iter := fieldVal.MapRange()
		for iter.Next() {
			key := iter.Key()
			val := iter.Value()

			labels := []string{}
			if key.Kind() == reflect.String {
				labels = append(labels, key.String())
			} else {
				labels = append(labels, fmt.Sprintf("%v", key.Interface()))
			}

			if val.CanInterface() {
				labels = append(labels, extractLabels(val.Interface())...)
			}

			block := body.AppendNewBlock(fieldName, labels)
			blockBody := block.Body()
			processStructFields(blockBody, val.Interface())
		}
		return
	}

	vals := make(map[string]cty.Value)
	iter := fieldVal.MapRange()

	for iter.Next() {
		key := iter.Key()
		val := iter.Value()

		var keyStr string
		if key.Kind() == reflect.String {
			keyStr = key.String()
		} else {
			keyStr = fmt.Sprintf("%v", key.Interface())
		}

		var ctyVal cty.Value
		if val.Kind() == reflect.String {
			ctyVal = cty.StringVal(val.String())
		} else {
			ctyVal = cty.StringVal(fmt.Sprintf("%v", val.Interface()))
		}

		vals[keyStr] = ctyVal
	}

	body.SetAttributeValue(fieldName, cty.MapVal(vals))
}
