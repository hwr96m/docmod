

function Init() {
	GetDirs();
	GetHLStyles()
	//InitStyleSelect(GetHLStyles())
}

//------ Дерево каталогов ---------------------------------------------------------
function InitJSTree(js){
	$('#div0').on('select_node.jstree', function (e, data) {
		let path = data.instance.get_path(data.node,'/');
		GetContent(path)
		console.log('Selected: ' + path); 
	  }).jstree({
		'core' : {
			'data' : js
		},
		"plugins" : [  "state"]
	});
}
function GetDirs() {		
	let formData = new FormData();
    let xhr = new XMLHttpRequest();
	formData.append("command", "GetTree"); 
    xhr.open("POST", "/info");    
    xhr.send(formData);   
    xhr.onreadystatechange = function() { 					// Ждём ответа от сервера
	    if (xhr.readyState == 4){ 							// возвращает текущее состояние объекта(0-4)			
	    	if(xhr.status == 200) {							// код 200 (если страница не найдена вернет 404)
				//alert(xhr.responseText);			
				var js = JSON.parse(xhr.responseText)
				InitJSTree(js)
			} 
	    }		
	}
};

//------ Контент ---------------------------------------------------------
function GetContent(path) {				// открываем файлы из директории
	//alert(d);
	var formData = new FormData();
	formData.append("path", path); 
	formData.append("command", "GetContent"); 
    var xhr = new XMLHttpRequest();
    xhr.open("POST", "/info");    
    xhr.send(formData);   
    xhr.onreadystatechange = function() { 					// Ждём ответа от сервера
	    if (xhr.readyState == 4){ 							// возвращает текущее состояние объекта(0-4)			
	    	if(xhr.status == 200) {							// код 200 (если страница не найдена вернет 404)
				json = JSON.parse(xhr.responseText)	
				let string = ""		//строка с файлами и их содержимым
				let string1 = "<h2>"+htmlEscape((json.Path))+"</h2>"	// строка с индексами
				string1 += '<ul id="indexlist">'
				for (x of Object.keys(json.Files)) { 	
					//let s = (json.Path + json.Files[x].Name).replace("\\","").replace(".","")	//id элементов для скролла
					let s = encodeURI(json.Path + json.Files[x].Name)	//id элементов для скролла
					string1 += '<li>'
					string1 += "<a href=\"javascript:void(0)\" onclick=\"GoTo('"+s+"');\">"+htmlEscape((json.Files[x].Name))+"</a>"
					string1 += '</li>'

					string += '<h3 id="'+s+'">'
					string += htmlEscape((json.Files[x].Name))	//название файла
					string += "&nbsp;&nbsp;&nbsp;"
					string += "<a href=\"javascript:void(0)\" onclick=\"GoTo('div0');\">Наверх</a>"	//кнопка наверх
					string += '</h3>'
					string += '<pre ><code class="'+json.Files[x].Name.split('.').pop()+'">'		// указываем расширение файла для определения языка
					string += htmlEscape((json.Files[x].Body))		//сам файл
					string += "</code></pre>"
				}				
				string1 += '</ul>'
				document.getElementById("div1").innerHTML = string1 + string;
				hljs.highlightAll();
			} 
	    }		
	}
}
function GoTo(id){			// автоматический скролл по странице
	//alert(id)
	document.getElementById(id).scrollIntoView({behavior: 'auto',block: 'start', inline: 'start'	});
}
//------  HighLighter стили ---------------------------------------------------------
function GetHLStyles() {			// загружает список стилей highlighter
	var formData = new FormData();
    var xhr = new XMLHttpRequest();
	var json
	formData.append("command", "GetHLStyles"); 
    xhr.open("POST", "/info");    
    xhr.send(formData);   
    xhr.onreadystatechange = function() { 					// Ждём ответа от сервера
	    if (xhr.readyState == 4){ 							// возвращает текущее состояние объекта(0-4)			
	    	if(xhr.status == 200) {							// код 200 (если страница не найдена вернет 404)
				json = JSON.parse(xhr.responseText)					
				var select = document.getElementById("hlstyle");		//добавляем строки в селект
				for(index in json) {					
					select.options[select.options.length] = new Option(json[index], json[index]);
				}
				if (getCookie('hlstyle') != undefined) {
					select.value = decodeURI(getCookie('hlstyle'))
					SetStyle()
				}
			} 
	    }		
	}
}
function SetStyle() {		
	let select = document.getElementById("hlstyle");
	setCookie('hlstyle', encodeURI(select.value), {'max-age':3600*24*30})
	document.querySelectorAll('link[rel="stylesheet hl"]').forEach(disableStylesheet);	//отключаем все CSS для HL
	LoadCSSFile(select.value)		//скачиваем CSS
}
function LoadCSSFile(filename) {		//загрузка файла css
	  var fileref = document.createElement("link")
	  fileref.setAttribute("rel", "stylesheet hl")
	  fileref.setAttribute("type", "text/css")
	  fileref.setAttribute("href", filename)
	  document.head.appendChild(fileref)		//добавляем CSS на страницу

} 
function enableStylesheet (node) {		//подключение файла css
	node.disabled = false;
}  
function disableStylesheet (node) {		//отключение CSS
	node.disabled = true;
}

//------ Общие функции ---------------------------------------------------------
function htmlEscape(str) {			// экранирование html спецсиволов
	return str
		.replace(/&/g, '&amp;')
		.replace(/"/g, '&quot;')
		.replace(/'/g, '&#39;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;')
		.replace(/\//g, '&#x2F;')
		.replace(/=/g,  '&#x3D;')
		.replace(/`/g, '&#x60;');
}

//------ cookie ---------------------------------------------------------
function getCookie(name) {							// возвращает куки с указанным name,
	let matches = document.cookie.match(new RegExp(   // или undefined, если ничего не найдено
	  "(?:^|; )" + name.replace(/([\.$?*|{}\(\)\[\]\\\/\+^])/g, '\\$1') + "=([^;]*)"
	));
	return matches ? decodeURIComponent(matches[1]) : undefined;
}
function setCookie(name, value, options = {}) {		
  
	options = {
	  path: '/',
	  // при необходимости добавьте другие значения по умолчанию
	  ...options
	};
  
	if (options.expires instanceof Date) {
	  options.expires = options.expires.toUTCString();
	}
  
	let updatedCookie = encodeURIComponent(name) + "=" + encodeURIComponent(value);
  
	for (let optionKey in options) {
	  updatedCookie += "; " + optionKey;
	  let optionValue = options[optionKey];
	  if (optionValue !== true) {
		updatedCookie += "=" + optionValue;
	  }
	}
  
	document.cookie = updatedCookie;
}
  function deleteCookie(name) {
	setCookie(name, "", {
	  'max-age': -1
	})
}
  
 










