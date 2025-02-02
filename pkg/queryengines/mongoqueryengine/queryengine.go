package mongoqueryengine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"slashbase.com/backend/internal/models"
	"slashbase.com/backend/pkg/queryengines/mongoqueryengine/mongoutils"
	"slashbase.com/backend/pkg/queryengines/queryconfig"
	"slashbase.com/backend/pkg/sbsql"
	"slashbase.com/backend/pkg/sshtunnel"
)

type MongoQueryEngine struct {
	openClients map[string]mongoClientInstance
}

func InitMongoQueryEngine() *MongoQueryEngine {
	return &MongoQueryEngine{
		openClients: map[string]mongoClientInstance{},
	}
}

func (mqe *MongoQueryEngine) RunQuery(dbConn *models.DBConnection, query string, config *queryconfig.QueryConfig) (map[string]interface{}, error) {
	port, _ := strconv.Atoi(string(dbConn.DBPort))
	if dbConn.UseSSH != models.DBUSESSH_NONE {
		remoteHost := string(dbConn.DBHost)
		if remoteHost == "" {
			remoteHost = "localhost"
		}
		sshTun := sshtunnel.GetSSHTunnel(dbConn.ID, dbConn.UseSSH,
			string(dbConn.SSHHost), remoteHost, port, string(dbConn.SSHUser),
			string(dbConn.SSHPassword), string(dbConn.SSHKeyFile),
		)
		dbConn.DBHost = sbsql.CryptedData("localhost")
		dbConn.DBPort = sbsql.CryptedData(fmt.Sprintf("%d", sshTun.GetLocalEndpoint().Port))
	}
	port, _ = strconv.Atoi(string(dbConn.DBPort))
	conn, err := mqe.getConnection(dbConn.ID, string(dbConn.DBScheme), string(dbConn.DBHost), uint16(port), string(dbConn.DBUser), string(dbConn.DBPassword))
	if err != nil {
		return nil, err
	}
	db := conn.Database(string(dbConn.DBName))
	queryType := mongoutils.GetMongoQueryType(query)

	queryTypeRead := mongoutils.IsQueryTypeRead(queryType.QueryType)
	if !queryTypeRead && config.ReadOnly {
		return nil, errors.New("not allowed run this query")
	}

	if queryType.QueryType == mongoutils.QUERY_FINDONE {
		result := db.Collection(queryType.CollectionName).
			FindOne(context.Background(), queryType.Args[0])
		if result.Err() != nil {
			return nil, result.Err()
		}
		keys, data := mongoutils.MongoSingleResultToJson(result)
		return map[string]interface{}{
			"keys": keys,
			"data": data,
		}, nil
	} else if queryType.QueryType == mongoutils.QUERY_FIND {
		cursor, err := db.Collection(queryType.CollectionName).
			Find(context.Background(), queryType.Args[0], &options.FindOptions{Limit: queryType.Limit, Skip: queryType.Skip, Sort: queryType.Sort})
		if err != nil {
			return nil, err
		}
		defer cursor.Close(context.Background())
		keys, data := mongoutils.MongoCursorToJson(cursor)
		if config.CreateLogFn != nil {
			config.CreateLogFn(query)
		}
		return map[string]interface{}{
			"keys": keys,
			"data": data,
		}, nil
	} else if queryType.QueryType == mongoutils.QUERY_INSERTONE {
		result, err := db.Collection(queryType.CollectionName).
			InsertOne(context.Background(), queryType.Args[0])
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"keys": []string{"insertedId"},
			"data": []map[string]interface{}{
				{
					"insertedId": result.InsertedID,
				},
			},
		}, nil
	} else if queryType.QueryType == mongoutils.QUERY_INSERT {
		result, err := db.Collection(queryType.CollectionName).
			InsertMany(context.Background(), queryType.Args[0].(bson.A))
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"keys": []string{"insertedIDs"},
			"data": []map[string]interface{}{
				{
					"insertedIDs": result.InsertedIDs,
				},
			},
		}, nil
	} else if queryType.QueryType == mongoutils.QUERY_DELETEONE {
		result, err := db.Collection(queryType.CollectionName).
			DeleteOne(context.Background(), queryType.Args[0])
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"keys": []string{"deletedCount"},
			"data": []map[string]interface{}{
				{
					"deletedCount": result.DeletedCount,
				},
			},
		}, nil
	} else if queryType.QueryType == mongoutils.QUERY_DELETEMANY {
		result, err := db.Collection(queryType.CollectionName).
			DeleteMany(context.Background(), queryType.Args[0].(bson.D))
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"keys": []string{"deletedCount"},
			"data": []map[string]interface{}{
				{
					"deletedCount": result.DeletedCount,
				},
			},
		}, nil
	} else if queryType.QueryType == mongoutils.QUERY_UPDATEONE {
		result, err := db.Collection(queryType.CollectionName).
			UpdateOne(context.Background(), queryType.Args[0], queryType.Args[1])
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"keys": []string{"updatedCount", "upsertedCount"},
			"data": []map[string]interface{}{
				{
					"updatedCount":  result.ModifiedCount,
					"upsertedCount": result.UpsertedCount,
				},
			},
		}, nil
	} else if queryType.QueryType == mongoutils.QUERY_UPDATEMANY {
		result, err := db.Collection(queryType.CollectionName).
			UpdateMany(context.Background(), queryType.Args[0], queryType.Args[1])
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"keys": []string{"updatedCount", "upsertedCount"},
			"data": []map[string]interface{}{
				{
					"updatedCount":  result.ModifiedCount,
					"upsertedCount": result.UpsertedCount,
				},
			},
		}, nil
	} else if queryType.QueryType == mongoutils.QUERY_RUNCMD {
		result := db.RunCommand(context.Background(), queryType.Args[0])
		if result.Err() != nil {
			return nil, result.Err()
		}
		keys, data := mongoutils.MongoSingleResultToJson(result)
		if config.CreateLogFn != nil {
			config.CreateLogFn(query)
		}
		return map[string]interface{}{
			"keys": keys,
			"data": data,
		}, nil
	} else if queryType.QueryType == mongoutils.QUERY_GETINDEXES {
		cursor, err := db.RunCommandCursor(context.Background(), bson.M{
			"listIndexes": queryType.CollectionName,
		})
		if err != nil {
			return nil, err
		}
		defer cursor.Close(context.Background())
		keys, data := mongoutils.MongoCursorToJson(cursor)
		if config.CreateLogFn != nil {
			config.CreateLogFn(query)
		}
		return map[string]interface{}{
			"keys": keys,
			"data": data,
		}, nil
	} else if queryType.QueryType == mongoutils.QUERY_LISTCOLLECTIONS {
		list, err := db.ListCollectionNames(context.Background(), queryType.Args[0])
		if err != nil {
			return nil, err
		}
		if config.CreateLogFn != nil {
			config.CreateLogFn(query)
		}
		data := []map[string]interface{}{}
		for _, name := range list {
			data = append(data, map[string]interface{}{"collectionName": name})
		}
		return map[string]interface{}{
			"keys": []string{"collectionName"},
			"data": data,
		}, nil
	} else if queryType.QueryType == mongoutils.QUERY_COUNT {
		count, err := db.Collection(queryType.CollectionName).
			CountDocuments(context.Background(), queryType.Args[0])
		if err != nil {
			return nil, err
		}
		if config.CreateLogFn != nil {
			config.CreateLogFn(query)
		}
		return map[string]interface{}{
			"keys": []string{"count"},
			"data": []map[string]interface{}{
				{
					"count": count,
				},
			},
		}, nil
	} else if queryType.QueryType == mongoutils.QUERY_AGGREGATE {
		cursor, err := db.Collection(queryType.CollectionName).
			Aggregate(context.Background(), queryType.Args[0])
		if err != nil {
			return nil, err
		}
		defer cursor.Close(context.Background())
		keys, data := mongoutils.MongoCursorToJson(cursor)
		if config.CreateLogFn != nil {
			config.CreateLogFn(query)
		}
		return map[string]interface{}{
			"keys": keys,
			"data": data,
		}, nil
	}
	return nil, errors.New("unknown query")
}

func (mqe *MongoQueryEngine) TestConnection(dbConn *models.DBConnection, config *queryconfig.QueryConfig) bool {
	query := "db.runCommand({ping: 1})"
	data, err := mqe.RunQuery(dbConn, query, config)
	if err != nil {
		return false
	}
	test := data["data"].([]map[string]interface{})[0]["ok"].(float64)
	return test == 1
}

func (mqe *MongoQueryEngine) GetDataModels(dbConn *models.DBConnection, config *queryconfig.QueryConfig) ([]map[string]interface{}, error) {
	query := "db.getCollectionNames()"
	data, err := mqe.RunQuery(dbConn, query, config)
	if err != nil {
		return nil, err
	}
	rdata := data["data"].([]map[string]interface{})
	return rdata, nil
}

func (mqe *MongoQueryEngine) GetSingleDataModelFields(dbConn *models.DBConnection, name string, config *queryconfig.QueryConfig) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`db.%s.aggregate([{$sample: {size: 1000}}])`, name)
	data, err := mqe.RunQuery(dbConn, query, config)
	if err != nil {
		return nil, err
	}
	returnedKeys := data["keys"].([]string)
	returnedData := data["data"].([]map[string]interface{})
	return mongoutils.AnalyseFieldsSchema(returnedKeys, returnedData), err
}

func (mqe *MongoQueryEngine) GetSingleDataModelIndexes(dbConn *models.DBConnection, name string, config *queryconfig.QueryConfig) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`db.%s.getIndexes()`, name)
	data, err := mqe.RunQuery(dbConn, query, config)
	if err != nil {
		return nil, err
	}
	returnedData := data["data"].([]map[string]interface{})
	return mongoutils.GetCollectionIndexes(returnedData), err
}

func (mqe *MongoQueryEngine) AddSingleDataModelKey(dbConn *models.DBConnection, schema, name, columnName, dataType string) (map[string]interface{}, error) {
	return nil, errors.New("not supported yet")
}

func (mqe *MongoQueryEngine) DeleteSingleDataModelKey(dbConn *models.DBConnection, schema, name, columnName string, config *queryconfig.QueryConfig) (map[string]interface{}, error) {
	query := fmt.Sprintf(`db.%s.updateMany({}, {$unset: {%s: ""}})`, name, columnName)
	data, err := mqe.RunQuery(dbConn, query, config)
	if err != nil {
		return nil, err
	}
	return data, err
}

func (mqe *MongoQueryEngine) GetData(dbConn *models.DBConnection, name string, limit int, offset int64, fetchCount bool, filter []string, sort []string, config *queryconfig.QueryConfig) (map[string]interface{}, error) {
	query := fmt.Sprintf(`db.%s.find().limit(%d).skip(%d)`, name, limit, offset)
	countQuery := fmt.Sprintf(`db.%s.count()`, name)
	if len(filter) == 1 && strings.HasPrefix(filter[0], "{") && strings.HasSuffix(filter[0], "}") {
		query = fmt.Sprintf(`db.%s.find(%s).limit(%d).skip(%d)`, name, filter[0], limit, offset)
		countQuery = fmt.Sprintf(`db.%s.count(%s)`, name, filter[0])
	}
	if len(sort) == 1 {
		query = query + fmt.Sprintf(`.sort(%s)`, sort[0])
	}
	data, err := mqe.RunQuery(dbConn, query, config)
	if err != nil {
		return nil, err
	}
	if fetchCount {
		countData, err := mqe.RunQuery(dbConn, countQuery, config)
		if err != nil {
			return nil, err
		}
		data["count"] = countData["data"].([]map[string]interface{})[0]["count"]
	}
	return data, err
}

func (mqe *MongoQueryEngine) UpdateSingleData(dbConn *models.DBConnection, name string, underscoreID string, documentData string, config *queryconfig.QueryConfig) (map[string]interface{}, error) {
	query := fmt.Sprintf(`db.%s.updateOne({_id: ObjectId("%s")}, {$set: %s } )`, name, underscoreID, documentData)
	data, err := mqe.RunQuery(dbConn, query, config)
	if err != nil {
		return nil, err
	}
	updatedCount := data["data"].([]map[string]interface{})[0]["updatedCount"]
	data = map[string]interface{}{
		"updatedCount": updatedCount,
	}
	return data, err
}

func (mqe *MongoQueryEngine) AddData(dbConn *models.DBConnection, schema string, name string, data map[string]interface{}, config *queryconfig.QueryConfig) (map[string]interface{}, error) {
	dataStr, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`db.%s.insertOne(%s)`, name, string(dataStr))
	rData, err := mqe.RunQuery(dbConn, query, config)
	if err != nil {
		return nil, err
	}
	insertedID := rData["data"].([]map[string]interface{})[0]["insertedId"]
	rData = map[string]interface{}{
		"insertedId": insertedID,
	}
	return rData, err
}

func (mqe *MongoQueryEngine) DeleteData(dbConn *models.DBConnection, name string, underscoreIds []string, config *queryconfig.QueryConfig) (map[string]interface{}, error) {
	for i, id := range underscoreIds {
		underscoreIds[i] = fmt.Sprintf(`ObjectId("%s")`, id)
	}
	underscoreIdsStr := strings.Join(underscoreIds, ", ")
	query := fmt.Sprintf(`db.%s.deleteMany({ _id : { "$in" : [%s] }})`, name, underscoreIdsStr)
	fmt.Println(query)
	return mqe.RunQuery(dbConn, query, config)
}
