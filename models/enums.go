package models

// Gender 性别类型
type Gender string

const (
	GenderMale    Gender = "male"
	GenderFemale  Gender = "female"
	GenderUnknown Gender = "unknown"
)

// Foot 惯用脚类型
type Foot string

const (
	FootRight Foot = "right"
	FootLeft  Foot = "left"
	FootBoth  Foot = "both"
)

// String 方法返回字符串表示
func (g Gender) String() string {
	return string(g)
}

func (f Foot) String() string {
	return string(f)
}
