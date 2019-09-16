package cmd

import (
	"flag"
	"fmt"
	"io/ioutil"
	"time"
)

func Init() {
	name := flag.String("name", "hello", "migration name")
	flag.Parse()
	typ := flag.Arg(0)

	switch typ {
	case "table":
		_ = ioutil.WriteFile(fmt.Sprintf("%s_create_table_%s.sql", time.Now().UTC().Format("2006_01_02_15_04_05"), *name), []byte(`create table HELLO();`), 0644)
		break
	}
}
