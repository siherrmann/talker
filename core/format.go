package core

import (
	"encoding/json"

	"github.com/siherrmann/talker/model"
	"github.com/siherrmann/validator"
)

func ValidateJSON(output string, v *validator.Validator) error {
	var target model.TargetJSONOutput
	err := json.Unmarshal([]byte(output), &target)
	if err != nil {
		return err
	}
	return v.Validate(&target)
}
