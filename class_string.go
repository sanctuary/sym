// Code generated by "stringer -linecomment -type Class"; DO NOT EDIT.

package sym

import "strconv"

const (
	_Class_name_0 = "STRTAG"
	_Class_name_1 = "TPDEF"
)

func (i Class) String() string {
	switch {
	case i == 10:
		return _Class_name_0
	case i == 13:
		return _Class_name_1
	default:
		return "Class(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
