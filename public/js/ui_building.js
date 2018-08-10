/* global bag, $, ws*/
/* global escapeHtml, toTitleCase, formatDate, known_companies, transfer_marble, record_company, show_tx_step, refreshHomePanel, auditingMarble*/
/* exported build_marble, record_company, build_user_panels, build_company_panel, build_notification, populate_users_marbles*/
/* exported build_a_tx, marbles */

var marbles = {};

// =================================================================================
//	UI Building
// =================================================================================
//build a marble
function build_marble(marble) {
	var html = '';
	var colorClass = '';
	var size = 'largeMarble';
	var auditing = '';

	marbles[marble.id] = marble;

	marble.id = escapeHtml(marble.id);
	marble.color = escapeHtml(marble.color);
	marble.owner.id = escapeHtml(marble.owner.id);
	marble.owner.username = escapeHtml(marble.owner.username);
	marble.owner.company = escapeHtml(marble.owner.company);
	var full_owner = escapeHtml(marble.owner.username.toLowerCase() + '.' + marble.owner.company);

	console.log('[ui] building marble: ', marble.color, full_owner, marble.id.substring(0, 4) + '...');
	if (marble.size == 16) size = 'smallMarble';
	if (marble.color) colorClass = marble.color.toLowerCase() + 'bg';

	if (auditingMarble && marble.id === auditingMarble.id) auditing = 'auditingMarble';

	//获取当前等待执行的动作
	var confirm,reject;
	for (var i=0;i<marble.check.length;i++){
		if(!(marble.check[i].date)){
			confirm = translateAction(i,2);
			reject = translateAction(i,3);
			break;
		}
	}
	var loc = getCompany(marble);
	// html += '<span id="' + marble.id + '" class="ball ' + size + ' ' + colorClass + ' ' + auditing + ' title="' + marble.id + '"';
	// html += ' username="' + marble.owner.username + '" company="' + marble.owner.company + '" owner_id="' + marble.owner.id + '"></span>';
	html += '<tr id="' + marble.id + '" class="item" loc="' + loc + '" reject="' + reject + '" confirm="' + confirm + '" title="' + marble.title + '" username="' + marble.owner.username + '" company="' + marble.owner.company + '" owner_id="' + marble.owner.id + '" balance="' + marble.balance + '"  contact="' + marble.contact + '"> ';
	html += '<td>'+marble.title+'</td><td>'+marble.balance+'</td><td>'+marble.owner.company+'</td><td>'+marble.contact+'</td><td>'+marble.check[0].date+'</td><td class="updateMarbleButton" onclick="OpenUpdatePanel(this)">deal</td></tr>';

	$('.marblesWrap[company="' + getCompany(marble) + '"]').first().find('.innerMarbleContainer').append(html);
	$('.marblesWrap[company="' + getCompany(marble) + '"]').first().find('.noMarblesMsg').hide();
	return html;
}

function OpenUpdatePanel(obj) {
	$('#tint').fadeIn();
	$('#updatePanel').fadeIn();
	var title = $(obj).parents('.item').attr('title');
	var owner_id = $(obj).parents('.item').attr('owner_id');
	var balance = $(obj).parents('.item').attr('balance');
	var contact = $(obj).parents('.item').attr('contact');
	var username = $(obj).parents('.item').attr('username');
	var company = $(obj).parents('.item').attr('company');
	var confirm = $(obj).parents('.item').attr('confirm');
	var reject = $(obj).parents('.item').attr('reject');
	var loc = $(obj).parents('.item').attr('loc');
	var id = $(obj).parents('.item').attr('id');
	// $('select[name="user"]').html('<option value="' + username + '">' + toTitleCase(username) + '</option>');
	$('input[name="title"]').val(title);
	$('input[name="balance"]').val(balance);
	$('input[name="contact"]').val(contact);
	$('input[name="username"]').val(username);
	$('input[name="company"]').val(company);
	$('#confirmButton').text(confirm);
	$('#rejectButton').text(reject);
	$('input[name="owner_id"]').val(owner_id);
	$('input[name="id"]').val(id);
	//权限判断
	if (loc !=Cookies.get('username') ){
		$('#confirmButton').hide();
		$('#rejectButton').hide();
		$('input[name="comment"]').parents('legend').hide();
	}else{
		$('#confirmButton').show();
		$('#rejectButton').show();
		$('input[name="comment"]').parents('legend').show();
	}
	var html = build_a_marble(window.AllMarbles[id]);
	$('#updateInnerWrapLeft').html(html);
	return false;
}


//translate history check status
// function translateHistory(checkStatus){
// 	var ret = [];
// 	for (i)
// }
//build a tx history div
function build_a_marble(data) {
	var html = '';
	var username = '-';
	var company = '-';
	var id = '-';
	if (data  && data.user){
		data.owner=data.user;
	}
	if (data && data.owner && data.owner.username) {
		username = data.owner.username;
		company = data.owner.company;
		id = data.owner.id;
	}
	var history = data.check;
	html += `
				<h4 style="margin-top:20px">Operation logs:</h4>
				<ul style="display:block;width:90%;">
			`;
	for (var i=0;i<data.check.length;i++){
		if (history[i] && history[i].date){
			html+=(
				//userid,name,date,review,comment
				'<li style="margin-top:10px">'+history[i].company+' '+translateAction(i,history[i].review)+' on '+ formatDate(history[i].date,'Y-M-d')+
				' with comment: '+history[i].comment+ '</li>'
			);
		}
	}
	html +=	`	</ul>
			   `;
			// 	<div class="rightHalf">
			// 		<p>
			// 			<div class="marbleLegend">Owner: </div>
			// 			<div class="marbleName">` + username + `</div>
			// 		</p>
			// 		<p>
			// 			<div class="marbleLegend">Company: </div>
			// 			<div class="marbleName">` + company + `</div>
			// 		</p>
			// 		<p>
			// 			<div class="marbleLegend">Ower Id: </div>
			// 			<div class="marbleName">` + id + `</div>
			// 		</p>
			// 	</div>
			// </div>`;
	return html;
}

//redraw the user's marbles
function populate_users_marbles(msg) {

	//reset
	console.log('[ui] clearing marbles for user ' + msg.owner_id);
	$('.marblesWrap[company="' + getCompany(msg) + '"]').first().find('.innerMarbleWrap').html(
	`
			<table class="innerMarbleContainer">
				<tr>
					<th>title</th>
					<th>balance</th>
					<th>company</th>
					<th>contact</th>
					<th>create date</th>
					<th>action</th>
				</tr>
			</table>
	`);
	$('.marblesWrap[company="' + getCompany(msg) + '"]').first().find('.noMarblesMsg').show();

	for (var i in msg.marbles) {
		build_marble(msg.marbles[i]);
	}
}

//crayp resize - dsh to do, dynamic one
function size_user_name(name) {
	var style = '';
	if (name.length >= 10) style = 'font-size: 22px;';
	if (name.length >= 15) style = 'font-size: 18px;';
	if (name.length >= 20) style = 'font-size: 15px;';
	if (name.length >= 25) style = 'font-size: 11px;';
	return style;
}

//build all user panels
function build_user_panels(data) {

	//reset
	console.log('[ui] clearing all user panels');
	$('.ownerWrap').html('');
	for (var x in known_companies) {
		known_companies[x].count = 0;
		known_companies[x].visible = 0;							//reset visible counts
	}
	for (var i=0;i<data.length;i++){
		var u = Cookies.get('username');
		if (data[i].company===u){
			var t = data[0];
			data[0]=data[i];
			data[i]=t;
			break;
		}
	}
	i=0;
	for (var i in data) {
		var html = '';
		var colorClass = '';
		data[i].id = escapeHtml(data[i].id);
		data[i].username = escapeHtml(data[i].username);
		data[i].company = escapeHtml(data[i].company);
		record_company(data[i].company);
		known_companies[data[i].company].count++;
		known_companies[data[i].company].visible++;

		console.log('[ui] building owner panel ' + data[i].id);

		let disableHtml = '';
		if (data[i].company  === escapeHtml(bag.marble_company)) {
			disableHtml = '<span class="fa fa-trash disableOwner" title="Disable Owner"></span>';
		}

		html += `<div id="user` + i + `wrap" username="` + data[i].username + `" company="` + data[i].company +
			`" owner_id="` + data[i].id + `" class="marblesWrap ` + colorClass + `">
					<div class="legend" style="` + size_user_name(data[i].username) + `">Loan Workflow
						` + //toTitleCase(data[i].username) + 
						// `
						// <span class="fa fa-thumb-tack marblesFix" title="Never Hide Owner"></span>
						// ` + disableHtml + 
						`
						<i class="fa fa-plus addMarbleButtion marblesFix"></i>
					</div>
					<div class="innerMarbleWrap">
						<table class="innerMarbleContainer">
							<tr>
								<th>title</th>
								<th>balance</th>
								<th>company</th>								
								<th>contact</th>
								<th>create date</th>
								<th>action</th>
							</tr>
						</table>
					</div>
					<div class="noMarblesMsg hint">you have no workflow in progress</div>
				</div>`;

		$('.companyPanel[company="' + data[i].company + '"]').find('.ownerWrap').append(html);
		$('.companyPanel[company="' + data[i].company + '"]').find('.companyVisible').html(known_companies[data[i].company].visible);
		$('.companyPanel[company="' + data[i].company + '"]').find('.companyCount').html(known_companies[data[i].company].count);
	}

	//drag and drop marble
	$('.innerMarbleWrap').sortable({ connectWith: '.innerMarbleWrap', items: 'span' }).disableSelection();
	// $('.innerMarbleWrap').droppable({
	// 	drop:
	// 	function (event, ui) {
	// 		var marble_id = $(ui.draggable).attr('id');

	// 		//  ------------ Delete Marble ------------ //
	// 		if ($(event.target).attr('id') === 'trashbin') {
	// 			console.log('removing marble', marble_id);
	// 			show_tx_step({ state: 'building_proposal' }, function () {
	// 				var obj = {
	// 					type: 'delete_marble',
	// 					id: marble_id,
	// 					v: 1
	// 				};
	// 				ws.send(JSON.stringify(obj));
	// 				$(ui.draggable).addClass('invalid bounce');
	// 				refreshHomePanel();
	// 			});
	// 		}

	// 		//  ------------ Transfer Marble ------------ //
	// 		else {
	// 			var dragged_owner_id = $(ui.draggable).attr('owner_id');
	// 			var dropped_owner_id = $(event.target).parents('.marblesWrap').attr('owner_id');

	// 			console.log('dropped a marble', dragged_owner_id, dropped_owner_id);
	// 			if (dragged_owner_id != dropped_owner_id) {										//only transfer marbles that changed owners
	// 				$(ui.draggable).addClass('invalid bounce');
	// 				transfer_marble(marble_id, dropped_owner_id);
	// 				return true;
	// 			}
	// 		}
	// 	}
	// });

	//user count
	$('#foundUsers').html(data.length);
	$('#totalUsers').html(data.length);
}

//build company wrap
function build_company_panel(company) {
	company = escapeHtml(company);
	console.log('[ui] building company panel ' + company);

	var mycss = '';
	if (company === escapeHtml(Cookies.get('username'))) mycss = 'myCompany';

	var html = `<div class="companyPanel" company="` + company + `">
					<div class="companyNameWrap ` + mycss + `">
					<span class="companyName">` + company + `&nbsp;-&nbsp;</span>
					<span class="companyVisible">0</span>/<span class="companyCount">0</span>`;
	if (company === escapeHtml(bag.marble_company)) {
		html += '<span class="fa fa-exchange floatRight"></span>';
	} else {
		html += '<span class="fa fa-long-arrow-left floatRight"></span>';
	}
	html += `	</div>
				<div class="ownerWrap ` + mycss + `"></div>
			</div>`;
	$('#allUserPanelsWrap').append(html);
}

//build a notification msg, `error` is boolean
function build_notification(error, msg) {
	var html = '';
	var css = '';
	var iconClass = 'fa-check';
	if (error) {
		css = 'warningNotice';
		iconClass = 'fa-minus-circle';
	}

	html += `<div class="notificationWrap ` + css + `">
				<span class="fa ` + iconClass + ` notificationIcon"></span>
				<span class="noticeTime">` + formatDate(Date.now(), `%M/%d %I:%m:%s`) + `&nbsp;&nbsp;</span>
				<span>` + escapeHtml(msg) + `</span>
				<span class="fa fa-close closeNotification"></span>
			</div>`;
	return html;
}


//build a tx history div
function build_a_tx(data, pos) {
	var html = '';
	var username = '-';
	var company = '-';
	var id = '-';
	if (data && data.value && data.value.user){
		data.value.owner=data.value.user;
	}
	if (data && data.value && data.value.owner && data.value.owner.username) {
		username = data.value.owner.username;
		company = data.value.owner.company;
		id = data.value.owner.id;
	}

	html += `<div class="txDetails">
				<div class="txCount">TX ` + (Number(pos) + 1) + `</div>
				<p>
					<div class="marbleLegend">Transaction: </div>
					<div class="marbleName txId">` + data.txId.substring(0, 14) + `...</div>
				</p>
				<p>
					<div class="marbleLegend">Owner: </div>
					<div class="marbleName">` + username + `</div>
				</p>
				<p>
					<div class="marbleLegend">Company: </div>
					<div class="marbleName">` + company + `</div>
				</p>
				<p>
					<div class="marbleLegend">Ower Id: </div>
					<div class="marbleName">` + id + `</div>
				</p>
			</div>`;
	return html;
}

