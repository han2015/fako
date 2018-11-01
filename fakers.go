package fako

import (
	"math/rand"
	"reflect"
	"strings"
	"time"

	"math"

	"github.com/icrowley/fake"
	"github.com/serenize/snaker"
)

var customGenerators = map[string]func() string{}
var typeMapping = map[string]func() string{
	"Brand":                    fake.Brand,
	"Character":                fake.Character,
	"Characters":               fake.Characters,
	"City":                     fake.City,
	"Color":                    fake.Color,
	"Company":                  fake.Company,
	"Continent":                fake.Continent,
	"Country":                  fake.Country,
	"CreditCardType":           fake.CreditCardType,
	"Currency":                 fake.Currency,
	"CurrencyCode":             fake.CurrencyCode,
	"Digits":                   fake.Digits,
	"DomainName":               fake.DomainName,
	"DomainZone":               fake.DomainZone,
	"EmailAddress":             fake.EmailAddress,
	"EmailBody":                fake.EmailBody,
	"EmailSubject":             fake.EmailSubject,
	"FemaleFirstName":          fake.FemaleFirstName,
	"FemaleFullName":           fake.FemaleFullName,
	"FemaleFullNameWithPrefix": fake.FemaleFullNameWithPrefix,
	"FemaleFullNameWithSuffix": fake.FemaleFullNameWithSuffix,
	"FemaleLastName":           fake.FemaleLastName,
	"FemalePatronymic":         fake.FemalePatronymic,
	"FirstName":                fake.FirstName,
	"FullName":                 fake.FullName,
	"FullNameWithPrefix":       fake.FullNameWithPrefix,
	"FullNameWithSuffix":       fake.FullNameWithSuffix,
	"Gender":                   fake.Gender,
	"GenderAbbrev":             fake.GenderAbbrev,
	"HexColor":                 fake.HexColor,
	"HexColorShort":            fake.HexColorShort,
	"IPv4":                     fake.IPv4,
	"Industry":                 fake.Industry,
	"JobTitle":                 fake.JobTitle,
	"Language":                 fake.Language,
	"LastName":                 fake.LastName,
	"LatitudeDirection":        fake.LatitudeDirection,
	"LongitudeDirection":       fake.LongitudeDirection,
	"MaleFirstName":            fake.MaleFirstName,
	"MaleFullName":             fake.MaleFullName,
	"MaleFullNameWithPrefix":   fake.MaleFullNameWithPrefix,
	"MaleFullNameWithSuffix":   fake.MaleFullNameWithSuffix,
	"MaleLastName":             fake.MaleLastName,
	"MalePatronymic":           fake.MalePatronymic,
	"Model":                    fake.Model,
	"Month":                    fake.Month,
	"MonthShort":               fake.MonthShort,
	"Paragraph":                fake.Paragraph,
	"Paragraphs":               fake.Paragraphs,
	"Patronymic":               fake.Patronymic,
	"Phone":                    fake.Phone,
	"Product":                  fake.Product,
	"ProductName":              fake.ProductName,
	"Sentence":                 fake.Sentence,
	"Sentences":                fake.Sentences,
	"SimplePassword":           fake.SimplePassword,
	"State":                    fake.State,
	"StateAbbrev":              fake.StateAbbrev,
	"Street":                   fake.Street,
	"StreetAddress":            fake.StreetAddress,
	"Title":                    fake.Title,
	"TopLevelDomain":           fake.TopLevelDomain,
	"UserName":                 fake.UserName,
	"WeekDay":                  fake.WeekDay,
	"WeekDayShort":             fake.WeekDayShort,
	"Word":                     fake.Word,
	"Words":                    fake.Words,
	"Zip":                      fake.Zip,
}

var validArrayTypes = []string{
	"[]string",
	"[]int",
	"[]int32",
	"[]int64",
	"[]float32",
	"[]float64",
}

// Register allows user to add his own data generators for special cases
// that we could not cover with the generators that fako includes by default.
func Register(identifier string, generator func() string) {
	fakeType := snaker.SnakeToCamel(identifier)
	customGenerators[fakeType] = generator
}

// Fuzz Fills passed interface with random data based on the struct field type,
// take a look at fuzzValueFor for details on supported data types.
func Fuzz(e interface{}) {
	ty := reflect.TypeOf(e)

	if ty.Kind() == reflect.Ptr {
		ty = ty.Elem()
	}

	if ty.Kind() == reflect.Slice {
		structs := reflect.MakeSlice(ty, 0, 3)
		_struct := ty.Elem()

		for i := 0; i < 3; i++ {
			_field := reflect.New(_struct)
			Fuzz(_field.Interface())
			structs = reflect.Append(structs, reflect.Indirect(_field))
		}

		reflect.ValueOf(e).Elem().Set(reflect.Indirect(structs))
		return
	}

	if ty.Kind() == reflect.Struct {
		value := reflect.ValueOf(e).Elem()
		for i := 0; i < ty.NumField(); i++ {
			field := value.Field(i)
			setValueForField(field)
		}
	}
}

func setValueForField(field reflect.Value) {
	if field.CanSet() {
		fieldType := field.Type().String()
		if fieldType == "time.Duration" {
			field.Set(randTimeDurationValue())
		} else if strings.Contains(fieldType, "Time") {
			ft := time.Now().Add(time.Duration(rand.Int63()))
			if strings.HasPrefix(fieldType, "*") {
				field.Set(reflect.ValueOf(&ft))
			} else {
				field.Set(reflect.ValueOf(ft))
			}
		} else if field.Kind() == reflect.Array || field.Kind() == reflect.Slice {
			//Not support struct array
			if strInArray(fieldType, validArrayTypes) {
				field.Set(randArray(fieldType))
				return
			}

			//TODO: to support struct array
			_field := reflect.New(field.Type())
			Fuzz(_field.Interface())
			field.Set(reflect.Indirect(_field))
		} else if field.Kind() == reflect.Struct {
			_field := reflect.New(field.Type())
			Fuzz(_field.Interface())
			field.Set(reflect.Indirect(_field))
		} else {
			_type := field.Type().Name()
			if _type == field.Kind().String() {
				field.Set(fuzzValueFor(field.Kind()))
				return
			}
			//to fix string alias: e.g json.Number
			field.Set(fuzzValueFor(field.Kind()).Convert(field.Type()))
		}
	}
}

func randTimeDurationValue() reflect.Value {
	randTime := rand.Int31n(3000)
	return reflect.ValueOf(time.Duration(randTime) * time.Millisecond)
}

func allGenerators() map[string]func() string {
	dst := typeMapping
	for k, v := range customGenerators {
		dst[k] = v
	}

	return dst
}

//findFakeFunctionFor returns a faker function for a fako identifier
func findFakeFunctionFor(fako string) func() string {
	result := func() string { return "" }

	for kind, function := range allGenerators() {
		if fako == kind {
			result = function
			break
		}
	}

	return result
}

// fuzzValueFor Generates random values for the following types:
// string, bool, int, int32, int64, float32, float64
func fuzzValueFor(kind reflect.Kind) reflect.Value {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	switch kind {
	case reflect.String:
		return reflect.ValueOf(randomString(25))
	case reflect.Int:
		return reflect.ValueOf(r.Intn(math.MaxInt32))
	case reflect.Int32:
		return reflect.ValueOf(r.Int31())
	case reflect.Int64:
		return reflect.ValueOf(r.Int63())
	case reflect.Float32:
		return reflect.ValueOf(r.Float32())
	case reflect.Float64:
		return reflect.ValueOf(r.Float64())
	case reflect.Uint:
		return reflect.ValueOf(uint(r.Uint32()))
	case reflect.Uint8:
		return reflect.ValueOf(uint8(r.Uint32()))
	case reflect.Uint16:
		return reflect.ValueOf(uint16(r.Uint32()))
	case reflect.Uint32:
		return reflect.ValueOf(r.Uint32())
	case reflect.Uint64:
		return reflect.ValueOf(r.Uint64())
	case reflect.Bool:
		val := r.Intn(2) > 0
		return reflect.ValueOf(val)
	}

	return reflect.ValueOf("")
}

func randArray(fieldType string) reflect.Value {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	numberOfElements := rand.Intn(20)
	switch fieldType {
	case "[]string":
		var data []string
		for i := 0; i < numberOfElements; i++ {
			data = append(data, randomString(10))
		}
		return reflect.ValueOf(data)
	case "[]int":
		var data []int
		for i := 0; i < numberOfElements; i++ {
			data = append(data, r.Int())
		}
		return reflect.ValueOf(data)
	case "[]int32":
		var data []int32
		for i := 0; i < numberOfElements; i++ {
			data = append(data, r.Int31())
		}
		return reflect.ValueOf(data)
	case "[]int64":
		var data []int64
		for i := 0; i < numberOfElements; i++ {
			data = append(data, r.Int63())
		}
		return reflect.ValueOf(data)
	case "[]float32":
		var data []float32
		for i := 0; i < numberOfElements; i++ {
			data = append(data, r.Float32())
		}
		return reflect.ValueOf(data)
	case "[]float64":
		var data []float64
		for i := 0; i < numberOfElements; i++ {
			data = append(data, r.Float64())
		}
		return reflect.ValueOf(data)
	}

	return reflect.ValueOf("")
}
