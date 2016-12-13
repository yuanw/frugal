package java

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/Workiva/frugal/compiler/generator"
	"github.com/Workiva/frugal/compiler/parser"
	"github.com/Workiva/frugal/compiler/plugin"
)

const (
	generateEnumValue        = "GenerateEnumValue"
	generateEnumFields       = "GenerateEnumFields"
	generateEnumConstructors = "GenerateEnumConstructors"
	generateEnumMethods      = "GenerateEnumMethods"
)

// hooks contains the symbols to load from plugins.
var hooks = []string{
	generateEnumValue,
	generateEnumFields,
	generateEnumConstructors,
	generateEnumMethods,
}

// Behavior exposes hooks for augmenting Java code generation behavior. Frugal
// provides a default behavior and users can provide additional behavior
// through plugins. If a Behavior hook is implemented in a plugin, Frugal will
// call it. Otherwise it will fall back to the default Behavior.
//
// Since multiple plugins can be provided, some hooks may be called more than
// once. Hooks which can only be called once are documented. If multiple
// plugins implement these hooks, precedence is determined by the order in
// which the plugins are provided. For example, if plugins A, B, and C all
// implement GenerateEnumValue, only A's will be invoked.
type Behavior interface {
	// GenerateEnumValue is called when generating enum values, e.g.
	//
	// public enum MyEnum implements TEnum {
	//     FOO(1), // produced by a call to GenerateEnumValue
	//     BAR(2),
	//     BAZ(3);
	//
	//     ...
	//
	// If multiple plugins implement this hook, only one will be invoked. Note
	// that Frugal requires the int value for wire-level compatibility.
	GenerateEnumValue(enum *parser.Enum, value *parser.EnumValue) (EnumValue, error)

	// GenerateEnumFields is called when generating enum fields, e.g.
	//
	// private final int value; // produced by a call to GenerateEnumFields
	//
	// Note that Frugal requires the int value for wire-level compatibility.
	GenerateEnumFields(enum *parser.Enum) ([]Field, error)

	// GenerateEnumConstructors is called when generating enum constructors,
	// e.g.
	//
	// MyEnum(int value) {      // produced by
	//     this.value = value;  // a call to
	// }                        // GenerateEnumConstructors
	//
	// Note that Frugal requires the int value for wire-level compatibility.
	GenerateEnumConstructors(enum *parser.Enum) ([]EnumConstructor, error)

	// GenerateEnumMethods is called when generating enum methods, e.g.
	//
	// public int getValue() {
	//     return value;
	// }
	//
	// Note that Frugal requires the int value for wire-level compatibility.
	GenerateEnumMethods(enum *parser.Enum) ([]Method, error)
}

// defaultBehavior implements the Behavior interface and provides default
// generate behavior for Java.
type defaultBehavior struct{}

// GenerateEnumValue is called when generating enum values, i.e.
//
// public enum MyEnum implements TEnum {
//     FOO(1), // produced by a call to GenerateEnumValue
//     BAR(2),
//     BAZ(3);
//
//     ...
//
// If multiple plugins implement this hook, only one will be invoked. Note that
// Frugal requires the int value for wire-level compatibility.
func (d *defaultBehavior) GenerateEnumValue(enum *parser.Enum, value *parser.EnumValue) (EnumValue, error) {
	// FOO(1), BAR(2), etc.
	return EnumValue{Name: value.Name, Arguments: []string{strconv.Itoa(value.Value)}}, nil
}

// GenerateEnumFields is called when generating enum fields, e.g.
//
// private final int value; // produced by a call to GenerateEnumFields
//
// Note that Frugal requires the int value for wire-level compatibility.
func (d *defaultBehavior) GenerateEnumFields(enum *parser.Enum) ([]Field, error) {
	// private final int value
	value := Field{
		Modifier: "private",
		Final:    true,
		Type:     "int",
		Name:     "value",
	}
	return []Field{value}, nil
}

// GenerateEnumConstructors is called when generating enum constructors, e.g.
//
// MyEnum(int value) {      // produced by
//     this.value = value;  // a call to
// }                        // GenerateEnumConstructors
//
// Note that Frugal requires the int value for wire-level compatibility.
func (d *defaultBehavior) GenerateEnumConstructors(enum *parser.Enum) ([]EnumConstructor, error) {
	// Default value constructor.
	printer := new(generator.Printer)
	printer.Println("this.value = value;")
	constructor := EnumConstructor{
		Modifier:  "private",
		Arguments: []Argument{Argument{Type: "int", Name: "value"}},
		Body:      printer,
	}
	return []EnumConstructor{constructor}, nil
}

// GenerateEnumMethods is called when generating enum methods, e.g.
//
// public int getValue() {
//     return value;
// }
//
// Note that Frugal requires the int value for wire-level compatibility.
func (d *defaultBehavior) GenerateEnumMethods(enum *parser.Enum) ([]Method, error) {
	methods := make([]Method, 2)
	// getValue() method.
	printer := new(generator.Printer)
	printer.Println("return value;")
	methods[0] = Method{
		Modifier:   "public",
		Name:       "getValue",
		ReturnType: "int",
		Body:       printer,
	}

	// static findByValue() method.
	printer = new(generator.Printer)
	printer.Println("switch (value) {").ScopeUp()
	for _, value := range enum.Values {
		printer.Println(fmt.Sprintf("case %d:", value.Value))
		printer.ScopeUp().Println(fmt.Sprintf("return %s;", value.Name)).ScopeDown()
	}
	printer.Println("default:")
	printer.ScopeUp().Println("return null;")
	printer.ScopeDown().ScopeDown().Println("}")
	methods[1] = Method{
		Modifier:   "public",
		Static:     true,
		Name:       "findByValue",
		ReturnType: enum.Name,
		Arguments:  []Argument{Argument{Type: "int", Name: "value"}},
		Body:       printer,
	}

	return methods, nil
}

// pluginBehavior implements the Behavior interface augmented by user-provided
// plugins.
type pluginBehavior struct {
	symbolTable     map[string][]interface{}
	defaultBehavior Behavior
}

// newPluginBehavior returns a Behavior which is augmented by the given
// FrugalPlugins.
func newPluginBehavior(plugins []*plugin.FrugalPlugin, defaultBehavior Behavior) Behavior {
	b := &pluginBehavior{
		symbolTable:     make(map[string][]interface{}),
		defaultBehavior: defaultBehavior,
	}

	// Load symbols from plugins.
	for _, plugin := range plugins {
		for _, hook := range hooks {
			if symbol := plugin.Lookup(hook); symbol != nil {
				symbols := b.symbolTable[hook]
				if symbols == nil {
					symbols = []interface{}{}
				}
				symbols = append(symbols, symbol)
				b.symbolTable[hook] = symbols
			}
		}
	}

	return b
}

// resolveSymbolOnce returns one instance of the specified symbol. Symbol
// precedence is determined by the order in which the plugins are provided,
// meaning if multiple plugins implement the symbol, the first is taken.
func (p *pluginBehavior) resolveSymbolOnce(symbol string) interface{} {
	if symbols := p.symbolTable[symbol]; len(symbols) > 0 {
		return symbols[0]
	}
	return nil
}

// resolveSymbol returns all instances of the specified symbol.
func (p *pluginBehavior) resolveSymbol(symbol string) []interface{} {
	return p.symbolTable[symbol]
}

// GenerateEnumValue is called when generating enum values, i.e.
//
// public enum MyEnum implements TEnum {
//     FOO(1), // produced by a call to GenerateEnumValue
//     BAR(2),
//     BAZ(3);
//
//     ...
//
// If multiple plugins implement this hook, only one will be invoked. Note that
// Frugal requires the int value for wire-level compatibility.
func (p *pluginBehavior) GenerateEnumValue(enum *parser.Enum, value *parser.EnumValue) (EnumValue, error) {
	f := p.resolveSymbolOnce(generateEnumValue)
	if f == nil {
		return p.defaultBehavior.GenerateEnumValue(enum, value)
	}
	generator, ok := f.(func(*parser.Enum, *parser.EnumValue) (EnumValue, error))
	if !ok {
		return EnumValue{}, fmt.Errorf("%s is %s, not func(*parser.Enum, *parser.Value) (string, error)",
			generateEnumValue, reflect.TypeOf(f))
	}
	return generator(enum, value)
}

// GenerateEnumFields is called when generating enum fields, e.g.
//
// private final int value; // produced by a call to GenerateEnumFields
//
// Note that Frugal requires the int value for wire-level compatibility.
func (p *pluginBehavior) GenerateEnumFields(enum *parser.Enum) ([]Field, error) {
	// Include default fields (private final int value).
	fields, err := p.defaultBehavior.GenerateEnumFields(enum)
	if err != nil {
		return nil, err
	}
	for _, f := range p.resolveSymbol(generateEnumFields) {
		generator, ok := f.(func(*parser.Enum) ([]Field, error))
		if !ok {
			return nil, fmt.Errorf("%s is %s, not func(*parser.Enum) ([]java.Field, error)",
				generateEnumFields, reflect.TypeOf(f))
		}
		pluginFields, err := generator(enum)
		if err != nil {
			return nil, err
		}
		fields = append(fields, pluginFields...)
	}
	return fields, nil
}

// GenerateEnumConstructors is called when generating enum constructors, e.g.
//
// MyEnum(int value) {      // produced by
//     this.value = value;  // a call to
// }                        // GenerateEnumConstructors
//
// Note that Frugal requires the int value for wire-level compatibility.
func (p *pluginBehavior) GenerateEnumConstructors(enum *parser.Enum) ([]EnumConstructor, error) {
	// Include default constructors (private value constructor).
	constructors, err := p.defaultBehavior.GenerateEnumConstructors(enum)
	if err != nil {
		return nil, err
	}
	for _, f := range p.resolveSymbol(generateEnumConstructors) {
		generator, ok := f.(func(*parser.Enum) ([]EnumConstructor, error))
		if !ok {
			return nil, fmt.Errorf("%s is %s, not func(*parser.Enum) ([]java.EnumConstructor, error)",
				generateEnumConstructors, reflect.TypeOf(f))
		}
		pluginConstructors, err := generator(enum)
		if err != nil {
			return nil, err
		}
		constructors = append(constructors, pluginConstructors...)
	}
	return constructors, nil
}

// GenerateEnumMethods is called when generating enum methods, e.g.
//
// public int getValue() {
//     return value;
// }
//
// Note that Frugal requires the int value for wire-level compatibility.
func (p *pluginBehavior) GenerateEnumMethods(enum *parser.Enum) ([]Method, error) {
	// Include default methods (public int getValue()).
	methods, err := p.defaultBehavior.GenerateEnumMethods(enum)
	if err != nil {
		return nil, err
	}
	for _, f := range p.resolveSymbol(generateEnumMethods) {
		generator, ok := f.(func(*parser.Enum) ([]Method, error))
		if !ok {
			return nil, fmt.Errorf("%s is %s, not func(*parser.Enum) ([]java.Method, error)",
				generateEnumMethods, reflect.TypeOf(f))
		}
		pluginMethods, err := generator(enum)
		if err != nil {
			return nil, err
		}
		methods = append(methods, pluginMethods...)
	}
	return methods, nil
}