package main

import (
	"log"

	"github.com/SentimensRG/sigctx"
	"github.com/lthibault/portal/protocol/pair"
)

func main() {
	prt0 := pair.New()
	defer prt0.Close()

	prt1 := pair.New()
	defer prt1.Close()

	if err := prt0.Bind("/test"); err != nil {
		log.Fatal("could not bind portal to addr: ", err)
	}

	if err := prt1.Connect("/test"); err != nil {
		log.Fatal("could not connect: ", err)
	}

	prt0.Send("payload")
	log.Println("SENT")

	log.Println(prt1.Recv())

	log.Println("OK")
	<-sigctx.New().Done()
}
