package async

import "fmt"

type Errors []error

func (es Errors) Error() string {
	return fmt.Sprint([]error(es))
}
