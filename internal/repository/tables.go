package repository

func (ar *AppRepo) CreateTableUsers() error {
	_, err := ar.db.Exec("CREATE TABLE IF NOT EXISTS users(id SERIAL PRIMARY KEY, login VARCHAR(350) UNIQUE NOT NULL, password VARCHAR(350) NOT NULL)")
	if err != nil {
		ar.logger.EasyLogFatal("repository", "failed to create users table", "", err)
		return err
	}
	return nil
}

func (ar *AppRepo) CreateTableUserOrders() error {
	_, err := ar.db.Exec("CREATE TABLE IF NOT EXISTS user_orders(id SERIAL PRIMARY KEY, user_id SERIAL NOT NULL, orders_number VARCHAR(350) UNIQUE NOT NULL,  orders_status VARCHAR (350) NOT NULL, uploaded_at VARCHAR(350) UNIQUE NOT NULL, accrual INTEGER, CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users (id))")
	if err != nil {
		ar.logger.EasyLogFatal("repository", "failed to create user_orders table", "", err)
		return err
	}
	return nil
}

func (ar *AppRepo) CreateTableUserBalance() error {
	_, err := ar.db.Exec("CREATE TABLE IF NOT EXISTS user_balance(id SERIAL PRIMARY KEY, user_id SERIAL NOT NULL, current_balance INTEGER, withdraw INTEGER, CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users (id))")
	if err != nil {
		ar.logger.EasyLogFatal("repository", "failed to create user_balance table", "", err)
		return err
	}
	return nil
}

func (ar *AppRepo) CreateTableWithdrawHistory() error {
	_, err := ar.db.Exec("CREATE TABLE IF NOT EXISTS withdraw_history(id SERIAL PRIMARY KEY, user_id SERIAL NOT NULL, orders_number VARCHAR(350) UNIQUE NOT NULL, sum INTEGER, procesed_at VARCHAR(350) UNIQUE NOT NULL, CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users (id), CONSTRAINT fk_orders_number FOREIGN KEY (orders_number) REFERENCES user_orders (orders_number))")
	if err != nil {
		ar.logger.EasyLogFatal("repository", "failed to create withdraw history_table", "", err)
		return err
	}
	return nil
}

func (ar *AppRepo) DropAllTables() error {
	_, err := ar.db.Exec("DROP TABLE users, user_orders, user_balance, withdraw_history CASCADE;")
	if err != nil {
		ar.logger.EasyLogFatal("repository", "failed to drop users table", "", err)
		return err
	}
	return nil
}

func (ar *AppRepo) DropTableUserOrders() error {
	_, err := ar.db.Exec("DROP TABLE user_orders CASCADE;")
	if err != nil {
		ar.logger.EasyLogFatal("repository", "failed to drop user_orders table", "", err)
		return err
	}
	return nil
}

func (ar *AppRepo) DropTableUserBalance() error {
	_, err := ar.db.Exec("DROP TABLE user_balance;")
	if err != nil {
		ar.logger.EasyLogFatal("repository", "failed to drop user_balance table", "", err)
		return err
	}
	return nil
}

func (ar *AppRepo) DropTableWithdrawHistory() error {
	_, err := ar.db.Exec("DROP TABLE withdraw_history;")
	if err != nil {
		ar.logger.EasyLogFatal("repository", "failed to create withdraw history_table", "", err)
		return err
	}
	return nil
}
