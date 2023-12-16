package parser

import (
	"io"

	"github.com/PuerkitoBio/goquery"
)

type Entry struct {
	NumStr  string
	Content string
}

type ParsedHtml struct {
	Entries []Entry
}

func ParseHtml(body io.ReadCloser) (ParsedHtml, error) {
	defer body.Close()

	res := ParsedHtml{}

	doc, err := goquery.NewDocumentFromReader(body)

	if err != nil {
		return res, err
	}

	doc.Find(".dict_Sense").Each(func(i int, sel *goquery.Selection) {
		numStr := sel.Find(".dict_SenseNumber").First().Text()
		content := sel.Find(".dict_Gloss").First().Text()

		res.Entries = append(res.Entries, Entry{numStr, content})
	})

	return res, nil
}
