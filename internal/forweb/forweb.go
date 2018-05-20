package forweb

import (
    "encoding/json"
    "io/ioutil"
)

func Save(hist interface{}, f string) {
    out, err := json.Marshal(hist)
    if err != nil {
        panic(err)
    }

    err = ioutil.WriteFile(f, []byte(string(out)), 0644)
    if err != nil {
        panic(err)
    }
}