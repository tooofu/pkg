package binding

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

type StructValidator interface {
	// NOTE: If the received type is not a struct, any validation should be skipped and nil must be returned.
	ValidateStruct(interface{}) error
	// NOTE: if the key already exists, the previous validation function will be replaced.
	RegisterValidation(string, validator.Func) error
}

type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

var _ StructValidator = &defaultValidator{}

var Validator StructValidator = &defaultValidator{}

func (v *defaultValidator) ValidateStruct(obj interface{}) error {
	// kindOfData
	value := reflect.ValueOf(obj)
	valueType := value.Kind()
	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}

	// validate struct
	if valueType == reflect.Struct {
		v.lazyinit()
		if err := v.validate.Struct(obj); err != nil {
			// return err
			for _, e := range err.(validator.ValidationErrors) {
				// NOTE: only return first
				return errors.New(fmt.Sprintf("%v field %v", strings.ToLower(e.Field()), e.Tag()))
			}
		}
	}

	// default nil
	return nil
}

func (v *defaultValidator) RegisterValidation(key string, fn validator.Func) error {
	v.lazyinit()
	return v.validate.RegisterValidation(key, fn)
}

func (v *defaultValidator) lazyinit() {
	v.once.Do(func() {
		v.validate = validator.New()
	})
}

func validate(obj interface{}) error {
	if Validator == nil {
		return nil
	}
	return Validator.ValidateStruct(obj)
}
