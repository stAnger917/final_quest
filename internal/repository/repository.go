package repository

import (
	"context"
	"database/sql"
	"final_quest/internal/models"
	"final_quest/pkg/logging"
	"fmt"
	"github.com/golang-module/carbon/v2"
	_ "github.com/lib/pq"
	"log"
)

type AppRepo struct {
	db     *sql.DB
	logger *logging.Logger
}

func NewDBClient(connectionString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("failed to connect to DB", err)
	}
	log.Println("DB connected")
	return db, nil
}

func InitAppDB(db *sql.DB, logger *logging.Logger) *AppRepo {
	return &AppRepo{
		db:     db,
		logger: logger,
	}
}

func (ar *AppRepo) InitTables() error {
	err := ar.CreateTableUsers()
	if err != nil {
		return err
	}
	err = ar.CreateTableUserOrders()
	if err != nil {
		return err
	}
	err = ar.CreateTableUserBalance()
	if err != nil {
		return err
	}
	err = ar.CreateTableWithdrawHistory()
	return err
}

func (ar *AppRepo) DropTables() error {
	err := ar.DropAllTables()
	return err
}

func (ar *AppRepo) CheckIfUserExists(ctx context.Context, userLogin string) (bool, error) {
	var id int
	var login string
	sqlString := fmt.Sprintf("SELECT id, login FROM users where login = '%s';", userLogin)
	err := ar.db.QueryRowContext(ctx, sqlString).Scan(&id, &login)
	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}

func (ar *AppRepo) CreateNewUser(ctx context.Context, userLogin, password string) error {
	sqlString := fmt.Sprintf("insert into users (login, password) values ('%s', '%s')", userLogin, password)
	_, err := ar.db.ExecContext(ctx, sqlString)
	return err
}

func (ar *AppRepo) GetUserByLogin(ctx context.Context, userLogin string) (models.UserData, error) {
	var result models.UserData
	sqlString := fmt.Sprintf("SELECT id, login, password FROM users WHERE login = '%s';", userLogin)
	rows, err := ar.db.QueryContext(ctx, sqlString)
	if err != nil {
		return models.UserData{}, err
	}
	defer rows.Close()
	for rows.Next() {
		item := models.UserData{}
		err = rows.Scan(&item.Id, &item.Login, &item.Password)
		if err != nil {
			return models.UserData{}, err
		}
		result = models.UserData{
			Id:       item.Id,
			Login:    item.Login,
			Password: item.Password,
		}
	}
	return result, nil
}

func (ar *AppRepo) CheckIfOrderExists(ctx context.Context, userId int, orderNumber int) (bool, error) {
	var id int
	var number int
	sqlString := fmt.Sprintf("SELECT user_id, orders_number FROM user_orders where user_id = '%v' AND orders_number = %v;", userId, orderNumber)
	err := ar.db.QueryRowContext(ctx, sqlString).Scan(&id, &number)
	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}

func (ar *AppRepo) CheckIfOrderBelongsToUser(ctx context.Context, userID int, orderNumber string) (bool, error) {
	var data models.SingleOrderData
	sqlString := fmt.Sprintf("SELECT user_id, orders_number FROM user_orders where orders_number = %v;", orderNumber)
	rows, err := ar.db.QueryContext(ctx, sqlString)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		item := models.SingleOrderData{}
		err = rows.Scan(&item.UserID, &item.Number)
		if err != nil {
			return false, err
		}
		data = models.SingleOrderData{
			Number: item.Number,
			UserID: item.UserID,
		}
	}
	if data.Number != "" {
		if data.UserID != userID && data.Number == orderNumber {
			return false, nil
		}
	}
	return true, nil
}

func (ar *AppRepo) SaveOrder(ctx context.Context, userId, ordersNumber int) error {
	defaultStatus := "NEW"
	uploadedAt := carbon.Now().ToRfc3339String()
	sqlString := fmt.Sprintf("insert into user_orders (user_id, orders_number, orders_status, uploaded_at) values ('%v', '%v', '%s', '%s')", userId, ordersNumber, defaultStatus, uploadedAt)
	_, err := ar.db.ExecContext(ctx, sqlString)
	return err
}

func (ar *AppRepo) GetOrdersByUserID(ctx context.Context, userID int) ([]models.OrderData, error) {
	var data models.OrderData
	var result []models.OrderData
	sqlString := fmt.Sprintf("SELECT orders_number, orders_status, uploaded_at, accrual FROM users WHERE user_id = '%v';", userID)
	rows, err := ar.db.QueryContext(ctx, sqlString)
	if err != nil {
		return []models.OrderData{}, err
	}
	defer rows.Close()
	for rows.Next() {
		item := models.OrderData{}
		err = rows.Scan(&item.Number, &item.Status, &item.Accrual, &item.UploadedAt)
		if err != nil {
			return []models.OrderData{}, err
		}
		data = models.OrderData{
			Number:     item.Number,
			Status:     item.Status,
			Accrual:    item.Accrual,
			UploadedAt: item.UploadedAt,
		}
		result = append(result, data)
	}
	return result, nil
}
