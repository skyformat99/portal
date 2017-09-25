package main

import (
	"log"

	"github.com/lthibault/portal"
	"github.com/lthibault/portal/protocol/pair"
)

func main() {
	prt0 := pair.New(portal.Cfg{})
	prt1 := pair.New(portal.Cfg{})

	if err := prt0.Bind("/test"); err != nil {
		log.Fatal("could not bind portal to addr: ", err)
	}

	if err := prt1.Connect("/test"); err != nil {
		log.Fatal("could not connect: ", err)
	}

	prt0.Send("0 to 1")
	log.Println(prt1.Recv())

	prt1.Send("1 to 0")
	log.Println(prt0.Recv())

	log.Println("OK")
	prt0.Close()
	prt1.Close()
}
