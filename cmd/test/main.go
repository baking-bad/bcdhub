package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcdast"
)

// Data -
type Data struct {
	Code    json.RawMessage `json:"code"`
	Storage json.RawMessage `json:"storage"`
}

func main() {
	f, err := os.Open("contract1.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var contract Data
	if err := json.NewDecoder(f).Decode(&contract); err != nil {
		panic(err)
	}

	ts := time.Now()
	script, err := bcdast.NewScript(contract.Code)
	if err != nil {
		panic(err)
	}

	fmt.Println("------ANNOTS-------")
	fmt.Printf("Parameter: %s\n", script.Storage.Annotations())
	fmt.Println("")

	parameter, err := script.Parameter.ToTypedAST()
	if err != nil {
		panic(err)
	}
	fmt.Println("------ENTRYPOINTS-------")
	fmt.Println(parameter.GetEntrypoints())

	storage, err := script.Storage.ToTypedAST()
	if err != nil {
		panic(err)
	}

	storageData, err := bcdast.NewUntypedAST(contract.Storage)
	if err != nil {
		panic(err)
	}

	if err := storage.Settle(storageData); err != nil {
		panic(err)
	}

	miguel, err := storage.ToMiguel()
	if err != nil {
		panic(err)
	}
	fmt.Println("")
	fmt.Println("------STORAGE MIGUEL-------")
	fmt.Println(miguel)

	fmt.Printf("Spent: %d ms\n", time.Since(ts).Milliseconds())
}