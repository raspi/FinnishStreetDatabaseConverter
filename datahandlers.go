package main

import (
	"path"
	"encoding/json"
	"log"
	"io/ioutil"
	"os"
	"strings"
	"io"
	"errors"
	"github.com/djimenez/iconv-go"
	"encoding/binary"
	"bytes"
	"time"
)

// Basic Address File Record Description
// Raw data
type RawLineStructure struct {
	RecordIdentifier      [5]byte  // #1 Record identifier, "KATUN"
	RunningDate           [8]byte  // #2 Running date, numeric date yyyymmdd
	PostalCode            [5]byte  // #3 Postal code, numeric
	PostalCodeNameFi      [30]byte // #4 Postal code name in Finnish
	PostalCodeNameSe      [30]byte // #5 Postal code name in Swedish, optional
	PostalCodeShortNameFi [12]byte // #6 Postal code name abbreviation in Finnish
	PostalCodeShortNameSe [12]byte // #7 Postal code name abbreviation in Swedish, optional
	StreetNameFi          [30]byte // #8 Street (location) name in Finnish
	StreetNameSe          [30]byte // #9 Street (location) name in Swedish, optional
	Blank1                [12]byte // #10 Blank
	Blank2                [12]byte // #11 Blank
	BuildingDataType      [1]byte  // #12 Building data type, 1 = odd 2 = even

	// #13 (skipped) Smallest building number (information about an odd/even building)
	Smallest_BuildingNumber1         [5]byte // #14 Building number 1, optional
	Smallest_BuildingDeliveryLetter1 [1]byte // #15 Building delivery letter 1, optional
	Smallest_PunctuationMark         [1]byte // #16 Punctuation mark, optional
	Smallest_BuildingNumber2         [5]byte // #17 Building number 2, optional
	Smallest_BuildingDeliveryLetter2 [1]byte // #18 Building delivery letter 2, optional

	// #19 (skipped) Highest building number (information about an odd/even building)
	Highest_BuildingNumber1         [5]byte // #20 Building number 1, optional
	Highest_BuildingDeliveryLetter1 [1]byte // #21 Building delivery letter 1, optional
	Highest_PunctuationMark         [1]byte // #22 Punctuation mark, optional
	Highest_BuildingNumber2         [5]byte // #23 Building number 2, optional
	Highest_BuildingDeliveryLetter2 [1]byte // #24 Building delivery letter 2, optional

	MunicipalityCode   [3]byte  // #25 Municipality code, numeric
	MunicipalityNameFi [20]byte // #26 Municipality name in Finnish
	MunicipalityNameSe [20]byte // #27 Municipality name in Swedish, optional
}

// structured
type EvenOdd uint8 // #12 Even / odd

// #12 Even / odd
const (
	NOT_USED EvenOdd = 0
	ODD      EvenOdd = 1
	EVEN     EvenOdd = 2
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

// JSON structures

type StreetJSON struct {
	Fi  string `json:"fi,omitempty"`  // Street name in Finnish
	Se  string `json:"se,omitempty"`  // Street name in Swedish
	Min int64  `json:"min,omitempty"` // Minimum number
	Max int64  `json:"max,omitempty"` // Maximum number
}

type PostnumberJSON struct {
	Fi    string `json:"fi,omitempty"`  // Post number name in Finnish
	Se    string `json:"se,omitempty"`  // Post number name in Swedish
	FiLyh string `json:"fil,omitempty"` // Shortened post number name in Finnish
	SeLyh string `json:"sel,omitempty"` // Shortened post number name in Swedish
}

type MunicipalityJSON struct {
	Fi string `json:"fi,omitempty"` // Municipality name in Finnish
	Se string `json:"se,omitempty"` // Municipality name in Swedish
}

// Converters
func StringToEvenOddConst(s string) EvenOdd {
	if s == "1" {
		return ODD
	} else if s == "2" {
		return EVEN
	} else {
		return NOT_USED
	}
}

// Find min and max building number
func (src StreetAddress) StreetNumberMinMax(arr []int64) (min int64, max int64) {
	var nums []int64 = []int64{src.SmallestBuilding.BuildingNumber1, src.SmallestBuilding.BuildingNumber2, src.HighestBuilding.BuildingNumber1, src.HighestBuilding.BuildingNumber2}
	nums = append(nums, arr...)
	return GetMinMaxArray(nums, -1)
}

// Street address JSON
func (src StreetAddress) NewStreetJSON() StreetJSON {
	min, max := src.StreetNumberMinMax([]int64{})

	return StreetJSON{
		Fi:  src.StreetNameFi,
		Se:  src.StreetNameSe,
		Min: min,
		Max: max,
	}
}

// Post number 
func (src StreetAddress) NewPostnumberJSON() PostnumberJSON {
	return PostnumberJSON{
		Fi:    src.PostalCodeNameFi,
		Se:    src.PostalCodeNameSe,
		FiLyh: src.PostalCodeShortNameFi,
		SeLyh: src.PostalCodeShortNameSe,
	}
}

// Municipality
func (src StreetAddress) NewMunicipalityJSON() MunicipalityJSON {
	return MunicipalityJSON{
		Fi: src.MunicipalityNameFi,
		Se: src.MunicipalityNameSe,
	}
}

// Possible debug info
func (src StreetAddress) write_info() {
	//log.Printf("%s (%s) %s (%s) %s\n", src.MunicipalityNameFi, src.MunicipalityCode, src.PostalCodeNameFi, src.PostalCode, src.StreetNameFi)
}

// Write to file
func (src StreetAddress) write_street(dir string) {

	if src.StreetNameFi == "" {
		return
	}

	src.write_info()

	filename := path.Join(dir, src.MunicipalityCode, src.PostalCode, "street.json")

	var data []StreetJSON
	err := UnmarshalJSONFromFile(filename, &data)
	if err != nil {
		panic(err)
	}

	var found bool = false
	for idx, k := range data {
		if k.Fi == src.StreetNameFi {
			min, max := src.StreetNumberMinMax([]int64{k.Min, k.Max})
			k.Min = min
			k.Max = max
			data[idx] = k
			found = true
			break
		}
	}

	if !found {
		data = append(data, src.NewStreetJSON())
	}

	writebytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(filename, writebytes, os.FileMode(0600))
	if err != nil {
		panic(err)
	}

}

// Write to file
func (src StreetAddress) write_postnumber(dir string) {
	if src.PostalCodeNameFi == "" {
		return
	}

	src.write_info()

	filename := path.Join(dir, src.MunicipalityCode, src.PostalCode, "postnumber.json")

	var data []PostnumberJSON
	err := UnmarshalJSONFromFile(filename, &data)
	if err != nil {
		panic(err)
	}

	var found bool = false
	for idx, k := range data {
		if k.Fi == src.PostalCodeNameFi {
			data[idx] = k
			found = true
			break
		}
	}

	if !found {
		data = append(data, src.NewPostnumberJSON())
	}

	writebytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(filename, writebytes, os.FileMode(0600))
	if err != nil {
		panic(err)
	}

}

// Write to file
func (src StreetAddress) write_municipality(dir string) {
	if src.MunicipalityNameFi == "" {
		return
	}

	src.write_info()

	filename := path.Join(dir, src.MunicipalityCode, "municipality.json")

	var data []MunicipalityJSON
	err := UnmarshalJSONFromFile(filename, &data)
	if err != nil {
		panic(err)
	}

	var found bool = false
	for idx, k := range data {
		if k.Fi == src.MunicipalityNameFi {
			data[idx] = k
			found = true
			break
		}
	}

	if !found {
		data = append(data, src.NewMunicipalityJSON())
	}

	writebytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(filename, writebytes, os.FileMode(0600))
	if err != nil {
		panic(err)
	}

}

func(src *RawLineStructure) ToStreet(converter *iconv.Converter) StreetAddress {
	smallest := Building{
		BuildingNumber1:         StringToInt64(BytesToString(src.Smallest_BuildingNumber1[:], converter)),        // 14
		BuildingDeliveryLetter1: StringToByte(BytesToString(src.Smallest_BuildingDeliveryLetter1[:], converter)), // 15
		PunctuationMark:         StringToByte(BytesToString(src.Smallest_PunctuationMark[:], converter)),         // 16
		BuildingNumber2:         StringToInt64(BytesToString(src.Smallest_BuildingNumber2[:], converter)),        // 17
		BuildingDeliveryLetter2: StringToByte(BytesToString(src.Smallest_BuildingDeliveryLetter2[:], converter)), // 18
	}

	highest := Building{
		BuildingNumber1:         StringToInt64(BytesToString(src.Highest_BuildingNumber1[:], converter)),        // 20
		BuildingDeliveryLetter1: StringToByte(BytesToString(src.Highest_BuildingDeliveryLetter1[:], converter)), // 21
		PunctuationMark:         StringToByte(BytesToString(src.Highest_PunctuationMark[:], converter)),         // 22
		BuildingNumber2:         StringToInt64(BytesToString(src.Highest_BuildingNumber2[:], converter)),        // 23
		BuildingDeliveryLetter2: StringToByte(BytesToString(src.Highest_BuildingDeliveryLetter2[:], converter)), // 24
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
func ConvertFile(sourcefile string, targetdir string) (err error){
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

	finfo, err := f.Stat()
	if err != nil {
		return err
	}

	var sourceTotalSizeBytes int64 = finfo.Size()
	var sourceReadedBytes int64 = 0

	// Ticker for stats
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

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

		// Convert to proper structs

		p := raw.ToStreet(converter)

		p.write_street(targetdir)
		p.write_postnumber(targetdir)
		p.write_municipality(targetdir)

		// Report stats
		select {
		case <-ticker.C:
			percent := ( float64(sourceReadedBytes) * float64(100.0) ) / float64(sourceTotalSizeBytes)
			log.Printf("%v / %v %07.3f%%", sourceReadedBytes, sourceTotalSizeBytes, percent)
		default:
			// do nothing
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

	return nil
}
