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
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// ============================================================================================================================
// write() - genric write variable into ledger
// 
// Shows Off PutState() - writting a key/value into the ledger
//
// Inputs - Array of strings
//    0   ,    1
//   key  ,  value
//  "abc" , "test"
// ============================================================================================================================
func write(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key, value string
	var err error
	fmt.Println("starting write")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2. key of the variable and value to set")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	key = args[0]                                   //rename for funsies
	value = args[1]
	err = stub.PutState(key, []byte(value))         //write the variable into the ledger
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end write")
	return shim.Success(nil)
}

// ============================================================================================================================
// delete_marble() - remove a marble from state and from marble index
// 
// Shows Off DelState() - "removing"" a key/value from the ledger
//
// Inputs - Array of strings
//      0      ,         1
//     id      ,  authed_by_company
// "m999999999", "united marbles"
// ============================================================================================================================
func delete_marble(stub shim.ChaincodeStubInterface, args []string) (pb.Response) {
	fmt.Println("starting delete_marble")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// input sanitation
	err := sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	id := args[0]
	authed_by_company := args[1]

	// get the marble
	marble, err := get_marble(stub, id)
	if err != nil{
		fmt.Println("Failed to find marble by id " + id)
		return shim.Error(err.Error())
	}

	// check authorizing company (see note in set_user() about how this is quirky)
	if marble.User.Company != authed_by_company{
		return shim.Error("The company '" + authed_by_company + "' cannot authorize deletion for '" + marble.User.Company + "'.")
	}

	// remove the marble
	err = stub.DelState(id)                                                 //remove the key from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	fmt.Println("- end delete_marble")
	return shim.Success(nil)
}
// ============================================================================================================================
// Init User - create a new owner aka end user, store into chaincode state
//
// Shows off building key's value from GoLang Structure
//
// Inputs - Array of Strings
//           0     ,     1   ,   2
//      owner id   , username, company
// "o9999999999999",     bob", "united marbles"
// ============================================================================================================================
func init_owner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting init_owner")

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	//input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var user User
	user.ObjectType = "marble_owner"
	user.Id =  args[0]
	user.Username = strings.ToLower(args[1])
	user.Company = args[2]
	user.Enabled = true

	fmt.Println(user)

	//check if user already exists
	_, err = get_user(stub, user.Id)
	if err == nil {
		fmt.Println("This user already exists - " + user.Id)
		return shim.Error("This user already exists - " + user.Id)
	}

	//store user
	ownerAsBytes, _ := json.Marshal(user)      //convert to array of bytes
	err = stub.PutState(user.Id, ownerAsBytes) //store user by its Id
	if err != nil {
		fmt.Println("Could not store user")
		return shim.Error(err.Error())
	}

	fmt.Println("- end init_owner marble")
	return shim.Success(nil)
}

// ============================================================================================================================
// Disable Marble User
//
// Shows off PutState()
//
// Inputs - Array of Strings
//       0     ,        1      
//  owner id       , company that auth the transfer
// "o9999999999999", "united_mables"
// ============================================================================================================================
func disable_owner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting disable_owner")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var owner_id = args[0]
	var authed_by_company = args[1]

	// get the marble owner data
	owner, err := get_user(stub, owner_id)
	if err != nil {
		return shim.Error("This owner does not exist - " + owner_id)
	}

	// check authorizing company
	if owner.Company != authed_by_company {
		return shim.Error("The company '" + authed_by_company + "' cannot change another companies marble owner")
	}

	// disable the owner
	owner.Enabled = false
	jsonAsBytes, _ := json.Marshal(owner)         //convert to array of bytes
	err = stub.PutState(args[0], jsonAsBytes)     //rewrite the owner
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end disable_owner")
	return shim.Success(nil)
}

//新建一个申请()
//       0     ,                       1       2
//  marble id（此次申请的ID）  ,       color    Balance
// "m999999999",                    "read",

//      0      ,    1  ,  2  ,           3          ,      4     ,
//     id      ,  contract, balance,     owner id    ,     authorizing_company,
// "m999999999", "blue", "35", "o9999999999999",        "inter",
func init_marble(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting init_marble")

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	//input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	id := args[0]
	contact := strings.ToLower(args[1])
	user_id := args[3]
	authed_by_company := args[4]
	balance, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	//check if new user exists
	user, err := get_user(stub, user_id)
	if err != nil {
		fmt.Println("Failed to find user - " + user_id)
		return shim.Error(err.Error())
	}

	//check authorizing company (see note in set_user() about how this is quirky)
	if user.Company != authed_by_company{
		return shim.Error("The company '" + authed_by_company + "' cannot authorize creation for '" + user.Company + "'.")
	}

	//check if marble id already exists
	marble, err := get_marble(stub, id)
	if err == nil {
		fmt.Println("This marble already exists - " + id)
		fmt.Println(marble)
		return shim.Error("This marble already exists - " + id)  //all stop a marble by this id exists
	}

	str := `{
		"docType":"marble", 
		"id": "` + id + `", 
		"contact": "` + contact + `", 
		"balance": ` + strconv.Itoa(balance) + `, 
		"user": {
			"id": "` + user_id + `", 
			"username": "` + user.Username + `", 
			"company": "` + user.Company + `"
		},
		"check":{
			{"id":"` + id + `","name":"` + user.Username + `","review":"2"},
			{"id":"` + id + `","name":"` + user.Username + `","review":"1"},
			{"id":"","name":"","review":"0"},
			{"id":"","name":"","review":"0"},
			{"id":"","name":"","review":"0"},
			{"id":"","name":"","review":"0"},
			{"id":"","name":"","review":"0"},
			{"id":"","name":"","review":"0"},
		},
		"where":"`+ user_id +`"
	}`
	err = stub.PutState(id, []byte(str))                         //store marble with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end init_marble")
	return shim.Success(nil)
}

//  操作:如果通过提交到下一环节进行复审，如果不通过则返回上一环节
//      0                1    ,    2        ，    3      ，     4               5
//    marbleId          id        name      ，   是否通过   ，    next
//  "09999999999"     "UserId"，"UserName"  ，  “2or3”   ，   下一步UserId    执行阶段
//
func  review_marble(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	var err error
	fmt.Println("starting submit_marble")
	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	//input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	marbleId := args[0]
	id := args[1]
	name := args[2]
	next := args[4]
	step,err := strconv.Atoi(args[5])
	state,err :=strconv.Atoi(args[3])
	if err != nil || step > stepNum || step < 0{
		fmt.Println("当前步骤无效")
	}
	marble:= getMarblesById(stub,marbleId)
//	user :=  getUserById(stub,id)

	if marble.Check[step].Id != id{
		fmt.Println("当前用户不可审核这次交易，userName： " + name)
	}
	//判断交易是否已经被决绝过，防止重复提交
	if marble.Check[0].Review == Disable{
		fmt.Println("本次申请己经提交过，请勿重复提交")
	}

	if marble.Check[step].Review != Wait{
		fmt.Println("本次交易 未处于等待处理状态 :",marble.Check[step].Review)
	}
	if state == Success{  //成功
		//marble.Check[step].Id = id
		marble.Check[step].Name = name
		marble.Check[step].Review = Success
		if next != ""{
			marble.Check[step+1].Id = next
		}else{
			marble.Check[step+1].Id = id
		}
		marble.Check[step+1].Review = Success

	}else if state == Failure{  //失败
		//marble.Check[step].Id = id
		marble.Check[step].Name = name
		marble.Check[step].Review = Failure
		marble.Check[EndOf].Review = Failure
		marble.Check[EndOf].Id = id
		marble.Check[EndOf].Name = name
	}else {
		return shim.Error("")
	}

	jsonAsBytes, _ := json.Marshal(marble)         //convert to array of bytes
	err = stub.PutState(marbleId, jsonAsBytes)     //rewrite the owner
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}


