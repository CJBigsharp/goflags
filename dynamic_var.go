package goflags

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type dynamicFlag struct {
	field        interface{}
	defaultValue interface{}
	name         string
}

func (df *dynamicFlag) Set(value string) error {
	fieldKind := reflect.TypeOf(df.field).Elem().Kind()
	var isBoolValue bool
	if _, err := strconv.ParseBool(value); err == nil {
		isBoolValue = true
	}
	if fieldKind == reflect.Bool && isBoolValue {
		boolField := df.field.(*bool)
		*boolField = true
		return nil
	}
	switch fieldKind {
	case reflect.Int:
		intField := df.field.(*int)
		if isBoolValue {
			*intField = df.defaultValue.(int)
			return nil
		}
		newValue, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		*intField = newValue
	case reflect.Float64:
		floatField := df.field.(*float64)
		if isBoolValue {
			*floatField = df.defaultValue.(float64)
			return nil
		}
		newValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		*floatField = newValue
	case reflect.String:
		stringField := df.field.(*string)
		if isBoolValue {
			*stringField = df.defaultValue.(string)
			return nil
		}
		*stringField = value
	default:
		return errors.New("unsupported type")
	}
	return nil
}

func (df *dynamicFlag) IsBoolFlag() bool {
	return true
}

func (df *dynamicFlag) String() string {
	return df.name
}

// DynamicVar acts as flag with a default value or a option with value
// example:
//   var titleSize int
//   flagSet.DynamicVar(&titleSize, "title", 50, "first N characters of the title")
// > go run ./examples/basic -title or go run ./examples/basic -title=100
// In case of `go run ./examples/basic -title` it will use default value 50
func (flagSet *FlagSet) DynamicVar(field interface{}, long string, defaultValue interface{}, usage string) *FlagData {
	return flagSet.DynamicVarP(field, long, "", defaultValue, usage)
}

// DynamicVarP same as DynamicVar but with short name
func (flagSet *FlagSet) DynamicVarP(field interface{}, long, short string, defaultValue interface{}, usage string) *FlagData {
	// validate field and defaultValue
	if reflect.TypeOf(field).Kind() != reflect.Ptr {
		panic(fmt.Errorf("-%v flag field must be a pointer", long))
	}
	if reflect.TypeOf(field).Elem().Kind() != reflect.TypeOf(defaultValue).Kind() {
		panic(fmt.Errorf("-%v flag field and defaultValue mismatch: fied type is %v and defaultValue Type is %T", long, reflect.TypeOf(field).Elem().Kind(), defaultValue))
	}
	if field == nil {
		panic(fmt.Errorf("field cannot be nil for flag -%v", long))
	}

	var dynamicFlag dynamicFlag
	dynamicFlag.field = field
	dynamicFlag.name = long
	if defaultValue != nil {
		dynamicFlag.defaultValue = defaultValue
	}

	flagData := &FlagData{
		usage:        usage,
		long:         long,
		defaultValue: defaultValue,
	}
	if short != "" {
		flagData.short = short
		flagSet.CommandLine.Var(&dynamicFlag, short, usage)
		flagSet.flagKeys.Set(short, flagData)
	}
	flagSet.CommandLine.Var(&dynamicFlag, long, usage)
	flagSet.flagKeys.Set(long, flagData)
	return flagData
}
