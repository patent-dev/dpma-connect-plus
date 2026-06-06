package dpmaconnect_test

import (
	"fmt"
	"time"

	dpma "github.com/patent-dev/dpma-connect-plus"
)

func ExampleNewClient() {
	config := dpma.DefaultConfig()
	config.Username = "your-username"
	config.Password = "your-password"

	client, err := dpma.NewClient(config)
	if err != nil {
		panic(err)
	}
	fmt.Println(client != nil)
	// Output: true
}

func ExampleValidatePatentQuery() {
	// A valid query against the patent field codes returns nil.
	fmt.Println(dpma.ValidatePatentQuery("TI=Elektrofahrzeug AND INH=Siemens"))
	// An unknown field code is rejected.
	fmt.Println(dpma.ValidatePatentQuery("MARKE=test") != nil)
	// Output:
	// <nil>
	// true
}

func ExampleFormatPublicationWeek() {
	pubWeek, err := dpma.FormatPublicationWeek(2024, 45)
	if err != nil {
		panic(err)
	}
	fmt.Println(pubWeek)
	// Output: 202445
}

func ExampleFormatDate() {
	date := time.Date(2024, 10, 23, 0, 0, 0, 0, time.UTC)
	fmt.Println(dpma.FormatDate(date))
	// Output: 2024-10-23
}
