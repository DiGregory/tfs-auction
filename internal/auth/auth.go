package auth

import (
	"time"
	"encoding/json"
	"github.com/DiGregory/tfs-auction/internal/errors"
	"fmt"
	"crypto/rand"
	"github.com/DiGregory/tfs-auction/internal/session"
	//postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/DiGregory/tfs-auction/internal/user"
	"github.com/jinzhu/gorm"
)

type Token struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
}

func MakeToken() (uuid string) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	uuid = fmt.Sprintf("%X%X%X%X", b[0:4], b[4:6], b[8:10], b[10:])
	return
}
func CheckEmailAndPass(currUser *user.User, db *gorm.DB) (int64, bool) {
	var u user.User
	db.Where("email = ? AND Password = ?", currUser.Email, currUser.Password).First(&u)
	if u.Password != "" && u.Email != "" {
		return u.ID, true
	}
	return -1, false
}

func CheckAuth(jsonData []byte) (*Token, error) {
	db, err := gorm.Open("postgres", "user=postgres password=1234 dbname=test sslmode=disable")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()
	db.AutoMigrate(&session.Session{})

	u := new(user.User)
	err = json.Unmarshal(jsonData, &u)
	if err != nil {
		return nil, errors.ErrInvalidEmailOrPass
	}
	UID, GoodData := CheckEmailAndPass(u, db)
	if !GoodData {
		return nil, errors.ErrInvalidEmailOrPass
	}

	T := Token{
		TokenType:   "bearer",
		AccessToken: MakeToken(),
	}
	NextDay:=time.Now().Add(time.Hour * 24)
	s := session.Session{
		SessionID:  T.AccessToken,
		UserID:     UID,
		ValidUntil: &NextDay,
	}

	db.Save(&s)

	return &T, nil

}

func CheckToken(t *Token) (int64, bool) {
	if t == nil {
		return -1, false
	}
	db, err := gorm.Open("postgres", "user=postgres password=1234 dbname=test sslmode=disable")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()
	if t.TokenType != "bearer" {
		return -1, false
	}
	var s session.Session
	db.Where("Session_ID = ?", t.AccessToken).First(&s)
	if t.AccessToken == s.SessionID {
		return s.UserID, true
	}
	return -1, false
}
