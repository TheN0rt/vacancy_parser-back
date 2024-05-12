package main

import (
	"flag"
	"fmt"
	"log"

	"runtime"

	"sync"
	"vacancy-parser/internal/app/apiserver"

	"vacancy-parser/internal/app/store"

	"vacancy-parser/internal/app/parser"

	"github.com/BurntSushi/toml"
)

// Создать папку в internal/lib parser
// Создать логику для парсинга habr vacancy
// Перенести всю логику парсинга в модуль store в методы для парсинга

const page = 20

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "A:/go/vacancy-parser/configs/apiserver.toml", "path to config file") // Изменить путь до конфига
}

func main() {
	flag.Parse()
	config := apiserver.NewConfig()
	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		log.Fatal(err)
	}

	startParser(config, "JavaScript")

	s := apiserver.New(config)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}

func startParser(config *apiserver.Config, language string) {
	store := store.New(config.Store)

	err := store.Open()
	if err != nil {
		log.Fatal(err)
	}

	defer store.Close()
	var repo = store.Vacancy()

	deleteAllVacancy(repo)

	var URLSlice = parser.GetURLS(page, language)
	for i := 0; i < len(URLSlice); i++ {
		fmt.Println(URLSlice[i])
	}

	var wg sync.WaitGroup
	wg.Add(len(URLSlice))
	sem := make(chan struct{}, runtime.GOMAXPROCS(12)) // Использовать все доступные ядра
	var mu sync.Mutex
	for i := 0; i < len(URLSlice); i++ {
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() {
				<-sem
			}()
			vacancyInfo := parser.GetInfoFromUrl(URLSlice[i], language)
			if vacancyInfo == nil {
				fmt.Println("Skip url:", URLSlice[i])
				return
			}
			mu.Lock()
			_, err = repo.InsertVacancy(vacancyInfo)
			mu.Unlock()
			if err != nil {
				log.Fatal(err)
			}
		}(i)
	}
	wg.Wait()
}

func deleteAllVacancy(repo *store.VacancyRepository) {
	_, err := repo.DeleteAllVacancy()
	if err != nil {
		log.Fatal(err)
	}
}
