package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"reflect"
	"time"

	"github.com/olivere/elastic"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/colinzuo/tunip/logp"
	"github.com/colinzuo/tunip/logp/configure"
)

// Tweet is a structure used for serializing/deserializing data in Elasticsearch
type Tweet struct {
	User     string                `json:"user"`
	Message  string                `json:"message"`
	Retweets int                   `json:"retweets"`
	Image    string                `json:"image,omitempty"`
	Created  time.Time             `json:"created,omitempty"`
	Tags     []string              `json:"tags,omitempty"`
	Location string                `json:"location,omitempty"`
	Suggest  *elastic.SuggestField `json:"suggest_field,omitempty"`
}

const mapping = `
{
	"settings":{
		"number_of_shards": 1,
		"number_of_replicas": 0
	},
	"mappings":{
		"tweet":{
			"properties":{
				"user":{
					"type":"keyword"
				},
				"message":{
					"type":"text",
					"store": true,
					"fielddata": true
				},
				"image":{
					"type":"keyword"
				},
				"created":{
					"type":"date"
				},
				"tags":{
					"type":"keyword"
				},
				"location":{
					"type":"geo_point"
				},
				"suggest_field":{
					"type":"completion"
				}
			}
		}
	}
}
`

func main() {
	appName := "elastic_demo"

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetConfigName(appName)
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	configure.Logging(appName)
	logger := logp.NewLogger("main")

	ctx := context.Background()
	serverAddr := "http://172.26.0.41:9200"

	client, err := elastic.NewClient(elastic.SetURL(serverAddr),
		elastic.SetErrorLog(logger),
		elastic.SetInfoLog(logger),
		elastic.SetSniff(false))
	if err != nil {
		panic(err)
	}
	defer client.Stop()

	info, code, err := client.Ping(serverAddr).Do(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Elasticsearch returned with code %d and version %s\n",
		code, info.Version.Number)

	esversion, err := client.ElasticsearchVersion(serverAddr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Elasticsearch version %s\n", esversion)

	exists, err := client.IndexExists("twitter").Do(ctx)
	if err != nil {
		panic(err)
	}

	if !exists {
		createIndex, err := client.CreateIndex("twitter").BodyString(mapping).Do(ctx)
		if err != nil {
			panic(err)
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
		}
	}

	// Index a tweet (using JSON serialization)
	tweet1 := Tweet{User: "olivere", Message: "Take Five", Retweets: 0}
	put1, err := client.Index().
		Index("twitter").
		Type("tweet").
		Id("1").
		BodyJson(tweet1).
		Do(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Index tweet %s to index %s, type %s\n", put1.Id,
		put1.Index, put1.Type)

	// Index a second tweet (by string)
	tweet2 := `{"user": "olivere", "message": "It's a Raggy Waltz"}`
	put2, err := client.Index().
		Index("twitter").
		Type("tweet").
		Id("2").
		BodyString(tweet2).
		Do(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Index tweet %s to index %s, type %s\n",
		put2.Id, put2.Index, put2.Type)

	// Get tweet with specified ID
	get1, err := client.Get().
		Index("twitter").
		Type("tweet").
		Id("1").
		Do(ctx)
	if err != nil {
		panic(err)
	}
	if get1.Found {
		fmt.Printf("Got document %s in version %d from index %s, type %s\n",
			get1.Id, *get1.Version, get1.Index, get1.Type)
	}

	// Flush to make sure the documents got written
	_, err = client.Flush().Index("twitter").Do(ctx)
	if err != nil {
		panic(err)
	}

	// Search with a term query
	termQuery := elastic.NewTermQuery("user", "olivere")
	searchResult, err := client.Search().
		Index("twitter").
		Query(termQuery).
		Sort("user", true).
		From(0).Size(10).
		Pretty(true).
		Do(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)

	var ttyp Tweet
	for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
		if t, ok := item.(Tweet); ok {
			fmt.Printf("Tweet by %s: %s\n", t.User, t.Message)
		}
	}
	fmt.Printf("Found a total of %d tweets\n", searchResult.TotalHits())

	if searchResult.Hits.TotalHits > 0 {
		fmt.Printf("Found a total of %d tweets\n", searchResult.Hits.TotalHits)

		for _, hit := range searchResult.Hits.Hits {
			var t Tweet
			err := json.Unmarshal(*hit.Source, &t)
			if err != nil {
				fmt.Printf("Unmarshal failed %v", *hit.Source)
				continue
			}
			fmt.Printf("Tweet by %s: %s\n", t.User, t.Message)
		}
	} else {
		fmt.Print("Found no tweets\n")
	}

	update, err := client.Update().
		Index("twitter").
		Type("tweet").
		Id("1").
		Script(elastic.NewScriptInline("ctx._source.retweets += params.num").
			Lang("painless").Param("num", 1)).
		Upsert(map[string]interface{}{"retweets": 0}).
		Do(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("New version of tweet %q is now %d\n", update.Id, update.Version)

	deleteIndex, err := client.DeleteIndex("twitter").Do(ctx)
	if err != nil {
		panic(err)
	}
	if !deleteIndex.Acknowledged {
		fmt.Print("Not acknowledged")
	}
}
