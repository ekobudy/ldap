// generated by stringer -type=Deref; DO NOT EDIT

package ldap

import "fmt"

const _Deref_name = "NeverDerefAliasesDerefInSearchingDerefFindingBaseObjDerefAlways"

var _Deref_index = [...]uint8{0, 17, 33, 52, 63}

func (i Deref) String() string {
	if i >= Deref(len(_Deref_index)-1) {
		return fmt.Sprintf("Deref(%d)", i)
	}
	return _Deref_name[_Deref_index[i]:_Deref_index[i+1]]
}
