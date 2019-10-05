package user

import (
	"time"
	"encoding/json"
	"github.com/DiGregory/tfs-auction/internal/errors"
	//postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/jinzhu/gorm"
)

type User struct {
	ID        int64      `json:"id"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	BirthDay  string     `json:"birthday,omitempty"`
	Email     string     `json:"email,omitempty"`
	Password  string     `json:"password,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"-"`
}
type UsersStorage struct {
	DB *gorm.DB
}


func StorageConnect(source string) (*UsersStorage, error) {

	db, err := gorm.Open("postgres", source)
	if err != nil {
		return nil, err
	}

	return &UsersStorage{DB: db,}, nil

}

func (s UsersStorage) getLastID() (id int64) {
	db := s.DB
	var u User
	db.Last(&u)
	if id == -1 {
		return
	}
	return u.ID

}
func (s UsersStorage) existEmail(email string) bool {
	var u User
	s.DB.Where("email = ?", email).First(&u)

	if u.Email == "" {
		return true
	}
	return false

}
func (s UsersStorage) findAndUpdate(user *User) (User) {
	var u User
	db := s.DB
	db.Where("ID = ?", user.ID).First(&u)
	if user.ID == u.ID {
		if user.FirstName != "" {
			u.FirstName = user.FirstName
		}
		if user.LastName != "" {
			u.LastName = user.LastName
		}
		if user.BirthDay != "" {
			u.BirthDay = user.BirthDay
		}
		u.UpdatedAt = user.UpdatedAt
		db.Debug().Save(&u)
		return u
	}

	return User{}
}

func (s UsersStorage) RegUser(jsonData []byte) error {
	db := s.DB

	var u interface{}
	err := json.Unmarshal(jsonData, &u)
	if err != nil {
		return errors.ErrBadReq
	}
	jsonMap := u.(map[string]interface{})

	User := &User{}

	User.ID = s.getLastID() + 1

	var ok bool
	if User.FirstName, ok = jsonMap["first_name"].(string); !ok || User.FirstName == "" {
		return errors.ErrBadReq
	}
	if User.LastName, ok = jsonMap["last_name"].(string); !ok || User.LastName == "" {
		return errors.ErrBadReq
	}
	if User.Password, ok = jsonMap["password"].(string); !ok || User.Password == "" {
		return errors.ErrBadReq
	}
	if User.Email, ok = jsonMap["email"].(string); !ok || User.Email == "" {
		return errors.ErrBadReq
	}

	IsGoodEmail := s.existEmail(User.Email)
	if !IsGoodEmail {
		return errors.ErrEmailExist
	}
	if _, ok = jsonMap["birthday"]; ok {
		User.BirthDay = jsonMap["birthday"].(string)
	}

	db.Create(&User)

	return  nil

}

func (s UsersStorage) UpdateUser(jsonData []byte, uid int64) (*User, error) {

	var u interface{}
	err := json.Unmarshal(jsonData, &u)

	if err != nil {
		return nil, errors.ErrBadReq
	}
	jsonMap := u.(map[string]interface{})

	User := &User{}
	var ok bool

	if _, ok = jsonMap["id"]; ok {
		return nil, errors.ErrBadReq
	}
	User.ID = uid
	if _, ok = jsonMap["email"]; ok {
		return nil, errors.ErrBadReq
	}
	if _, ok = jsonMap["password"]; ok {
		return nil, errors.ErrBadReq
	}
	if _, ok = jsonMap["created_at"]; ok {
		return nil, errors.ErrBadReq
	}
	if _, ok = jsonMap["first_name"]; ok {
		User.FirstName = jsonMap["first_name"].(string)
	}
	if _, ok = jsonMap["last_name"]; ok {
		User.LastName = jsonMap["last_name"].(string)
	}
	if _, ok = jsonMap["birthday"]; ok {
		User.BirthDay = jsonMap["birthday"].(string)
	}
	Now, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	User.UpdatedAt = &Now
	UpdatedUser := s.findAndUpdate(User)
	UpdatedUser.Password = ""

	return &UpdatedUser, nil
}

func (s UsersStorage) GetUser(id int) (*User, error) {

	db := s.DB
	User := &User{}

	db.Where("id = ? ", id).First(User)
	if User.ID == 0 {
		return nil, errors.ErrBadReq
	}
	User.Password = ""
	return User, nil
}
