package model

type Vacancy struct {
	Title        string   `json:"title"`
	Link         string   `json:"link"`
	Location     string   `json:"location"`
	Company      string   `json:"company"`
	HardSkills   []string `json:"hardSkills"`
	Site         string   `json:"site"`
	Date         string   `json:"date"`
	Salary       string   `json:"salary"`
	Experience   string   `json:"experience"`
	MainLanguage string   `json:"mainLanguage"`
}
