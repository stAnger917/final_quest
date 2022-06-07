package users

import (
	"context"
	"final_quest/internal/errs"
	"final_quest/internal/models"
	"final_quest/internal/repository"
	"final_quest/internal/usecase/loyality"
	"final_quest/pkg/hasher"
	"final_quest/pkg/logging"
	"fmt"
	"github.com/golang-module/carbon/v2"
	"github.com/theplant/luhn"
	"strconv"
	"strings"
)

type Users struct {
	repository        *repository.AppRepo
	logger            *logging.Logger
	accountingService *loyality.AccountingService
}

func NewUsersUseCase(repo *repository.AppRepo, logger *logging.Logger, accountingService *loyality.AccountingService) *Users {
	return &Users{
		repository:        repo,
		logger:            logger,
		accountingService: accountingService,
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
	return usersData.ID, nil
}

func (u *Users) SaveOrderNumber(ctx context.Context, userID int, orderNumber string) error {
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
	err = u.repository.CheckOrder(ctx, userID, orderNumber)
	if err != nil {
		return err
	}
	// saving order
	err = u.repository.SaveOrder(ctx, userID, orderNumber)
	if err != nil {
		return err
	}
	err = u.accountingService.GetPointsInfoByOrder(ctx, orderNumber)
	return err
}

func (u *Users) GetUserOrders(ctx context.Context, userID int) ([]models.OrderData, error) {
	data, err := u.repository.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return []models.OrderData{}, err
	}
	for _, v := range data {
		err = u.accountingService.GetPointsInfoByOrder(ctx, v.Number)
		if err != nil {
			return []models.OrderData{}, err
		}
	}
	updatedData, err := u.repository.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return []models.OrderData{}, err
	}
	sortOrdersByDateInc(updatedData)
	return updatedData, nil
}

func (u *Users) GetUserBalance(ctx context.Context, userID int) (models.UserBalanceInfo, error) {
	data, err := u.repository.GetUserBalanceByID(ctx, userID)
	if err != nil {
		return models.UserBalanceInfo{}, err
	}
	u.logger.EasyLogInfo("use case", "got balance data for user: ", fmt.Sprintf("user: %v, balance: %v", userID, data.Current))
	return data, nil
}

func (u *Users) MakeWithdraw(ctx context.Context, userID int, orderNumber string, sum float32) error {
	// checking order`s number - it must belong to current user
	err := u.repository.CheckOrderForWithdraw(ctx, userID, orderNumber)
	if err != nil {
		return err
	}
	err = u.repository.MakeWithdraw(ctx, userID, sum, orderNumber)
	return err
}

func (u *Users) GetWithdrawals(ctx context.Context, userID int) (models.Withdrawals, error) {
	res, err := u.repository.GetUserWithdrawals(ctx, userID)
	if err != nil {
		return models.Withdrawals{}, err
	}
	sortOrdersByDateDesc(res.Data)
	return res, nil
}

func (u *Users) GetUserOrder(ctx context.Context, orderNum string) (models.SingleOrder, error) {
	data, err := u.repository.GetOrder(ctx, orderNum)
	if err != nil {
		return models.SingleOrder{}, err
	}
	return data, nil
}

func swapOrders(ar []models.OrderData, i, j int) {
	tmp := ar[i]
	ar[i] = ar[j]
	ar[j] = tmp
}

func sortOrdersByDateInc(data []models.OrderData) {
	for i := 0; i < len(data); i++ {
		for j := i; j < len(data); j++ {
			if carbon.Parse(data[i].UploadedAt).Compare(">", carbon.Parse(data[j].UploadedAt)) {
				swapOrders(data, i, j)
			}
		}
	}
}

func sortOrdersByDateDesc(data []models.WithdrawInfo) {
	for i := 0; i < len(data); i++ {
		for j := i; j < len(data); j++ {
			if carbon.Parse(data[i].ProcessedAt).Compare("<", carbon.Parse(data[j].ProcessedAt)) {
				swapWithdrawals(data, i, j)
			}
		}
	}
}

func swapWithdrawals(ar []models.WithdrawInfo, i, j int) {
	tmp := ar[i]
	ar[i] = ar[j]
	ar[j] = tmp
}
