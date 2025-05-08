package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"

	"github.com/scttfrdmn/snoozebot/pkg/plugin"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ifacecheck <file.go>")
		os.Exit(1)
	}

	filepath := os.Args[1]
	
	// Parse the Go file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filepath, nil, 0)
	if err != nil {
		fmt.Printf("Error parsing file: %v\n", err)
		os.Exit(1)
	}

	// Get all CloudProvider interface methods
	interfaceMethods := getAllInterfaceMethods(reflect.TypeOf((*plugin.CloudProvider)(nil)).Elem())
	
	// Check if all provider types implement CloudProvider
	errors := checkProviderImplementations(file, interfaceMethods)
	
	if len(errors) > 0 {
		fmt.Println("Interface implementation errors found:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
		os.Exit(1)
	}
	
	fmt.Println("All provider types correctly implement the CloudProvider interface")
}

// getAllInterfaceMethods returns all method names from an interface type
func getAllInterfaceMethods(interfaceType reflect.Type) []string {
	var methods []string
	for i := 0; i < interfaceType.NumMethod(); i++ {
		methods = append(methods, interfaceType.Method(i).Name)
	}
	return methods
}

// checkProviderImplementations checks if all provider types in the file
// implement all methods of the CloudProvider interface
func checkProviderImplementations(file *ast.File, interfaceMethods []string) []string {
	var errors []string
	
	// Find all struct types with "Provider" in the name
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			
			typeName := typeSpec.Name.Name
			if !strings.Contains(typeName, "Provider") {
				continue
			}
			
			// Found a provider type, now check if it implements all methods
			methodsImplemented := findMethodsForType(file, typeName)
			missingMethods := findMissingMethods(interfaceMethods, methodsImplemented)
			
			if len(missingMethods) > 0 {
				errors = append(errors, fmt.Sprintf(
					"Type %s is missing methods: %s",
					typeName,
					strings.Join(missingMethods, ", "),
				))
			}
		}
	}
	
	return errors
}

// findMethodsForType finds all methods implemented by a type
func findMethodsForType(file *ast.File, typeName string) []string {
	var methods []string
	
	receiverName := fmt.Sprintf("*%s", typeName)
	
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil {
			continue
		}
		
		// Check if this function is a method of our type
		for _, receiver := range funcDecl.Recv.List {
			var receiverType string
			
			// Get the receiver type as a string
			switch expr := receiver.Type.(type) {
			case *ast.StarExpr:
				if ident, ok := expr.X.(*ast.Ident); ok {
					receiverType = fmt.Sprintf("*%s", ident.Name)
				}
			case *ast.Ident:
				receiverType = expr.Name
			}
			
			if receiverType == receiverName {
				methods = append(methods, funcDecl.Name.Name)
			}
		}
	}
	
	return methods
}

// findMissingMethods finds methods that are in required but not in implemented
func findMissingMethods(required, implemented []string) []string {
	var missing []string
	
	implementedMap := make(map[string]bool)
	for _, m := range implemented {
		implementedMap[m] = true
	}
	
	for _, m := range required {
		if !implementedMap[m] {
			missing = append(missing, m)
		}
	}
	
	return missing
}