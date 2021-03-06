package models

import (
	"errors"
	"time"

	"github.com/astaxie/beego"
	"github.com/jinzhu/gorm"
)

func Register(form *UserRegisterForm) error {
	// check name
	var count int
	DB.Model(&User{}).Where(&User{Email: form.Email}).Count(&count)
	if count != 0 {
		return errors.New("这个邮箱已经注册过账号了！")
	}

	DB.Model(&User{}).Where(&User{Name: form.Name}).Count(&count)
	if count != 0 {
		return errors.New("昵称重复了，换一个吧~")
	}

	DB.Model(&Page{}).Where(&Page{Domain: form.Domain}).Count(&count)
	if count != 0 {
		return errors.New("个性域名重复了，换一个吧~")
	}

	user := new(User)
	user.Name = form.Name
	user.Password = AddSalt(form.Password)
	user.Email = form.Email
	user.Avatar = beego.AppConfig.String("default_avatar") // default avatar

	// create page
	page := new(Page)
	page.Domain = form.Domain
	page.Intro = "问你想问的"
	page.Background = beego.AppConfig.String("default_background") // default background

	tx := DB.Begin()
	if tx.Create(&page).RowsAffected != 1 {
		tx.Rollback()
		return errors.New("注册失败，好像是服务器坏了...")
	}
	user.PageID = page.ID

	if tx.Create(&user).RowsAffected != 1 {
		tx.Rollback()
		return errors.New("注册失败，好像是服务器坏了...")
	}
	tx.Commit()
	return nil
}

func Login(form *UserLoginForm) (*User, error) {
	user := new(User)
	DB.Model(&User{}).Where(&User{Email: form.Email}).Find(&user)
	if user.Email == "" {
		return &User{}, errors.New("")
	}

	if user.Password == AddSalt(form.Password) {
		return user, nil
	}

	return &User{}, errors.New("")
}

func GetUserByPage(pageId uint) (*User, error) {
	user := new(User)
	DB.Model(&User{}).Where(&User{PageID: pageId}).Find(&user)
	if user.Name == "" {
		return &User{}, errors.New("")
	}
	return user, nil
}

func GetUserByEmail(email string) (*User, error) {
	user := new(User)
	DB.Model(&User{}).Where(&User{Email: email}).Find(&user)
	if user.ID == 0 {
		return nil, errors.New("")
	}
	return user, nil
}

func ValidateEmailCode(code string) (*EmailValidation, error) {
	mail := new(EmailValidation)
	DB.Model(&EmailValidation{}).Where("`code` = ?", code).Find(&mail)
	if mail.ID == 0 {
		return nil, errors.New("")
	}
	if mail.CreatedAt.Add(30 * time.Minute).Before(time.Now()) {
		return nil, errors.New("")
	}
	return mail, nil
}

func DeleteEmailCode(code string) {
	tx := DB.Begin()
	if tx.Delete(&EmailValidation{}, "`code` = ?", code).RowsAffected != 1 {
		tx.Rollback()
		return
	}
	tx.Commit()
}

func ResetUserPassword(userID uint, password string) {
	tx := DB.Begin()
	if tx.Model(&User{}).Where("`id` = ?", userID).Update(&User{Password: AddSalt(password)}).RowsAffected != 1 {
		tx.Rollback()
		return
	}
	tx.Commit()
}

func UpdateUser(id uint, u *User) {
	tx := DB.Begin()
	if tx.Model(&User{}).Where(&User{Model: gorm.Model{ID: id}}).Update(u).RowsAffected != 1 {
		tx.Rollback()
		return
	}
	tx.Commit()
}
