package main

import (
	"C"
	"log"

	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/tidwall/gjson"
)

//export GetMetadata
func GetMetadata(contract *C.char) *C.char {
	converter := JSONConverter{}

	data := gjson.Parse(C.GoString(contract))
	metadata, err := meta.ParseMetadata(data)
	if err != nil {
		return converter.EncodeError(err)
	}
	return converter.Encode(metadata)
}

//export MichelineToMiguel
func MichelineToMiguel(micheline *C.char, metadataBytes *C.char) *C.char {
	converter := JSONConverter{}

	data := gjson.Parse(C.GoString(micheline))

	var metadata meta.Metadata
	if err := converter.Decode([]byte(C.GoString(metadataBytes)), &metadata); err != nil {
		return converter.EncodeError(err)
	}

	miguel, err := newmiguel.MichelineToMiguel(data, metadata)
	if err != nil {
		return converter.EncodeError(err)
	}
	return converter.Encode(miguel)
}

//export MichelineToMichelson
func MichelineToMichelson(micheline *C.char, inline C.char, lineSize C.int) *C.char {
	converter := JSONConverter{}

	data := gjson.Parse(C.GoString(micheline))
	michelson, err := formatter.MichelineToMichelson(data, inline == 0, int(lineSize))
	if err != nil {
		return converter.EncodeError(err)
	}
	return converter.EncodeBytes([]byte(michelson))
}

//export ParameterToMiguel
func ParameterToMiguel(parameters *C.char, metadataBytes *C.char) *C.char {
	converter := JSONConverter{}

	var metadata meta.Metadata
	if err := converter.Decode([]byte(C.GoString(metadataBytes)), &metadata); err != nil {
		return converter.EncodeError(err)
	}

	data := gjson.Parse(C.GoString(parameters))
	miguel, err := newmiguel.ParameterToMiguel(data, metadata)
	if err != nil {
		return converter.EncodeError(err)
	}
	return converter.Encode(miguel)
}

//export Hello
func Hello() {
	log.Print("Hello from BCD library!")
}

func main() {}
