package localvalidator

import (
	"gitlab.com/TitanInd/lumerin/lumerinlib"
)

func BoilerPlateFunc() (string, error) {
	msg := "Local Validator Package"
	return lumerinlib.BoilerPlateLibFunc(msg), nil // always returns no error
}