# Finnish street database converter

Convert Finnish post office's (Posti) street database file (`BAF_yyyymmdd.dat`) to JSON.

## Sources:
* English: https://www.posti.fi/business/help-and-support/postal-code-services/postal-code-files.html
* Finnish: https://www.posti.fi/yritysasiakkaat/apu-ja-tuki/postinumeropalvelut/postinumerotiedostot.html
* Database: http://www.posti.fi/webpcode/

## Basic Address File Record Description
File format (`BAF_yyyymmdd.dat`): ISO-8859-1 formatted text file separated by newlines (`\n`).

| #   | Position | Length | Optional | Description                                                       | Example            |
|-----|----------|--------|----------|-------------------------------------------------------------------|--------------------|
|  1. |        1 |      5 |          | Record identifier                                                 | "`KATUN`"          |
|  2. |        6 |      8 |          | Running date                                                      | yyyymmdd, 20171231 |
|  3. |       14 |      5 |          | Postal code                                                       | "`40100`"          |
|  4. |       19 |     30 |          | Postal code name in Finnish                                       | "`Jyväskylä`"      |
|  5. |       49 |     30 |      [x] | Postal code name in Swedish                                       |                    |
|  6. |       79 |     12 |          | Postal code name abbreviation in Finnish                          | "`jkl`"            |
|  7. |       91 |     12 |      [x] | Postal code name abbreviation in Swedish                          |                    |
|  8. |      103 |     30 |          | Street (location) name in Finnish                                 | "`vapaudenkatu`"   |
|  9. |      133 |     30 |      [x] | Street (location) name in Swedish                                 |                    |
| 10. |      163 |     12 |          | Blank                                                             | " " * 12           |
| 11. |      175 |     12 |          | Blank                                                             | " " * 12           |
| 12. |      187 |      1 |      [x] | Building data type                                                | 1 = odd, 2 = even  |
| 13. |          |        |          | Smallest building number (information about an odd/even building) |                    |
| 14. |      188 |      5 |      [x] | Smallest building number 1                                        | "1"                |
| 15. |      193 |      1 |      [x] | Smallest building delivery letter 1                               | "A"                |
| 16. |      194 |      1 |      [x] | Smallest building punctuation mark                                | "/"                |
| 17. |      195 |      5 |      [x] | Smallest building number 2                                        | "10"               |
| 18. |      200 |      1 |      [x] | Smallest building delivery letter 2                               | "C"                |
| 19. |          |        |          | Highest building number (information about an odd/even building)  |                    |
| 20. |      201 |      5 |      [x] | Highest building number 1                                         | "10"               |
| 21. |      206 |      1 |      [x] | Highest building delivery letter 1                                | "F"                |
| 22. |      207 |      1 |      [x] | Highest building punctuation mark                                 | "-"                |
| 23. |      208 |      5 |      [x] | Highest building number 2                                         | "11"               |
| 24. |      213 |      1 |      [x] | Highest building delivery letter 2                                | "E"                |
| 25. |      214 |      3 |          | Municipality code                                                 |                    |
| 26. |      217 |     20 |          | Municipality name in Finnish                                      |                    |
| 27. |      237 |     20 |      [x] | Municipality name in Swedish                                      |                    |
| 28. |      257 |      1 |          | New line                                                          | "\n"               |
