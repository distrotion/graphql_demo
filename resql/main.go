package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
)

type Foo struct {
	Name string
}

var FieldFooType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Foo",
	Fields: graphql.Fields{
		"name": &graphql.Field{Type: graphql.String},
	},
})

type Bar struct {
	Name string
}

var FieldBarType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Bar",
	Fields: graphql.Fields{
		"name": &graphql.Field{Type: graphql.String},
	},
})

// QueryType fields: `concurrentFieldFoo` and `concurrentFieldBar` are resolved
// concurrently because they belong to the same field-level and their `Resolve`
// function returns a function (thunk).
var QueryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"concurrentFieldFoo": &graphql.Field{
			Type: FieldFooType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				var foo = Foo{Name: "Foo's name"}
				return func() (interface{}, error) {
					return &foo, nil
				}, nil
			},
		},
		"concurrentFieldBar": &graphql.Field{
			Type: FieldBarType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				type result struct {
					data interface{}
					err  error
				}
				ch := make(chan *result, 1)
				go func() {
					defer close(ch)
					bar := &Bar{Name: "hello PARIN"}
					ch <- &result{data: bar, err: nil}
				}()
				return func() (interface{}, error) {
					r := <-ch
					return r.data, r.err
				}, nil
			},
		},
	},
})

type data struct {
	Query string `bson:"query,omitempty"`
}

func main() {
	r := gin.Default()

	r.POST("/", func(c *gin.Context) {

		var input data
		//var input interface{}
		//var input string
		c.ShouldBind(&input)

		//-----------------------------------------
		schema, err := graphql.NewSchema(graphql.SchemaConfig{
			Query: QueryType,
		})
		if err != nil {
			log.Fatal(err)
		}
		// query := `
		// 	query { concurrentFieldFoo { name } concurrentFieldBar {name} }`

		fmt.Println("My input =", input.Query)
		query := input.Query

		result := graphql.Do(graphql.Params{
			RequestString: query,
			Schema:        schema,
		})
		// b, err := json.Marshal(result)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// fmt.Printf("%s", b)

		//=========================================

		c.JSON(200, result)
	})

	r.Run(":7500")
}
