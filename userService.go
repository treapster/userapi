package main

import (
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"strconv"
	"time"
)

type (
	UserStore struct {
		Increment int      `json:"increment"`
		List      UserList `json:"list"`
	}
	DefaultUserService struct {
		Filename string
	}
	UserService interface {
		CreateUser(req CreateUserRequest) string
		FindAllUsers() UserList
		GetUser(id string) (User, error)
		UpdateUser(req UpdateUserRequest, id string) error
		DeleteUser(id string) error
	}
)

func (us *DefaultUserService) getStore() UserStore {
	f, _ := ioutil.ReadFile(us.Filename)
	s := UserStore{}
	_ = json.Unmarshal(f, &s)
	return s
}
func (us *DefaultUserService) setStore(s UserStore) {
	b, _ := json.Marshal(&s)
	_ = ioutil.WriteFile(us.Filename, b, fs.ModePerm)
}
func NewDefaultUserService(name string) *DefaultUserService {
	return &DefaultUserService{Filename: name}
}

func (us *DefaultUserService) FindAllUsers() UserList {
	s := us.getStore()
	return s.List
}
func (us *DefaultUserService) CreateUser(req CreateUserRequest) string {
	s := us.getStore()
	s.Increment++
	u := User{
		CreatedAt:   time.Now(),
		DisplayName: req.DisplayName,
		Email:       req.DisplayName,
	}
	id := strconv.Itoa(s.Increment)
	s.List[id] = u
	us.setStore(s)
	return id
}
func (us *DefaultUserService) GetUser(id string) (User, error) {
	s := us.getStore()
	if _, ok := s.List[id]; !ok {
		return User{}, UserNotFound
	}
	return s.List[id], nil
}
func (us *DefaultUserService) UpdateUser(req UpdateUserRequest, id string) error {
	s := us.getStore()
	if _, ok := s.List[id]; !ok {
		return UserNotFound
	}
	u := s.List[id]
	u.DisplayName = req.DisplayName
	s.List[id] = u
	us.setStore(s)
	return nil
}

func (us *DefaultUserService) DeleteUser(id string) error {
	s := us.getStore()
	if _, ok := s.List[id]; !ok {
		return UserNotFound
	}
	delete(s.List, id)

	us.setStore(s)
	return nil
}
