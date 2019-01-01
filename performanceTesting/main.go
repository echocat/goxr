package main

import (
	"flag"
	"fmt"
	"github.com/c2h5oh/datasize"
)

var (
	filesFlag = append(staticFiles, temporaryFileBy(1*datasize.MB))
)

func init() {
	flag.Var(&filesFlag, "files", fmt.Sprintf(`Could be either one of the static defined files from %s
    or a unit definition like 1K, 1M, ...`, filesDirectory))
}

func main() {
	flag.Parse()

	fmt.Println(filesFlag)
	filesFlag.ensure()
}
