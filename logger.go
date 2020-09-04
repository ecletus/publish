package publish

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/moisespsena-go/aorm"
)

// LoggerInterface logger interface used to print publish logs
type LoggerInterface interface {
	Print(...interface{})
}

// Logger default logger used to print publish logs
var Logger LoggerInterface

func init() {
	Logger = log.New(os.Stdout, "\r\n", 0)
}

func stringify(object interface{}) string {
	if obj, ok := object.(interface {
		Stringify() string
	}); ok {
		return obj.Stringify()
	}

	ms := aorm.StructOf(object)
	if field := ms.FirstFieldValue(object, "Description", "Name", "Title", "Code"); field != nil {
		return fmt.Sprintf("%v", field.Field.Interface())
	}

	if id := ms.GetID(object); id != nil {
		if id.IsZero() {
			return ""
		}
		return fmt.Sprintf("%v#%s", ms.Type.Name(), id)
	}

	return fmt.Sprint(reflect.Indirect(reflect.ValueOf(object)).Interface())
}

func stringifyPrimaryValues(primaryValues [][][]interface{}, columns ...string) string {
	var values []string
	for _, primaryValue := range primaryValues {
		var primaryKeys []string
		for _, value := range primaryValue {
			if len(columns) == 0 {
				primaryKeys = append(primaryKeys, fmt.Sprint(value[1]))
			} else {
				for _, column := range columns {
					if column == fmt.Sprint(value[0]) {
						primaryKeys = append(primaryKeys, fmt.Sprint(value[1]))
					}
				}
			}
		}
		if len(primaryKeys) > 1 {
			values = append(values, fmt.Sprintf("[%v]", strings.Join(primaryKeys, ", ")))
		} else {
			values = append(values, primaryKeys...)
		}
	}
	return strings.Join(values, "; ")
}
