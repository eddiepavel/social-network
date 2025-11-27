package utils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateInt(value interface{}, key string) error {
	switch val := value.(type) {
	case int, int8, int16, int32, int64:
		return nil
	case float32, float64:
		// JSON numbers become float64 by default; allow ints encoded as float64 (e.g., 42 -> 42.0)
		// Reject if it isn't an integer value.
		f := any(val).(float64)
		if f == float64(int64(f)) {
			return nil
		}
		return errors.New(key + " value is not a valid integer")
	case json.Number:
		if _, err := val.Int64(); err == nil {
			return nil
		}
	case string:
		s := strings.TrimSpace(val)
		if _, err := strconv.Atoi(s); err == nil {
			return nil
		}
	}
	return errors.New(key + " value is not a valid integer")
}

func (v *Validator) ValidateEmail(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return errors.New("value is not a string")
	}
	str = strings.TrimSpace(str)
	if _, err := mail.ParseAddress(str); err != nil {
		return errors.New("invalid email format")
	}
	return nil
}

func (v *Validator) Required(value interface{}, key string) error {
	if value == nil {
		return errors.New(key + " is required")
	}
	switch t := value.(type) {
	case string:
		if strings.TrimSpace(t) == "" {
			return errors.New(key + " is required")
		}
	case []any:
		if len(t) == 0 {
			return errors.New(key + " is required")
		}
	case map[string]any:
		if len(t) == 0 {
			return errors.New(key + " is required")
		}
	}
	return nil
}

func (v *Validator) IsBoolean(value interface{}, key string) error {

	_, ok := value.(bool)

	if !ok {
		return errors.New(key + " must be a boolean")
	}

	return nil
}

// helper: fetch nested keys like "user.email"
func getByPath(m map[string]any, path string) (any, bool) {
	cur := any(m)
	parts := strings.Split(path, ".")
	for _, p := range parts {
		obj, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		v, ok := obj[p]
		if !ok {
			return nil, false
		}
		cur = v
	}
	return cur, true
}

// unchanged: ValidateInput (the version that returns []string)
// (If you currently return a single error, you can keep that;
// just adapt the call below accordingly.)

func (v *Validator) ValidateInput(value interface{}, rules []interface{}, key string, hold map[string]interface{}) []string {
	var errs []string
	for _, rule := range rules {
		switch r := rule.(type) {
		case string:
			switch {
			case r == "string":
				if _, ok := value.(string); !ok {
					errs = append(errs, key+" value is not a valid string")
				}
			case r == "boolean":
				if err := v.IsBoolean(value, key); err != nil {
					errs = append(errs, err.Error())
				}
			case r == "int":
				if err := v.ValidateInt(value, key); err != nil {
					errs = append(errs, err.Error())
				}
			case r == "email":
				if err := v.ValidateEmail(value); err != nil {
					errs = append(errs, err.Error())
				}
			case r == "required":
				if err := v.Required(value, key); err != nil {
					errs = append(errs, err.Error())
				}
			case r == "sometimes":
				if value == nil {
					rules = nil
					return nil
				}
			case r == "base64":
				str, ok := value.(string)
				if !ok {
					errs = append(errs, "image must be base64")
					break
				}
				if idx := strings.Index(str, ","); idx != -1 {
					str = str[idx+1:]
				}
				if _, err := base64.StdEncoding.DecodeString(str); err != nil {
					errs = append(errs, "image must be base64")
				}
			case strings.HasPrefix(r, "same:"):
				other := strings.TrimPrefix(r, "same:")
				if value != hold[other] {
					errs = append(errs, key+" must match "+other)
				}
			case strings.HasPrefix(r, "min:"):
				n, _ := strconv.Atoi(strings.TrimPrefix(r, "min:"))
				s, ok := value.(string)
				if !ok || len(s) < n {
					errs = append(errs, key+" must be at least "+strconv.Itoa(n)+" characters")
				}
			case strings.HasPrefix(r, "max:"):
				n, _ := strconv.Atoi(strings.TrimPrefix(r, "max:"))
				s, ok := value.(string)
				if !ok || len(s) > n {
					errs = append(errs, key+" must be at most "+strconv.Itoa(n)+" characters")
				}
			default:
				errs = append(errs, "unknown validation rule: "+r)
			}
		case func(interface{}) error:
			if err := r(value); err != nil {
				errs = append(errs, err.Error())
			}
		default:
			errs = append(errs, "invalid validation rule type")
		}
	}
	return errs
}

// NEW: validate against JSON body instead of form values.
func Validate(r *http.Request, inputs map[string][]interface{}, req any) (bool, map[string][]string) {

	bodyBytes, err := io.ReadAll(r.Body)

	if err != nil {
		return false, map[string][]string{"_json": {"bad request"}}
	}

	var body map[string]any

	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return false, map[string][]string{"_json": {"bad request"}}
	}

	defer r.Body.Close()

	if err := json.Unmarshal(bodyBytes, req); err != nil {
		return false, map[string][]string{"_json": {"bad request"}}
	}

	v := NewValidator()
	errs := make(map[string][]string)
	hold := make(map[string]interface{})

	// Collect all referenced values into hold (for same: rules)
	for key := range inputs {
		if val, ok := getByPath(body, key); ok {
			hold[key] = val
		} else {
			hold[key] = nil
		}
	}

	for key, rules := range inputs {
		val, _ := getByPath(body, key) // nil if missing
		fieldErrs := v.ValidateInput(val, rules, key, hold)
		if len(fieldErrs) > 0 {
			errs[key] = fieldErrs
		}
	}

	return len(errs) == 0, errs
}
