package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
)

// WriteAndReturn writes log and return value
func WriteAndReturn(w http.ResponseWriter, msg string) error {
	logrus.Println(msg)

	_, err := fmt.Fprint(w, msg)
	if err != nil {
		return err
	}

	return nil
}

// WriteError writes error only
func WriteError(err error) {
	logrus.Errorln(err.Error())
}
