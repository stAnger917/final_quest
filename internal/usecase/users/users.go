package users

import (
	"context"
	"errors"
	"final_quest/internal/errs"
	"final_quest/internal/models"
	"final_quest/internal/repository"
	"final_quest/pkg/hasher"
	"final_quest/pkg/logging"
	"fmt"
	"github.com/golang-module/carbon/v2"
	"github.com/theplant/luhn"
	"strconv"
)

type Users struct {
	repository *repository.AppRepo
	logger     *logging.Logger
}

func NewUsersUseCase(repo *repository.AppRepo, logger *logging.Logger) *Users {
	return &Users{
		repository: repo,
		logger:     logger,
	}
}

func (u *Users) CreateNewUser(ctx context.Context, login, password string) error {
	isUserExists, err := u.repository.CheckIfUserExists(ctx, login)
	if err != nil {
		return err
	}
	if isUserExists {
		return errors.New("user already exists")
	}
	hashedPassword := hasher.HashPassword(password)
	err = u.repository.CreateNewUser(ctx, login, hashedPassword)
	return err
}

func (u *Users) LoginUser(ctx context.Context, login, password string) error {
	usersData, err := u.repository.GetUserByLogin(ctx, login)
	if err != nil {
		return err
	}
	if usersData.Login == "" {
		return errs.ErrUserNotFound
	}
	matchPasswordsStatus := hasher.CheckPasswordHash(password, usersData.Password)
	fmt.Println("Login/password status match: ", matchPasswordsStatus)
	if !matchPasswordsStatus {
		return errs.ErrLoginMismatch
	}
	return err
}

func (u *Users) GetUserID(ctx context.Context, login string) (int, error) {
	usersData, err := u.repository.GetUserByLogin(ctx, login)
	if err != nil {
		return 0, err
	}
	return usersData.Id, nil
}

func (u *Users) SaveOrderNumber(ctx context.Context, userId int, orderNumber string) error {
	i, err := strconv.Atoi(orderNumber)
	if err != nil {
		return err
	}
	// checking is order num valid
	isOrderNumberValid := luhn.Valid(i)
	if !isOrderNumberValid {
		return errs.ErrInvalidOrderNumber
	}
	// checking is order already exist
	isOrderExists, err := u.repository.CheckIfOrderExists(ctx, userId, i)
	if err != nil {
		return err
	}
	if isOrderExists {
		return errs.ErrOrderAlreadyExists
	}
	isAnotherUser, err := u.repository.CheckIfOrderBelongsToUser(ctx, userId, orderNumber)
	if err != nil {
		return err
	}
	if !isAnotherUser {
		return errs.ErrOrderBelongsToAnotherUser
	}
	// saving order
	err = u.repository.SaveOrder(ctx, userId, i)
	return err
}

func (u *Users) GetUserOrders(ctx context.Context, userID int) ([]models.OrderData, error) {
	data, err := u.repository.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return []models.OrderData{}, err
	}
	sort(data)
	return data, nil
}

func swap(ar []models.OrderData, i, j int) {
	tmp := ar[i]
	ar[i] = ar[j]
	ar[j] = tmp
}

func sort(data []models.OrderData) {
	for i := 0; i < len(data); i++ {
		for j := i; j < len(data); j++ {
			if carbon.Parse(data[i].UploadedAt).Compare(">", carbon.Parse(data[j].UploadedAt)) {
				swap(data, i, j)
			}
		}
	}
}
