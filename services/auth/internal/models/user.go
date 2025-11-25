package models

import "time"

// User represents a user in the auth service (no foreign keys to other services)
type User struct {
	ID            int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	FirstName     string     `gorm:"column:first_name;type:varchar(255);not null" json:"first_name"`
	LatestName    string     `gorm:"column:latest_name;type:varchar(255);not null" json:"latest_name"`
	FatherName    string     `gorm:"column:father_name;type:varchar(255);not null" json:"father_name"`
	NationalityID *int       `gorm:"column:nationality_id" json:"nationality_id,omitempty"`
	DateOfBirth   *time.Time `gorm:"column:dob;type:date" json:"dob,omitempty"`
	WorkEmail     string     `gorm:"column:work_email;type:varchar(255);not null;uniqueIndex" json:"work_email"`
	PrivateEmail  *string    `gorm:"column:private_email;type:varchar(255)" json:"private_email,omitempty"`
	Password      string     `gorm:"column:password;type:varchar(255);not null" json:"-"`
	WorkMobile    *string    `gorm:"column:work_mobile;type:varchar(255)" json:"work_mobile,omitempty"`
	PrivateMobile *string    `gorm:"column:private_mobile;type:varchar(255)" json:"private_mobile,omitempty"`
	IDNumber      *string    `gorm:"column:id_number;type:varchar(255)" json:"id_number,omitempty"`
	IDType        *int       `gorm:"column:id_type" json:"id_type,omitempty"`
	ContactID     *int       `gorm:"column:contact_id" json:"contact_id,omitempty"`
	Active        *int       `gorm:"column:active;default:1" json:"active,omitempty"`
	EmailVerified bool       `gorm:"column:email_verified;default:false" json:"email_verified"`
	PhoneVerified bool       `gorm:"column:phone_verified;default:false" json:"phone_verified"`
	CreatedAt     time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships (no foreign keys, just for GORM queries)
	Roles       []Role       `gorm:"many2many:user_role;foreignKey:ID;joinForeignKey:UserID;References:ID;joinReferences:RoleID" json:"roles,omitempty"`
	Permissions []Permission `gorm:"many2many:user_special_permission;foreignKey:ID;joinForeignKey:UserID;References:ID;joinReferences:PermissionID" json:"permissions,omitempty"`
}

func (User) TableName() string {
	return `"User"`
}

// UserRole represents the user_role join table
type UserRole struct {
	UserID int `gorm:"column:user_id;primaryKey" json:"user_id"`
	RoleID int `gorm:"column:role_id;primaryKey" json:"role_id"`
}

func (UserRole) TableName() string {
	return "user_role"
}

// UserPermission represents the user_special_permission join table
type UserPermission struct {
	UserID       int `gorm:"column:user_id;primaryKey" json:"user_id"`
	PermissionID int `gorm:"column:permission_id;primaryKey" json:"permission_id"`
}

func (UserPermission) TableName() string {
	return "user_special_permission"
}

