package encodedTime

import (
	"bytes"
	"fmt"
	"time"
	"encoding/json"
	"testing"
)


func TestMillisecsUnmarshall(t *testing.T) {

	v := struct {
		Timestamp Millisecs
	}{}

	err := json.Unmarshal([]byte(`{"Timestamp":1449808143436}`), &v)
	if err != nil {
		t.Fatal(err)
	}

	if n:=time.Time(v.Timestamp).Sub(time.Unix(1449808143436/1000, 0));n != 0 {
		t.Fatal(fmt.Errorf("times not equal:%d",n))
	}
}


// SSBQuirk
func TestFloats(t *testing.T) {

	v := struct {
		Timestamp Millisecs
	}{}
	
	err := json.Unmarshal([]byte(`{"Timestamp":1553708494043.0059}`), &v)
	if err != nil {
		t.Fatal(err)
	}

	if n:=time.Time(v.Timestamp).Sub(time.Unix(15537084940430059/1000, 0));n != 0 {
		t.Fatal(fmt.Errorf("times not equal:%d",n))
	}
}

func TestMillisecsMarshal(t *testing.T) {

	v := struct {
		Date Millisecs
	}{Millisecs(time.Unix(12345, 0))}

	out, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(out, []byte(`{"Date":12345000}`)) != 0 {
		t.Fatal(fmt.Errorf("times not equal - got %q", out))
	}

}
