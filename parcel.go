package main

import (
	"database/sql"
	"errors"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	// number — автоинкрементный, поэтому в INSERT его не указываем
	res, err := s.db.Exec(
		"INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)",
		p.Client, p.Status, p.Address, p.CreatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	// верните идентификатор последней добавленной записи
	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка
	row := s.db.QueryRow(
		"SELECT number, client, status, address, created_at FROM parcel WHERE number = ?",
		number,
	)

	// заполните объект Parcel данными из таблицы
	p := Parcel{}
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return Parcel{}, err
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк
	rows, err := s.db.Query(
		"SELECT number, client, status, address, created_at FROM parcel WHERE client = ? ORDER BY number",
		client,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// заполните срез Parcel данными из таблицы
	var res []Parcel

	for rows.Next() {
		p := Parcel{}
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}

		res = append(res, p)
	}

	// rows.Err() проверяем после цикла - так мы поймаем ошибки, возникшие во время итерации
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	res, err := s.db.Exec("UPDATE parcel SET status = ? WHERE number = ?", status, number)
	if err != nil {
		return err
	}

	// RowsAffected() показывает, сколько строк реально обновилось.
	// Если 0 - значит посылки с таким number нет (или нечего было обновлять).
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		// для интеграционных тестов удобно возвращать sql.ErrNoRows, если запись не найдена
		return sql.ErrNoRows
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	res, err := s.db.Exec(
		"UPDATE parcel SET address = ? WHERE status = ? AND number = ?",
		address, ParcelStatusRegistered, number,
	)
	if err != nil {
		return err
	}

	// RowsAffected() покажет, было ли обновление.
	// Если n == 0, значит:
	// 1) либо посылки с таким number нет,
	// 2) либо посылка есть, но она не в статусе registered (тогда менять адрес нельзя).
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		// Проверяем, существует ли вообще посылка.
		// getErr — это ошибка из Get(number): если записи нет, вернётся sql.ErrNoRows.
		_, getErr := s.Get(number)
		if getErr != nil {
			return getErr
		}
		// Если запись существует, но обновление не прошло — значит статус не registered.
		return errors.New("cannot change address: parcel is not registered")
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	res, err := s.db.Exec(
		"DELETE FROM parcel WHERE status = ? AND number = ?",
		ParcelStatusRegistered, number,
	)
	if err != nil {
		return err
	}

	// RowsAffected() покажет, было ли удаление.
	// Если n == 0, значит:
	// 1) либо посылки с таким number нет,
	// 2) либо посылка есть, но она не в статусе registered (тогда удалять нельзя).
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		// Проверяем, существует ли вообще посылка.
		// getErr — это ошибка из Get(number): если записи нет, вернётся sql.ErrNoRows.
		_, getErr := s.Get(number)
		if getErr != nil {
			return getErr
		}
		// Если запись существует, но удаление не прошло — значит статус не registered.
		return errors.New("cannot delete: parcel is not registered")
	}

	return nil
}
