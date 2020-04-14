package main

import (
	"fmt"
	"github.com/tjbrockmeyer/cypher"
	_ "github.com/tjbrockmeyer/cypher/neohttp"
	"log"
)

func main() {
	var err error
	var db cypher.DB
	checkErr := func() {
		if err != nil {
			fmt.Printf("%+v\n\n", err)
			panic(err)
		}
	}
	// cypher.Debug = true
	db, err = cypher.Connect("neohttp", "http://localhost:7474", "neo4j", "neo4j", "graph")
	checkErr()
	// 	for i := 0; i < 99; i++ {
	// 		params := map[string]interface{}{
	// 			"name": fmt.Sprint("abc", i),
	// 		}
	// 		err = db.Run(`
	// CREATE (:Label {name: $name})`, params).Results(func(result cypher.Result) (bool, error) {
	// 			_, err := result.Consume()
	// 			return true, err
	// 		})
	// 	}
	// 	checkErr()

	// 	err = db.Run(`MATCH (x) RETURN properties(x)`, nil).Results(func(result cypher.Result) (bool, error) {
	// 		return true, result.Rows(func(row cypher.Row) (bool, error) {
	// 			log.Printf("%+v\n", row)
	// 			return true, nil
	// 		})
	// 	})
	// 	checkErr()

	res := db.RunMany(
		`MATCH (x {name: $name}) RETURN x.name`, map[string]interface{}{"name": "abc90"},
		`MATCH (x) WHERE x.name CONTAINS $search RETURN x`, map[string]interface{}{"search": "8"})
	for res.NextResult() {
		r := res.GetResult()
		for r.NextRow() {
			log.Println(r.GetRow().GetAt(0))
		}
		err = r.Err()
		checkErr()
	}
	err = res.Err()
	checkErr()

	rows, err := cypher.Collect(db.Run(`MATCH (x) WHERE x.name CONTAINS $search RETURN x`, map[string]interface{}{"search": "1"}))
	checkErr()
	for _, r := range rows {
		log.Println(r.GetAt(0))
	}

	// params := map[string]interface{}{"name": "abc123"}
	// err = db.Run(`MATCH (x {name: $name}) DELETE x`, params).Consume()
	// checkErr()
}
