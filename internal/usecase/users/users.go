package users

import (
	"context"
	"final_quest/internal/errs"
	"final_quest/internal/models"
	"final_quest/internal/repository"
	"final_quest/pkg/hasher"
	"final_quest/pkg/logging"
	"github.com/golang-module/carbon/v2"
	"github.com/theplant/luhn"
	"strconv"
	"strings"
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
		return errs.ErrUserAlreadyExists
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
	number := strings.Replace(orderNumber, "\n", "", 1)
	i, err := strconv.Atoi(number)
	if err != nil {
		return err
	}
	// checking is order num valid
	isOrderNumberValid := luhn.Valid(i)
	if !isOrderNumberValid {
		return errs.ErrInvalidOrderNumber
	}
	// checking is order already exist
	err = u.repository.CheckOrder(ctx, userId, orderNumber)
	if err != nil {
		return err
	}
	// saving order
	err = u.repository.SaveOrder(ctx, userId, orderNumber)
	return err
}

func (u *Users) GetUserOrders(ctx context.Context, userID int) ([]models.OrderData, error) {
	data, err := u.repository.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return []models.OrderData{}, err
	}
	sortOrdersByDate(data)
	return data, nil
}

func (u *Users) GetUserBalance(ctx context.Context, userID int) (models.UserBalanceInfo, error) {
	data, err := u.repository.GetUserBalanceByID(ctx, userID)
	if err != nil {
		return models.UserBalanceInfo{}, err
	}
	return data, nil
}

func swapOrders(ar []models.OrderData, i, j int) {
	tmp := ar[i]
	ar[i] = ar[j]
	ar[j] = tmp
}

func sortOrdersByDate(data []models.OrderData) {
	for i := 0; i < len(data); i++ {
		for j := i; j < len(data); j++ {
			if carbon.Parse(data[i].UploadedAt).Compare(">", carbon.Parse(data[j].UploadedAt)) {
				swapOrders(data, i, j)
			}
		}
	}
}
