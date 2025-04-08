// Package diff - Type and Value Formatting
// This file contains functions for formatting Go types and values
// for better diff representation
package diff

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"sort"
	"strings"

	"github.com/walteh/yaml"
)

// ConvolutedFormatReflectValue formats a reflect.Value into a standardized string
// representation suitable for comparison. It converts the value to JSON and then
// uses a stable ordering to ensure consistent output.
func ConvolutedFormatReflectValueAsJSON(s reflect.Value) any {
	if !s.IsValid() {
		return s.String()
	}

	// Convert the value to JSON
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "\t")

	if err := enc.Encode(s.Interface()); err != nil {
		panic(err)
	}

	// Use ordered map to ensure stable field ordering
	mapd := yaml.NewOrderedMap()
	if err := json.Unmarshal(buf.Bytes(), mapd); err != nil {
		panic(err)
	}

	// Sort keys alphabetically
	ms := mapd.ToMapSlice()
	keys := extractAndSortKeys(ms)
	ms.SortKeys(keys...)

	// Re-encode to JSON with consistent ordering
	buf.Reset()
	enc2 := json.NewEncoder(buf)
	enc2.SetIndent("", "\t")
	if err := enc2.Encode(ms); err != nil {
		panic(err)
	}

	return buf.String()
}

// extractAndSortKeys extracts keys from a MapSlice and returns them sorted
func extractAndSortKeys(ms yaml.MapSlice) []string {
	keys := []string{}
	for _, key := range ms {
		if keyStr, ok := key.Key.(string); ok {
			keys = append(keys, keyStr)
		}
	}
	sort.Strings(keys)
	return keys
}

// ConvolutedFormatReflectType formats a reflect.Type into a standardized string
// representation.
// It handles struct types specially, formatting them with consistent field ordering
// and indentation for better readability.
func ConvolutedFormatReflectType(s reflect.Type) string {
	// Only process struct types specially
	if !strings.Contains(s.String(), "struct {") {
		return s.String()
	}

	// Create a valid Go file from the struct
	src := fmt.Sprintf("package p\ntype T %s", s)

	// Handle string escaping for proper parsing
	src = strings.ReplaceAll(src, "\\\"", "$$$$")
	src = strings.ReplaceAll(src, "\"", "`")
	src = strings.ReplaceAll(src, "$$$$", "\"")

	// Parse the file using Go's AST parser
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return s.String() // Return original if parsing fails
	}

	// Find the struct type in the AST
	var structType *ast.StructType
	ast.Inspect(file, func(n ast.Node) bool {
		if t, ok := n.(*ast.StructType); ok {
			structType = t
			return false
		}
		return true
	})

	if structType == nil {
		return s.String()
	}

	// Format the struct AST
	return formatStructAST(structType, 0)
}

// formatStructAST formats a struct AST node with proper indentation
// and field ordering.
func formatStructAST(structType *ast.StructType, depth int) string {
	if structType == nil || structType.Fields == nil {
		return ""
	}

	// Collect and sort fields
	var fields []string
	for _, field := range structType.Fields.List {
		fieldStr := formatField(field, depth+1)
		if fieldStr != "" {
			fields = append(fields, fieldStr)
		}
	}
	sort.Strings(fields)

	// Build the formatted struct
	var result strings.Builder
	result.WriteString("struct {\n")

	// Add each field with proper indentation
	for i, field := range fields {
		result.WriteString(strings.Repeat("\t", depth+1))
		result.WriteString(field)
		if i < len(fields)-1 {
			result.WriteString("\n")
		}
	}

	// Close the struct
	if len(fields) > 0 {
		result.WriteString("\n")
	}
	result.WriteString(strings.Repeat("\t", depth) + "}")

	return result.String()
}

// formatField formats a single struct field from AST
func formatField(field *ast.Field, depth int) string {
	if field == nil {
		return ""
	}

	// Extract field name
	var name string
	if len(field.Names) > 0 {
		name = field.Names[0].Name
	}

	// Format the type
	typeStr := formatType(field.Type, depth)

	// Format the tag if present
	var tagStr string
	if field.Tag != nil {
		tagStr = field.Tag.Value
	}

	// Build the field string
	var parts []string
	if name != "" {
		parts = append(parts, name)
	}
	parts = append(parts, typeStr)
	if tagStr != "" {
		parts = append(parts, tagStr)
	}

	return strings.Join(parts, " ")
}

// formatType formats a type AST node
func formatType(expr ast.Expr, depth int) string {
	if expr == nil {
		return ""
	}

	switch t := expr.(type) {
	case *ast.Ident:
		// Simple identifier (like 'string', 'int', etc.)
		return t.Name

	case *ast.StarExpr:
		// Pointer type (like '*string')
		return "*" + formatType(t.X, depth)

	case *ast.ArrayType:
		// Array or slice type (like '[]string')
		return "[]" + formatType(t.Elt, depth)

	case *ast.MapType:
		// Map type (like 'map[string]int')
		return fmt.Sprintf("map[%s]%s",
			formatType(t.Key, depth),
			formatType(t.Value, depth))

	case *ast.StructType:
		// Nested struct
		return formatStructAST(t, depth)

	default:
		// Other types (interfaces, channels, etc.)
		return fmt.Sprintf("%#v", expr)
	}
}
