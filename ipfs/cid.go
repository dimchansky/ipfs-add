package ipfs

import (
	"encoding/json"
	"fmt"
)

// Cid represents a self-describing content adressed
type Cid string

// String returns the default string representation of a Cid.
func (c Cid) String() string {
	return string(c)
}

// Defined returns true if a Cid is defined
// Calling any other methods on an undefined Cid will result in
// undefined behavior.
func (c Cid) Defined() bool {
	return c != ""
}

type cid struct {
	CidTarget string `json:"/"`
}

// MarshalJSON procudes a JSON representation of a Cid, which looks as follows:
//
//    { "/": "<cid-string>" }
//
// Note that this formatting comes from the IPLD specification
// (https://github.com/ipld/specs/tree/master/ipld)
func (c Cid) MarshalJSON() ([]byte, error) {
	if !c.Defined() {
		return []byte("null"), nil
	}
	return json.Marshal(cid{
		CidTarget: string(c),
	})
}

// UnmarshalJSON parses the JSON representation of a Cid.
func (c *Cid) UnmarshalJSON(b []byte) error {
	if len(b) < 2 {
		return fmt.Errorf("invalid cid json blob")
	}
	var obj cid
	objptr := &obj
	err := json.Unmarshal(b, &objptr)
	if err != nil {
		return err
	}
	if objptr == nil {
		*c = Cid("")
		return nil
	}

	if obj.CidTarget == "" {
		return fmt.Errorf("cid was incorrectly formatted")
	}

	*c = Cid(obj.CidTarget)

	return nil
}

// ToLink creates link from Cid
func (c *Cid) ToLink(name string, size uint64) Link { return Link{Cid: *c, Name: name, Size: size} }
