package ldap

import (
	"errors"
	"fmt"
	"github.com/hsoj/asn1-ber"
)

const (
	ModAdd     = 0
	ModDelete  = 1
	ModReplace = 2
)

var ModMap map[uint8]string = map[uint8]string{
	ModAdd:     "add",
	ModDelete:  "delete",
	ModReplace: "replace",
}

/* Reuse search struct, should Values be a [][]byte
type EntryAttribute struct {
	Name   string
	Values []string
}
*/
type Mod struct {
	ModOperation uint8
	Modification EntryAttribute
}

type ModifyRequest struct {
	DN   string
	Mods []Mod
}

/* Example...

func modifyTest(l *ldap.Conn){
    var modDNs []string = []string{"cn=test,ou=People,dc=example,dc=com"}
    var modAttrs []string = []string{"cn"}
    var modValues []string = []string{"aaa", "bbb", "ccc"}
	modreq := ldap.NewModifyRequest(modDNs[0])
	mod := ldap.NewMod(ldap.ModAdd, modAttrs[0], modValues)
	modreq.AddMod(mod)
    err := l.Modify(modreq)
	if err != nil {
        fmt.Printf("Modify : %s : result = %d\n",modDNs[0],err.ResultCode)
        return
    }
    fmt.Printf("Modify Success")
}

*/

/*
   ModifyRequest ::= [APPLICATION 6] SEQUENCE {
         object          LDAPDN,
         changes         SEQUENCE OF change SEQUENCE {
              operation       ENUMERATED {
                   add     (0),
                   delete  (1),
                   replace (2),
                   ...  },
              modification    PartialAttribute } }
*/

func (l *Conn) Modify(modReq *ModifyRequest) *Error {
	messageID := l.nextMessageID()

	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Request")
	packet.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimative, ber.TagInteger, messageID, "MessageID"))
	packet.AppendChild(encodeModifyRequest(modReq))

	if l.Debug {
		ber.PrintPacket(packet)
	}

	channel, err := l.sendMessage(packet)

	if err != nil {
		return err
	}

	if channel == nil {
		return NewError(ErrorNetwork, errors.New("Could not send message"))
	}

	defer l.finishMessage(messageID)
	if l.Debug {
		fmt.Printf("%d: waiting for response\n", messageID)
	}

	packet = <-channel

	if l.Debug {
		fmt.Printf("%d: got response %p\n", messageID, packet)
	}

	if packet == nil {
		return NewError(ErrorNetwork, errors.New("Could not retrieve message"))
	}

	if l.Debug {
		if err := addLDAPDescriptions(packet); err != nil {
			return NewError(ErrorDebugging, err)
		}
		ber.PrintPacket(packet)
	}

	result_code, result_description := getLDAPResultCode(packet)

	if result_code != 0 {
		return NewError(result_code, errors.New(result_description))
	}

	if l.Debug {
		fmt.Printf("%d: returning\n", messageID)
	}
	// success
	return nil
}

func (req *ModifyRequest) Bytes() []byte {
	return encodeModifyRequest(req).Bytes()
}

func encodeModifyRequest(req *ModifyRequest) (p *ber.Packet) {
	modpacket := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ApplicationModifyRequest, nil, ApplicationMap[ApplicationModifyRequest])
	modpacket.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimative, ber.TagOctetString, req.DN, "LDAP DN"))
	seqOfChanges := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Changes")
	for _, mod := range req.Mods {
		modification := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Modification")
		op := ber.NewInteger(ber.ClassUniversal, ber.TypePrimative, ber.TagEnumerated, uint64(mod.ModOperation), "Modify Op ("+ModMap[mod.ModOperation]+")")
		modification.AppendChild(op)
		partAttr := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "PartialAttribute")

		partAttr.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimative, ber.TagOctetString, mod.Modification.Name, "AttributeDescription"))
		valuesSet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSet, nil, "Attribute Value Set")
		for _, val := range mod.Modification.Values {
			value := ber.NewString(ber.ClassUniversal, ber.TypePrimative, ber.TagOctetString, val, "AttributeValue")
			valuesSet.AppendChild(value)
		}
		partAttr.AppendChild(valuesSet)
		modification.AppendChild(partAttr)
		seqOfChanges.AppendChild(modification)
	}
	modpacket.AppendChild(seqOfChanges)

	return modpacket
}

func NewModifyRequest(dn string) (req *ModifyRequest) {
	req = &ModifyRequest{DN: dn, Mods: make([]Mod, 0, 5)}
	return
}

// Basic LDIF dump, no formating, etc
func (req *ModifyRequest) DumpModRequest() (dump string) {
	dump = fmt.Sprintf("dn: %s\n", req.DN)
	for _, mod := range req.Mods {
		dump += mod.DumpMod()
	}
	return
}

// Basic LDIF dump, no formating, etc
func (mod *Mod) DumpMod() (dump string) {
	dump = fmt.Sprintf("changetype: modify\n")
	dump += fmt.Sprintf("%s: %s\n", ModMap[mod.ModOperation], mod.Modification.Name)
	for _, val := range mod.Modification.Values {
		dump += fmt.Sprintf("%s: %s\n", mod.Modification.Name, val)
	}
	dump += "-\n"
	return dump
}

func NewMod(modType uint8, attr string, values []string) (mod *Mod) {
	partEntryAttr := EntryAttribute{Name: attr, Values: values}
	mod = &Mod{ModOperation: modType, Modification: partEntryAttr}
	return
}

func (req *ModifyRequest) AddMod(mod *Mod) {
	req.Mods = append(req.Mods, *mod)
}

func (req *ModifyRequest) AddMods(mods []Mod) {
	req.Mods = append(req.Mods, mods...)
}