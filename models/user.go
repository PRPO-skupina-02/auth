package models

import (
	"errors"
	"time"

	"github.com/PRPO-skupina-02/common/request"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleEmployee UserRole = "employee"
	RoleAdmin    UserRole = "admin"
)

func (r UserRole) IsValid() bool {
	switch r {
	case RoleCustomer, RoleEmployee, RoleAdmin:
		return true
	}
	return false
}

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null" json:"-"`
	FirstName    string
	LastName     string
	Role         UserRole `gorm:"type:user_role;default:'customer'"`
	Active       bool     `gorm:"default:true"`
}

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

func (u *User) Create(tx *gorm.DB) error {
	if err := tx.Create(u).Error; err != nil {
		return err
	}
	return nil
}

func (u *User) Save(tx *gorm.DB) error {
	if err := tx.Save(u).Error; err != nil {
		return err
	}
	return nil
}

func (u *User) Delete(tx *gorm.DB) error {
	if err := tx.Delete(u).Error; err != nil {
		return err
	}
	return nil
}

func GetUser(tx *gorm.DB, id uuid.UUID) (User, error) {
	var user User
	if err := tx.Where("id = ?", id).First(&user).Error; err != nil {
		return user, err
	}
	return user, nil
}

func GetUserByEmail(tx *gorm.DB, email string) (User, error) {
	var user User
	if err := tx.Where("email = ?", email).First(&user).Error; err != nil {
		return user, err
	}
	return user, nil
}

func GetUsers(tx *gorm.DB, pagination *request.PaginationOptions, sort *request.SortOptions) ([]User, int64, error) {
	var users []User
	var total int64

	query := tx.Model(&User{})

	if err := query.Count(&total).Error; err != nil {
		return users, 0, err
	}

	if pagination != nil {
		query = query.Offset(pagination.Offset).Limit(pagination.Limit)
	}

	if sort != nil && sort.Column != "" {
		order := sort.Column
		if sort.Desc {
			order += " DESC"
		}
		query = query.Order(order)
	}

	if err := query.Find(&users).Error; err != nil {
		return users, 0, err
	}

	return users, total, nil
}

func UserExists(tx *gorm.DB, email string) (bool, error) {
	var count int64
	if err := tx.Model(&User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func ValidateCredentials(tx *gorm.DB, email, password string) (*User, error) {
	user, err := GetUserByEmail(tx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	if !user.Active {
		return nil, errors.New("user account is inactive")
	}

	if err := user.CheckPassword(password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}
