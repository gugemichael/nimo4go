package nimo

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const DefaultSeparator = ";"

var onlyAlpha = regexp.MustCompile("^[a-zA-Z0-9_-]+")

const (
	GenericTypeRaw  = 1
	GenericTypeDate = 2
)

// ConfigLoader load config from file
type ConfigLoader struct {
	conf *os.File

	// date format string
	format string
	// string slice separator
	separator string
}

// NewConfigLoader new loader
func NewConfigLoader(file *os.File) *ConfigLoader {
	return &ConfigLoader{conf: file, separator: DefaultSeparator}
}

func (loader *ConfigLoader) SetDateFormat(format string) {
	loader.format = format
}

func (loader *ConfigLoader) SetSliceSeparator(sep string) {
	loader.separator = sep
}

// Load load config from file to target's field by reflect
func (loader *ConfigLoader) Load(target interface{}) error {
	if target == nil {
		return errors.New("target is nil error")
	}
	// fetch all target's field member
	instance := reflect.ValueOf(target).Elem()
	injected := 0

	if !instance.IsValid() || instance.Kind() != reflect.Struct {
		return errors.New("target is not a valid struct type")
	}

	// file content reader. handled line by line
	reader := bufio.NewReader(loader.conf)
	for {
		if buffer, _, err := reader.ReadLine(); err == nil {
			// reach the end of File EOF
			line := strings.Trim(string(buffer), " \r\n")

			// skip start with "#". it's comment
			if !isComment(line) && len(line) != 0 {
				pair := strings.SplitN(line, "=", 2)
				key := strings.Trim(pair[0], " ")
				if len(pair) != 2 || !onlyAlpha.MatchString(key) {
					return fmt.Errorf("error confiugre key line found. key is %s", key)
				}
				value := strings.Trim(pair[1], " ")
				// key is good. look up the matching tag
				var err error
				var found string
				var genericType int
				if found, genericType, err = lookupTagMember(target, key); err != nil {
					// not found in struct, give error
					return fmt.Errorf("couldn't found structure key's tag '%s'. one must be tagged", key)
				}

				field := instance.FieldByName(found)
				if !field.IsValid() || !field.CanSet() {
					return fmt.Errorf("structure key '%s' is not accessable", key)
				}

				// set config's value
				switch field.Kind() {
				case reflect.String: // for string
					field.SetString(value)
				case reflect.Bool: // for bool
					if v, err := strconv.ParseBool(value); err == nil {
						field.SetBool(v)
					} else {
						return fmt.Errorf("boolean key %s wrong format %v", key, value)
					}
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64: // for numberic
					if genericType == GenericTypeDate {
						if t, e := time.ParseInLocation(loader.format, value, time.Local); e == nil {
							field.SetInt(t.Unix())
						} else {
							return fmt.Errorf("integer conver to date format [%s] from (%s)", loader.format, value)
						}
					} else if v, err := strconv.ParseInt(value, 10, 64); err == nil {
						field.SetInt(v)
					} else {
						return fmt.Errorf("integer key %s wrong format %v", key, value)
					}
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					if v, err := strconv.ParseUint(value, 10, 64); err == nil {
						field.SetUint(v)
					} else {
						return fmt.Errorf("unsigned integer key %s wrong format %v", key, value)
					}
				case reflect.Slice: // for slice
					if items, err := loader.extractList(value); err == nil {
						filtered := make([]string, 0, len(items))
						for _, item := range items {
							if s := strings.TrimSpace(item); len(s) != 0 {
								filtered = append(filtered, s)
							}
						}
						field.Set(reflect.ValueOf(filtered))
					} else {
						return err
					}
				}
				injected++
			} else {
				if err == io.EOF {
					if injected == 0 {
						return errors.New("file is empty without any configuration")
					}
				}
			}
		}
	}
}

func (loader *ConfigLoader) extractList(value string) ([]string, error) {
	segments := make([]string, 0)
	if strings.HasPrefix(value, "@@") {
		// we need to fetch the content from specific file
		if ref, err := os.Open(fmt.Sprintf("%s%c%s", filepath.Dir(loader.conf.Name()),
			filepath.Separator, value[2:])); err == nil {
			reader := bufio.NewReader(ref)
			for {
				line, end, err := reader.ReadLine()
				if end || err != nil {
					break
				}
				if !isComment(strings.TrimSpace(string(line))) {
					segments = append(segments, string(line))
				}
			}
		} else {
			return nil, err
		}
	} else {
		segments = strings.Split(value, DefaultSeparator)
	}
	return segments, nil
}

func lookupTagMember(target interface{}, tag string) (string, int, error) {
	total := reflect.ValueOf(target).Elem().NumField()
	types := reflect.TypeOf(target)
	for i := 0; i != total; i++ {
		if types.Elem().Field(i).Tag.Get("config") == tag {
			if types.Elem().Field(i).Tag.Get("type") == "date" {
				return types.Elem().Field(i).Name, GenericTypeDate, nil
			}
			return types.Elem().Field(i).Name, GenericTypeRaw, nil
		}
	}
	return "", -1, fmt.Errorf("no tagged field [%s] found", tag)
}

func isComment(line string) bool {
	return strings.HasPrefix(line, "#")
}
