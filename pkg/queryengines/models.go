package queryengines

import (
	"slashbase.com/backend/internal/models"
)

type DBDataModel struct {
	Name       string             `json:"name"`
	SchemaName string             `json:"schemaName"`
	Fields     []DBDataModelField `json:"fields"`
	Indexes    []DBDataModelIndex `json:"indexes"`
}

type DBDataModelField struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	IsPrimary  bool     `json:"isPrimary"`
	IsNullable bool     `json:"isNullable"`
	Tags       []string `json:"tags"`
}

type DBDataModelIndex struct {
	Name     string `json:"name"`
	IndexDef string `json:"indexDef"`
}

func BuildDBDataModel(dbConn *models.DBConnection, tableData map[string]interface{}) *DBDataModel {
	if dbConn.Type == models.DBTYPE_POSTGRES {
		view := DBDataModel{
			Name:       tableData["0"].(string),
			SchemaName: tableData["1"].(string),
		}
		return &view
	} else if dbConn.Type == models.DBTYPE_MONGO {
		view := DBDataModel{
			Name: tableData["collectionName"].(string),
		}
		return &view
	}
	return nil
}

func BuildDBDataModelField(dbConn *models.DBConnection, fieldData map[string]interface{}) *DBDataModelField {
	if dbConn.Type == models.DBTYPE_POSTGRES {
		view := DBDataModelField{
			Name:       fieldData["name"].(string),
			Type:       fieldData["type"].(string),
			IsNullable: fieldData["isNullable"].(bool),
			IsPrimary:  fieldData["isPrimary"].(bool),
			Tags:       fieldData["tags"].([]string),
		}
		return &view
	} else if dbConn.Type == models.DBTYPE_MONGO {
		view := DBDataModelField{
			Name:       fieldData["name"].(string),
			Type:       fieldData["types"].(string),
			IsNullable: fieldData["isNullable"].(bool),
			IsPrimary:  fieldData["isPrimary"].(bool),
		}
		return &view
	}
	return nil
}

func BuildDBDataModelIndex(dbConn *models.DBConnection, fieldData map[string]interface{}) *DBDataModelIndex {
	if dbConn.Type == models.DBTYPE_POSTGRES {
		view := DBDataModelIndex{
			Name:     fieldData["0"].(string),
			IndexDef: fieldData["1"].(string),
		}
		return &view
	} else if dbConn.Type == models.DBTYPE_MONGO {
		view := DBDataModelIndex{
			Name:     fieldData["name"].(string),
			IndexDef: fieldData["key"].(string),
		}
		return &view
	}
	return nil
}
