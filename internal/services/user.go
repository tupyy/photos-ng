package services

import "git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (u *UserService) GetUsers() []entity.User {
	return []entity.User{}
}
