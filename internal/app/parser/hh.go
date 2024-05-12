package parser

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"vacancy-parser/internal/app/model"

	"github.com/PuerkitoBio/goquery"
)

func GetURLS(page int, language string) []string {
	var URLSlice []string
	var lastPageInt int

	// Send an HTTP GET request to the HH job openings page
	resp, err := http.Get("https://hh.ru/search/vacancy?text=Javascript&from=suggest_post&area=1&hhtmFrom=main&hhtmFromLabel=vacancy_search_line&page=0")
	if err != nil {
		log.Fatal(err)
	}
	resp.Request.URL.RawQuery = strings.Replace(resp.Request.URL.RawQuery, "korolev.", "", 1)
	defer resp.Body.Close()

	// Create a new goquery document from the HTML response
	if page == 0 {
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		lastPage := doc.Find("div.pager a").Slice(-2, -1).First().Text()
		lastPageInt, err = strconv.Atoi(lastPage)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Last page:", lastPage)
		fmt.Println("URLS is got")
	} else {
		lastPageInt = page
	}

	for i := page; i < lastPageInt; i++ {
		// .vacancy-serp-item_clickme - avoid this links
		// Send an HTTP GET request to the HH job openings page
		resp, err := http.Get("https://hh.ru/search/vacancy?text=" + language + "&from=suggest_post&area=1&hhtmFrom=main&hhtmFromLabel=vacancy_search_line&page=" + strconv.Itoa(i))
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		// Create a new goquery document from the HTML response
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		doc.Find("div.serp-item_link").Each(func(i int, s *goquery.Selection) {
			if !s.HasClass("vacancy-serp-item_clickme") {
				url := s.Find("a.bloko-link").AttrOr("href", "")
				URLSlice = append(URLSlice, url)
			}
		})
	}

	return URLSlice
}

func GetInfoFromUrl(url string, language string) *model.Vacancy {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		if strings.Contains(err.Error(), "connectex: A connection attempt failed because the connected party did not properly respond after a period of time, or established connection failed because connected host has failed to respond") {
			return nil
		}
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	// Create a new goquery document from the HTML response
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil
	}

	title := doc.Find("div.vacancy-title").Text()
	if len(title) == 0 {
		title = doc.Find("h1[data-qa='vacancy-title']").First().Text()
	}

	location := doc.Find("span[data-qa='vacancy-view-raw-address']").First().Text()
	if len(location) == 0 {
		location = doc.Find("p[data-qa='vacancy-view-location']").First().Text()
	}

	var hardSkillSlice []string
	doc.Find("span.bloko-tag__section_text").Each(func(i int, s *goquery.Selection) {
		hardSkillSlice = append(hardSkillSlice, s.Text())
	})

	date := doc.Find("p.vacancy-creation-time-redesigned span").Text()

	company := doc.Find("span.vacancy-company-name").First().Text()
	salary := doc.Find("div[data-qa='vacancy-salary']").Text()

	experience := doc.Find("span[data-qa='vacancy-experience']").Text()

	fmt.Println("Title:", title)
	fmt.Println("Location:", location)
	fmt.Println("Hard skills:", hardSkillSlice)
	fmt.Println("URL:", url)

	return &model.Vacancy{
		Title:        title,
		Location:     location,
		HardSkills:   hardSkillSlice,
		Link:         url,
		Company:      company,
		Site:         "hh.ru",
		Date:         date,
		Salary:       salary,
		Experience:   experience,
		MainLanguage: language,
	}
}
