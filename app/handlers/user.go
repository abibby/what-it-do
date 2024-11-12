package handlers

import (
	"context"

	"github.com/abibby/salusa/database"
	"github.com/abibby/salusa/request"
	"github.com/abibby/what-it-do/app/models"
	"github.com/jmoiron/sqlx"
)

type ListUserRequest struct {
	Read database.Read   `inject:""`
	Ctx  context.Context `inject:""`
}
type ListUserResponse struct {
	Users []*models.User `json:"users"`
}

var UserList = request.Handler(func(r *ListUserRequest) (*ListUserResponse, error) {
	var users []*models.User
	var err error
	err = r.Read(func(tx *sqlx.Tx) error {
		users, err = models.UserQuery(r.Ctx).Get(tx)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &ListUserResponse{
		Users: users,
	}, nil
})

type GetUserRequest struct {
	User *models.User `inject:"id"`
}
type GetUserResponse struct {
	User *models.User `json:"user"`
}

var UserGet = request.Handler(func(r *GetUserRequest) (*GetUserResponse, error) {
	return &GetUserResponse{
		User: r.User,
	}, nil
})
