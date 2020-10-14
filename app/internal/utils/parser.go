package utils

import (
	"avito/internal"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/xerrors"
	"net/http"
	"regexp"
	"strconv"
)

func ParseAd(link string) (*internal.Ad, error) {
	ad := &internal.Ad{}

	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	closedWarning := doc.Find(".item-closed-warning").First().Text()
	if len(closedWarning) != 0 {
		ad.Removed = true
		return ad, nil
	}

	ad.Name = doc.Find(".title-info-title-text").First().Text()
	price := regexp.MustCompile(`\D+`).ReplaceAllString(doc.Find("[itemprop=\"price\"]").First().Text(), "")
	if ad.Name == "" && price == "" {
		return nil, xerrors.Errorf("Advert name or price is empty")
	}

	ad.Price, err = strconv.Atoi(price)
	if err != nil {
		return nil, err
	}

	ad.Link = link

	return ad, nil
}
