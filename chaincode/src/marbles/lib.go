/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"fmt"
)

// ============================================================================================================================
// Get Marble - get a marble asset from ledger
// ============================================================================================================================
func get_marble(stub shim.ChaincodeStubInterface, id string) (Marble, error) {
	var marble Marble
	marbleAsBytes, err := stub.GetState(id)                  //getState retreives a key/value from the ledger
	if err != nil {                                          //this seems to always succeed, even if key didn't exist
		return marble, errors.New("Failed to find marble - " + id)
	}
	json.Unmarshal(marbleAsBytes, &marble)                   //un stringify it aka JSON.parse()

	if marble.Id != id {                                     //test if marble is actually here or just nil
		return marble, errors.New("Marble does not exist - " + id)
	}

	return marble, nil
}

// ============================================================================================================================
// Get User - get the owner asset from ledger
// ============================================================================================================================
func get_user(stub shim.ChaincodeStubInterface, id string) (User, error) {


	var owner User
	ownerAsBytes, err := stub.GetState(id)                     //getState retreives a key/value from the ledger
	if err != nil {                                            //this seems to always succeed, even if key didn't exist
		return owner, errors.New("Failed to get owner - " + id)
	}
	json.Unmarshal(ownerAsBytes, &owner)                       //un stringify it aka JSON.parse()

	if len(owner.Username) == 0 {                              //test if owner is actually here or just nil
		return owner, errors.New("User does not exist - " + id + ", '" + owner.Username + "' '" + owner.Company + "'")
	}
	
	return owner, nil
}

// ========================================================
// Input Sanitation - dumb input checking, look for empty strings
// ========================================================
func sanitize_arguments(strs []string) error{
	for i, val:= range strs {
		if len(val) <= 0 {
			return errors.New("Argument " + strconv.Itoa(i) + " must be a non-empty string")
		}
		if len(val) > 32 {
			return errors.New("Argument " + strconv.Itoa(i) + " must be <= 32 characters")
		}
	}
	return nil
}



func getMarblesById(stub shim.ChaincodeStubInterface,id string)( marble Marble){

	var err error
	fmt.Println("starting read")
	valAsbytes, err := stub.GetState(id)           //get the var from ledger
	if err != nil {
		fmt.Sprintf("{\"Error\":\"Failed to get state for " + id + "\"}")
	}

	json.Unmarshal(valAsbytes, &marble)
	return marble
}
func getAllMarbles(stub shim.ChaincodeStubInterface)(marbles []Marble,err error){

	//var marble []Marble

	// ---- Get All Marbles ---- //
	resultsIterator, err := stub.GetStateByRange("m0", "m9999999999999999999")
	if err != nil {
		fmt.Println("get All marbles error !")

	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		aKeyValue, err := resultsIterator.Next()
		if err != nil {
			return marbles,err
		}
		queryKeyAsStr := aKeyValue.Key
		queryValAsBytes := aKeyValue.Value
		fmt.Println("on marble id - ", queryKeyAsStr)
		var marble Marble
		json.Unmarshal(queryValAsBytes, &marble)                  //un stringify it aka JSON.parse()
		marbles = append(marbles, marble)   //add this marble to the list
	}
	fmt.Println("marble array - ", marbles)
	return marbles,nil
}