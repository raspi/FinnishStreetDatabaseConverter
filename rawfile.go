package main

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
	SmallestBuildingNumber1         [5]byte // #14 Building number 1, optional
	SmallestBuildingDeliveryLetter1 [1]byte // #15 Building delivery letter 1, optional
	SmallestPunctuationMark         [1]byte // #16 Punctuation mark, optional
	SmallestBuildingNumber2         [5]byte // #17 Building number 2, optional
	SmallestBuildingDeliveryLetter2 [1]byte // #18 Building delivery letter 2, optional

	// #19 (skipped) Highest building number (information about an odd/even building)
	HighestBuildingNumber1         [5]byte // #20 Building number 1, optional
	HighestBuildingDeliveryLetter1 [1]byte // #21 Building delivery letter 1, optional
	HighestPunctuationMark         [1]byte // #22 Punctuation mark, optional
	HighestBuildingNumber2         [5]byte // #23 Building number 2, optional
	HighestBuildingDeliveryLetter2 [1]byte // #24 Building delivery letter 2, optional

	MunicipalityCode   [3]byte  // #25 Municipality code, numeric
	MunicipalityNameFi [20]byte // #26 Municipality name in Finnish
	MunicipalityNameSe [20]byte // #27 Municipality name in Swedish, optional
}
