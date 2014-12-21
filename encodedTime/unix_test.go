package encodedTime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestUnixUnmarshall(t *testing.T) {

	v := struct {
		Date Unix
	}{}

	err := json.Unmarshal([]byte(`{"Date":12345}`), &v)
	if err != nil {
		t.Fatal(err)
	}

	if time.Time(v.Date).Sub(time.Unix(12345, 0)) != 0 {
		t.Fatal(fmt.Errorf("times not equal"))
	}

}

func TestUnixMarshal(t *testing.T) {

	v := struct {
		Date Unix
	}{Unix(time.Unix(12345, 0))}

	out, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(out, []byte(`{"Date":12345}`)) != 0 {
		t.Fatal(fmt.Errorf("times not equal - got %q", out))
	}

}
