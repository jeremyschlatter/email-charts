package app

import (
	"log"
)

func checkError(err error) {
	if err != nil {
		log.Panicf("Letting this request die because of the following error:\n%s", err.Error())
	}
}

var exitFuncs []func()

func RunAtExit(fn func()) {
	exitFuncs = append(exitFuncs, fn)
}

func CallExitFuncs() {
	log.Println("Running exit functions.")
	for _, fn := range exitFuncs {
		fn()
	}
}
