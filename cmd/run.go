package kenginecmd

import (
	"flag"
	"log"

	"github.com/khulnasoft/kengine"
)

// Main executes the main function of the kengine command.
func Main() {
	flag.Parse()

	err := kengine.StartAdmin(*listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer kengine.StopAdmin()

	log.Println("Kengine 2 admin endpoint listening on", *listenAddr)

	select {}
}

// TODO: for dev only
var listenAddr = flag.String("listen", ":1234", "The admin endpoint listener address")
