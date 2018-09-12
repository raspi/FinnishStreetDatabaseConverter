package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/djimenez/iconv-go"
	"github.com/spf13/afero"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

// structured
type EvenOdd uint8 // #12 Even / odd

// #12 Even / odd
const (
	NOTUSED EvenOdd = iota
	ODD
	EVEN
)

type Building struct {
	BuildingNumber1         int64 // #14 & #20
	BuildingDeliveryLetter1 byte  // #15 & #21
	PunctuationMark         byte  // #16 & #22
	BuildingNumber2         int64 // #17 & #23
	BuildingDeliveryLetter2 byte  // #18 & #24
}

type StreetAddress struct {
	//RecordIdentifier              string // #1
	//RunningDate                    string // #2
	PostalCode            string // #3 Postal code, numeric
	PostalCodeNameFi      string // #4 Postal code name in Finnish
	PostalCodeNameSe      string // #5 Postal code name in Swedish
	PostalCodeShortNameFi string // #6 Postal code name abbreviation in Finnish
	PostalCodeShortNameSe string // #7 Postal code name abbreviation in Swedish
	StreetNameFi          string // #8 Street (location) name in Finnish
	StreetNameSe          string // #9 Street (location) name in Swedish
	//Blank1                    string // #10 Blank
	//Blank2                    string // #11 Blank
	BuildingDataTypeEvenOdd EvenOdd // #12 Building data type, odd / even

	SmallestBuilding Building
	HighestBuilding  Building

	MunicipalityCode   string // #25 Municipality code, numeric
	MunicipalityNameFi string // #26 Municipality name in Finnish
	MunicipalityNameSe string // #27 Municipality name in Swedish
}

// Converters

func StringToEvenOddConst(s string) EvenOdd {
	if s == "1" {
		return ODD
	} else if s == "2" {
		return EVEN
	} else {
		return NOTUSED
	}
}

// Find min and max building number
func (src StreetAddress) StreetNumberMinMax(arr []int64) (min int64, max int64) {
	var numbers = []int64{src.SmallestBuilding.BuildingNumber1, src.SmallestBuilding.BuildingNumber2, src.HighestBuilding.BuildingNumber1, src.HighestBuilding.BuildingNumber2}
	numbers = append(numbers, arr...)
	return GetMinMaxArray(numbers, -1)
}

func (src *RawLineStructure) ToStreet(converter *iconv.Converter) StreetAddress {
	smallest := Building{
		BuildingNumber1:         StringToInt64(BytesToString(src.SmallestBuildingNumber1[:], converter)),        // 14
		BuildingDeliveryLetter1: StringToByte(BytesToString(src.SmallestBuildingDeliveryLetter1[:], converter)), // 15
		PunctuationMark:         StringToByte(BytesToString(src.SmallestPunctuationMark[:], converter)),         // 16
		BuildingNumber2:         StringToInt64(BytesToString(src.SmallestBuildingNumber2[:], converter)),        // 17
		BuildingDeliveryLetter2: StringToByte(BytesToString(src.SmallestBuildingDeliveryLetter2[:], converter)), // 18
	}

	highest := Building{
		BuildingNumber1:         StringToInt64(BytesToString(src.HighestBuildingNumber1[:], converter)),        // 20
		BuildingDeliveryLetter1: StringToByte(BytesToString(src.HighestBuildingDeliveryLetter1[:], converter)), // 21
		PunctuationMark:         StringToByte(BytesToString(src.HighestPunctuationMark[:], converter)),         // 22
		BuildingNumber2:         StringToInt64(BytesToString(src.HighestBuildingNumber2[:], converter)),        // 23
		BuildingDeliveryLetter2: StringToByte(BytesToString(src.HighestBuildingDeliveryLetter2[:], converter)), // 24
	}

	p := StreetAddress{
		PostalCode:              BytesToString(src.PostalCode[:], converter),                             // 3
		PostalCodeNameFi:        strings.ToLower(BytesToString(src.PostalCodeNameFi[:], converter)),      // 4
		PostalCodeNameSe:        strings.ToLower(BytesToString(src.PostalCodeNameSe[:], converter)),      // 5
		PostalCodeShortNameFi:   strings.ToLower(BytesToString(src.PostalCodeShortNameFi[:], converter)), // 6
		PostalCodeShortNameSe:   strings.ToLower(BytesToString(src.PostalCodeShortNameSe[:], converter)), // 7
		StreetNameFi:            strings.ToLower(BytesToString(src.StreetNameFi[:], converter)),          // 8
		StreetNameSe:            strings.ToLower(BytesToString(src.StreetNameSe[:], converter)),          // 9
		BuildingDataTypeEvenOdd: StringToEvenOddConst(BytesToString(src.BuildingDataType[:], converter)), // 12
		SmallestBuilding:        smallest,                                                                // 14-18
		HighestBuilding:         highest,                                                                 // 20-24
		MunicipalityCode:        BytesToString(src.MunicipalityCode[:], converter),                       // 25
		MunicipalityNameFi:      strings.ToLower(BytesToString(src.MunicipalityNameFi[:], converter)),    // 26
		MunicipalityNameSe:      strings.ToLower(BytesToString(src.MunicipalityNameSe[:], converter)),    // 27
	}

	return p
}

// Convert file to multiple JSON files
func ConvertFile(sourcefile string, targetdir string) (err error) {
	converter, err := iconv.NewConverter("iso-8859-1", "utf-8")
	if err != nil {
		return err
	}

	f, err := os.Open(sourcefile)
	defer f.Close()

	if err != nil {
		return err
	}

	var raw RawLineStructure

	buffer := make([]byte, binary.Size(raw))
	nl := make([]byte, 1) // new line

	fInfo, err := f.Stat()
	if err != nil {
		return err
	}

	var sourceTotalSizeBytes = fInfo.Size()
	var sourceReadedBytes int64 = 0

	// Ticker for stats
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	// Create new in-memory filesystem
	fSystem := &afero.Afero{
		Fs: afero.NewMemMapFs(),
	}

	// Read source file line by line
	for {
		_, err := f.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		// Read from file
		r := bytes.NewReader(buffer)

		// Get position
		pos, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}

		sourceReadedBytes = pos

		// Read to struct
		err = binary.Read(r, binary.BigEndian, &raw)
		if err != nil {
			return err
		}

		// Convert to proper struct
		streetAddr := raw.ToStreet(converter)

		err = ConvertMunicipality(fSystem, streetAddr)
		if err != nil {
			panic(err)
		}

		err = ConvertPostalCode(fSystem, streetAddr)
		if err != nil {
			panic(err)
		}

		err = ConvertStreet(fSystem, streetAddr)
		if err != nil {
			panic(err)
		}

		// Report stats
		select {
		case <-ticker.C:
			percent := (float64(sourceReadedBytes) * float64(100.0)) / float64(sourceTotalSizeBytes)
			log.Printf("%v / %v %07.3f%%", sourceReadedBytes, sourceTotalSizeBytes, percent)
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			log.Printf(`%v %v`, bytesToHuman(m.Alloc), bytesToHuman(m.TotalAlloc))
		default:

		}

		_, err = f.Read(nl)

		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		if nl[0] != '\n' {
			return errors.New("Not newline")
		}

	}

	log.Printf(`Generating directories..`)
	// Create directories
	err = fSystem.Walk(`/`, func(filepath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			// Skip files
			return nil
		}

		dirPath := path.Join(targetdir, filepath)

		err = os.MkdirAll(dirPath, os.FileMode(0700))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Memory files to actual files
	log.Printf(`Saving files..`)
	err = fSystem.Walk(`/`, func(filepath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip directories
			return nil
		}

		b, err := fSystem.ReadFile(filepath)
		if err != nil {
			return err
		}

		dirPath := path.Join(targetdir, filepath)
		err = ioutil.WriteFile(dirPath, b, os.FileMode(0600))
		if err != nil {
			return err
		}

		return nil
	})

	return nil
}

type PostnumberJSON struct {
	Fi    string `json:"fi,omitempty"`  // Post number name in Finnish
	Se    string `json:"se,omitempty"`  // Post number name in Swedish
	FiLyh string `json:"fil,omitempty"` // Shortened post number name in Finnish
	SeLyh string `json:"sel,omitempty"` // Shortened post number name in Swedish
}

func ConvertFromFile(fName string, fs *afero.Afero, v interface{}) error {
	_, err := fs.Stat(fName)
	if err != nil {
		if os.IsNotExist(err) {
			f, err := fs.Create(fName)
			if err != nil {
				return err
			}

			f.WriteString(`[]`) // Empty array
			f.Close()
		} else {
			return err
		}
	}

	b, err := fs.ReadFile(fName)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	return nil

}

func ConvertPostalCode(fs *afero.Afero, addr StreetAddress) error {
	fName := path.Join(string(os.PathSeparator), addr.MunicipalityCode, addr.PostalCode, `postnumber.json`)

	var data []PostnumberJSON

	err := ConvertFromFile(fName, fs, &data)
	if err != nil {
		return err
	}

	var found = false
	for idx, k := range data {
		if k.Fi == addr.PostalCodeNameFi {
			data[idx] = k
			found = true
			break
		}
	}

	if !found {
		data = append(data, PostnumberJSON{
			Fi:    addr.PostalCodeNameFi,
			FiLyh: addr.PostalCodeShortNameFi,
			Se:    addr.PostalCodeNameSe,
			SeLyh: addr.PostalCodeShortNameSe,
		})
	}

	err = SaveData(fs, fName, data)
	if err != nil {
		return err
	}

	return nil
}

func SaveData(fs *afero.Afero, s string, v interface{}) error {
	writebytes, err := json.Marshal(v)
	if err != nil {
		return err
	}

	err = fs.WriteFile(s, writebytes, os.FileMode(0600))
	if err != nil {
		return err
	}

	return nil

}

type MunicipalityJSON struct {
	Fi string `json:"fi,omitempty"` // Municipality name in Finnish
	Se string `json:"se,omitempty"` // Municipality name in Swedish
}

func ConvertMunicipality(fs *afero.Afero, addr StreetAddress) error {
	fName := path.Join(string(os.PathSeparator), addr.MunicipalityCode, `municipality.json`)

	var data []MunicipalityJSON

	err := ConvertFromFile(fName, fs, &data)
	if err != nil {
		return err
	}

	var found = false
	for idx, k := range data {
		if k.Fi == addr.MunicipalityNameFi {
			data[idx] = k
			found = true
			break
		}
	}

	if !found {
		data = append(data, MunicipalityJSON{
			Fi: addr.MunicipalityNameFi,
			Se: addr.MunicipalityNameSe,
		})
	}

	err = SaveData(fs, fName, data)
	if err != nil {
		return err
	}

	return nil
}

type StreetJSON struct {
	Fi  string `json:"fi,omitempty"`  // Street name in Finnish
	Se  string `json:"se,omitempty"`  // Street name in Swedish
	Min int64  `json:"min,omitempty"` // Minimum number
	Max int64  `json:"max,omitempty"` // Maximum number
}

func ConvertStreet(fs *afero.Afero, addr StreetAddress) error {
	if addr.StreetNameFi == `` {
		return nil
	}

	fName := path.Join(string(os.PathSeparator), addr.MunicipalityCode, addr.PostalCode, `street.json`)

	var data []StreetJSON

	err := ConvertFromFile(fName, fs, &data)
	if err != nil {
		return err
	}

	var found = false
	for idx, k := range data {
		if k.Fi == addr.StreetNameFi {
			min, max := addr.StreetNumberMinMax([]int64{k.Min, k.Max})
			k.Min = min
			k.Max = max
			data[idx] = k
			found = true
			break
		}
	}

	if !found {
		min, max := addr.StreetNumberMinMax([]int64{})
		data = append(data, StreetJSON{
			Fi:  addr.StreetNameFi,
			Se:  addr.StreetNameSe,
			Min: min,
			Max: max,
		})
	}

	err = SaveData(fs, fName, data)
	if err != nil {
		return err
	}

	return nil
}
