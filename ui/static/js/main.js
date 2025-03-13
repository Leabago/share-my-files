var approximateMaxFileSize = approximate(maxFileSize);
document.getElementById("maxFileSize").textContent = approximateMaxFileSize;
const bigFileMessage = "The files are too big, no more than " + approximateMaxFileSize + " allowed";
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

	var uploadData = new FormData();
	var count = 0
	var errors = document.getElementById("errors")
	errors.textContent = "";
	const rejectedFiles = [];


	// Create a Set
	const fileNameSet = new Set();
	for (const file_dt of dt.files) {
		fileNameSet.add(file_dt.name)
	}

	for (const file of fileList) {
		if (file.size = 0) {
			continue
		}

		fileName = setFileName(fileNameSet, file.name)
		var fileNew = new File([file], fileName, { type: file.type });
		Object.defineProperty(fileNew, 'size', { value: file.size })

		let numberOfBytes = 0;

		for (const file_dt of dt.files) {
			numberOfBytes += file_dt.size;
		}

		numberOfBytes += fileNew.size


		// check maxFileSize
		if (numberOfBytes <= maxFileSize) {
			dt.items.add(fileNew);
			uploadData.append("files", fileNew);
			count += 1
		} else {
			rejectedFiles.push(file.name)
			continue
		}
	}

	if (rejectedFiles.length > 0) {
		errors.textContent = `Files [ ${rejectedFiles.map(file => `"${file}"`).join(", ")} ] not loaded. ` + bigFileMessage;
	}

	if (count != 0) {
		updateAll()
		uploadAll(uploadData)
	}

	this.value = null
}

// setFileName return file name, add prefix if same name exists
function setFileName(fileNameSet, fileName) {
	updatedfileName = fileName
	count = 0
	while (fileNameSet.has(updatedfileName)) {
		updatedfileName = "(" + count + ")" + fileName
		count = count + 1
	}

	if (updatedfileName != fileName) {
		fileNameSet.add(updatedfileName)
	}

	return updatedfileName
}


// const output = document.getElementById("output");
// const input = document.getElementById("input");



function updateAll() {
	console.log("updateAll")
	var submit = document.getElementById("submit")
	// submit.disabled = true;

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
		var fileName = document.createElement("div");
		fileName.classList.add("col-8")

		// spinner
		var statusIcon = document.createElement("div");
		statusIcon.classList.add("col-1")


		var trashIcon = document.createElement("div");
		trashIcon.classList.add("col-1")

		li.appendChild(divRow);
		divRow.appendChild(statusIcon);
		divRow.appendChild(fileName);
		divRow.appendChild(trashIcon);

		var fileNameText = document.createElement("p");
		fileNameText.textContent = setName(file.name);
		fileNameText.className = "truncate-text"
		fileName.appendChild(fileNameText)

		//   button
		var btn = document.createElement("span");
		btn.id = i;

		btn.classList.add("bi-trash");
		trashIcon.appendChild(btn)
		output.appendChild(li);

		i = i + 1;
		btn.addEventListener("click", function (e) {
			//delete the parent li
			var ul = document.getElementById("output");
			while (ul.firstChild) ul.removeChild(ul.firstChild);

			// remove from list
			dt.items.remove(this.id);


			console.log("remove :", file.name)

			console.dir("files updateAll: ");
			console.dir(dt.files);


			fetch(ddnsAddress + "/delete/" + file.name, {
				method: "post",
			})
				.catch((error) => ("Something went wrong!", error))
				.then((response) => response.text().then(function (text) {
					console.log("response text: " + text)
				}))

			// update after remove
			updateAll()
			console.dir(dt.files);
		});



		// Create the div element
		let spinner = document.createElement("div");
		spinner.className = "spinner-border spinner-border-sm";
		spinner.setAttribute("role", "status");

		// Create the span element
		let span = document.createElement("span");
		span.className = "visually-hidden";
		span.textContent = "Loading...";

		// Append the span inside the div
		spinner.appendChild(span);
		statusIcon.appendChild(spinner)

	}

	// Approximate to the closest prefixed unit
	var outputSize = approximate(numberOfBytes)
	document.getElementById("fileSize").textContent = outputSize;
	// size 

	// checkSize(numberOfBytes)
}


function uploadAll(uploadData) {
	console.log("uploadAll")

	fetch(ddnsAddress + "/upload", {
		method: "post",
		body: uploadData,
	})
		.catch((error) => ("Something went wrong!", error))
		.then((response) => response.text().then(function (text) {
			console.log("response text: (" + text + ")")
		}))
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



	var submit = document.getElementById("submit")
	submit.disabled = true;

	const K = 1024
	let numberOfBytes = 0;
	for (const file of dt.files) {
		numberOfBytes += file.size;
		formData.append("files", file);
	}



	fetch(ddnsAddress + "/archive", {
		method: "post",
		body: formData,
	})
		.catch((error) => ("Something went wrong!", error))
		.then((response) => response.text().then(function (text) {
			console.log("response text: " + text)
			window.location.href = ddnsAddress + "/archive/" + text;
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


