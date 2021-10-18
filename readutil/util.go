package readutil

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

func ReaderToJSON(b io.Reader, a interface{}) error {
	body, err := ioutil.ReadAll(b)
	if err != nil {
		return fmt.Errorf("error read body: %w", err)
	}
	err = json.Unmarshal(body, a)
	if err != nil {
		return fmt.Errorf("error json Unmarshal(%s): %w", string(body), err)
	}
	return nil
}
