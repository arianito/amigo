package amigo

import (
	"bufio"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var dialect = "mysql"
var migrationTable = "amigo_migrations"

func SetTable(name string) {
	migrationTable = name
}

func SetDialect(name string) {
	dialect = name
}
func readFile(name string) (string, string) {
	file, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var up, down string
	flg := 0
	for scanner.Scan() {
		txt := strings.Trim(scanner.Text(), " \n")
		if strings.Contains(txt, "migrate_up") {
			flg = 0
		} else if strings.Contains(txt, "migrate_down") {
			flg = 1
		} else {
			if flg == 0 {
				up += txt + "\n"
			} else if flg == 1 {
				down += txt + "\n"
			}
		}
	}
	up = strings.Trim(up, " \n")
	down = strings.Trim(down, " \n")
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return up, down
}

func Transact(txFunc func() error) error {
	return txFunc()
}
func Migrate(path, action, option string, db *sql.DB) {
	switch action {
	case "create":
		propertyName := option
		if propertyName == "" {
			propertyName = "some"
		}
		propertyName = dashify(propertyName)
		isTable := strings.Contains(propertyName, "table")

		var data []byte
		if isTable {
			data = []byte(`/* -- migrate_up -- */
create table TABLE_NAME(
	id int auto_increment,
	constraint primary key (id)
);
/* -- migrate_down -- */
drop table TABLE_NAME;`)
		} else  {
			data = []byte(`/* -- migrate_up -- */
/* -- migrate_down -- */`)
		}
		_ = ioutil.WriteFile(
			fmt.Sprintf("%s/%s_create_%s.sql", path, time.Now().UTC().Format("2006_01_02_15_04_05"), propertyName),
			data,
			0644,
		)
		return
	case "up":
		err := Transact(func() error {

			createMigrationTable(db)
			saved := retrieveMigratedList(db)
			files, err := ioutil.ReadDir(path)
			if err != nil {
				return err
			}
			var names []string
			for _, f := range files {
				nm := f.Name()
				if !f.IsDir() && nm[0] != '.' {
					names = append(names, f.Name())
				}
			}
			sort.Strings(names)
			savedLen := len(saved)

			for i, name := range names {
				if i < savedLen && saved[i] == name {
					log.Println("> already migrated: ", name)
				} else {
					up, _ := readFile(path + "/" + name)
					if up != "" {
						err = exec(db, up)
						if err != nil {
							return err
						}
					}
					err := addMigration(db, name, i)
					if  err != nil {
						return err
					}
					log.Println(">> succeed : ", name)
				}
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
		return
	case "down":
		err := Transact(func() error {
			createMigrationTable(db)
			saved := retrieveMigratedList(db)
			savedLen := len(saved)
			for i := savedLen - 1; i >= 0; i-- {
				name := saved[i]
				_, down := readFile(path + "/" + name)
				if down != "" {
					if err := exec(db, down); err != nil {
						return err
					}
				}
				if err := removeMigration(db, i); err != nil {
					return err
				}
				log.Println(">> rolled-back : ", name)
			}
			return nil
		})

		if err != nil {
			log.Fatal(err)
		}
		return
	case "rollback":
		stepsArg := option
		if stepsArg == "" {
			stepsArg = "1"
		}
		steps, _ := strconv.Atoi(stepsArg)
		err := Transact(func() error {
			createMigrationTable(db)
			saved := retrieveMigratedList(db)
			savedLen := len(saved)
			k := 0
			for i := savedLen - 1; i >= 0; i-- {
				name := saved[i]
				_, down := readFile(path + "/" + name)
				if down != "" {
					if err := exec(db, down); err != nil {
						return err
					}
				}
				if err := removeMigration(db, i); err != nil {
					return err
				}
				log.Println(">> rolled-back : ", name)
				k++
				if k == steps {
					break
				}
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
		return
	}
}

func dashify(in string) string {
	return strings.ReplaceAll(strings.ToLower(in), " ", "_")
}

func createMigrationTable(db *sql.DB) {
	query := ""
	if dialect == "mysql" {
		query = `create table if not exists `+migrationTable+` (
	id int not null auto_increment,
	name varchar(255) not null,
	priority int not null,
	created_at timestamp not null default current_timestamp,
	primary key (id),
	unique (name)
);`
	}else if dialect == "sqlite3" {
		query = `create table if not exists `+migrationTable+` (
	id integer primary key autoincrement,
	name varchar(255) not null,
	priority int not null,
	created_at timestamp not null default current_timestamp,
	unique (name)
);`
	}
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func retrieveMigratedList(db *sql.DB) []string {
	var names []string
	rows, err := db.Query(`select name from `+migrationTable+` order by priority;`)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			log.Fatal(err)
		}
		names = append(names, name)
	}
	return names
}

func addMigration(db *sql.DB, name string, priority int) error {
	_, err := db.Exec(`insert into `+migrationTable+` (name, priority) values (?, ?);`, name, priority)
	if err != nil {
		return err
	}
	return nil
}
func removeMigration(db *sql.DB, priority int) error {
	_, err := db.Exec(`delete from `+migrationTable+` where priority=?;`, priority)
	if err != nil {
		return err
	}
	return nil
}

func exec(db *sql.DB, query string, args ...interface{}) error {
	_, err := db.Exec(query, args...)
	if err != nil {
		return err
	}
	return nil
}
