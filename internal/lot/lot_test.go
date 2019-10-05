package lot

import (
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
	"log"
	"github.com/DiGregory/tfs-auction/internal/errors"
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
	rep := LotsStorage{gdb}
	s := Suite{DB: gdb, mock: mock, rep: rep}
	return &s, nil
}

type Suite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock
	rep  LotsStorage
}
type CreateUserCase struct {
	Input     []byte
	InputID   int
	OutputErr error
	OutputLot *Lot
}

type GetLotsCase struct {
	InputID     int
	InputStatus string
	OutputErr   error
	OutputLots  []Lot
}

var GetLotsTestCases = []GetLotsCase{
	{InputID: 1, InputStatus: statusActive, OutputErr: errors.ErrBadReq, OutputLots: []Lot{},},
	{InputID: 1, InputStatus: "badstatus", OutputErr: errors.ErrBadReq, OutputLots: []Lot{},},
}

var CreateLotTestCases = []CreateUserCase{
	{Input: []byte(` {
	"title": "Apple iPhone 16",
	"description": "Новый, подарили, торгую за ненадобностью",
	"buy_price":6,
	"min_price": 2,
	"price_step": 3,
	"end_at": "2019-04-29T20:04:44.275Z",
	"status": "created"
}`), InputID: 1, OutputLot: nil, OutputErr: nil},

	{Input: []byte(` {

	"description": "Новый, подарили, торгую за ненадобностью",
	"min_price": 2,
	"price_step": 3,
	"end_at": "2019-04-29T20:04:44.275Z",
	"status": "created"
}`),
InputID: 1, OutputLot: nil, OutputErr: errors.ErrBadReq},

	{Input: []byte(` {
	"title": "Apple iPhone 16",
	"description": "Новый, подарили, торгую за ненадобностью",
	"min_price": "0",
	"price_step": 3,
	"end_at": "2019-04-29T20:04:44.275Z",
	"status": "created"
}`), InputID: 1, OutputLot: nil, OutputErr: errors.ErrBadReq},


	{Input: []byte(` {
	"title": "Apple iPhone 16",
	"description": "Новый, подарили, торгую за ненадобностью",
	"buy_price": 6,
	"min_price": 1,
	"price_step": 4,
	"end_at": "2019-04-29T20:04:44.275Z",
	"status": "created"
}`), InputID: 1, OutputLot: nil, OutputErr: errors.ErrBadReq},
}

func TestCreateLot(t *testing.T) {
	s, err := suitCreate()
	if err != nil {
		log.Fatal(err)
	}
	t.Run("CreateLotTest:", func(t *testing.T) {
		for i, v := range CreateLotTestCases {

			s.mock.ExpectQuery(
				`"INSERT INTO "users"`).WillReturnError(nil)

			if _, err := s.rep.CreateLot(v.Input, int64(v.InputID)); err != v.OutputErr {
				t.Errorf("Bad output in %v testcase; want [%v] - got [%v]", i, v.OutputErr, err)
			}
		}
	})
}
func TestGetUserLots(t *testing.T) {
	s, err := suitCreate()
	if err != nil {
		log.Fatal(err)
	}
	t.Run("GetUserLots:", func(t *testing.T) {
		for i, v := range GetLotsTestCases {

			s.mock.ExpectQuery(
				`"INSERT INTO "users"`).WillReturnError(nil)

			if _, err := s.rep.GetUserLots(v.InputID, v.InputStatus); err != v.OutputErr {
				t.Errorf("Bad output in %v testcase; want [%v] - got [%v]", i, v.OutputErr, err)
			}
		}
	})
}
