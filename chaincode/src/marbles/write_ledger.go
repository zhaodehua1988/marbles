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
	"time"
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
	user.ObjectType = "marble_user"
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
	err = stub.PutState(user.Id, ownerAsBytes) //store user by its UserID
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
//      0      ,      1  ,           2  ,     3                4        ,           5,
//     id      ,    contact,      balance,   title           user id    ,     authorizing_company,
// "m999999999", "13188888888",     "35",    "title"       "o9999999999999",        "inter",
func init_marble(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting init_marble")

	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	//input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	id := args[0]
	contact := args[1]
	balance, err := strconv.Atoi(args[2])
	title := args[3]
	user_id := args[4]
	authed_by_company := args[5]

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
	v, err := get_marble(stub, id)
	if err == nil {
		fmt.Println("This marble already exists - " + id)
		fmt.Println(v)
		return shim.Error("This marble already exists - " + id)  //all stop a marble by this id exists
	}

	var marble Marble

	marble.ObjectType = "marble"
	marble.Id = id
	marble.Contact = contact
	marble.Balance = balance
	marble.Title = title
	marble.User.Id = user_id
	marble.User.Username = user.Username
	marble.User.Company = user.Company
	marble.Check[New].UserID = user_id
	marble.Check[New].Company = user.Company
	marble.Check[New].Review=Success
	marble.Check[New].Date = time.Now().Format("2006-01-02 15:04:05")
	marble.Check[New].Comment = "new marbles"
	marble.Check[SuppApply].UserID = user_id
	marble.Check[SuppApply].Company   = user.Company
	marble.Check[SuppApply].Review = Wait
	marble.Check[SuppApply].Comment = ""
	for i:=2;i< StepNum;i++{
		marble.Check[i].UserID=""
		marble.Check[i].Company   = ""
		marble.Check[i].Review = Disable
		marble.Check[i].Comment = ""
	}
/*
	str := `{
		"docType":"marble", 
		"id": "` + id + `", 
		"contact": "` + contact + `", 
		"balance": "` + strconv.Itoa(balance) + `", 
		"title"  : "`+title+`"
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
*/
	jsonAsBytes, _ := json.Marshal(marble)         //convert to array of bytes
	//fmt.Println(jsonAsBytes)
	err = stub.PutState(id, jsonAsBytes)     //rewrite the owner
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("- end init_marble")
	return shim.Success(jsonAsBytes)
}

//  操作:如果通过提交到下一环节进行复审，如果不通过则返回上一环节
//      0                1    ,           2     ，             3                       4             5
//    marbleId          userID            step   ，             state                 next         comment
//  "09999999999"     "UserId"，         “step”   ，     "2/3(success/failure)"     "nextUser"      "comment"
//
func  tx_marble(stub shim.ChaincodeStubInterface, args []string) pb.Response{
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
	userID := args[1]
	//name := args[2]
	step,err := strconv.Atoi(args[2])
	state,err :=strconv.Atoi(args[3])
	next := args[4]
	commont := args[5]

	user, err := get_user(stub, userID)
	if err != nil {
		fmt.Println("Failed to find user - " + userID)
		return shim.Error(err.Error())
	}

	if !user.Enabled{
		fmt.Println("user is disable -"+userID)
		return shim.Error(err.Error())
	}

	if err != nil || step > StepNum || step < 0{
		fmt.Println("当前步骤无效")
		return shim.Error(err.Error())
	}
	marble,err:= getMarblesById(stub,marbleId)

	if err != nil{
		return shim.Error("invalid marble id:"+marbleId)
	}

	if marble.Check[step].UserID != userID{
		return shim.Error("user :"+userID+"no competence to review this marble")
	}

	if marble.Check[step].Review != Wait{
		fmt.Println("本次交易 未处于等待处理状态 :",marble.Check[step].Review)
		return shim.Error("invalid,the marble is not waiting state"+strconv.Itoa(marble.Check[step].Review))
	}
	if state == Success{  //成功
		//marble.Check[step].UserID = userID
		marble.Check[step].Company = user.Company
		marble.Check[step].Review = Success
		marble.Check[step].Date = time.Now().Format("2006-01-02 15:04:05")
		marble.Check[step].Comment = commont
		if next != ""{
			marble.Check[step+1].UserID = next
		}else{
			marble.Check[step+1].UserID = userID
		}

		marble.Check[step+1].Review = Wait
		if step == BankRecv{ //如果是银行确认收款成功，设置最后结束的状态
			marble.Check[EndOf].Review = Success
			marble.Check[EndOf].Date = time.Now().Format("2006-01-02 15:04:05")
			marble.Check[EndOf].Comment = "the marbles is end success"
			marble.Check[EndOf].Company = user.Company
		}


	}else if state == Failure{  //失败
		//marble.Check[step].UserID = userID
		marble.Check[step].Company = user.Company
		marble.Check[step].Review = Failure
		marble.Check[step].Date = time.Now().Format("2006-01-02 15:04:05")
		marble.Check[EndOf].Review = Failure
		marble.Check[EndOf].UserID = userID
		marble.Check[EndOf].Company = user.Company
		marble.Check[EndOf].Comment=commont
		marble.Check[EndOf].Date = time.Now().Format("2006-01-02 15:04:05")
	}else {
		return shim.Error("the marbles state is wrong")
	}

	jsonAsBytes, _ := json.Marshal(marble)         //convert to array of bytes
	err = stub.PutState(marbleId, jsonAsBytes)     //rewrite the owner
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}


//  操作:如果通过提交到下一环节进行复审，如果不通过则结束
//      0               1        ，          2      ，                3
//   marbleId        userid               state                   comment
//  "09999999999"    "011111"    ， "2/3(success/failure)"        "comment"
//
func  review_marble(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	var err error
	fmt.Println("starting submit_marble")
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	//input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	marbleId := args[0]
	userID := args[1]
	state,err :=strconv.Atoi(args[2])
	commont := args[3]

	var step int
	var next User
	user, err := get_user(stub, userID)
	if err != nil {
		fmt.Println("Failed to find user - " + userID)
		return shim.Error(err.Error())
	}

	if !user.Enabled{
		fmt.Println("user is disable -"+userID)
		return shim.Error(err.Error())
	}

//	if err != nil || step > StepNum || step < 0{
//		fmt.Println("当前步骤无效")
//		return shim.Error(err.Error())
//	}
	marble,err:= getMarblesById(stub,marbleId)
	if err != nil{
		return shim.Error("invalid marble id:"+marbleId)
	}
	for i:=1;i<StepNum;i++{
		if marble.Check[i].Review == Wait{
			step = i
			if user.Company != Step_company[step]{
				return shim.Error("you don't have the permissions to this step")
			}
			 next,err = getUserByCompany(stub,Step_company[step+1]);if err != nil{
				return shim.Error("can not get the next step user !!")
			}
			break
		}
	}

	if marble.Check[step].UserID != userID{
		return shim.Error("user :"+userID+"no competence to review this marble")
	}

	if marble.Check[step].Review != Wait{
		return shim.Error("invalid,the marble is not waiting state="+strconv.Itoa(marble.Check[step].Review))
	}
	if state == Success{  //成功
		//marble.Check[step].UserID = userID
		marble.Check[step].Company = user.Company
		marble.Check[step].Review = Success
		marble.Check[step].Date = time.Now().Format("2006-01-02 15:04:05")
		marble.Check[step].Comment = commont
		if next.Id != ""{
			marble.Check[step+1].UserID = next.Id
		}

		marble.Check[step+1].Review = Wait
		if step == BankRecv{ //如果是银行确认收款成功，设置最后结束的状态
			marble.Check[EndOf].Review = Success
			marble.Check[EndOf].Date = time.Now().Format("2006-01-02 15:04:05")
			marble.Check[EndOf].Comment = "the marble is end success !"
			marble.Check[EndOf].Company = user.Company
		}


	}else if state == Failure{  //失败
		//marble.Check[step].UserID = userID
		marble.Check[step].Company = user.Company
		marble.Check[step].Review = Failure
		marble.Check[step].Date = time.Now().Format("2006-01-02 15:04:05")
		marble.Check[EndOf].Review = Failure
		marble.Check[EndOf].UserID = userID
		marble.Check[EndOf].Company = user.Company
		marble.Check[EndOf].Comment="the marble is end failure "
		marble.Check[EndOf].Date = time.Now().Format("2006-01-02 15:04:05")
	}else {
		return shim.Error("the marbles state is wrong")
	}

	jsonAsBytes, _ := json.Marshal(marble)         //convert to array of bytes
	err = stub.PutState(marbleId, jsonAsBytes)     //rewrite the owner
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}
