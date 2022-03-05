package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

const store = `users.json`

type (
	User struct {
		CreatedAt   time.Time `json:"created_at"`
		DisplayName string    `json:"display_name"`
		Email       string    `json:"email"`
	}
	UserList          map[string]User
	CreateUserRequest struct {
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
	}
	UpdateUserRequest struct {
		DisplayName string `json:"display_name"`
	}
	ErrResponse struct {
		Err            error `json:"-"`
		HTTPStatusCode int   `json:"-"`

		StatusText string `json:"status"`
		AppCode    int64  `json:"code,omitempty"`
		ErrorText  string `json:"error,omitempty"`
	}
)

var (
	UserNotFound = errors.New("user_not_found")
)

func initRouter(userRouter func(r chi.Router)) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(time.Now().String()))
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Route("/users", userRouter)
		})
	})
	return r
}
func main() {
	us := NewDefaultUserService(store)
	r := initRouter(us.NewRouter())
	err := http.ListenAndServe(":3333", r)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func findAllUsers(w http.ResponseWriter, r *http.Request, us UserService) {

	render.JSON(w, r, us.FindAllUsers())
}

func (c *CreateUserRequest) Bind(r *http.Request) error { return nil }

func createUser(w http.ResponseWriter, r *http.Request, us UserService) {
	request := CreateUserRequest{}
	if err := render.Bind(r, &request); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"user_id": us.CreateUser(request),
	})
}

func getUser(w http.ResponseWriter, r *http.Request, us UserService) {

	id := chi.URLParam(r, "id")
	u, err := us.GetUser(id)
	if err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.JSON(w, r, u)
}

func (c *UpdateUserRequest) Bind(r *http.Request) error { return nil }

func updateUser(w http.ResponseWriter, r *http.Request, us UserService) {

	request := UpdateUserRequest{}

	if err := render.Bind(r, &request); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	id := chi.URLParam(r, "id")

	if err := us.UpdateUser(request, id); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Status(r, http.StatusNoContent)
}

func deleteUser(w http.ResponseWriter, r *http.Request, us UserService) {

	id := chi.URLParam(r, "id")

	if err := us.DeleteUser(id); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Status(r, http.StatusNoContent)
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}
