package main

import (
	"log"

	"github.com/lthibault/portal/protocol/pair"
)

func main() {
	prt0 := pair.New()
	defer prt0.Close()

	prt1 := pair.New()
	defer prt1.Close()

	if err := prt0.Bind("/test"); err != nil {
		log.Fatal(err)
	}

	if err := prt1.Connect("/test"); err != nil {
		log.Fatal(err)
	}

	prt0.Send("oh hai")
	prt1.Recv()

	log.Println("OK")
}
