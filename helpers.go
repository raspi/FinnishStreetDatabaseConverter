package main

import (
	"errors"
	"fmt"
	"github.com/djimenez/iconv-go"
	"math"
	"strconv"
	"strings"
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

// Convert 1024 to '1 KiB' etc
func bytesToHuman(src uint64) string {
	if src < 10 {
		return fmt.Sprintf("%d B", src)
	}

	s := float64(src)
	base := float64(1024)
	sizes := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}

	e := math.Floor(math.Log(s) / math.Log(base))
	suffix := sizes[int(e)]
	val := math.Floor(s/math.Pow(base, e)*10+0.5) / 10
	f := "%.0f %s"
	if val < 10 {
		f = "%.1f %s"
	}

	return fmt.Sprintf(f, val, suffix)
}
