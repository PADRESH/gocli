package gocli

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type Argument struct {
	Alias       string
	Argument    string
	Description string
	Type        reflect.Type
}

func parseArgument(arg string, t reflect.Type) Argument {
	pairs := strings.Split(arg, ",")
	m := make(map[string]string)
	for _, pair := range pairs {
		values := strings.Split(pair, "=")
		m[values[0]] = values[1]
	}
	return Argument{
		Alias:       m["alias"],
		Argument:    m["argument"],
		Description: m["description"],
		Type:        t,
	}
}

func find(m map[string]Argument, filter func(*Argument) bool) (string, error) {
	for key, arg := range m {
		if filter(&arg) {
			return key, nil
		}
	}
	return "", errors.New("not found")
}

func PrintHelp[A interface{}](name string, description string, data *A) error {
	arguments, err := parseArguments(data)
	if err != nil {
		return err
	}

	fmt.Printf("Usage: %s [parameter]\n", name)
	fmt.Printf("%s\n", description)
	fmt.Printf("parameters:\n")

	for _, arg := range arguments {
		var cnt int = 0
		if len(arg.Alias) > 0 {
			fmt.Printf("%*s%s", 5, "-", arg.Alias)
			cnt += 3
		} else {
			fmt.Printf("   ")
		}

		if len(arg.Argument) > 0 {
			fmt.Printf(" --%s", arg.Argument)
			cnt += len(arg.Argument) + 3
			if arg.Type.Kind() != reflect.Bool {
				fmt.Printf("=%s", "WORD")
				cnt += 5
			}
		}
		if len(arg.Description) > 0 {
			fmt.Printf("%*s\n", 18-cnt+len(arg.Description), arg.Description)
		}
	}
	return nil
}

func parseArguments[A interface{}](data *A) (map[string]Argument, error) {

	dataType := reflect.TypeOf(data)
	if dataType.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("arguments must be a pointer")
	}

	dataSource := dataType.Elem()
	if dataSource.Kind() != reflect.Struct {
		return nil, fmt.Errorf("config must be a struct")
	}

	arguments := make(map[string]Argument)
	for i := 0; i < dataSource.NumField(); i++ {
		fieldTag, ok := dataSource.Field(i).Tag.Lookup("args")
		if !ok {
			continue
		}
		arguments[dataSource.Field(i).Name] = parseArgument(fieldTag, dataSource.Field(i).Type)
	}
	return arguments, nil
}

func LoadArguments[A interface{}](data *A) (*A, error) {

	arguments, err := parseArguments(data)
	if err != nil {
		return nil, err
	}

	dataSource := reflect.TypeOf(data).Elem()
	args := os.Args

	counter := 1
	for counter < len(args) {
		value := args[counter]
		var current string
		var err error
		if strings.HasPrefix(value, "--") {
			current, err = find(arguments, func(argument *Argument) bool { return argument.Argument == value[2:] })
		} else if strings.HasPrefix(value, "-") {
			current, err = find(arguments, func(argument *Argument) bool { return argument.Alias == value[1:] })
		} else {
			return nil, errors.New(fmt.Sprintf("unknown attribute %s", value))
		}
		if err == nil {
			field, found := dataSource.FieldByName(current)
			if found {
				fieldValue := reflect.ValueOf(data).Elem().FieldByIndex(field.Index)
				if field.Type.Kind() == reflect.Bool {
					fieldValue.SetBool(true)
				} else if field.Type.Kind() == reflect.Int {
					counter++
					value = args[counter]
					val, err := strconv.ParseInt(value, 10, 32)
					if err == nil {
						fieldValue.SetInt(val)
					}
				} else if field.Type.Kind() == reflect.String {
					counter++
					value = args[counter]
					fieldValue.SetString(value)
				} else {
					return nil, errors.New("unknown argument type")
				}
			} else {
				return nil, errors.New(fmt.Sprintf("unknown filed %s", current))
			}
		}
		counter++
	}
	return data, nil
}
