package repository

import (
	"context"
	"database/sql"
	"final_quest/internal/errs"
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
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			ar.logger.EasyLogCloseRowsErr(err)
		}
	}(rows)
	if err != nil {
		return models.UserData{}, err
	}
	for rows.Next() {
		item := models.UserData{}
		err = rows.Scan(&item.ID, &item.Login, &item.Password)
		if err != nil {
			return models.UserData{}, err
		}
		result = models.UserData{
			ID:       item.ID,
			Login:    item.Login,
			Password: item.Password,
		}
	}
	return result, nil
}

func (ar *AppRepo) CheckOrder(ctx context.Context, userID int, orderNumber string) error {
	var data models.OrderInfo
	sqlString := fmt.Sprintf("SELECT user_id, orders_number FROM user_orders where orders_number = '%s';", orderNumber)
	rows, err := ar.db.QueryContext(ctx, sqlString)
	if err != nil {
		return err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			ar.logger.EasyLogCloseRowsErr(err)
		}
	}(rows)
	for rows.Next() {
		item := models.OrderInfo{}
		err = rows.Scan(&item.UserID, &item.Number)
		if err != nil {
			return err
		}
		data = models.OrderInfo{
			Number: item.Number,
			UserID: item.UserID,
		}
	}
	if data.UserID == 0 {
		return nil
	}
	switch {
	case data.Number == orderNumber && data.UserID == userID:
		return errs.ErrOrderAlreadyExists
	case data.Number == orderNumber && data.UserID != userID:
		return errs.ErrOrderBelongsToAnotherUser
	}
	return nil
}

func (ar *AppRepo) CheckIfOrderBelongsToUser(ctx context.Context, userID int, orderNumber string) (bool, error) {
	var data models.SingleOrderData
	sqlString := fmt.Sprintf("SELECT user_id, orders_number FROM user_orders where orders_number = %v;", orderNumber)
	rows, err := ar.db.QueryContext(ctx, sqlString)
	if err != nil {
		return false, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			ar.logger.EasyLogCloseRowsErr(err)
		}
	}(rows)
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

func (ar *AppRepo) SaveOrder(ctx context.Context, userID int, ordersNumber string) error {
	defaultStatus := "REGISTERED"
	uploadedAt := carbon.Now().ToRfc3339String()
	sqlString := fmt.Sprintf("insert into user_orders (user_id, orders_number, orders_status, uploaded_at) values ('%v', '%s', '%s', '%s')", userID, ordersNumber, defaultStatus, uploadedAt)
	_, err := ar.db.ExecContext(ctx, sqlString)
	return err
}

func (ar *AppRepo) GetOrdersByUserID(ctx context.Context, userID int) ([]models.OrderData, error) {
	var data models.OrderData
	var result []models.OrderData
	sqlString := fmt.Sprintf("SELECT orders_number, orders_status, uploaded_at, accrual FROM user_orders WHERE user_id = '%v';", userID)
	rows, err := ar.db.QueryContext(ctx, sqlString)
	if err != nil {
		return []models.OrderData{}, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			ar.logger.EasyLogCloseRowsErr(err)
		}
	}(rows)
	for rows.Next() {
		item := models.OrderData{}
		err = rows.Scan(&item.Number, &item.Status, &item.UploadedAt, &item.Accrual)
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

func (ar *AppRepo) GetUserBalanceByID(ctx context.Context, userID int) (models.UserBalanceInfo, error) {
	var data models.UserBalanceInfo
	fmt.Println("Got repository call for balance")
	sqlString := fmt.Sprintf("SELECT user_id, current_balance, withdraw FROM user_balance WHERE user_id = %v;", userID)
	rows, err := ar.db.QueryContext(ctx, sqlString)
	if err != nil {
		return models.UserBalanceInfo{}, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			ar.logger.EasyLogCloseRowsErr(err)
		}
	}(rows)
	for rows.Next() {
		item := models.UserBalance{}
		err = rows.Scan(&item.UserID, &item.Withdraw, &item.Current)
		if err != nil {
			return models.UserBalanceInfo{}, err
		}
		fmt.Println("BALANCE DATA: ", item.UserID, item.Withdraw, item.Current)
		data = models.UserBalanceInfo{
			Current:  item.Current,
			Withdraw: item.Withdraw,
		}

	}
	fmt.Println("DATA: ", data)
	return data, nil
}

func (ar *AppRepo) MakeWithdraw(ctx context.Context, userID int, withdrawSum float32, orderNum string) error {
	tx, err := ar.db.Begin()
	if err != nil {
		return err
	}
	// checking user`s balance
	balanceInfo, err := ar.GetUserBalanceByID(ctx, userID)
	if err != nil {
		return err
	}
	// if ok - make withdraw
	if balanceInfo.Current < withdrawSum {
		return errs.ErrNotEnoughFounds
	}
	newBalance := balanceInfo.Current - withdrawSum
	newWithdrawBalance := balanceInfo.Withdraw + withdrawSum
	// setting new values in user_balance table
	sqlString := fmt.Sprintf("UPDATE user_balance "+
		"SET current_balance = %v, "+
		"withdraw = %v WHERE user_id = %v;", newBalance, newWithdrawBalance, userID)
	_, err = ar.db.ExecContext(ctx, sqlString)
	if err != nil {
		tx.Rollback()
		return err
	}
	processedAt := carbon.Now().ToRfc3339String()
	sqlStringForWithdrawHistory := fmt.Sprintf("INSERT INTO withdraw_history "+
		"(user_id, orders_number, sum, processed_at)"+
		"VALUES (%v, '%s', %v, '%s')", userID, orderNum, withdrawSum, processedAt)
	_, err = ar.db.ExecContext(ctx, sqlStringForWithdrawHistory)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}

func (ar *AppRepo) CheckOrderForWithdraw(ctx context.Context, userID int, orderNumber string) error {
	var data models.OrderInfo
	sqlString := fmt.Sprintf("SELECT user_id, orders_number FROM user_orders where orders_number = '%s';", orderNumber)
	rows, err := ar.db.QueryContext(ctx, sqlString)
	if err != nil {
		return err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			ar.logger.EasyLogCloseRowsErr(err)
		}
	}(rows)
	for rows.Next() {
		item := models.OrderInfo{}
		err = rows.Scan(&item.UserID, &item.Number)
		if err != nil {
			return err
		}
		data = models.OrderInfo{
			Number: item.Number,
			UserID: item.UserID,
		}
	}
	if data.UserID == 0 {
		return errs.ErrOrderNotFound
	}

	if data.Number == orderNumber && data.UserID != userID {
		return errs.ErrOrderBelongsToAnotherUser
	}
	return nil
}

func (ar *AppRepo) AddAccrualPoints(ctx context.Context, userID int, sum float32) error {
	//tx, err := ar.db.Begin()
	//if err != nil {
	//	return err
	//}
	// checking user`s balance
	checkBalance, err := ar.CheckIfBalanceExist(ctx, userID)
	if !checkBalance {
		fmt.Println("Creating new record in balance")
		sqlString := fmt.Sprintf("INSERT INTO user_balance (current_balance, user_id, withdraw) VALUES (%v, %v, 0);", sum, userID)
		_, err = ar.db.ExecContext(ctx, sqlString)
		if err != nil {
			fmt.Println("ERROR", err)
		}
		return err
	}
	balanceInfo, err := ar.GetUserBalanceByID(ctx, userID)
	if err != nil {
		return err
	}
	// if ok - make withdraw
	newBalance := balanceInfo.Current + sum
	fmt.Println("NEW USER BALANCE: ", newBalance)
	// setting new values in user_balance table
	sqlString := fmt.Sprintf("UPDATE user_balance SET current_balance = %v WHERE user_id = %v;", newBalance, userID)
	_, err = ar.db.ExecContext(ctx, sqlString)
	if err != nil {
		fmt.Println("ERROR", err)
		return err
	}
	//err = tx.Commit()
	return err
}

func (ar *AppRepo) ChangeOrderStatusByOrderNum(ctx context.Context, orderNum, status string) error {
	sqlString := fmt.Sprintf("UPDATE user_orders SET orders_status = '%s' WHERE orders_number = '%s';", status, orderNum)
	_, err := ar.db.QueryContext(ctx, sqlString)
	return err
}

func (ar *AppRepo) GetUserIDByOrderNum(ctx context.Context, orderNum string) (models.OrderOwner, error) {
	var userID models.OrderOwner
	sqlString := fmt.Sprintf("SELECT user_id FROM user_orders WHERE orders_number = '%s';", orderNum)
	rows, err := ar.db.QueryContext(ctx, sqlString)
	if err != nil {
		return models.OrderOwner{}, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			ar.logger.EasyLogCloseRowsErr(err)
		}
	}(rows)
	for rows.Next() {
		item := models.OrderOwner{}
		err = rows.Scan(&item.UserID)
		if err != nil {
			return models.OrderOwner{}, err
		}
		userID = models.OrderOwner{UserID: item.UserID}
	}
	return userID, nil
}

func (ar *AppRepo) GetAllOpenedOrders(ctx context.Context) ([]string, error) {
	var result []string
	sqlString := "SELECT orders_number FROM user_orders WHERE orders_status = 'REGISTERED' OR  orders_status = 'PROCESSING';"
	rows, err := ar.db.QueryContext(ctx, sqlString)
	if err != nil {
		return []string{}, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			ar.logger.EasyLogCloseRowsErr(err)
		}
	}(rows)
	for rows.Next() {
		item := models.OrderNumber{}
		err = rows.Scan(&item.Number)
		if err != nil {
			return []string{}, err
		}
		result = append(result, item.Number)
	}
	return result, nil
}

func (ar *AppRepo) GetUserWithdrawals(ctx context.Context, userID int) (models.Withdrawals, error) {
	var data models.Withdrawals
	sqlString := fmt.Sprintf("SELECT orders_number, sum, processed_at FROM withdraw_history WHERE user_id = %v;", userID)
	rows, err := ar.db.QueryContext(ctx, sqlString)
	if err != nil {
		return models.Withdrawals{}, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			ar.logger.EasyLogCloseRowsErr(err)
		}
	}(rows)
	for rows.Next() {
		item := models.WithdrawInfo{}
		err = rows.Scan(&item.Order, &item.Sum, &item.ProcessedAt)
		if err != nil {
			return models.Withdrawals{}, err
		}
		data.Data = append(data.Data, item)
	}
	return data, nil
}

func (ar *AppRepo) GetOrder(ctx context.Context, orderNum string) (models.SingleOrder, error) {
	var data models.SingleOrder
	sqlString := fmt.Sprintf("SELECT orders_number, orders_status, accrual FROM user_orders WHERE orders_number = '%s';", orderNum)
	rows, err := ar.db.QueryContext(ctx, sqlString)
	if err != nil {
		return models.SingleOrder{}, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			ar.logger.EasyLogCloseRowsErr(err)
		}
	}(rows)
	for rows.Next() {
		item := models.SingleOrder{}
		err = rows.Scan(&item.Order, &item.Status, &item.Accrual)
		if err != nil {
			return models.SingleOrder{}, err
		}
		data = models.SingleOrder{
			Order:   item.Order,
			Status:  item.Status,
			Accrual: item.Accrual,
		}
	}
	return data, nil
}

func (ar *AppRepo) CheckIfBalanceExist(ctx context.Context, userID int) (bool, error) {
	var id int
	var login string
	sqlString := fmt.Sprintf("SELECT id, FROM user_balance  WHERE user_id = %v;", userID)
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
