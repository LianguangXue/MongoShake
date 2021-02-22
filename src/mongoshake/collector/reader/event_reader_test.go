package sourceReader

import (
	"testing"
	"fmt"

	"mongoshake/unit_test_common"
	"mongoshake/common"
	"mongoshake/collector/configure"

	"github.com/vinllen/mongo-go-driver/bson"
	bsonMgo "github.com/vinllen/mgo/bson"
	"github.com/stretchr/testify/assert"
	"time"
	"mongoshake/oplog"
	"sync"
)

const (
	testMongoAddressCs = unit_test_common.TestUrlServerlessTenant
)

func TestEventReader(t *testing.T) {
	// test EventReader

	utils.InitialLogger("", "", "all", true, true)

	var nr int
	// normal: test 2 db
	{
		fmt.Printf("TestEventReader case %d.\n", nr)
		nr++

		conn, err := utils.NewMongoCommunityConn(testMongoAddressCs, "primary", true, "", "")
		assert.Equal(t, nil, err, "should be equal")
		err = conn.Client.Database("db1").Drop(nil)
		assert.Equal(t, nil, err, "should be equal")
		err = conn.Client.Database("db2").Drop(nil)
		assert.Equal(t, nil, err, "should be equal")

		// test two databases
		_, err = conn.Client.Database("db1").Collection("c1").InsertOne(nil, bson.M{"yy": 1})
		assert.Equal(t, nil, err, "should be equal")
		_, err = conn.Client.Database("db2").Collection("c1").InsertOne(nil, bson.M{"yy": 1})
		assert.Equal(t, nil, err, "should be equal")

		er := NewEventReader(testMongoAddressCs, "ut_event_reader")
		conf.Options.SpecialSourceDBFlag = utils.VarSpecialSourceDBFlagAliyunServerless
		er.StartFetcher()

		flag := false
		startIndex := 0
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				in, err := er.Next()
				if err == TimeoutError {
					time.Sleep(2 * time.Second)
					fmt.Printf("timeout, resumeToken: %v\n", er.client.ResumeToken())
					continue
				}
				assert.Equal(t, nil, err, "should be equal")

				var val oplog.Event
				err = bsonMgo.Unmarshal(in, &val)
				// err = bson.Unmarshal(in, &val)
				fmt.Printf("%v\n", val)
				assert.Equal(t, nil, err, "should be equal")

				if !flag {
					if val.Ns["db"] != "db1" && val.Ns["db"] != "db2" && val.Ns["coll"] != "c1" {
						fmt.Printf("timeout because of unexpect ns[%v] event[%v]\n", val.Ns, val)
						time.Sleep(2 * time.Second)
						continue
					} else {
						flag = true
						m, _ := oplog.ConvertBsonD2M(val.FullDocument)
						startIndex = m["x"].(int)
						if startIndex < 1 {
							fmt.Printf("timeout because of unexpect x[%v] fullDocument[%v]\n", startIndex, m)
							time.Sleep(2 * time.Second)
							continue
						}
					}
				}

				m, _ := oplog.ConvertBsonD2M(val.FullDocument)
				innerVal := m["x"].(int)
				assert.Equal(t, startIndex, innerVal, "should be equal")
				fmt.Printf("match x[%v]\n", startIndex)
				startIndex++
				if startIndex >= 100 {
					break
				}
			}
		}()

		time.Sleep(10 * time.Second)
		for i := 1; i <= 100; i ++ {
			db := "db1"
			if i%2 == 0 {
				db = "db2"
			}
			_, err = conn.Client.Database(db).Collection("c1").InsertOne(nil, bson.M{"x": i})
			assert.Equal(t, nil, err, "should be equal")
		}
		wg.Wait()
	}

	// normal: test 1 db
	{
		fmt.Printf("TestEventReader case %d.\n", nr)
		nr++

		conn, err := utils.NewMongoCommunityConn(testMongoAddressCs, "primary", true, "", "")

		// drop all databases
		dbs, err := conn.Client.ListDatabaseNames(nil, bson.M{})
		assert.Equal(t, nil, err, "should be equal")
		for _, db := range dbs {
			err = conn.Client.Database(db).Drop(nil)
			assert.Equal(t, nil, err, "should be equal")
		}

		// test one database
		_, err = conn.Client.Database("db1").Collection("c1").InsertOne(nil, bson.M{"yy": 1})
		assert.Equal(t, nil, err, "should be equal")

		er := NewEventReader(testMongoAddressCs, "ut_event_reader")
		conf.Options.SpecialSourceDBFlag = utils.VarSpecialSourceDBFlagAliyunServerless
		er.StartFetcher()

		flag := false
		startIndex := 0
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				in, err := er.Next()
				if err == TimeoutError {
					time.Sleep(2 * time.Second)
					fmt.Printf("timeout, resumeToken: %v\n", er.client.ResumeToken())
					continue
				}
				assert.Equal(t, nil, err, "should be equal")

				var val oplog.Event
				err = bsonMgo.Unmarshal(in, &val)
				// err = bson.Unmarshal(in, &val)
				fmt.Printf("%v\n", val)
				assert.Equal(t, nil, err, "should be equal")

				if !flag {
					if val.Ns["db"] != "db1" && val.Ns["coll"] != "c1" {
						fmt.Printf("timeout because of unexpect ns[%v] event[%v]\n", val.Ns, val)
						time.Sleep(2 * time.Second)
						continue
					} else {
						flag = true
						m, _ := oplog.ConvertBsonD2M(val.FullDocument)
						startIndex = m["x"].(int)
						if startIndex < 1 {
							fmt.Printf("timeout because of unexpect x[%v] fullDocument[%v]\n", startIndex, m)
							time.Sleep(2 * time.Second)
							continue
						}
					}
				}

				m, _ := oplog.ConvertBsonD2M(val.FullDocument)
				innerVal := m["x"].(int)
				assert.Equal(t, startIndex, innerVal, "should be equal")
				fmt.Printf("match x[%v]\n", startIndex)
				startIndex++
				if startIndex >= 100 {
					break
				}
			}
		}()

		time.Sleep(10 * time.Second)
		for i := 1; i <= 100; i ++ {
			db := "db1"
			_, err = conn.Client.Database(db).Collection("c1").InsertOne(nil, bson.M{"x": i})
			assert.Equal(t, nil, err, "should be equal")
		}
		wg.Wait()
	}
}