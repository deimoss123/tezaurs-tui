package parser

import (
	"fmt"
	"io"
	"strconv"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
)

type Entry struct {
	NumStr  string
	Content string
}

type TableColItem struct {
	Width int
}

type TableRowItem struct {
	Text      string
	ColSpan   int
	ColId     int
	IsThead   bool
	IsHeading bool
}

type ConjugationTable struct {
	ColumnCount int
	ColItems    []TableColItem
	RowItems    [][]TableRowItem
}

type ParsedHtml struct {
	Entries       []Entry
	Verbalisation string
	ConjTable     ConjugationTable
	TestText      string
}

func ParseHtml(body io.ReadCloser) (ParsedHtml, error) {
	defer body.Close()

	res := ParsedHtml{}

	doc, err := goquery.NewDocumentFromReader(body)

	if err != nil {
		return res, err
	}

	// definīcijas
	doc.Find(".dict_Sense").Each(func(i int, sel *goquery.Selection) {
		numStr := sel.Find(".dict_SenseNumber").First().Text()
		content := sel.Find(".dict_Gloss").First().Text()

		res.Entries = append(res.Entries, Entry{numStr, content})
	})

	res.Verbalisation = doc.Find(".dict_Verbalization").First().Text()

	table := doc.Find(".dict_MorphoTable table.inflections").First()

	if table.Length() > 0 {
		// saskaitīt kolonnu skaitu
		colCount := 0
		table.Find("tr").First().Children().Each(func(i int, sel *goquery.Selection) {
			colspan, exists := sel.First().Attr("colspan")
			if exists {
				num, err := strconv.Atoi(colspan)
				if err != nil {
					colCount += 1
				} else {
					colCount += num
				}
			} else {
				colCount += 1
			}
		})

		res.ConjTable.ColumnCount = colCount
		res.ConjTable.ColItems = make([]TableColItem, colCount)

		// saliekam rindas un to elementus
		table.Find("tr").Each(func(_ int, sel *goquery.Selection) {
			// row := append(res.ConjTable.RowItems, []TableRowItem{})

			parentNodeName := goquery.NodeName(sel.Parent().First())
			isThead := parentNodeName == "thead"

			row := make([]TableRowItem, colCount)
			colIdx := 0

			sel.Children().Each(func(_ int, col *goquery.Selection) {
				nodeName := goquery.NodeName(col)

				colspanStr, exists := col.Attr("colspan")

				colspan := 1

				if exists {
					num, err := strconv.Atoi(colspanStr)
					if err == nil {
						colspan = num
					}
				}

				// res.ConjTable.RowItems = append(
				// 	res.ConjTable.RowItems,
				// 	TableRowItem{Text: col.Text(), ColSpan: colspan, ColId: colIdx},
				// )

				row[colIdx] = TableRowItem{
					Text:      col.Text(),
					ColSpan:   colspan,
					ColId:     colIdx,
					IsThead:   isThead,
					IsHeading: nodeName == "th",
				}

				colIdx += colspan
			})

			res.ConjTable.RowItems = append(res.ConjTable.RowItems, row)
		})

		// izejam cauri kolonnām un iegūstam to platumus
		for colIndex := range res.ConjTable.ColItems {
			maxWidth := 0
			for j := 0; j < len(res.ConjTable.RowItems); j++ {
				length := utf8.RuneCountInString(res.ConjTable.RowItems[j][colIndex].Text)
				if length > maxWidth {
					maxWidth = length
				}
			}
			res.ConjTable.ColItems[colIndex].Width = maxWidth
		}

		res.TestText = "ir tabula, kolonnu sk. " + strconv.Itoa(colCount) + fmt.Sprintf("\n\n%+v", res.ConjTable)
	} else {
		res.TestText = "nav tabulas"
	}

	return res, nil
}
