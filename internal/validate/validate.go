package validate

import (
	"fmt"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/errors"
	"reflect"
	"strings"
)

type Validator struct {
	validator *validator.Validate
	ut        ut.Translator
}

func New() *Validator {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	english := en.New()
	uni := ut.New(english, english)
	trans, _ := uni.GetTranslator("en")

	err := enTranslations.RegisterDefaultTranslations(v, trans)
	if err != nil {
		panic(err)
	}

	return &Validator{validator: v, ut: trans}
}

func (v *Validator) Struct(s interface{}) error {
	err := v.validator.Struct(s)
	if err != nil {
		var msg string
		for _, err := range err.(validator.ValidationErrors) {
			msg += fmt.Sprintf("%v\n", err.Translate(v.ut))
		}

		return errors.E(errors.Validation, msg)
	}

	return nil
}
