package apiserver

import (
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"vacancy-parser/internal/app/model"
	"vacancy-parser/internal/app/store"

	"github.com/gorilla/mux"
)

type APIServer struct {
	config *Config
	logger slog.Logger
	router *mux.Router
	store  *store.Store
}

type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
	// Meta  map[string]interface{} `json:"meta,omitempty"`
}

func New(config *Config) *APIServer {
	return &APIServer{
		config: config,
		logger: *slog.New(slog.NewJSONHandler(os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug})),
		router: mux.NewRouter(),
	}
}

func (s *APIServer) Start() error {
	s.configureRouter()

	if err := s.configureStore(); err != nil {
		return err
	}

	s.logger.Info("starting server")
	s.logger.Info("port listening", slog.String("port", s.config.BindAddr))
	return http.ListenAndServe(s.config.BindAddr, s.router)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func (s *APIServer) configureRouter() {
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			enableCors(&w)
			if r.Method == "OPTIONS" {
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	s.router.HandleFunc("/hello", s.handleHello())
	s.router.HandleFunc("/user", s.CreateUser).Methods(http.MethodPost)
	s.router.HandleFunc("/user/{email}", s.GetUserByEmail).Methods(http.MethodGet)
	s.router.HandleFunc("/users", s.GetAllUsers).Methods(http.MethodGet)
	s.router.HandleFunc("/user/{email}", s.UpdateUserByEmail).Methods(http.MethodPut)
	s.router.HandleFunc("/user/{email}", s.DeleteUserByEmail).Methods(http.MethodDelete)
	s.router.HandleFunc("/users", s.DeleteAllUsers).Methods(http.MethodDelete)
	s.router.HandleFunc("/vacancy", s.InsertVacancy).Methods(http.MethodPost)
	s.router.HandleFunc("/vacancies/count/", s.GetAllVacanciesCount).Methods(http.MethodGet)
	s.router.HandleFunc("/vacancies/hardSkills/", s.GetAllHardSkills).Methods(http.MethodGet)
	s.router.HandleFunc("/vacancies/{page:[0-9]+}/{limit:[0-9]+}/", s.GetVacancies).Methods(http.MethodGet)
	s.router.HandleFunc("/vacancies/", s.GetAllVacancies).Methods(http.MethodGet)
}

func (s *APIServer) configureStore() error {
	st := store.New(s.config.Store)

	if err := st.Open(); err != nil {
		return err
	}

	s.store = st

	return nil
}

func (s *APIServer) handleHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello")
	}
}

func (s *APIServer) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	res := &Response{}
	defer json.NewEncoder(w).Encode(res)

	var user model.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("invalid request body", err)
		res.Error = err.Error()
		return
	}

	repo := s.store.User()
	_, err = repo.CreateUser(&user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("cannot create user", err)
		res.Error = err.Error()
		return
	}

	res.Data = user.Email
	w.WriteHeader(http.StatusOK)
	log.Println("created user", user.Email)
}
func (s *APIServer) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	res := &Response{}
	defer json.NewEncoder(w).Encode(res)

	email := mux.Vars(r)["email"]
	log.Println("email:", email)

	repo := s.store.User()
	user, err := repo.FindByEmail(email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("cannot find user", err)
		res.Error = err.Error()
		return
	}

	res.Data = user
	w.WriteHeader(http.StatusOK)
	log.Println("found user", user.Email)
}
func (s *APIServer) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	res := &Response{}
	defer json.NewEncoder(w).Encode(res)

	repo := s.store.User()
	users, err := repo.FindAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("cannot find users", err)
		res.Error = err.Error()
		return
	}

	res.Data = users
	w.WriteHeader(http.StatusOK)
	log.Println("found users")
}
func (s *APIServer) UpdateUserByEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	res := &Response{}
	defer json.NewEncoder(w).Encode(res)

	email := mux.Vars(r)["email"]

	repo := s.store.User()
	user, err := repo.FindByEmail(email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("cannot find user", err)
		res.Error = err.Error()
		return
	}

	var updUser model.User
	err = json.NewDecoder(r.Body).Decode(&updUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("invalid request body", err)
		res.Error = err.Error()
		return
	}

	user.Email = email
	count, err := repo.UpdateUserByEmail(email, &updUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("cannot update user", err)
		res.Error = err.Error()
		return
	}

	res.Data = count

	w.WriteHeader(http.StatusOK)
	log.Println("updated user", email)
}
func (s *APIServer) DeleteUserByEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	res := &Response{}

	email := mux.Vars(r)["email"]

	repo := s.store.User()

	count, err := repo.DeleteUserByEmail(email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("cannot delete user", err)
		res.Error = err.Error()
		return
	}

	res.Data = count

	w.WriteHeader(http.StatusOK)
	log.Println("deleted user", email)
}
func (s *APIServer) DeleteAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	res := &Response{}

	repo := s.store.User()

	count, err := repo.DeleteAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("cannot delete users", err)
		res.Error = err.Error()
		return
	}

	res.Data = count

	w.WriteHeader(http.StatusOK)
	log.Println("deleted all users")
}

func (s *APIServer) InsertVacancy(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	res := &Response{}

	var vac model.Vacancy
	err := json.NewDecoder(r.Body).Decode(&vac)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("invalid request body", err)
		res.Error = err.Error()
		return
	}

	repo := s.store.Vacancy()

	_, err = repo.InsertVacancy(&vac)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("cannot insert vacancy", err)
		res.Error = err.Error()
		return
	}

	res.Data = vac

	w.WriteHeader(http.StatusOK)
	log.Println("inserted vacancy")
}

func (s *APIServer) GetVacancies(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	res := &Response{}
	defer json.NewEncoder(w).Encode(res)

	vars := mux.Vars(r)
	page, err := strconv.Atoi(vars["page"])
	if err != nil {
		log.Fatal(err)
	}
	limit, err := strconv.Atoi(vars["limit"])
	if err != nil {
		log.Fatal(err)
	}

	repo := s.store.Vacancy()

	vacs, err := repo.GetVacancies(int64(page), int64(limit))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("cannot find vacancies", err)
		res.Error = err.Error()
		return
	}

	res.Data = vacs

	w.WriteHeader(http.StatusOK)
	log.Println("page", page)
	log.Println("limit", limit)
	log.Println("found vacancies")
}

func (s *APIServer) GetAllVacancies(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	res := &Response{}
	defer json.NewEncoder(w).Encode(res)

	repo := s.store.Vacancy()

	vacs, err := repo.FindAllVacancy()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("cannot find vacancies", err)
		res.Error = err.Error()
		return
	}

	res.Data = vacs

	w.WriteHeader(http.StatusOK)
	log.Println("found vacancies")
}

func (s *APIServer) GetAllVacanciesCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	res := &Response{}
	defer json.NewEncoder(w).Encode(res)

	repo := s.store.Vacancy()

	count, err := repo.GetAllVacanciesCount()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("cannot find vacancies", err)
		res.Error = err.Error()
		return
	}

	res.Data = count

	w.WriteHeader(http.StatusOK)
	log.Println("found count of vacancies")
}

func (s *APIServer) GetAllHardSkills(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	res := &Response{}
	defer json.NewEncoder(w).Encode(res)

	repo := s.store.Vacancy()

	skills, err := repo.GetAllHardSkills()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("cannot find vacancies", err)
		res.Error = err.Error()
		return
	}

	res.Data = skills

	w.WriteHeader(http.StatusOK)
	log.Println("found skills")
}
