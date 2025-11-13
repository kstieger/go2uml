package main

import (
	"strings"
	"testing"
)

func TestConvertToMermaid(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: `classDiagram`,
		},
		{
			name: "simple class",
			input: `@startuml
class "User" << (S,Aquamarine) >> {
    + ID int
    + Name string
}
@enduml`,
			expected: `classDiagram
    class User {
        <<struct>>
        +ID int
        +Name string
    }`,
		},
		{
			name: "simple interface",
			input: `@startuml
interface "UserService" {
    + GetUser(id int) error
    + CreateUser(user User) error
}
@enduml`,
			expected: `classDiagram
    class UserService {
        <<interface>>
        +GetUser(id int) error
        +CreateUser(user User) error
    }`,
		},
		{
			name: "class with namespace",
			input: `@startuml
namespace example {
    class "User" << (S,Aquamarine) >> {
        + ID int
        + Name string
    }
}
@enduml`,
			expected: `classDiagram
    class User {
        <<struct>>
        +ID int
        +Name string
    }`,
		},
		{
			name: "interface with namespace",
			input: `@startuml
namespace example {
    interface "UserService" {
        + GetUser(id int) error
    }
}
@enduml`,
			expected: `classDiagram
    class UserService {
        <<interface>>
        +GetUser(id int) error
    }`,
		},
		{
			name: "inheritance relationship",
			input: `@startuml
namespace example {
    interface "UserService" {
        + GetUser(id int) error
    }
    class "DatabaseUserService" << (S,Aquamarine) >> {
        + GetUser(id int) error
    }
}
"example.UserService" <|-- "example.DatabaseUserService"
@enduml`,
			expected: `classDiagram
    class UserService {
        <<interface>>
        +GetUser(id int) error
    }
    class DatabaseUserService {
        <<struct>>
        +GetUser(id int) error
    }
    DatabaseUserService --|> UserService`,
		},
		{
			name: "composition relationship",
			input: `@startuml
namespace example {
    class "User" << (S,Aquamarine) >> {
        + ID int
    }
    class "Profile" << (S,Aquamarine) >> {
        + Bio string
    }
}
"example.User" *-- "example.Profile"
@enduml`,
			expected: `classDiagram
    class User {
        <<struct>>
        +ID int
    }
    class Profile {
        <<struct>>
        +Bio string
    }
    User *-- Profile`,
		},
		{
			name: "multiple field types",
			input: `@startuml
class "TestClass" << (S,Aquamarine) >> {
    + PublicField string
    - privateField int
    # protectedField bool
}
@enduml`,
			expected: `classDiagram
    class TestClass {
        <<struct>>
        +PublicField string
        -privateField int
        #protectedField bool
    }`,
		},
		{
			name: "generic type parameters",
			input: `@startuml
interface "Generic" << [T, K] >> {
    + Process(t T) K
}
@enduml`,
			expected: `classDiagram
    class Generic {
        <<interface>>
        +Process(t T) K
    }`,
		},
		{
			name: "enum stereotype",
			input: `@startuml
class "Status" << (E,Yellow) >> {
    + ACTIVE
    + INACTIVE
}
@enduml`,
			expected: `classDiagram
    class Status {
        <<enum>>
        +ACTIVE
        +INACTIVE
    }`,
		},
		{
			name: "html color tags removal",
			input: `@startuml
class "User" << (S,Aquamarine) >> {
    + Data <font color=blue>map</font>[string]interface{}
}
@enduml`,
			expected: `classDiagram
    class User {
        <<struct>>
        +Data map[string]interface{}
    }`,
		},
		{
			name: "constraints handling",
			input: `@startuml
class "T" <<type parameter>> {
    constraints: Comparable
}
@enduml`,
			expected: `classDiagram
    class T {
        <<type parameter>>
    }`,
		},
		{
			name: "dependency relationship",
			input: `@startuml
class "Client" << (S,Aquamarine) >> {
}
class "Service" << (S,Aquamarine) >> {
}
"Client" <-- "Service"
@enduml`,
			expected: `classDiagram
    class Client {
        <<struct>>
    }
    class Service {
        <<struct>>
    }
    Client <-- Service`,
		},
		{
			name: "association relationship",
			input: `@startuml
class "User" << (S,Aquamarine) >> {
}
class "Group" << (S,Aquamarine) >> {
}
"User" -- "Group"
@enduml`,
			expected: `classDiagram
    class User {
        <<struct>>
    }
    class Group {
        <<struct>>
    }
    User -- Group`,
		},
		{
			name: "multiple namespaces and complex relationships",
			input: `@startuml
namespace api {
    interface "Handler" {
        + Handle() error
    }
}
namespace impl {
    class "UserHandler" << (S,Aquamarine) >> {
        + Handle() error
    }
}
"api.Handler" <|-- "impl.UserHandler"
@enduml`,
			expected: `classDiagram
    class Handler {
        <<interface>>
        +Handle() error
    }
    class UserHandler {
        <<struct>>
        +Handle() error
    }
    UserHandler --|> Handler`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertToMermaid(tt.input)
			if err != nil {
				t.Fatalf("ConvertToMermaid() error = %v", err)
			}

			// Normalize whitespace for comparison
			expected := strings.TrimSpace(tt.expected)
			actual := strings.TrimSpace(result)

			if actual != expected {
				t.Errorf("ConvertToMermaid() mismatch:\nExpected:\n%s\n\nActual:\n%s", expected, actual)
			}
		})
	}
}

// TestConvertToMermaidErrors tests error scenarios
func TestConvertToMermaidErrors(t *testing.T) {
	// Test with malformed PlantUML that might cause issues
	plantUML := `this is not valid plantuml`

	result, err := ConvertToMermaid(plantUML)
	if err != nil {
		t.Fatalf("ConvertToMermaid() should not return error for malformed input, got: %v", err)
	}

	// Should return minimal Mermaid syntax even for malformed input
	expected := "classDiagram"
	if strings.TrimSpace(result) != expected {
		t.Errorf("ConvertToMermaid() for malformed input = %v, want %v", result, expected)
	}
}

func TestExtractClassName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "quoted class name",
			input:    `class "User" << (S,Aquamarine) >> {`,
			expected: "User",
		},
		{
			name:     "quoted interface name",
			input:    `interface "UserService" {`,
			expected: "UserService",
		},
		{
			name:     "class with alias",
			input:    `class "ResponseBuilder" as ResponseBuilder_generic_D_I << [D, I] >> {`,
			expected: "ResponseBuilder",
		},
		{
			name:     "interface with alias and generics",
			input:    `interface "AllowedResponseTypes" as AllowedResponseTypes_generic_D_I << [D, I] >> {`,
			expected: "AllowedResponseTypes",
		},
		{
			name:     "unquoted class name",
			input:    `class User {`,
			expected: "User",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "malformed input",
			input:    `something else`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractClassName(tt.input)
			if result != tt.expected {
				t.Errorf("extractClassName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractStereotype(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "struct stereotype",
			input:    `class "User" << (S,Aquamarine) >> {`,
			expected: "struct",
		},
		{
			name:     "interface stereotype",
			input:    `class "Handler" << (I,Blue) >> {`,
			expected: "interface",
		},
		{
			name:     "enum stereotype",
			input:    `class "Status" << (E,Yellow) >> {`,
			expected: "enum",
		},
		{
			name:     "generic type parameters",
			input:    `interface "Generic" << [T, K] >> {`,
			expected: "generic: T  K",
		},
		{
			name:     "custom stereotype",
			input:    `class "Custom" << custom >> {`,
			expected: "custom",
		},
		{
			name:     "type parameter stereotype",
			input:    `class "T" << type parameter >> {`,
			expected: "type parameter",
		},
		{
			name:     "no stereotype",
			input:    `class "User" {`,
			expected: "",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractStereotype(tt.input)
			if result != tt.expected {
				t.Errorf("extractStereotype() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCleanClassName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "User",
			expected: "User",
		},
		{
			name:     "namespaced name",
			input:    "example.User",
			expected: "example_User",
		},
		{
			name:     "quoted name",
			input:    "\"User\"",
			expected: "User",
		},
		{
			name:     "name with generics",
			input:    "Generic<T,K>",
			expected: "Generic_T_K_",
		},
		{
			name:     "name with brackets",
			input:    "Array[int]",
			expected: "Array_int_",
		},
		{
			name:     "name with spaces",
			input:    "My Class Name",
			expected: "My_Class_Name",
		},
		{
			name:     "name with parentheses",
			input:    "Function(int,string)",
			expected: "Function_int_string_",
		},
		{
			name:     "complex name",
			input:    "\"example.Generic<T,K>[Value]\"",
			expected: "example_Generic_T_K__Value_",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanClassName(tt.input)
			if result != tt.expected {
				t.Errorf("cleanClassName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertFieldOrMethod(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "public field",
			input:    "+ ID int",
			expected: "+ID int",
		},
		{
			name:     "private field",
			input:    "- db interface{}",
			expected: "-db interface{}",
		},
		{
			name:     "protected field",
			input:    "# config Config",
			expected: "#config Config",
		},
		{
			name:     "public method",
			input:    "+ GetUser(id int) (*User, error)",
			expected: "+GetUser(id int) (*User, error)",
		},
		{
			name:     "field with html color tags",
			input:    "+ Data <font color=blue>map</font>[string]interface{}",
			expected: "+Data map[string]interface{}",
		},
		{
			name:     "complex field with multiple tags",
			input:    "- internal <font color=red>chan</font> <font color=blue>struct</font>{}",
			expected: "-internal <font color=red>chan struct{}",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "malformed input",
			input:    "not a field",
			expected: "not a field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertFieldOrMethod(tt.input)
			if result != tt.expected {
				t.Errorf("convertFieldOrMethod() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertRelationshipWithMapping(t *testing.T) {
	classNameMapping := map[string]string{
		"example_UserService":         "UserService",
		"example_DatabaseUserService": "DatabaseUserService",
		"example_User":                "User",
		"example_Profile":             "Profile",
		"api_Handler":                 "Handler",
		"impl_UserHandler":            "UserHandler",
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "inheritance relationship",
			input:    `"example.UserService" <|-- "example.DatabaseUserService"`,
			expected: "DatabaseUserService --|> UserService",
		},
		{
			name:     "composition relationship",
			input:    `"example.User" *-- "example.Profile"`,
			expected: "User *-- Profile",
		},
		{
			name:     "dependency relationship",
			input:    `"example.User" <-- "example.DatabaseUserService"`,
			expected: "User <-- DatabaseUserService",
		},
		{
			name:     "association relationship",
			input:    `"example.User" -- "example.Profile"`,
			expected: "User -- Profile",
		},
		{
			name:     "cross-namespace inheritance",
			input:    `"api.Handler" <|-- "impl.UserHandler"`,
			expected: "UserHandler --|> Handler",
		},
		{
			name:     "unknown classes fallback",
			input:    `"unknown.ClassA" <|-- "unknown.ClassB"`,
			expected: "unknown_ClassB --|> unknown_ClassA",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "malformed relationship",
			input:    "not a relationship",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertRelationshipWithMapping(tt.input, classNameMapping)
			if result != tt.expected {
				t.Errorf("convertRelationshipWithMapping() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkConvertToMermaid(b *testing.B) {
	plantUML := `@startuml
namespace example {
    interface "UserService" {
        + GetUser(id int) (*User, error)
        + CreateUser(user *User) error
        + UpdateUser(id int, user *User) error
        + DeleteUser(id int) error
    }
    class "User" << (S,Aquamarine) >> {
        + ID int
        + Name string
        + Email string
        + CreatedAt time.Time
        + UpdatedAt time.Time
    }
    class "DatabaseUserService" << (S,Aquamarine) >> {
        - db <font color=blue>interface</font>{}
        + GetUser(id int) (*User, error)
        + CreateUser(user *User) error
        + UpdateUser(id int, user *User) error
        + DeleteUser(id int) error
    }
}
"example.UserService" <|-- "example.DatabaseUserService"
@enduml`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ConvertToMermaid(plantUML)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExtractClassName(b *testing.B) {
	testLine := `class "DatabaseUserService" << (S,Aquamarine) >> {`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractClassName(testLine)
	}
}

func BenchmarkCleanClassName(b *testing.B) {
	testName := "example.Generic<T,K>[Value]"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cleanClassName(testName)
	}
}
