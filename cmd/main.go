package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	goplantuml "github.com/jfeliu007/goplantuml/parser"
)

// RenderingOptionSlice will implements the sort interface
type RenderingOptionSlice []goplantuml.RenderingOption

// Len is the number of elements in the collection.
func (as RenderingOptionSlice) Len() int {
	return len(as)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (as RenderingOptionSlice) Less(i, j int) bool {
	return as[i] < as[j]
}

// Swap swaps the elements with indexes i and j.
func (as RenderingOptionSlice) Swap(i, j int) {
	as[i], as[j] = as[j], as[i]
}

func main() {
	recursive := flag.Bool("recursive", false, "walk all directories recursively")
	ignore := flag.String("ignore", "", "comma separated list of folders to ignore")
	maxDepth := flag.Int("max-depth", 0, "maximum nesting depth for packages (0 = unlimited)")
	showAggregations := flag.Bool(
		"show-aggregations",
		false,
		"renders public aggregations even when -hide-connections is used (do not render by default)",
	)
	hideFields := flag.Bool("hide-fields", false, "hides fields")
	hideMethods := flag.Bool("hide-methods", false, "hides methods")
	hideConnections := flag.Bool("hide-connections", false, "hides all connections in the diagram")
	showCompositions := flag.Bool("show-compositions", false, "Shows compositions even when -hide-connections is used")
	showImplementations := flag.Bool(
		"show-implementations",
		false,
		"Shows implementations even when -hide-connections is used",
	)
	showAliases := flag.Bool("show-aliases", false, "Shows aliases even when -hide-connections is used")
	showConnectionLabels := flag.Bool(
		"show-connection-labels",
		false,
		"Shows labels in the connections to identify the connections types (e.g. extends, implements, aggregates, alias of",
	)
	title := flag.String("title", "", "Title of the generated diagram")
	notes := flag.String("notes", "", "Comma separated list of notes to be added to the diagram")
	output := flag.String("output", "", "output file path. If omitted, then this will default to standard output")
	showOptionsAsNote := flag.Bool(
		"show-options-as-note",
		false,
		"Show a note in the diagram with the none evident options ran with this CLI",
	)
	aggregatePrivateMembers := flag.Bool(
		"aggregate-private-members",
		false,
		"Show aggregations for private members. Ignored if -show-aggregations is not used.",
	)
	hidePrivateMembers := flag.Bool("hide-private-members", false, "Hide private fields and methods")
	format := flag.String("format", "plantuml", "output format: plantuml or mermaid (mermaid support is experimental)")
	flag.Parse()
	renderingOptions := map[goplantuml.RenderingOption]any{
		goplantuml.RenderConnectionLabels:  *showConnectionLabels,
		goplantuml.RenderFields:            !*hideFields,
		goplantuml.RenderMethods:           !*hideMethods,
		goplantuml.RenderAggregations:      *showAggregations,
		goplantuml.RenderTitle:             *title,
		goplantuml.AggregatePrivateMembers: *aggregatePrivateMembers,
		goplantuml.RenderPrivateMembers:    !*hidePrivateMembers,
	}
	if *hideConnections {
		renderingOptions[goplantuml.RenderAliases] = *showAliases
		renderingOptions[goplantuml.RenderCompositions] = *showCompositions
		renderingOptions[goplantuml.RenderImplementations] = *showImplementations

	}
	noteList := []string{}
	if *showOptionsAsNote {
		legend, err := getLegend(renderingOptions)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		noteList = append(noteList, legend)
	}
	if *notes != "" {
		noteList = append(noteList, "", "<b><u>Notes</u></b>")
	}
	split := strings.Split(*notes, ",")
	for _, note := range split {
		trimmed := strings.TrimSpace(note)
		if trimmed != "" {
			noteList = append(noteList, trimmed)
		}
	}
	renderingOptions[goplantuml.RenderNotes] = strings.Join(noteList, "\n")
	dirs, err := getDirectories()

	if err != nil {
		slog.Error("DIR Must be a valid directory", "usage", "goplantuml <DIR>", "error", err)
		os.Exit(1)
	}
	ignoredDirectories, err := getIgnoredDirectories(*ignore)
	if err != nil {

		slog.Error(
			"DIRLIST Must be a valid comma separated list of existing directories",
			"usage",
			"goplantuml [-ignore=<DIRLIST>]",
			"error",
			err,
		)
		os.Exit(1)
	}

	result, err := goplantuml.NewClassDiagramWithMaxDepth(dirs, ignoredDirectories, *recursive, *maxDepth)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	if result == nil {
		fmt.Fprintln(os.Stderr, "No classes found to generate diagram")
		os.Exit(1)
	}
	_ = result.SetRenderingOptions(renderingOptions)

	rendered := result.Render()
	switch strings.ToLower(*format) {
	case "plantuml":
		// do nothing, plantuml is the default
	case "mermaid":
		rendered, err = ConvertToMermaid(rendered)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	default:
		fmt.Println("usage:\ngoplantuml [-format=plantuml|mermaid]\nformat must be plantuml or mermaid")
		fmt.Fprintln(os.Stderr, "format must be plantuml or mermaid")
		os.Exit(1)
	}

	var writer io.Writer
	if *output != "" {
		writer, err = os.Create(*output)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	} else {
		writer = os.Stdout
	}
	_, _ = fmt.Fprint(writer, rendered)
}

func getDirectories() ([]string, error) {

	args := flag.Args()
	if len(args) < 1 {
		return nil, errors.New("DIR missing")
	}
	dirs := []string{}
	for _, dir := range args {
		fi, err := os.Stat(dir)
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("could not find directory %s", dir)
		}
		if !fi.Mode().IsDir() {
			return nil, fmt.Errorf("%s is not a directory", dir)
		}
		dirAbs, err := filepath.Abs(dir)
		if err != nil {
			return nil, fmt.Errorf("could not find directory %s", dir)
		}
		dirs = append(dirs, dirAbs)
	}
	return dirs, nil
}

func getIgnoredDirectories(list string) ([]string, error) {
	result := []string{}
	list = strings.TrimSpace(list)
	if list == "" {
		return result, nil
	}
	split := strings.Split(list, ",")
	for _, dir := range split {
		dirAbs, err := filepath.Abs(strings.TrimSpace(dir))
		if err != nil {
			return nil, fmt.Errorf("could not find directory %s", dir)
		}
		result = append(result, dirAbs)
	}
	return result, nil
}

func getLegend(ro map[goplantuml.RenderingOption]any) (string, error) {
	result := "<u><b>Legend</b></u>\n"
	orderedOptions := RenderingOptionSlice{}
	for o := range ro {
		orderedOptions = append(orderedOptions, o)
	}
	sort.Sort(orderedOptions)
	for _, option := range orderedOptions {
		val := ro[option]
		switch option {
		case goplantuml.RenderAggregations:
			result = fmt.Sprintf("%sRender Aggregations: %t\n", result, val.(bool))
		case goplantuml.RenderAliases:
			result = fmt.Sprintf("%sRender Connections: %t\n", result, val.(bool))
		case goplantuml.RenderCompositions:
			result = fmt.Sprintf("%sRender Compositions: %t\n", result, val.(bool))
		case goplantuml.RenderFields:
			result = fmt.Sprintf("%sRender Fields: %t\n", result, val.(bool))
		case goplantuml.RenderImplementations:
			result = fmt.Sprintf("%sRender Implementations: %t\n", result, val.(bool))
		case goplantuml.RenderMethods:
			result = fmt.Sprintf("%sRender Methods: %t\n", result, val.(bool))
		case goplantuml.AggregatePrivateMembers:
			result = fmt.Sprintf("%sPrivate Aggregations: %t\n", result, val.(bool))
		}
	}
	return strings.TrimSpace(result), nil
}

// ConvertToMermaid converts a PlantUML diagram string to a Mermaid diagram string
func ConvertToMermaid(plantUML string) (string, error) {
	lines := strings.Split(plantUML, "\n")
	var mermaidLines []string

	// Start with classDiagram
	mermaidLines = append(mermaidLines, "classDiagram")

	// Track classes and interfaces for relationship processing
	classTypes := make(map[string]string)       // className -> "class" or "interface"
	classNameMapping := make(map[string]string) // full name -> simple name
	insideClass := false
	currentNamespace := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "<<") && strings.Contains(line, ">>") {
			start := strings.Index(line, "<<")
			end := strings.Index(line, ">>")
			if start >= 0 && end > start {
				stereotype := strings.TrimSpace(line[start+2 : end])
				if strings.HasPrefix(stereotype, "S,") {
					line = line[:start] + line[end+2:]
				}
			}
		}

		// Skip PlantUML directives
		if strings.HasPrefix(line, "@startuml") || strings.HasPrefix(line, "@enduml") {
			continue
		}

		// Handle namespace
		if strings.HasPrefix(line, "namespace ") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				currentNamespace = strings.TrimSuffix(parts[1], " {")
			}
			continue
		}

		// Handle interface definitions
		if strings.Contains(line, "interface ") && strings.Contains(line, " {") {
			interfaceName := extractClassName(line)
			if interfaceName != "" {
				// Clean interface name from quotes and generics
				cleanName := cleanClassName(interfaceName)
				fullName := currentNamespace + "." + interfaceName
				classTypes[cleanName] = "interface"
				classNameMapping[cleanClassName(fullName)] = cleanName
				mermaidLines = append(mermaidLines, fmt.Sprintf("    class %s {", cleanName))
				mermaidLines = append(mermaidLines, "        <<interface>>")
				insideClass = true
			}
			continue
		}

		// Handle class definitions
		if strings.Contains(line, "class ") && strings.Contains(line, " {") {
			className := extractClassName(line)
			if className != "" {
				cleanName := cleanClassName(className)
				fullName := currentNamespace + "." + className
				classTypes[cleanName] = "class"
				classNameMapping[cleanClassName(fullName)] = cleanName

				// Check for stereotypes
				stereotype := extractStereotype(line)
				mermaidLines = append(mermaidLines, fmt.Sprintf("    class %s {", cleanName))
				if stereotype != "" {
					mermaidLines = append(mermaidLines, fmt.Sprintf("        <<%s>>", stereotype))
				}
				insideClass = true
			}
			continue
		}

		// Handle fields and methods inside class/interface definitions
		if insideClass &&
			(strings.HasPrefix(line, "+ ") || strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "# ")) {
			field := convertFieldOrMethod(line)
			if field != "" {
				mermaidLines = append(mermaidLines, fmt.Sprintf("        %s", field))
			}
			continue
		}

		// Handle constraints lines (for generic type parameters)
		if strings.Contains(line, "constraints:") {
			// Skip constraints in Mermaid as they don't have direct equivalent
			continue
		}

		// Handle closing braces
		if line == "}" {
			if insideClass {
				mermaidLines = append(mermaidLines, "    }")
				insideClass = false
			} else if currentNamespace != "" {
				// Exiting namespace
				currentNamespace = ""
			}
			continue
		}

		// Handle relationships (outside of class definitions)
		if !insideClass && (strings.Contains(line, "<|--") || strings.Contains(line, "*--") ||
			strings.Contains(line, "<--") || strings.Contains(line, "--") ||
			strings.Contains(line, "<..") || strings.Contains(line, "..>")) {
			relationship := convertRelationshipWithMapping(line, classNameMapping)
			if relationship != "" {
				mermaidLines = append(mermaidLines, fmt.Sprintf("    %s", relationship))
			}
			continue
		}
	}

	return strings.Join(mermaidLines, "\n"), nil
}

// extractClassName extracts the class name from a class or interface definition line
func extractClassName(line string) string {
	// Handle various patterns like:
	// class "ErrorObject" << (S,Aquamarine) >> {
	// interface "AllowedResponseTypes" as AllowedResponseTypes_generic_D_I <<[D, I]>> {
	// class "ResponseBuilder" as ResponseBuilder_generic_D_I <<[D, I]>> {

	if strings.Contains(line, "interface ") {
		parts := strings.Split(line, "interface ")
		if len(parts) > 1 {
			remaining := strings.TrimSpace(parts[1])
			if strings.HasPrefix(remaining, "\"") {
				// Extract quoted name
				end := strings.Index(remaining[1:], "\"")
				if end > 0 {
					return remaining[1 : end+1]
				}
			} else {
				// Extract unquoted name
				parts := strings.Fields(remaining)
				if len(parts) > 0 {
					return parts[0]
				}
			}
		}
	}

	if strings.Contains(line, "class ") {
		parts := strings.Split(line, "class ")
		if len(parts) > 1 {
			remaining := strings.TrimSpace(parts[1])
			if strings.HasPrefix(remaining, "\"") {
				// Extract quoted name
				end := strings.Index(remaining[1:], "\"")
				if end > 0 {
					return remaining[1 : end+1]
				}
			} else {
				// Extract unquoted name
				parts := strings.Fields(remaining)
				if len(parts) > 0 {
					return parts[0]
				}
			}
		}
	}

	return ""
}

// cleanClassName removes problematic characters from class names for Mermaid
func cleanClassName(name string) string {
	// Remove quotes, dots, and replace problematic characters
	name = strings.ReplaceAll(name, "\"", "")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "[", "_")
	name = strings.ReplaceAll(name, "]", "_")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, ",", "_")
	name = strings.ReplaceAll(name, "(", "_")
	name = strings.ReplaceAll(name, ")", "_")
	return name
}

// extractStereotype extracts stereotype information from class definition
func extractStereotype(line string) string {
	// Look for patterns like << (S,Aquamarine) >> or <<[D, I]>>
	start := strings.Index(line, "<<")
	end := strings.Index(line, ">>")

	if start >= 0 && end > start {
		stereotype := strings.TrimSpace(line[start+2 : end])

		// Handle different stereotype patterns for Mermaid compatibility
		if strings.Contains(stereotype, "(") && strings.Contains(stereotype, ")") {
			// Pattern like (S,Aquamarine) - extract only the meaningful part
			stereotype = strings.ReplaceAll(stereotype, "(", "")
			stereotype = strings.ReplaceAll(stereotype, ")", "")
			parts := strings.Split(stereotype, ",")
			if len(parts) > 0 {
				firstPart := strings.TrimSpace(parts[0])
				// For PlantUML class stereotypes like "S", convert to more descriptive Mermaid stereotype
				switch firstPart {
				case "S":
					return "struct"
				case "I":
					return "interface"
				case "E":
					return "enum"
				default:
					return firstPart
				}
			}
		} else if strings.Contains(stereotype, "[") && strings.Contains(stereotype, "]") {
			// Pattern like [D, I] - generic type parameters, keep as is but clean up
			stereotype = strings.ReplaceAll(stereotype, "[", "")
			stereotype = strings.ReplaceAll(stereotype, "]", "")
			stereotype = strings.ReplaceAll(stereotype, ",", " ")
			return "generic: " + strings.TrimSpace(stereotype)
		} else {
			// Simple stereotype without special characters
			return strings.TrimSpace(stereotype)
		}
	}

	return ""
}

// convertFieldOrMethod converts PlantUML field/method syntax to Mermaid
func convertFieldOrMethod(line string) string {
	// Remove leading + - # symbols and convert to Mermaid syntax
	line = strings.TrimSpace(line)

	// Remove HTML color tags
	line = strings.ReplaceAll(line, "<font color=blue>", "")
	line = strings.ReplaceAll(line, "</font>", "")

	if strings.HasPrefix(line, "+ ") {
		return "+" + strings.TrimSpace(line[2:])
	} else if strings.HasPrefix(line, "- ") {
		return "-" + strings.TrimSpace(line[2:])
	} else if strings.HasPrefix(line, "# ") {
		return "#" + strings.TrimSpace(line[2:])
	}

	return line
}

// convertRelationshipWithMapping converts PlantUML relationships to Mermaid syntax using class name mapping
func convertRelationshipWithMapping(line string, classNameMapping map[string]string) string {
	line = strings.TrimSpace(line)

	// Handle inheritance: A <|-- B becomes B --|> A
	if strings.Contains(line, "<|--") {
		parts := strings.Split(line, "<|--")
		if len(parts) == 2 {
			parentUnquoted := strings.TrimSpace(strings.ReplaceAll(parts[0], "\"", ""))
			childUnquoted := strings.TrimSpace(strings.ReplaceAll(parts[1], "\"", ""))

			parentFull := cleanClassName(parentUnquoted)
			childFull := cleanClassName(childUnquoted)

			// Map to simple names if available
			parent := classNameMapping[parentFull]
			child := classNameMapping[childFull]
			if parent == "" {
				parent = parentFull
			}
			if child == "" {
				child = childFull
			}

			return fmt.Sprintf("%s --|> %s", child, parent)
		}
	}

	// Handle composition: A *-- B becomes A *-- B
	if strings.Contains(line, "*--") {
		parts := strings.Split(line, "*--")
		if len(parts) == 2 {
			leftUnquoted := strings.TrimSpace(strings.ReplaceAll(parts[0], "\"", ""))
			rightUnquoted := strings.TrimSpace(strings.ReplaceAll(parts[1], "\"", ""))

			leftFull := cleanClassName(leftUnquoted)
			rightFull := cleanClassName(rightUnquoted)

			left := classNameMapping[leftFull]
			right := classNameMapping[rightFull]
			if left == "" {
				left = leftFull
			}
			if right == "" {
				right = rightFull
			}

			return fmt.Sprintf("%s *-- %s", left, right)
		}
	}

	// Handle dependency: A <-- B becomes A <-- B
	if strings.Contains(line, "<--") && !strings.Contains(line, "<|--") {
		parts := strings.Split(line, "<--")
		if len(parts) == 2 {
			leftUnquoted := strings.TrimSpace(strings.ReplaceAll(parts[0], "\"", ""))
			rightUnquoted := strings.TrimSpace(strings.ReplaceAll(parts[1], "\"", ""))

			leftFull := cleanClassName(leftUnquoted)
			rightFull := cleanClassName(rightUnquoted)

			left := classNameMapping[leftFull]
			right := classNameMapping[rightFull]
			if left == "" {
				left = leftFull
			}
			if right == "" {
				right = rightFull
			}

			return fmt.Sprintf("%s <-- %s", left, right)
		}
	}

	// Handle association: A -- B becomes A -- B
	if strings.Contains(line, "--") && !strings.Contains(line, "<--") && !strings.Contains(line, "*--") &&
		!strings.Contains(line, "<|--") {
		parts := strings.Split(line, "--")
		if len(parts) == 2 {
			leftUnquoted := strings.TrimSpace(strings.ReplaceAll(parts[0], "\"", ""))
			rightUnquoted := strings.TrimSpace(strings.ReplaceAll(parts[1], "\"", ""))

			leftFull := cleanClassName(leftUnquoted)
			rightFull := cleanClassName(rightUnquoted)

			left := classNameMapping[leftFull]
			right := classNameMapping[rightFull]
			if left == "" {
				left = leftFull
			}
			if right == "" {
				right = rightFull
			}

			return fmt.Sprintf("%s -- %s", left, right)
		}
	}

	return ""
}
