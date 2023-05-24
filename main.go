package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

type Item struct {
	Image string
	Brand string
	Name  string
	Url   string
	Sizes []string
	Price string
}

func main() {
	domain := "www.yoox.com"
	baseUrl := fmt.Sprintf("https://%s/uk/shoponline", domain)
	requiredSizes := []string{"13", "13.5", "14", "14.5"}
	var items []*Item
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.64 Safari/537.11"),
		colly.AllowedDomains(domain),
	)

	c.OnHTML(
		"div[data-srpage]", func(e *colly.HTMLElement) {
			e.DOM.Children().Each(
				func(_ int, itemNode *goquery.Selection) {
					var matchedSizes []string
					itemNode.Find(".aSize").Each(
						func(_ int, aSize *goquery.Selection) {
							size := aSize.Text()
							for _, requiredSize := range requiredSizes {
								if requiredSize == size {
									matchedSizes = append(matchedSizes, size)
								}
							}
						},
					)
					if len(matchedSizes) == 0 {
						return
					}
					item := &Item{Sizes: matchedSizes}
					itemLink := itemNode.Find(".itemlink")
					if path, ok := itemLink.Attr("href"); ok {
						itemUrl := fmt.Sprintf("%s%s", fmt.Sprintf("https://%s", domain), path)
						item.Url = itemUrl
					}
					imgSelection := itemNode.Find(".front.imgFormat_20_f")
					img, ok := imgSelection.Attr("data-original")
					if !ok {
						img = imgSelection.AttrOr("src", "")
					}
					item.Image = img
					if brand := itemLink.Find(".brand").Text(); brand != "" {
						item.Brand = strings.TrimSpace(brand)
					}
					if title := itemLink.Find(".title").Text(); title != "" {
						item.Name = strings.TrimSpace(title)
					}
					price := itemLink.Find(".retail-newprice").Text()
					if price == "" {
						price = itemLink.Find(".fullprice").Text()
					}
					item.Price = strings.TrimSpace(price)
					items = append(items, item)
				},
			)
		},
	)

	c.OnHTML(
		".next-page", func(e *colly.HTMLElement) {
			if rel, ok := e.DOM.Find("a").Attr("rel"); ok {
				relPath := strings.TrimPrefix(rel, "address:")
				nextUrl := fmt.Sprintf("%s?%s", baseUrl, relPath)
				c.Visit(nextUrl)
			}
		},
	)

	c.OnRequest(
		func(r *colly.Request) {
			fmt.Println("scraping", r.URL.String())
			fmt.Println()
		},
	)

	if err := c.Visit(fmt.Sprintf("%s%s", baseUrl, "?dept=shoesmen&gender=U&page=1&size=8")); err != nil {
		fmt.Printf("error is %s", err)
	}
	c.Wait()
	sort.Slice(
		items, func(i, j int) bool {
			return items[i].Brand < items[j].Brand
		},
	)
	for _, v := range items {
		fmt.Println(v)
	}
	fmt.Printf("scrape finished, %v items found.", len(items))
}
