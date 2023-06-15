package application

import (
	"app/internal/repository"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

type app struct {
	ctx   context.Context
	repo  *repository.Repository
	cache map[string]repository.User
}

func (a app) Routes(r *httprouter.Router) {
	r.ServeFiles("/public/*filepath", http.Dir("public"))
	r.GET("/", a.authorized(a.StartPage))
	r.GET("/login", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.LoginPage(rw, "")
	})
	r.POST("/login", a.Login)
	r.GET("/logout", a.Logout)
	r.GET("/signup", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SignupPage(rw, "")
	})
	r.POST("/signup", a.Signup)
}

func (a app) StartPage(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	motivation, err := a.repo.GetRandomMotivation(a.ctx)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	lp := filepath.Join("public", "html", "motivation.html.go")
	tmpl, err := template.ParseFiles(lp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	err = tmpl.ExecuteTemplate(rw, "motivation", motivation)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func NewApp(ctx context.Context, dbpool *pgxpool.Pool) *app {
	return &app{ctx, repository.NewRepository(dbpool), make(map[string]repository.User)}
}

func (a app) LoginPage(rw http.ResponseWriter, message string) {
	lp := filepath.Join("public", "html", "login.html.go")
	tmpl, err := template.ParseFiles(lp)
	if err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}

	type answer struct {
		Message string
	}
	data := answer{message}
	err = tmpl.ExecuteTemplate(rw, "login", data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a app) Login(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	login := r.FormValue("login")
	password := r.FormValue("password")
	if login == "" || password == "" {
		http.Error(rw, fmt.Sprintf("Необходимо указать логин и пароль!"), 400)
		return
	}

	hash := md5.Sum([]byte(password))
	hashedPassword := hex.EncodeToString(hash[:])
	user, err := a.repo.Login(a.ctx, login, hashedPassword)
	if err != nil {
		a.LoginPage(rw, fmt.Sprintf("Вы ввели неверный логин или пароль!: %s", err.Error()))
		return
	}

	// генерируем токен пишем его в кэш и cookie
	time64 := time.Now().Unix()
	timeInt := string(time64)
	token := login + password + timeInt
	hashToken := md5.Sum([]byte(token))
	hashedToken := hex.EncodeToString(hashToken[:])
	a.cache[hashedToken] = user
	livingTime := 60 * time.Minute
	expiration := time.Now().Add(livingTime)
	// кука будет жить 1 час
	cookie := http.Cookie{Name: "token", Value: url.QueryEscape(hashedToken), Expires: expiration}
	http.SetCookie(rw, &cookie)

	http.Redirect(rw, r, "/", http.StatusSeeOther)
}

func (a app) Logout(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	for _, v := range r.Cookies() {
		c := http.Cookie{
			Name:   v.Name,
			MaxAge: -1}
		http.SetCookie(rw, &c)
	}
	http.Redirect(rw, r, "/login", http.StatusSeeOther)
}

func (a app) SignupPage(rw http.ResponseWriter, message string) {
	lp := filepath.Join("public", "html", "signup.html.go")
	tmpl, err := template.ParseFiles(lp)
	if err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}

	type answer struct {
		Message string
	}
	data := answer{message}
	err = tmpl.ExecuteTemplate(rw, "signup", data)
	if err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}
}

func (a app) Signup(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// получаем данные формы
	name := strings.TrimSpace(r.FormValue("name"))
	surname := strings.TrimSpace(r.FormValue("surname"))
	login := strings.TrimSpace(r.FormValue("login"))
	password := strings.TrimSpace(r.FormValue("password"))
	password2 := strings.TrimSpace(r.FormValue("password2"))

	fmt.Println("name:", name)
	fmt.Println("surname:", surname)
	fmt.Println("login:", login)
	fmt.Println("password:", password)
	fmt.Println("password2:", password2)
	// проверяем валидность данных
	if name == "" || surname == "" || login == "" || password == "" {
		a.SignupPage(rw, "Все поля должны быть заполнены!")
		return
	}

	// сравниваем пароли
	if password != password2 {
		a.SignupPage(rw, "Пароли не совпадают! Попробуйте еще")
		return
	}

	//	хешируем пароли
	hash := md5.Sum([]byte(password))
	hashedPassword := hex.EncodeToString(hash[:])

	// добавляем нового пользователя
	err := a.repo.AddNewUser(a.ctx, name, surname, login, hashedPassword)
	if err != nil {
		a.SignupPage(rw, fmt.Sprintf("Ошибка создания пользователя %s", err.Error()))
		return
	}
	a.LoginPage(rw, fmt.Sprintf("%s, Вы успешно зарегистрированы!", name))
}

func (a app) authorized(next httprouter.Handle) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		token, err := readCookie("token", r)
		// если нет куки с токеном - пользователь отправляется на авторизацию
		if err != nil {
			http.Redirect(rw, r, "/login", http.StatusSeeOther)
			return
		}
		// если токен есть, но его нет в кеше - пользователь отправляется на авторизацию
		if _, ok := a.cache[token]; !ok {
			http.Redirect(rw, r, "/login", http.StatusSeeOther)
			return
		}
		next(rw, r, p)
	}
}

func readCookie(name string, r *http.Request) (value string, err error) {
	if name == "" {
		return value, errors.New("Вы пытаетесь прочитать пустой cookie")
	}
	cookie, err := r.Cookie(name)
	if err != nil {
		return
	}
	str := cookie.Value
	value, _ = url.QueryUnescape(str)
	return
}
