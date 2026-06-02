package types

type BloodGroup string

const (
	BloodGroupTypeApositive  BloodGroup = "A+"
	BloodGroupTypeANegative  BloodGroup = "A-"
	BloodGroupTypeBpositive  BloodGroup = "B+"
	BloodGroupTypeBnegative  BloodGroup = "B-"
	BloodGroupTypeABpositive BloodGroup = "AB+"
	BloodGroupTypeABnegative BloodGroup = "AB-"
	BloodGroupTypeOpositive  BloodGroup = "O+"
	BloodGroupTypeOnegative  BloodGroup = "O-"
	BloodGroupTypeUnknown    BloodGroup = "unknown"
)