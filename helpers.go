package main

import (
	"github.com/djimenez/iconv-go"
	"strings"
	"os"
	"strconv"
	"path"
	"encoding/json"
	"fmt"
	"errors"
)

func StringToInt64(s string) int64 {
	if s == "" {
		return -1
	}

	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}

	return val

}

func BytesToString(bytes []byte, converter *iconv.Converter) string {
	out, err := converter.ConvertString(strings.TrimSpace(string(bytes[:])))
	if err != nil {
		panic(err)
	}
	return out
}

func StringToByte(s string) byte {
	if len(s) > 0 {
		return []byte(s)[0]
	}

	return 0
}

func Min(x, y int64) int64 {
	if x < y {
		return x
	}

	return y
}

func Max(x, y int64) int64 {
	if x > y {
		return x
	}

	return y
}

func MinArray(arr []int64) (min int64) {
	for _, item := range arr {
		min = Min(item, min)
	}

	return min
}

func MaxArray(arr []int64) (max int64) {
	for _, item := range arr {
		max = Max(item, max)
	}

	return max
}

func MinMaxArray(arr []int64) (min int64, max int64) {
	min = MinArray(arr)
	max = MaxArray(arr)
	return min, max
}

func GetMinMaxArray(arr []int64, filter int64) (min int64, max int64) {
	var newarr []int64

	for _, item := range arr {
		if item != filter {
			newarr = append(newarr, item)
		}
	}

	return MinMaxArray(newarr)
}

// Read JSON file to struct
func UnmarshalJSONFromFile(filepath string, v interface{}) (err error) {
	err = os.MkdirAll(path.Dir(filepath), os.FileMode(0700))
	if err != nil {
		return err
	}

	// create file if not exists
	_, err = os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			file, err := os.Create(filepath)
			defer file.Close()
			if err != nil {
				return err
			}

			// Empty array
			file.Write([]byte("[]"))
		} else {
			return err
		}
	}

	file, err := os.Open(filepath)
	defer file.Close()
	if err != nil {
		return err
	}

	return json.NewDecoder(file).Decode(v)
}

func HasRequiredCommandLineArguments(required []string, seen map[string]bool) (err error) {
	err = nil
	var errs []string
	for _, req := range required {
		if !seen[req] {
			errs = append(errs, fmt.Sprintf("Error: Missing required -%s argument/flag.", req))
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return err
}
