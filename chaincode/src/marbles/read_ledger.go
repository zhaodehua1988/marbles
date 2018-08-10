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
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
)

// ============================================================================================================================
// Read - read a generic variable from ledger
//
// Shows Off GetState() - reading a key/value from the ledger
//
// Inputs - Array of strings
//  0
//  key
//  "abc"
// 
// Returns - string
// ============================================================================================================================
func read(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key, jsonResp string
	var err error
	fmt.Println("starting read")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting key of the var to query")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key)           //get the var from ledger
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	}

	fmt.Println("- end read")
	return shim.Success(valAsbytes)                  //send it onward
}

// ============================================================================================================================
// Get everything we need (owners + marbles + companies)
//
// Inputs - none
//
// Returns:
// {
//	"owners": [{
//			"id": "o99999999",
//			"company": "United Marbles"
//			"username": "alice"
//	}],
//	"marbles": [{
//		"id": "m1490898165086",
//		"color": "white",
//		"docType" :"marble",
//		"owner": {
//			"company": "United Marbles"
//			"username": "alice"
//		},
//		"size" : 35
//	}]
// }
// ============================================================================================================================
func read_everything(stub shim.ChaincodeStubInterface,args []string) pb.Response {
	type Everything struct {
		Owners   []User   `json:"owners"`
		Marbles  []Marble `json:"marbles"`
	}
	var everything Everything
	if len(args) > 1 {
		return shim.Error("Incorrect number of arguments. Expecting args num > 1")
	}

	if len(args) == 1{
		companyName := args[0]
		user,err:=getUserByCompany(stub,companyName)
		if err != nil {
			fmt.Println("Failed to find user - " + companyName)
			return shim.Error(err.Error())
		}

		if !user.Enabled{
			fmt.Println("user is disable -"+companyName)
			return shim.Error(err.Error())
		}
		//var needMarbles []Marble
		marbles,err:= getAllMarbles(stub)
		if err != nil{
			fmt.Println("getAllMarblesByUserID err :",err.Error())
			return shim.Error(err.Error())
		}

		marblesNum := len(marbles)
		for i:=0;i<marblesNum;i++{
			if marbles[i].User.Id == user.Id{
				everything.Marbles = append(everything.Marbles, marbles[i])
				continue
			}

			for j:=0;j<4;j++{
				if marbles[i].Check[j].UserID == user.Id{
					//needMarbles = append(needMarbles, marbles[i])
					everything.Marbles = append(everything.Marbles, marbles[i])
					continue
				}
			}
		}


	} else{
		// ---- Get All Marbles ---- //

		everything.Marbles,_= getAllMarbles(stub)
	}

	// ---- Get All Users ---- //
	ownersIterator, err := stub.GetStateByRange("o0", "o9999999999999999999")
	if err != nil {
		return shim.Error(err.Error())
	}
	defer ownersIterator.Close()

	for ownersIterator.HasNext() {
		aKeyValue, err := ownersIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		queryKeyAsStr := aKeyValue.Key
		queryValAsBytes := aKeyValue.Value
		fmt.Println("on owner id - ", queryKeyAsStr)
		var owner User
		json.Unmarshal(queryValAsBytes, &owner)                   //un stringify it aka JSON.parse()

		if owner.Enabled {                                        //only return enabled owners
			everything.Owners = append(everything.Owners, owner)  //add this marble to the list
		}
	}
	fmt.Println("owner array - ", everything.Owners)

	//change to array of bytes
	everythingAsBytes, _ := json.Marshal(everything)              //convert to array of bytes
	return shim.Success(everythingAsBytes)
}

// ============================================================================================================================
// Get history of asset
//
// Shows Off GetHistoryForKey() - reading complete history of a key/value
//
// Inputs - Array of strings
//  0
//  id
//  "m01490985296352SjAyM"
// ============================================================================================================================
func getHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	type AuditHistory struct {
		TxId    string   `json:"txId"`
		Value   Marble   `json:"value"`
	}
	var history []AuditHistory;
	var marble Marble

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	marbleId := args[0]
	fmt.Printf("- start getHistoryForMarble: %s\n", marbleId)

	// Get History
	resultsIterator, err := stub.GetHistoryForKey(marbleId)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		historyData, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		//historyData.Value
		var tx AuditHistory
		tx.TxId = historyData.TxId                     //copy transaction id over
		json.Unmarshal(historyData.Value, &marble)     //un stringify it aka JSON.parse()
		if historyData.Value == nil {                  //marble has been deleted
			var emptyMarble Marble
			tx.Value = emptyMarble                 //copy nil marble
		} else {
			json.Unmarshal(historyData.Value, &marble) //un stringify it aka JSON.parse()
			tx.Value = marble                      //copy marble over
		}
		history = append(history, tx)              //add this tx to the list
	}
	fmt.Printf("- getHistoryForMarble returning:\n%s", history)

	//change to array of bytes
	historyAsBytes, _ := json.Marshal(history)     //convert to array of bytes
	return shim.Success(historyAsBytes)
}

// ============================================================================================================================
// Get history of asset - performs a range query based on the start and end keys provided.
//
// Shows Off GetStateByRange() - reading a multiple key/values from the ledger
//
// Inputs - Array of strings
//       0     ,    1
//   startKey  ,  endKey
//  "marbles1" , "marbles5"
// ============================================================================================================================
func getMarblesByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	startKey := args[0]
	endKey := args[1]

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		aKeyValue, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		queryResultKey := aKeyValue.Key
		queryResultValue := aKeyValue.Value

		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResultKey)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResultValue))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getMarblesByRange queryResult:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

/*
//申请所处的各个阶段
const (
	New = iota     //0
	SuppApply      //供应商申请  supplier apply
	CompanyCheck   //核心企业审核
	BankCheck      //银行审核
	BankLoan       //银行放款
	SuppRecv       //供应商收款
	SuppRepayment  //供应商还款
	EndOf            //包括成功和失败两种情况
)
//申请的状态
const(
	Disable = iota //0
	Wait
	Success
	Failure
)

*/
//根据id查询所有相关的审核
//       0
//    userID
//
//
func  read_allmarble(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	userID := args[0]
	user, err := get_user(stub, userID)
	if err != nil {
		fmt.Println("Failed to find user - " + userID)
		return shim.Error(err.Error())
	}

	if !user.Enabled{
		fmt.Println("user is disable -"+userID)
		return shim.Error(err.Error())
	}
	var needMarbles []Marble
	marbles,err:= getAllMarbles(stub)
	if err != nil{
		fmt.Println("getAllMarblesByUserID err :",err.Error())
		return shim.Error(err.Error())
	}

	marblesNum := len(marbles)
	if marblesNum <=0{
		fmt.Println("There is no marbles")
		return shim.Error("There is no marbles")
	}

	for i:=0;i<marblesNum;i++{
		if marbles[i].User.Id == userID{
			needMarbles = append(needMarbles, marbles[i])
			continue
		}

		for j:=0;j<4;j++{
			if marbles[i].Check[j].UserID == userID{
				needMarbles = append(needMarbles, marbles[i])
				continue
			}
		}
	}
	marblesAsBytes, _:= json.Marshal(needMarbles)
	return shim.Success(marblesAsBytes)

}

/*
//申请所处的各个阶段
const (
	New = iota     //0
	SuppApply      //供应商申请  supplier apply
	CompanyCheck   //核心企业审核
	BankCheck      //银行审核
	BankLoan       //银行放款
	SuppRecv       //供应商收款
	SuppRepayment  //供应商还款
	EndOf            //包括成功和失败两种情况
)
//申请的状态
const(
	Disable = iota //0
	Wait
	Success
	Failure
)

*/
//根据userId查询需要审核的申请
//     0                    1                          2
//   userID              查询阶段                     申请的状态
//   "123456"        0～7（SuppApply）          1（waite） 2（success）3（failure）
//-------------------------------------------------------------------------------
//example 1
//  供应商查询银行所有审核通过的申请
//        0                 1                  2
//       userID           查询阶段            申请状态
//   "supplierID"      “BankCheck”          “Success”

//example 2
//银行查询所有未还款项
//      0               1                        3
//     userID         查询阶段                   状态
//    “bankID”      “SuppRepayment”           “Wait”
//
func  read_allstate(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	userID := args[0]
	stage,err:= strconv.Atoi(args[1])   //阶段
	state,err := strconv.Atoi(args[2])  //状态
	var needMarbles []Marble
	marbles,err:= getAllMarbles(stub)
	if err != nil{
		fmt.Println("getAllMarblesByUserID err :",err.Error())
		return shim.Error(err.Error())
	}

	marblesNum := len(marbles)
	if marblesNum <=0{
		fmt.Println("There is no marbles")
		return shim.Error("There is no marbles")
	}

	for i:=0;i<marblesNum;i++{
		if marbles[i].User.Id == userID{
			if marbles[i].Check[stage].Review == state{
				//查询到对应阶段的对应状态
				needMarbles = append(needMarbles, marbles[i])
			}
			continue
		}

		for j:=0;j<4;j++{
			if marbles[i].Check[j].UserID == userID{
				if marbles[i].Check[stage].Review == state{
					//查询到对应阶段的对应状态
					needMarbles = append(needMarbles, marbles[i])
				}
				continue
			}
		}
	}
	marblesAsBytes, _:= json.Marshal(needMarbles)
	return shim.Success(marblesAsBytes)

}
