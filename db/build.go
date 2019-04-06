package db

import (
	util "../util"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

/*
   DeploymentDetails represents the data for the construction of a testnet.
*/
type DeploymentDetails struct {
	/*
	   Servers: The ids of the servers to build on
	*/
	Servers []int `json:"servers"`
	/*
	   Blockchain: The blockchain to build.
	*/
	Blockchain string `json:"blockchain"`
	/*
	   Nodes:  The number of nodes to build
	*/
	Nodes int `json:"nodes"`
	/*
	   Image: The docker image to build off of
	*/
	Image string `json:"image"`
	/*
	   Params: The blockchain specific parameters
	*/
	Params map[string]interface{} `json:"params"`
	/*
	   Resources: The resources per node
	*/
	Resources []util.Resources `json:"resources"`

	Environments []map[string]string `json:"environments"`

	Files map[string]string `json:"files"`

	Logs map[string]string `json:"logs"`

	Extras map[string]interface{} `json:"extras"`
}

/*
   Get all of the builds done by a user
*/
func GetAllBuilds() ([]DeploymentDetails, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT servers,blockchain,nodes,image,params,resources,environment FROM %s", BuildsTable))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()
	builds := []DeploymentDetails{}

	for rows.Next() {
		var build DeploymentDetails
		var servers []byte
		var params []byte
		var resources []byte
		var environment []byte

		err = rows.Scan(&servers, &build.Blockchain, &build.Nodes, &build.Image, &params, &resources, &environment)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		err = json.Unmarshal(servers, &build.Servers)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		err = json.Unmarshal(params, &build.Params)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		err = json.Unmarshal(resources, &build.Resources)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		err = json.Unmarshal(environment, &build.Environments)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		builds = append(builds, build)
	}
	return builds, nil
}

/*
   Get the build paramters based off testnet id
*/
func GetBuildByTestnet(id string) (DeploymentDetails, error) {

	row := db.QueryRow(fmt.Sprintf("SELECT servers,blockchain,nodes,image,params,resources,environment FROM %s WHERE testnet = \"%s\"", BuildsTable, id))
	var build DeploymentDetails
	var servers []byte
	var params []byte
	var resources []byte
	var environment []byte

	err := row.Scan(&servers, &build.Blockchain, &build.Nodes, &build.Image, &params, &resources, &environment)
	if err != nil {
		log.Println(err)
		return DeploymentDetails{}, err
	}

	err = json.Unmarshal(servers, &build.Servers)
	if err != nil {
		log.Println(err)
		return DeploymentDetails{}, err
	}

	err = json.Unmarshal(params, &build.Params)
	if err != nil {
		log.Println(err)
		return DeploymentDetails{}, err
	}

	err = json.Unmarshal(resources, &build.Resources)
	if err != nil {
		log.Println(err)
		return DeploymentDetails{}, err
	}

	err = json.Unmarshal(environment, &build.Environments)
	if err != nil {
		log.Println(err)
		return DeploymentDetails{}, err
	}

	return build, nil
}

/*
   Insert a build
*/
func InsertBuild(dd DeploymentDetails, testnetId string) error {

	tx, err := db.Begin()

	if err != nil {
		log.Println(err)
		return err
	}

	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (testnet,servers,blockchain,nodes,image,params,resources,environment) VALUES (?,?,?,?,?,?,?,?)", BuildsTable))

	if err != nil {
		log.Println(err)
		return err
	}

	defer stmt.Close()

	servers, _ := json.Marshal(dd.Servers)
	params, _ := json.Marshal(dd.Params)
	resources, _ := json.Marshal(dd.Resources)
	environment, err := json.Marshal(dd.Environments)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = stmt.Exec(testnetId, string(servers), dd.Blockchain, dd.Nodes, dd.Image, string(params), string(resources), string(environment))

	if err != nil {
		log.Println(err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
