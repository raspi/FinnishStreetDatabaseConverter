package main

import (
	"path"
	"encoding/json"
	"log"
	"fmt"
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

// Raw data
type RawLineStructure struct {
	Tietuetunnus              [5]byte  // 1, "KATUN"
	Ajopvm                    [8]byte  // 2, numeric date yyyymmdd
	Postinumero               [5]byte  // 3, numeric
	Postinumeron_nimi_fi      [30]byte // 4
	Postinumeron_nimi_se      [30]byte // 5
	Postinumeron_lyhenne_fi   [12]byte // 6
	Postinumeron_lyhenne_se   [12]byte // 7
	Katu_fi                   [30]byte // 8
	Katu_se                   [30]byte // 9
	Tyhja1                    [12]byte // 10, empty
	Tyhja2                    [12]byte // 11, empty
	Kiinteistotietojen_tyyppi [1]byte  // 12

	// 13 skipped in doc
	Pienin_numero1      [5]byte // 14
	Pienin_jakokirjain1 [1]byte // 15
	Pienin_valimerkki   [1]byte // 16
	Pienin_numero2      [5]byte // 17
	Pienin_jakokirjain2 [1]byte // 18

	// 19 skipped in doc
	Suurin_numero1      [5]byte // 20
	Suurin_jakokirjain1 [1]byte // 21
	Suurin_valimerkki   [1]byte // 22
	Suurin_numero2      [5]byte // 23
	Suurin_jakokirjain2 [1]byte // 24

	Kunnan_koodi [3]byte  // 25, numeric
	Kunta_fi     [20]byte // 26
	Kunta_se     [20]byte // 27
}

// structured
type EvenOdd uint8

const (
	NOT_USED EvenOdd = 0
	ODD      EvenOdd = 1
	EVEN     EvenOdd = 2
)

type Kiinteisto struct {
	Numero1      int64
	Jakokirjain1 byte
	Valimerkki   byte
	Numero2      int64
	Jakokirjain2 byte
}

type StreetAddress struct {
	//Tietuetunnus              string // 1
	//Ajopvm                    string // 2
	Postinumero             string // 3
	Postinumeron_nimi_fi    string // 4
	Postinumeron_nimi_se    string // 5
	Postinumeron_lyhenne_fi string // 6
	Postinumeron_lyhenne_se string // 7
	Katu_fi                 string // 8
	Katu_se                 string // 9
	//Tyhja1                    string // 10
	//Tyhja2                    string // 11
	Kiinteistotietojen_tyyppi EvenOdd // 12

	Pienin Kiinteisto
	Suurin Kiinteisto

	Kunnan_koodi string // 25
	Kunta_fi     string // 26
	Kunta_se     string // 27
}

// JSON structures

type StreetJSON struct {
	Fi string `json:"fi,omitempty"`
	Se string `json:"se,omitempty"`
	Min int64 `json:"min,omitempty"`
	Max int64 `json:"max,omitempty"`
}

type PostnumberJSON struct {
	Fi string `json:"fi,omitempty"`
	Se string `json:"se,omitempty"`
	FiLyh string `json:"fil,omitempty"`
	SeLyh string `json:"sel,omitempty"`
}

type MunicipalityJSON struct {
	Fi string `json:"fi,omitempty"`
	Se string `json:"se,omitempty"`
}


// Converters
func strToConst(s string) EvenOdd {
	if s == "1" {
		return ODD
	} else if s == "2" {
		return EVEN
	} else {
		return NOT_USED
	}
}

func (posti StreetAddress) StreetNumberMinMax(arr []int64) (min int64, max int64) {
	var nums []int64 = []int64{posti.Pienin.Numero1, posti.Pienin.Numero2, posti.Suurin.Numero1, posti.Suurin.Numero2}
	nums = append(nums, arr...)
	return GetMinMaxArray(nums, -1)
}

func (posti StreetAddress) NewStreetJSON() StreetJSON {
	min,max := posti.StreetNumberMinMax([]int64{})

	return StreetJSON{
		Fi: posti.Katu_fi,
		Se: posti.Katu_se,
		Min: min,
		Max: max,
	}
}

func (posti StreetAddress) NewPostnumberJSON() PostnumberJSON {
	return PostnumberJSON{
		Fi: posti.Postinumeron_nimi_fi,
		Se: posti.Postinumeron_nimi_se,
		FiLyh:posti.Postinumeron_lyhenne_fi,
		SeLyh:posti.Postinumeron_lyhenne_se,
	}
}

func (posti StreetAddress) NewMunicipalityJSON() MunicipalityJSON {
	return MunicipalityJSON{
		Fi: posti.Kunta_fi,
		Se: posti.Kunta_se,
	}
}


func (posti StreetAddress) write_info() {
	//log.Printf("%s (%s) %s (%s) %s\n", posti.Kunta_fi, posti.Kunnan_koodi, posti.Postinumeron_nimi_fi, posti.Postinumero, posti.Katu_fi)
}

func (posti StreetAddress) write_street(dir string) {

	if posti.Katu_fi == "" {
		return
	}

	var nums []int64 = []int64{posti.Pienin.Numero1, posti.Pienin.Numero2, posti.Suurin.Numero1, posti.Suurin.Numero2}
	fmt.Printf("%v\n", nums)

	posti.write_info()

	filename := path.Join(dir, posti.Kunnan_koodi, posti.Postinumero, "street.json")

	var err error
	readbytes, err := ReadFileToByteArray(filename)

	var data []StreetJSON
	err = json.Unmarshal(readbytes, &data)
	if err != nil {
		log.Fatal(fmt.Sprintf("Filename: %[1]s Error: %[2]s (%[2]T)", filename, err))
		panic(err)
	}

	var found bool = false
	for idx, k := range data {
		if k.Fi == posti.Katu_fi {
			min,max := posti.StreetNumberMinMax([]int64{k.Min, k.Max})
			k.Min = min
			k.Max = max
			data[idx] = k
			found = true
			break
		}
	}

	if !found {
		data = append(data, posti.NewStreetJSON())
	}

	writebytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(filename, writebytes, os.FileMode(0600))
}

func (posti StreetAddress) write_postnumber(dir string) {
	if posti.Postinumeron_nimi_fi == "" {
		return
	}

	posti.write_info()

	filename := path.Join(dir, posti.Kunnan_koodi, posti.Postinumero, "postnumber.json")

	var err error
	readbytes, err := ReadFileToByteArray(filename)

	var data []PostnumberJSON
	err = json.Unmarshal(readbytes, &data)
	if err != nil {
		log.Fatal(fmt.Sprintf("Filename: %[1]s Error: %[2]s (%[2]T)", filename, err))
		panic(err)
	}

	var found bool = false
	for idx, k := range data {
		if k.Fi == posti.Postinumeron_nimi_fi {
			data[idx] = k
			found = true
			break
		}
	}

	if !found {
		data = append(data, posti.NewPostnumberJSON())
	}

	writebytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(filename, writebytes, os.FileMode(0600))

}

func (posti StreetAddress) write_municipality(dir string) {
	if posti.Kunta_fi == "" {
		return
	}

	posti.write_info()

	filename := path.Join(dir, posti.Kunnan_koodi, "municipality.json")

	var err error
	readbytes, err := ReadFileToByteArray(filename)

	var data []MunicipalityJSON
	err = json.Unmarshal(readbytes, &data)
	if err != nil {
		log.Fatal(fmt.Sprintf("Filename: %[1]s Error: %[2]s (%[2]T)", filename, err))
		panic(err)
	}

	var found bool = false
	for idx, k := range data {
		if k.Fi == posti.Kunta_fi {
			data[idx] = k
			found = true
			break
		}
	}

	if !found {
		data = append(data, posti.NewMunicipalityJSON())
	}

	writebytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(filename, writebytes, os.FileMode(0600))

}


func convertfile(sourcefile string, targetdir string){

	var err error

	converter, err := iconv.NewConverter("iso-8859-1", "utf-8")
	if err != nil {
		panic(err)
	}

	f, err := os.Open(sourcefile)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	buffer := make([]byte, 256)
	nl := make([]byte, 1) // new line

	finfo, err := f.Stat()
	var sourceTotalSizeBytes int64 = finfo.Size()
	var sourceReadedBytes int64 = 0

	// Ticker for stats
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		var raw RawLineStructure

		_, err := f.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				panic(err)
			}
		}

		// Read from file
		r := bytes.NewReader(buffer)
		pos, err := f.Seek(0, io.SeekCurrent)

		sourceReadedBytes = pos

		// Read to struct
		err = binary.Read(r, binary.BigEndian, &raw)
		if err != nil {
			panic(err)
		}

		// Convert to proper structs

		pienin := Kiinteisto{
			Numero1:      StringToInt64(BytesToString(raw.Pienin_numero1[:], converter)),     // 14
			Jakokirjain1: StringToByte(BytesToString(raw.Pienin_jakokirjain1[:], converter)), // 15
			Valimerkki:   StringToByte(BytesToString(raw.Pienin_valimerkki[:], converter)),   // 16
			Numero2:      StringToInt64(BytesToString(raw.Pienin_numero2[:], converter)),     // 17
			Jakokirjain2: StringToByte(BytesToString(raw.Pienin_jakokirjain2[:], converter)), // 18
		}

		suurin := Kiinteisto{
			Numero1:      StringToInt64(BytesToString(raw.Suurin_numero1[:], converter)),     // 20
			Jakokirjain1: StringToByte(BytesToString(raw.Suurin_jakokirjain1[:], converter)), // 21
			Valimerkki:   StringToByte(BytesToString(raw.Suurin_valimerkki[:], converter)),   // 22
			Numero2:      StringToInt64(BytesToString(raw.Suurin_numero2[:], converter)),     // 23
			Jakokirjain2: StringToByte(BytesToString(raw.Suurin_jakokirjain2[:], converter)), // 24
		}

		p := StreetAddress{
			Postinumero:               BytesToString(raw.Postinumero[:], converter),                              // 3
			Postinumeron_nimi_fi:      strings.ToLower(BytesToString(raw.Postinumeron_nimi_fi[:], converter)),    // 4
			Postinumeron_nimi_se:      strings.ToLower(BytesToString(raw.Postinumeron_nimi_se[:], converter)),    // 5
			Postinumeron_lyhenne_fi:   strings.ToLower(BytesToString(raw.Postinumeron_lyhenne_fi[:], converter)), // 6
			Postinumeron_lyhenne_se:   strings.ToLower(BytesToString(raw.Postinumeron_lyhenne_se[:], converter)), // 7
			Katu_fi:                   strings.ToLower(BytesToString(raw.Katu_fi[:], converter)),                 // 8
			Katu_se:                   strings.ToLower(BytesToString(raw.Katu_se[:], converter)),                 // 9
			Kiinteistotietojen_tyyppi: strToConst(BytesToString(raw.Kiinteistotietojen_tyyppi[:], converter)),    // 12
			Pienin:                    pienin,                                                                    // 14-18
			Suurin:                    suurin,                                                                    // 20-24
			Kunnan_koodi:              BytesToString(raw.Kunnan_koodi[:], converter),                             // 25
			Kunta_fi:                  strings.ToLower(BytesToString(raw.Kunta_fi[:], converter)),                // 26
			Kunta_se:                  strings.ToLower(BytesToString(raw.Kunta_se[:], converter)),                // 27
		}

		p.write_street(targetdir)
		p.write_postnumber(targetdir)
		p.write_municipality(targetdir)

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
				panic(err)
			}
		}

		if nl[0] != '\n' {
			panic(errors.New("Not newline"))
		}

	}

}