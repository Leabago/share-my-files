
var approximateMaxFileSize = approximate(maxFileSize);
document.getElementById("maxFileSize").textContent = approximateMaxFileSize;
const bigFileMessage = "File size is too large, no more than " + approximateMaxFileSize + "megabytes allowed";
const minFileMessage = "Please upload the file";
const smallFileMessage = "Please upload a larger file";
const unknownFileMessage = "Unknown error";

const maxFileNameSize = 30

var navLinks = document.querySelectorAll("nav a");
console.log('navLinks:', navLinks);
for (var i = 0; i < navLinks.length; i++) {
	var link = navLinks[i]
	if (link.getAttribute('href') == window.location.pathname) {
		link.classList.add("active");
		break;
	}
}

var submit = document.getElementById("submit")
submit.disabled = true;


// Создаем коллекцию файлов:
var dt = new DataTransfer();
var file_list = dt.files;
var formData = new FormData();


console.log('коллекция файлов создана:');
console.dir(file_list);
// Вставим созданную коллекцию в реальное поле:
document.querySelector('input[type="file"]').files = file_list;

const inputElement = document.getElementById("input");
inputElement.addEventListener("change", handleFiles, false);

function handleFiles() {
	const fileList = this.files; /* now you can work with the file list */

	for (const file of fileList) {
		var fileNew = new File([file], file.name, { type: file.type });
		Object.defineProperty(fileNew, 'size', { value: file.size })
		console.log("fileNew: ", fileNew.size)
		dt.items.add(fileNew);
	}
	console.dir(dt.files);

	updateAll()
}


const output = document.getElementById("output");
const input = document.getElementById("input");



function updateAll() {
	var submit = document.getElementById("submit")
	submit.disabled = true;

	//delete the parent li
	var ul = document.getElementById("output");
	while (ul.firstChild) ul.removeChild(ul.firstChild);

	var i = 0;
	let numberOfBytes = 0;

	for (const file of dt.files) {
		numberOfBytes += file.size;
		// li
		const li = document.createElement("li");

		var divRow = document.createElement("div");
		divRow.classList.add("row")
		var divCol1 = document.createElement("div");
		divCol1.classList.add("col-10")


		var divCol2 = document.createElement("div");
		divCol2.classList.add("col-2")

		li.appendChild(divRow);
		divRow.appendChild(divCol1);
		divRow.appendChild(divCol2);

		var a = document.createElement("a");
		a.id = "file_name"
		a.textContent = setName(file.name);
		divCol1.appendChild(a)

		//   button
		var btn = document.createElement("span");
		btn.id = i;

		btn.classList.add("bi-trash");
		divCol2.appendChild(btn)
		output.appendChild(li);

		i = i + 1;
		btn.addEventListener("click", function (e) {
			//delete the parent li
			var ul = document.getElementById("output");
			while (ul.firstChild) ul.removeChild(ul.firstChild);

			// remove from list
			dt.items.remove(this.id);

			console.dir("files updateAll: ");
			console.dir(dt.files);


			// update after remove
			updateAll()
			console.dir(dt.files);
		});
	}

	// Approximate to the closest prefixed unit
	var outputSize = approximate(numberOfBytes)
	document.getElementById("fileSize").textContent = outputSize;
	// size 



	checkSize(numberOfBytes)
}

function checkSize(numberOfBytes) {
	var checkMaxSize = numberOfBytes <= maxFileSize
	var checkMinSize = numberOfBytes > 0
	var submit = document.getElementById("submit")
	var errors = document.getElementById("errors")

	// check
	if (checkMaxSize && checkMinSize) {
		errors.textContent = "";
		submit.disabled = false;
		return true;
	} else {
		if (!checkMaxSize) {
			errors.textContent = bigFileMessage;
		} else if (!checkMinSize && dt.files.length == 0) {
			errors.textContent = minFileMessage;
		} else if (dt.files.length > 0 && !checkMinSize) {
			errors.textContent = smallFileMessage;
		} else {
			errors.textContent = unknownFileMessage;
		}
		submit.disabled = true;
		return false;
	}
}

// Approximate to the closest prefixed unit
function approximate(numberOfBytes) {
	const K = 1024
	const units = [
		"B",
		"KiB",
		"MiB",
		"GiB",
		"TiB",
		"PiB",
		"EiB",
		"ZiB",
		"YiB",
	];
	const exponent = Math.min(
		Math.floor(Math.log(numberOfBytes) / Math.log(1024)),
		units.length - 1
	);
	const approx = numberOfBytes / K ** exponent;
	var outputSize =
		exponent === 0
			? `${numberOfBytes} bytes`
			: `${approx.toFixed(3)} ${units[exponent]
			} (${numberOfBytes} bytes)`;

	if (numberOfBytes == 0) {
		outputSize = 0
	}

	return outputSize
}

function setName(name) {
	if (name.length > maxFileNameSize) {
		name = name.substring(0, maxFileNameSize);
		name = name + "...";
	}
	return name;
}

function validateMyForm() {

	console.log("")

	var submit = document.getElementById("submit")
	submit.disabled = true;

	const K = 1024
	let numberOfBytes = 0;
	for (const file of dt.files) {
		numberOfBytes += file.size;
		formData.append("files", file);
	}



	fetch("https://localhost:8080/upload", {
		method: "post",
		body: formData,
	})
		.catch((error) => ("Something went wrong!", error))
		.then((response) => response.text().then(function (text) {
			console.log("response text: " + text)
			window.location.href = "https://localhost:8080/archive/" + text;
		}))

}

function dropHandler(ev) {
	console.log("File(s) dropped");

	// Prevent default behavior (Prevent file from being opened)
	ev.preventDefault();

	if (ev.dataTransfer.items) {
		// Use DataTransferItemList interface to access the file( s)
		[...ev.dataTransfer.items].forEach((item, i) => {
			// If dropped items aren't files, reject them
			if (item.kind === "file") {
				const file = item.getAsFile();
				console.log(`… file[${i}].name = ${file.name}`);

				dt.items.add(file);

			}
		});
	} else {
		// Use DataTransfer interface to access the file(s)
		[...ev.dataTransfer.files].forEach((file, i) => {
			console.log(`… file[${i}].name = ${file.name}`);
			dt.items.add(file);
		});
	}

	updateAll()
	console.dir(dt.files);

}

function dragOverHandler(ev) {
	console.log("File(s) in drop zone");

	// Prevent default behavior (Prevent file from being opened)
	ev.preventDefault();
}

const fileSelect = document.getElementById("fileSelect");
const fileElem = document.getElementById("input");

fileSelect.addEventListener(
	"click",
	(e) => {
		if (fileElem) {
			fileElem.click();
		}
	},
	false
);


