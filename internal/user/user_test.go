package user

import (
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
	"github.com/DiGregory/tfs-auction/internal/errors"
	"log"
)

func suitCreate() (*Suite, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	gdb, err := gorm.Open("postgres", db)
	if err != nil {
		return nil, err
	}
	rep := UsersStorage{gdb}
	s := Suite{DB: gdb, mock: mock, rep: rep}
	return &s, nil
}

type Suite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock
	rep  UsersStorage

}
type RegUserCase struct {
	Input  []byte
	Output error
}
type GetUserCase struct {
	InputID    int
	OutputUser *User
	OutputErr  error
}
type UpdateUserCase struct {
	Input      []byte
	UID        int64
	OutputUser *User
	OutputErr  error
}

var UpdateUserTestCases = []UpdateUserCase{
	{Input: []byte(` {
	
  "email": "Павел",
  "last_name": "Дуров",
  "birthday": "1985-10-10"
	 
}`), UID: 1, OutputUser: nil, OutputErr: errors.ErrBadReq},
	{Input: []byte(` {
	
  "password": "Павел",
  "last_name": "Дуров",
  "birthday": "1985-10-10"
	 
}`), UID: 1, OutputUser: nil, OutputErr: errors.ErrBadReq},

}

var GetUserTestCases = []GetUserCase{
	{InputID: 1, OutputUser: nil, OutputErr: errors.ErrBadReq},
}
var RegUserTestCases = []RegUserCase{
	{
		Input: []byte(`{
			"first_name": "Павел",
				"last_name": "Дуров",
				"birthday":"1905-10-10",
				"email": "durov@telegram7.org",
				"password": "qwerty"
		}`),
		Output: nil,
	},
	{
		Input: []byte(`{
			"first_name": "Павел",
				"last_name": "Дуров",
				"birthday":"1905-10-10",
				 
				"password": "qwerty"
		}`),
		Output: errors.ErrBadReq,
	},
	{
		Input: []byte(`{
			"first_name": "",
				"last_name": "Дуров",
				"birthday":"1905-10-10",
				"email": "durov@telegram7.org",
				"password": "qwerty"
		}`),
		Output: errors.ErrBadReq,
	},
	{
		Input: []byte(`{
			"first_name": "Дуров",
				"last_name": "",
				"birthday":"1905-10-10",
				"email": "durov@telegram7.org",
				"password": "qwerty"
		}`),
		Output: errors.ErrBadReq,
	}, {
		Input: []byte(`{
			"first_name": "Дуров",
				"last_name": "Дуров",
				"birthday":"",
				"email": "durov@telegram7.org",
				"password": "qwerty"
		}`),
		Output: nil,
	}, {
		Input: []byte(`{
			"first_name": "Дуров",
				"last_name": "Дуров",
				"birthday":"1995-10-10",
				"email": "",
				"password": "qwerty"
		}`),
		Output: errors.ErrBadReq,
	},
	{
		Input: []byte(`{
			"first_name": "Дуров",
				"last_name": "Дуров",
				"birthday":"1995-10-10",
				"email": " durov@telegram7.org",
				"password": ""
		}`),
		Output: errors.ErrBadReq,
	},
}

func TestRegUser(t *testing.T) {
	s, err := suitCreate()
	if err != nil {
		log.Fatal(err)
	}
	t.Run("RegUserTests:", func(t *testing.T) {
		for i, v := range RegUserTestCases {

			s.mock.ExpectQuery(
				`"INSERT INTO "users"`).WillReturnError(nil)

			if err = s.rep.RegUser(v.Input); err != v.Output {
				t.Errorf("Bad output in %v testcase; want [%v] - got [%v]", i, v.Output, err)
			}
		}
	})
}
func TestGetUser(t *testing.T) {
	s, err := suitCreate()
	if err != nil {
		log.Fatal(err)
	}
	t.Run("RegUserTests:", func(t *testing.T) {
		for i, v := range GetUserTestCases {

			s.mock.ExpectQuery(
				`"SELECT * "users" WHERE`).WillReturnError(nil)

			if _, err := s.rep.GetUser(v.InputID); err != v.OutputErr {
				t.Errorf("Bad output in %v testcase; want [%v] - got [%v]", i, v.OutputErr, err)
			}
		}
	})
}
func TestUpdateUser(t *testing.T) {
	s, err := suitCreate()
	if err != nil {
		log.Fatal(err)
	}
	t.Run("UpdateUserTests:", func(t *testing.T) {
		for i, v := range UpdateUserTestCases {

			s.mock.ExpectQuery(
				`"UPDATE "users"`).WillReturnError(nil)

			if _,err = s.rep.UpdateUser(v.Input,v.UID); err != v.OutputErr {
				t.Errorf("Bad output in %v testcase; want [%v] - got [%v]", i, v.OutputErr, err)
			}
		}
	})
}
