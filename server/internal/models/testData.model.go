package models

import "github.com/google/uuid"

type TestData struct {
	BaseUUIDModel
	LoadTestID uuid.UUID `gorm:"type:uuid;not null;index" json:"loadTestId"`
	// Known date columns - these are validated and normalized to RFC3339 format with UTC timezone
	BirthDate     *string `gorm:"type:varchar(255)"               json:"birth_date"`     // Normalized RFC3339: "2006-01-02T15:04:05Z"
	StartDate     *string `gorm:"type:varchar(255)"               json:"start_date"`     // Normalized RFC3339: "2006-01-02T15:04:05Z"
	EndDate       *string `gorm:"type:varchar(255)"               json:"end_date"`       // Normalized RFC3339: "2006-01-02T15:04:05Z"
	// Meaningful columns (20 total for demographics, employment, and insurance data)
	FirstName        *string `gorm:"type:varchar(255)"               json:"first_name"`
	LastName         *string `gorm:"type:varchar(255)"               json:"last_name"`
	Email            *string `gorm:"type:varchar(255)"               json:"email"`
	Phone            *string `gorm:"type:varchar(255)"               json:"phone"`
	AddressLine1     *string `gorm:"type:varchar(255)"               json:"address_line_1"`
	AddressLine2     *string `gorm:"type:varchar(255)"               json:"address_line_2"`
	City             *string `gorm:"type:varchar(255)"               json:"city"`
	State            *string `gorm:"type:varchar(255)"               json:"state"`
	ZipCode          *string `gorm:"type:varchar(255)"               json:"zip_code"`
	Country          *string `gorm:"type:varchar(255)"               json:"country"`
	SocialSecurityNo *string `gorm:"type:varchar(255)"               json:"social_security_no"`
	Employer         *string `gorm:"type:varchar(255)"               json:"employer"`
	JobTitle         *string `gorm:"type:varchar(255)"               json:"job_title"`
	Department       *string `gorm:"type:varchar(255)"               json:"department"`
	Salary           *string `gorm:"type:varchar(255)"               json:"salary"`
	InsurancePlanID  *string `gorm:"type:varchar(255)"               json:"insurance_plan_id"`
	InsuranceCarrier *string `gorm:"type:varchar(255)"               json:"insurance_carrier"`
	PolicyNumber     *string `gorm:"type:varchar(255)"               json:"policy_number"`
	GroupNumber      *string `gorm:"type:varchar(255)"               json:"group_number"`
	MemberID         *string `gorm:"type:varchar(255)"               json:"member_id"`
}