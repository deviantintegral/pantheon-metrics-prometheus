// Package output provides formatting utilities for CLI output.
package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// Format represents output format types
type Format string

const (
	// FormatTable is table format
	FormatTable Format = "table"
	// FormatJSON is JSON format
	FormatJSON Format = "json"
	// FormatYAML is YAML format
	FormatYAML Format = "yaml"
	// FormatCSV is CSV format
	FormatCSV Format = "csv"
	// FormatList is simple list format (one item per line)
	FormatList Format = "list"
)

// Options configures output formatting
type Options struct {
	Format Format
	Fields []string
	Writer io.Writer
}

// DefaultOptions returns default output options
func DefaultOptions() *Options {
	return &Options{
		Format: FormatTable,
		Writer: os.Stdout,
	}
}

// Print prints data in the specified format
func Print(data interface{}, opts *Options) error {
	if opts == nil {
		opts = DefaultOptions()
	}

	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}

	switch opts.Format {
	case FormatTable:
		return printTable(data, opts)
	case FormatJSON:
		return printJSON(data, opts.Writer)
	case FormatYAML:
		return printYAML(data, opts.Writer)
	case FormatCSV:
		return printCSV(data, opts)
	case FormatList:
		return printList(data, opts)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

// printJSON prints data as JSON
func printJSON(data interface{}, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// printYAML prints data as YAML
func printYAML(data interface{}, w io.Writer) error {
	encoder := yaml.NewEncoder(w)
	defer func() { _ = encoder.Close() }()
	return encoder.Encode(data)
}

// printTable prints data as a table
func printTable(data interface{}, opts *Options) error {
	rows, headers := extractTableData(data, opts.Fields)
	if len(rows) == 0 {
		return nil
	}

	// For single item, use vertical layout (PHP terminus style)
	if len(rows) == 1 {
		return printVerticalTable(rows[0], headers, opts)
	}

	// For multiple items, use horizontal layout
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	for i, h := range headers {
		_, _ = fmt.Fprintf(opts.Writer, "%-*s", widths[i]+2, h)
	}
	_, _ = fmt.Fprintln(opts.Writer)

	// Print separator
	for _, w := range widths {
		_, _ = fmt.Fprintf(opts.Writer, "%s", strings.Repeat("-", w+2))
	}
	_, _ = fmt.Fprintln(opts.Writer)

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				_, _ = fmt.Fprintf(opts.Writer, "%-*s", widths[i]+2, cell)
			}
		}
		_, _ = fmt.Fprintln(opts.Writer)
	}

	return nil
}

// printVerticalTable prints a single item as a vertical table (PHP terminus style)
func printVerticalTable(row, headers []string, opts *Options) error {
	// Calculate column widths
	fieldWidth := 0
	valueWidth := 0

	// Convert headers to human-readable titles and calculate widths
	titles := make([]string, len(headers))
	for i, h := range headers {
		titles[i] = toHumanReadable(h)
		if len(titles[i]) > fieldWidth {
			fieldWidth = len(titles[i])
		}
	}

	for _, cell := range row {
		if len(cell) > valueWidth {
			valueWidth = len(cell)
		}
	}

	// Print top border
	_, _ = fmt.Fprintf(opts.Writer, " %s %s\n",
		strings.Repeat("-", fieldWidth),
		strings.Repeat("-", valueWidth))

	// Print each field-value pair
	for i, title := range titles {
		value := ""
		if i < len(row) {
			value = row[i]
		}
		_, _ = fmt.Fprintf(opts.Writer, "  %-*s   %s\n", fieldWidth, title, value)
	}

	// Print bottom border
	_, _ = fmt.Fprintf(opts.Writer, " %s %s\n",
		strings.Repeat("-", fieldWidth),
		strings.Repeat("-", valueWidth))

	return nil
}

// toHumanReadable converts a field name to a human-readable title
func toHumanReadable(fieldName string) string {
	// Handle special cases
	switch strings.ToLower(fieldName) {
	case "id":
		return "ID"
	case "firstname":
		return "First Name"
	case "lastname":
		return "Last Name"
	case "email":
		return "Email"
	case "profile":
		return "Profile"
	}

	// Default: capitalize first letter
	if fieldName == "" {
		return fieldName
	}

	// Split camelCase or handle simple cases
	words := splitCamelCase(fieldName)
	for i, word := range words {
		if word != "" {
			words[i] = strings.ToUpper(word[0:1]) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}

// splitCamelCase splits a camelCase or lowercase string into words
func splitCamelCase(s string) []string {
	// If string has underscores, split by them
	if strings.Contains(s, "_") {
		return strings.Split(s, "_")
	}

	// Simple case: all lowercase or uppercase
	if s == strings.ToLower(s) || s == strings.ToUpper(s) {
		return []string{s}
	}

	// Split camelCase
	var words []string
	var currentWord strings.Builder

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			words = append(words, currentWord.String())
			currentWord.Reset()
		}
		currentWord.WriteRune(r)
	}

	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

// printCSV prints data as CSV
func printCSV(data interface{}, opts *Options) error {
	rows, headers := extractTableData(data, opts.Fields)
	if len(rows) == 0 {
		return nil
	}

	w := csv.NewWriter(opts.Writer)
	defer w.Flush()

	// Write header
	if err := w.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rows
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// printList prints data as a simple list
func printList(data interface{}, opts *Options) error {
	rows, _ := extractTableData(data, opts.Fields)

	for _, row := range rows {
		if len(row) > 0 {
			_, _ = fmt.Fprintln(opts.Writer, row[0])
		}
	}

	return nil
}

// extractTableData extracts table data from various data types
func extractTableData(data interface{}, fields []string) (rows [][]string, headers []string) {
	v := reflect.ValueOf(data)

	// Dereference pointer
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// If no fields specified, check if the data implements DefaultFielder
	if len(fields) == 0 {
		fields = getDefaultFields(data)
	}

	// Handle slice/array
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		if v.Len() == 0 {
			return nil, nil
		}

		var rows [][]string
		var headers []string

		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)

			// Don't dereference here - let extractRow handle it
			// This allows extractRow to check if pointer types implement interfaces

			row, itemHeaders := extractRow(item, fields)
			if i == 0 {
				headers = itemHeaders
			}
			rows = append(rows, row)
		}

		return rows, headers
	}

	// Handle single item
	row, headers := extractRow(v, fields)
	return [][]string{row}, headers
}

// getDefaultFields checks if data implements DefaultFielder and returns default fields
func getDefaultFields(data interface{}) []string {
	// Check direct interface
	if df, ok := data.(DefaultFielder); ok {
		return df.DefaultFields()
	}

	// For slices, check the first element
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		if v.Len() > 0 {
			item := v.Index(0)
			if item.Kind() == reflect.Ptr {
				item = item.Elem()
			}
			if item.CanInterface() {
				if df, ok := item.Interface().(DefaultFielder); ok {
					return df.DefaultFields()
				}
			}
		}
	}

	return nil
}

// extractRow extracts a single row from a struct or map
func extractRow(v reflect.Value, fields []string) (row, headers []string) {
	// Check if the value (potentially a pointer) implements Serializer interface
	// Do this BEFORE dereferencing to handle pointer receivers
	if v.CanInterface() {
		if serializer, ok := v.Interface().(Serializer); ok {
			return extractFromSerializer(serializer, fields)
		}
	}

	// Dereference pointer
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Check again after dereferencing in case of value receivers
	if v.CanInterface() {
		if serializer, ok := v.Interface().(Serializer); ok {
			return extractFromSerializer(serializer, fields)
		}
	}

	switch v.Kind() {
	case reflect.Struct:
		return extractFromStruct(v, fields)
	case reflect.Map:
		return extractFromMap(v, fields)
	default:
		// For primitive types, just convert to string
		return []string{fmt.Sprintf("%v", v.Interface())}, []string{"Value"}
	}
}

// extractFromStruct extracts data from a struct
func extractFromStruct(v reflect.Value, fields []string) (row, headers []string) {
	t := v.Type()

	// Build field map from struct tags
	fieldMap := make(map[string]int)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Get JSON tag name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Remove omitempty and other options
		tagParts := strings.Split(jsonTag, ",")
		fieldName := tagParts[0]

		fieldMap[strings.ToLower(fieldName)] = i
		fieldMap[strings.ToLower(field.Name)] = i
	}

	// If fields specified, use them; otherwise use all fields
	if len(fields) > 0 {
		for _, fieldName := range fields {
			idx, ok := fieldMap[strings.ToLower(fieldName)]
			if !ok {
				// Field not found, add empty value
				row = append(row, "")
				headers = append(headers, fieldName)
				continue
			}

			field := t.Field(idx)
			value := v.Field(idx)

			headers = append(headers, getHeaderName(&field))
			row = append(row, formatValue(value))
		}
	} else {
		// Use all fields in order
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// Skip unexported fields
			if !field.IsExported() {
				continue
			}

			// Skip fields with json:"-"
			jsonTag := field.Tag.Get("json")
			if jsonTag == "-" {
				continue
			}

			value := v.Field(i)

			headers = append(headers, getHeaderName(&field))
			row = append(row, formatValue(value))
		}
	}

	return row, headers
}

// extractFromMap extracts data from a map
func extractFromMap(v reflect.Value, fields []string) (row, headers []string) {
	if len(fields) > 0 {
		for _, fieldName := range fields {
			key := reflect.ValueOf(fieldName)
			value := v.MapIndex(key)

			headers = append(headers, fieldName)
			if value.IsValid() {
				row = append(row, formatValue(value))
			} else {
				row = append(row, "")
			}
		}
	} else {
		// Use all keys
		for _, key := range v.MapKeys() {
			value := v.MapIndex(key)

			headers = append(headers, fmt.Sprintf("%v", key.Interface()))
			row = append(row, formatValue(value))
		}
	}

	return row, headers
}

// extractFromSerializer extracts data from a type that implements Serializer
func extractFromSerializer(serializer Serializer, fields []string) (row, headers []string) {
	serializedFields := serializer.Serialize()

	// If specific fields are requested, filter and reorder
	if len(fields) > 0 {
		// Build a map for quick lookup
		fieldMap := make(map[string]SerializedField)
		for _, sf := range serializedFields {
			fieldMap[strings.ToLower(sf.Name)] = sf
		}

		// Extract only requested fields in the requested order
		for _, fieldName := range fields {
			sf, ok := fieldMap[strings.ToLower(fieldName)]
			if !ok {
				// Field not found, add empty value
				headers = append(headers, fieldName)
				row = append(row, "")
				continue
			}

			headers = append(headers, sf.Name)
			row = append(row, formatSerializedValue(sf.Value))
		}
	} else {
		// Use all fields in the order returned by Serialize()
		for _, sf := range serializedFields {
			headers = append(headers, sf.Name)
			row = append(row, formatSerializedValue(sf.Value))
		}
	}

	return row, headers
}

// formatSerializedValue formats a serialized field value as a string
func formatSerializedValue(val interface{}) string {
	if val == nil {
		return ""
	}

	v := reflect.ValueOf(val)
	return formatValue(v)
}

// getHeaderName gets the display name for a struct field
func getHeaderName(field *reflect.StructField) string {
	// Try to get from json tag first
	jsonTag := field.Tag.Get("json")
	if jsonTag != "" && jsonTag != "-" {
		tagParts := strings.Split(jsonTag, ",")
		return tagParts[0]
	}

	// Use field name
	return field.Name
}

// formatValue formats a reflect.Value as a string
func formatValue(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}

	// Dereference pointer
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	// Handle different types
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return fmt.Sprintf("%t", v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", v.Float())
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return ""
		}
		// For slices, join elements
		var parts []string
		for i := 0; i < v.Len(); i++ {
			parts = append(parts, formatValue(v.Index(i)))
		}
		return strings.Join(parts, ", ")
	case reflect.Map, reflect.Struct:
		// For complex types, use JSON
		data, _ := json.Marshal(v.Interface())
		return string(data)
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}
