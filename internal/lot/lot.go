package lot

import (
	"time"
	"github.com/DiGregory/tfs-auction/internal/user"
	"encoding/json"
	"github.com/DiGregory/tfs-auction/internal/errors"
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"math"
	"fmt"
	"sync"
	"github.com/gorilla/websocket"
)

type Lot struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	BuyPrice    float64    `json:"buy_price"`
	MinPrice    float64    `json:"min_price"`
	PriceStep   float64    `json:"price_step"`
	Status      string     `json:"status"`
	EndAt       *time.Time `json:"end_at"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	DeletedAt   *time.Time `json:"-"`
	Creator     *user.User `json:"creator"`
	Buyer       *user.User `json:"buyer,omitempty"`
	CreatorID   uint       `json:"-"`
	BuyerID     uint       `json:"-"`
}

const (
	statusActive  = "active"
	statusCreated = "created"
)

var SingleLotCh = make(chan Lot)
var AllLotsCh = make(chan Lot)
var WsClients1 WSClients //для конкретного лота
var WsClients2 WSClients //для всех лотов

type WSClients struct {
	wsConn []*websocket.Conn
	sync.Mutex
}

type LotsStorage struct {
	DB *gorm.DB
}

func StorageConnect(source string) (*LotsStorage, error) {
	//	args := "user=postgres password=1234 dbname=test sslmode=disable"
	db, err := gorm.Open("postgres", source)
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Lot{}, &user.User{}, )
	return &LotsStorage{DB: db,}, nil

}

func (clients *WSClients) BroadcastMessage(message []byte) {
	for i, c := range clients.wsConn {
		err := c.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			fmt.Printf("can't broadcast message: %+v\n", err)
			clients.removeClientByID(i)
		}

	}
}

func (clients *WSClients) AddClient(conn *websocket.Conn) {
	clients.Mutex.Lock()
	clients.wsConn = append(clients.wsConn, conn)
	clients.Mutex.Unlock()
	fmt.Printf("added client, total clients: %d\n", len(clients.wsConn))
}

func (clients *WSClients) removeClientByID(id int) {
	clients.Mutex.Lock()
	clients.wsConn = append(clients.wsConn[:id], clients.wsConn[id+1:]...)
	clients.Mutex.Unlock()
	fmt.Printf("removed client #%d, total clients %d\n", id, len(clients.wsConn))
}

func (s LotsStorage) getLastID() (id int) {
	db := s.DB
	id = -1
	var l Lot
	db.Last(&l)
	if id == -1 {
		return
	}
	return int(l.ID)

}

func (s LotsStorage) CreateLot(jsonData []byte, uid int64) (*Lot, error) {
	db := s.DB
	db.AutoMigrate(&Lot{})

	var l interface{}
	err := json.Unmarshal(jsonData, &l)
	if err != nil {
		return nil, err
	}
	jsonMap := l.(map[string]interface{})

	Lot, err := scanCreatingLot(jsonMap, s)
	if err != nil {
		return nil, errors.ErrBadReq
	}
	u := user.User{}
	db.Where("ID = ?", uid).First(&u)

	Lot.CreatorID = uint(u.ID)

	Lot.BuyerID = 0

	db.Save(&Lot)

	Lot.Creator = &user.User{ID: u.ID, FirstName: u.FirstName, LastName: u.LastName, CreatedAt: nil}
	Lot.Buyer = nil
	if len(WsClients2.wsConn) != 0 {
		AllLotsCh <- *Lot
	}

	return Lot, nil

}
func scanCreatingLot(jsonMap map[string]interface{}, s LotsStorage) (*Lot, error) {

	Lot := &Lot{}
	Lot.ID = uint(s.getLastID() + 1)

	var ok bool
	if Lot.Title, ok = jsonMap["title"].(string); !ok || Lot.Title == "" {
		return nil, errors.ErrBadReq
	}
	if _, ok = jsonMap["description"].(string); ok {
		Lot.Description = jsonMap["description"].(string)
	}
	if Lot.MinPrice, ok = jsonMap["min_price"].(float64); !ok {
		fmt.Println(jsonMap["min_price"])
		return nil, errors.ErrBadReq
	}
	if Lot.BuyPrice, ok = jsonMap["buy_price"].(float64); !ok {

		return nil, errors.ErrBadReq
	}
	if Lot.MinPrice < 1 {
		return nil, errors.ErrBadReq
	}

	if _, ok = jsonMap["price_step"].(float64); ok {
		Lot.PriceStep = jsonMap["price_step"].(float64)
	}
	fmt.Println(Lot.PriceStep)
	fmt.Println(Lot.BuyPrice)
	fmt.Println(Lot.BuyPrice/Lot.PriceStep == math.Trunc(Lot.BuyPrice/Lot.PriceStep))
	//if !(Lot.BuyPrice/Lot.PriceStep == math.Trunc(Lot.BuyPrice/Lot.PriceStep)) {
	//	return nil, errors.ErrConflict
	//}

	if Lot.PriceStep < 1 {
		Lot.PriceStep = 1
	}

	if _, ok = jsonMap["end_at"].(string); ok {
		t, err := time.Parse(time.RFC3339, jsonMap["end_at"].(string))
		if err != nil {
			return nil, errors.ErrBadReq
		}
		Lot.EndAt = &t

	}
	if Lot.Status, ok = jsonMap["status"].(string); !ok {
		Lot.Status = statusCreated
	} else if Lot.Status != statusCreated && Lot.Status != statusActive {
		return nil, errors.ErrBadReq
	}

	return Lot, nil
}
func scanUpdatingLot(jsonMap map[string]interface{}, lotID int) (*Lot, error) {
	Lot := &Lot{}

	var ok bool
	if _, ok = jsonMap["id"]; ok {
		return nil, errors.ErrBadReq
	}
	Lot.ID = uint(lotID)
	if _, ok = jsonMap["title"]; ok {
		Lot.Title = jsonMap["title"].(string)
	}
	if _, ok = jsonMap["description"]; ok {
		Lot.Description = jsonMap["description"].(string)
	}
	if _, ok = jsonMap["min_price"].(float64); ok {
		Lot.MinPrice = jsonMap["min_price"].(float64)
	}
	if _, ok = jsonMap["price_step"].(float64); ok {
		Lot.PriceStep = jsonMap["price_step"].(float64)
	}
	if _, ok = jsonMap["end_at"].(string); ok {

		t, err := time.Parse(time.RFC3339, jsonMap["end_at"].(string))
		if err != nil {

			return nil, errors.ErrBadReq
		}
		Lot.EndAt = &t
	}
	if _, ok = jsonMap["status"]; ok {
		Lot.Status = jsonMap["status"].(string)
	}

	return Lot, nil
}

func (s LotsStorage) UpdateLot(jsonData []byte, uid int64, lotID int) (*Lot, error) {
	db := s.DB

	db.AutoMigrate(&Lot{})

	var l interface{}
	err := json.Unmarshal(jsonData, &l)
	if err != nil {
		return nil, errors.ErrBadReq
	}
	jsonMap := l.(map[string]interface{})

	Lot, err := scanUpdatingLot(jsonMap, lotID)
	if err != nil {
		return nil, err
	}
	Lot.CreatorID = uint(uid) //чтобы проверить, что лот принадлежит обновляющему
	Now, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	Lot.UpdatedAt = &Now

	lot, err := s.findAndUpdate(Lot)
	if err != nil {
		return nil, errors.ErrBadReq
	}

	c := s.getUser(lot.CreatorID)
	lot.Creator = &c
	if lot.BuyerID != 0 {
		b := s.getUser(lot.BuyerID)
		lot.Buyer = &b
	}
	if len(WsClients1.wsConn) != 0 {
		SingleLotCh <- lot
	}

	if len(WsClients2.wsConn) != 0 {
		AllLotsCh <- lot
	}

	fmt.Println("ch: ", AllLotsCh, " : ", lot)
	return &lot, nil
}

func (s LotsStorage) findAndUpdate(lot *Lot) (Lot, error) {
	db := s.DB
	var l Lot
	db.Where("ID = ?", lot.ID).First(&l)

	if lot.ID == l.ID && l.Status == "created" && l.CreatorID == lot.CreatorID && l.DeletedAt == nil {

		if lot.Title != "" {
			l.Title = lot.Title
		}
		if lot.Description != "" {
			l.Description = lot.Description
		}
		if lot.MinPrice != 0.0 {
			l.MinPrice = lot.MinPrice
		}
		if lot.PriceStep != 0.0 {
			l.PriceStep = lot.PriceStep
		}
		if !lot.EndAt.IsZero() {
			l.EndAt = lot.EndAt
		}

		if lot.Status == "active" || lot.Status == "created" {
			l.Status = lot.Status
		} else {
			return Lot{}, errors.ErrBadReq
		}
		l.UpdatedAt = lot.UpdatedAt
		db.Debug().Save(&l)
		return l, nil
	}

	return Lot{}, errors.ErrBadReq
}
func (s LotsStorage) getUser(uid uint) user.User {
	db := s.DB
	u := user.User{}
	db.Debug().Select("id,first_name,last_name").Where("id = ? ", uid).First(&u)
	return u
}
func (s LotsStorage) DeleteLot(uid int64, lotID uint) error {
	db := s.DB

	var l Lot
	db.Where("ID = ?", lotID).Find(&l)

	if int64(l.CreatorID) != uid || l.Status != "created" || l.ID != lotID {
		return errors.ErrNotFound
	}

	Now, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	l.UpdatedAt = &Now
	l.DeletedAt = &Now
	db.Save(l)

	if len(WsClients1.wsConn) != 0 {
		SingleLotCh <- Lot{ID: lotID,}
	}

	if len(WsClients2.wsConn) != 0 {
		AllLotsCh <- Lot{ID: lotID}
	}

	return nil
}
func (s LotsStorage) GetUserLots(uid int, lotsType string) ([]Lot, error) {
	db := s.DB
	var l []Lot
	switch lotsType {
	case "":
		db.Where("creator_id = ?", uid).Find(&l)
	case "own":
		db.Where("creator_id = ? AND status != ? ", uid, "finished").Find(&l)

	case "buyed":
		db.Where("creator_id = ? AND status = ? AND buyer_id != ? ", uid, "finished", 0).Find(&l)
	default:
		return nil, errors.ErrBadReq
	}

	if len(l) == 0 {
		return nil, errors.ErrNotFound
	}
	for i, v := range l {
		var c user.User
		var b user.User
		db.Where("id = ? ", v.CreatorID).Find(&c)
		db.Where(" id = ?", v.BuyerID).Find(&b)
		l[i].Creator = &user.User{ID: c.ID, FirstName: c.FirstName, LastName: c.LastName}
		if b.ID != 0 {
			l[i].Buyer = &user.User{ID: b.ID, FirstName: b.FirstName, LastName: b.LastName}
		} else {
			l[i].Buyer = nil
		}
	}
	return l, nil
}
func (s LotsStorage) GetAllLots(status string) ([]Lot, error) {
	db := s.DB
	var l []Lot
	switch status {
	case "":
		db.Find(&l)
	case "created":
		db.Where("status = ? ", "created").Find(&l)

	case "active":
		db.Where("status = ? ", "active").Find(&l)
	case "finished":
		db.Where("status = ? ", "finished").Find(&l)
	}

	if len(l) == 0 {
		return nil, errors.ErrNotFound
	}

	for i := range l {
		s.findLotCreator(&l[i])
	}
	return l, nil
}
func (s LotsStorage) findLotCreator(l *Lot) () {
	db := s.DB
	var c user.User
	var b user.User
	db.Where("id = ? ", l.CreatorID).Find(&c)
	db.Where(" id = ?", l.BuyerID).Find(&b)
	l.Creator = &user.User{ID: c.ID, FirstName: c.FirstName, LastName: c.LastName}
	if b.ID != 0 {
		l.Buyer = &user.User{ID: b.ID, FirstName: b.FirstName, LastName: b.LastName}
	} else {
		l.Buyer = nil
	}

}
func (s LotsStorage) GetLot(lotID int) (*Lot, error) {
	db := s.DB
	var l Lot
	db.Where("ID = ?", lotID).Find(&l)
	if l.ID == 0 {

		return nil, errors.ErrNotFound
	}

	s.findLotCreator(&l)

	return &l, nil
}
func (s LotsStorage) checkBuyingLot(uid int64, price float64, lotID int) (*Lot, error) {
	db := s.DB
	l := &Lot{}
	db.Where("ID = ?", lotID).Find(&l)
	switch {
	case l.ID == 0:
		return nil, errors.ErrBadReq
	case l.CreatorID == uint(uid) || l.BuyerID == uint(uid):
		return nil, errors.ErrConflict
	case l.Status != "active":
		return nil, errors.ErrConflict
	case l.BuyPrice >= price:
		return nil, errors.ErrConflict
	case l.MinPrice >= price:
		return nil, errors.ErrConflict
	case !(price/l.PriceStep == math.Trunc(price/l.PriceStep)):
		return nil, errors.ErrConflict
	}
	return l, nil
}

func (s LotsStorage) BuyLot(jsonData []byte, lotID int, uid int64) (*Lot, error) {
	db := s.DB

	var p map[string]float64
	err := json.Unmarshal(jsonData, &p)

	if err != nil {
		return nil, errors.ErrBadReq
	}
	if _, ok := p["price"]; !ok {
		return nil, errors.ErrBadReq
	}
	price := p["price"]
	l, err := s.checkBuyingLot(uid, price, lotID)
	if err != nil {
		return nil, err
	}
	l.BuyPrice = price
	l.BuyerID = uint(uid)
	db.Save(&l)

	var c user.User
	var b user.User
	db.Where("id = ? ", l.CreatorID).First(&c)
	db.Where(" id = ?", uid).First(&b)
	l.Creator = &user.User{ID: c.ID, FirstName: c.FirstName, LastName: c.LastName}

	l.Buyer = &user.User{ID: b.ID, FirstName: b.FirstName, LastName: b.LastName}
	if len(WsClients1.wsConn) != 0 {
		SingleLotCh <- *l
	}
	if len(WsClients2.wsConn) != 0 {
		AllLotsCh <- *l
	}
	return l, nil
}
func (s LotsStorage) CheckEndAt() {
	db := s.DB
	for {
		time.Sleep(time.Second)

		l := []Lot{}
		db.Find(&l)
		now := time.Now()

		for i := range l {
			if now.After(*l[i].EndAt) {
				l[i].Status = "finished"
				db.Save(&l[i])
				var c user.User
				var b user.User
				db.Where("id = ? ", l[i].CreatorID).First(&c)
				db.Where(" id = ?", l[i].BuyerID).First(&b)
				l[i].Creator = &user.User{ID: c.ID, FirstName: c.FirstName, LastName: c.LastName}
				if b.ID != 0 {
					l[i].Buyer = &user.User{ID: b.ID, FirstName: b.FirstName, LastName: b.LastName}
				} else {
					l[i].Buyer = nil
				}
				if len(WsClients1.wsConn) != 0 {
					SingleLotCh <- l[i]
				}
				if len(WsClients2.wsConn) != 0 {
					AllLotsCh <- l[i]
				}
			}

		}

	}
}
