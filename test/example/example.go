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

	_, err = db.RunMany(
		`MATCH (x {name: $name}) RETURN x.name`, map[string]interface{}{"name": "abc90"},
		`MATCH (x) WHERE x.name CONTAINS $search RETURN x`, map[string]interface{}{"search": "8"}).
		Results(func(result cypher.Result) (interface{}, error) {
			return result.Rows(func(row cypher.Row) (interface{}, error) {
				log.Println(row.GetAt(0))
				return row.GetAt(0), nil
			})
		})
	checkErr()

	// params := map[string]interface{}{"name": "abc123"}
	// err = db.Run(`MATCH (x {name: $name}) DELETE x`, params).Consume()
	// checkErr()
}
