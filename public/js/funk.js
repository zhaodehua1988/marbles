/* global document */
 /*exported in_array, formatDate, randStr, toTitleCase, nDig, escapeHtml, getRandomInt */
//if element is in array
function in_array(name, array){
	for(var i in array){
		if(array[i] == name) return true;
	}
	return false;
}
			
//make random string of set length
function randStr(length){
	var text = '';
	var possible = 'abcdefghijkmnpqrstuvwxyz0123456789';
	for(var i=0; i < length; i++ ) text += possible.charAt(Math.floor(Math.random() * possible.length));
	return text;
}	
	
//capital first letter of each word
function toTitleCase(str){
	return str.replace(/\w\S*/g, function(txt){return txt.charAt(0).toUpperCase() + txt.substr(1).toLowerCase();});
}

//random integer
function getRandomInt(min, max) {
	min = Math.ceil(min);
	max = Math.floor(max);
	return Math.floor(Math.random() * (max - min)) + min;
}

function formatDate(date, fmt) {
	date = new Date(date);
	function pad(value) {
		return (value.toString().length < 2) ? '0' + value : value;
	}
	return fmt.replace(/([a-zA-Z])/g, function (_, fmtCode) {
		var tmp;
		console.log(fmtCode);
		switch (fmtCode) {
		case 'Y':								//Year
			return date.getUTCFullYear();
		case 'M':								//Month 0 padded
			return pad(date.getUTCMonth() + 1);
		case 'd':								//Date 0 padded
			return pad(date.getUTCDate());
		case 'H':								//24 Hour 0 padded
			return pad(date.getUTCHours());
		case 'I':								//12 Hour 0 padded
			tmp = date.getUTCHours();
			if(tmp === 0) tmp = 12;				//00:00 should be seen as 12:00am
			else if(tmp > 12) tmp -= 12;
			return pad(tmp);
		case 'p':								//am / pm
			tmp = date.getUTCHours();
			if(tmp >= 12) return 'pm';
			return 'am';
		case 'P':								//AM / PM
			tmp = date.getUTCHours();
			if(tmp >= 12) return 'PM';
			return 'AM';
		case 'm':								//Minutes 0 padded
			return pad(date.getUTCMinutes());
		case 's':								//Seconds 0 padded
			return pad(date.getUTCSeconds());
		case 'r':								//Milliseconds 0 padded
			return pad(date.getUTCMilliseconds(), 3);
		case 'q':								//UTC timestamp
			return date.getTime();
		default:
			throw new Error('Unsupported format code: ' + fmtCode);
		}
	});
}

function nDig(n, digits){								//zero left pad to number of digits
	var ret = n;
	for(var i=0; i < digits; i++) ret = '0' + ret;		//add  up to max i would need
	ret = ret.substring(ret.length - digits);			//cut off what you don't need
	return ret;
}

function escapeHtml(str) {
	var ret = str;
	if(str && str.replace){
		str = str.replace(new RegExp('[<,>]', 'g'), '');
		var div = document.createElement('div');
		div.appendChild(document.createTextNode(str));
		ret = div.innerHTML;
	}
	return ret;
}

// const (
// 	New = iota     //0
// 	SuppApply      //供应商申请  supplier apply  1
// 	CompanyCheck   //核心企业审核  2
// 	BankCheck      //银行审核     3
// 	SuppRecv       //供应商收款   4
// 	CompanyRePayMent //核心企业还款   5
// 	SuppRepayment  //供应商还款       6
// 	BankRecv       //银行确认收款     7
// 	EndOf            //包括成功和失败两种情况 8
//  )
//  //确认阶段
//  const(
// 	Disable = iota //0
// 	Wait
// 	Success
// 	Failure
//  )
function translateAction(where,operation){
	var operations = [
		['','wait','create'],
		['','wait','confirm','reject'],
		['','wait','loan','reject'],
		['','wait','receive','reject'],
		['','wait','pay','reject'],
		['','wait','pay','reject'],
		['','wait','confirm','reject'],
		['','wait','successed','failed']
	];
	return operations[where][operation];
}

//var Step_company  =[StepNum]string {"supplier","core-enterprise","bank","supplier","core-enterprise","supplier","bank","bank"}
function getCompany(marble){
	var company = [
		"supplier","core-enterprise","bank","supplier","core-enterprise","supplier","bank","bank"
	];
	for (var i in marble.check){
		if (marble.check[i].review===1){
			return company[i];
		}
	}
	return '';
}