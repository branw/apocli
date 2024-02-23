package main

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"go-apo/pkg/apocli"
	"reflect"
	"strings"
)

type ConfigPathCmd struct {
}

func (cmd *ConfigPathCmd) Run(_ *Context) error {
	path, err := apocli.ConfigFilePath()
	if err != nil {
		return err
	}
	fmt.Println(path)

	return nil
}

type ConfigListCmd struct {
}

func (cmd *ConfigListCmd) Run(ctx *Context) error {
	x1 := reflect.ValueOf(ctx.Config).Elem()
	printAllStructValues(x1)

	return nil
}

type ConfigGetCmd struct {
	Key string `arg:""`
}

func (cmd *ConfigGetCmd) Run(ctx *Context) error {
	requestedKey := cli.Config.Get.Key
	key, value, err := getValueFromStruct(requestedKey, ctx.Config)
	if err != nil {
		return fmt.Errorf("no config key matching \"%s\"", requestedKey)
	}

	fmt.Printf("%s (%s) = %+v\n", key, value.Type(), value.Interface())

	return nil
}

type ConfigSetCmd struct {
	Key   string `arg:""`
	Value string `arg:""`
}

func (cmd *ConfigSetCmd) Run(ctx *Context) error {
	requestedKey := cli.Config.Set.Key
	key, value, err := getValueFromStruct(requestedKey, ctx.Config)
	if err != nil {
		return fmt.Errorf("no config key matching \"%s\"", requestedKey)
	}

	value.SetString(cli.Config.Set.Value)

	fmt.Printf("%s (%s) = %+v\n", key, value.Type(), value.Interface())

	err = ctx.Config.Save()
	if err != nil {
		return fmt.Errorf("failed to save config after updating: %+v", err)
	}

	return nil
}

type ConfigClearCmd struct {
	Key string `arg:""`
}

func (cmd *ConfigClearCmd) Run(ctx *Context) error {
	requestedKey := cli.Config.Clear.Key
	key, value, err := getValueFromStruct(requestedKey, ctx.Config)
	if err != nil {
		return fmt.Errorf("no config key matching \"%s\"", requestedKey)
	}

	_, defaultValue, err := getValueFromStruct(requestedKey, apocli.DefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to get default value for key: %+v", err)
	}

	value.Set(defaultValue)

	fmt.Printf("%s (%s) = %+v\n", key, value.Type(), value.Interface())

	err = ctx.Config.Save()
	if err != nil {
		return fmt.Errorf("failed to save config after updating: %+v", err)
	}

	return nil
}

type ConfigClearAllCmd struct {
}

func (cmd *ConfigClearAllCmd) Run(_ *Context) error {
	config := apocli.DefaultConfig()
	err := config.Save()
	if err != nil {
		return fmt.Errorf("failed to save config after updating: %+v", err)
	}

	return nil
}

func toCamelCase(key string) string {
	return strings.ReplaceAll(strcase.ToCamel(key), "Id", "ID")
}

func getValueFromStruct(keyWithDots string, object interface{}) (string, reflect.Value, error) {
	keySlice := strings.Split(keyWithDots, ".")
	v := reflect.ValueOf(object)
	for i, key := range keySlice {
		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			return "", v, fmt.Errorf("only accepts structs; got %T", v)
		}

		key = toCamelCase(key)
		keySlice[i] = key
		v = v.FieldByName(key)
	}
	zeroValue := reflect.Value{}
	if v == zeroValue {
		return "", v, fmt.Errorf("unknown value")
	}

	return strings.Join(keySlice, "."), v, nil
}

func printAllStructValues(v reflect.Value) {
	typeOfT := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		fmt.Printf("%s (%s) = %v\n", typeOfT.Field(i).Name, f.Type(), f.Interface())

		if f.Kind().String() == "struct" {
			x1 := reflect.ValueOf(f.Interface())
			fmt.Printf("type: %s\n", x1)
			printAllStructValues(x1)
		}
	}
}
