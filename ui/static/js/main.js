
// Создаем коллекцию файлов:
var dt = new DataTransfer();
// dt.items.add(file);
var file_list = dt.files;

console.log('Коллекция файлов создана:');
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
}


const output = document.getElementById("output");
const input = document.getElementById("input");
input.addEventListener("change", (event) => {
  updateAll()
});


function updateAll() {
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
	var divCol1 =  document.createElement("div");
	divCol1.classList.add("col-10")
  

	var divCol2 =  document.createElement("div");
	divCol2.classList.add("col-2")

	li.appendChild(divRow);
	divRow.appendChild(divCol1);
	divRow.appendChild(divCol2);

	var a =  document.createElement("a");
	a.id = "file_name"
	// a.textContent =   file.name;

	a.textContent =   setName(file.name);
	 

	divCol1.appendChild(a)
  


	// li.textContent = file.name;
	//   button
	var btn = document.createElement("span");
	btn.id = i;
	// btn.textContent = i;

	// var span = document.createElement("Span");
	btn.classList.add("bi-trash");
	// divCol2.appendChild(span);
	divCol2.appendChild(btn)


	output.appendChild(li);
	// li.appendChild(btn);

	i = i + 1;
	btn.addEventListener("click", function (e) {
	  // document.getElementById("fileSize").textContent = 0;

	  //delete the parent li
	  var ul = document.getElementById("output");
	  while (ul.firstChild) ul.removeChild(ul.firstChild);

	  // remove from list
	  dt.items.remove(this.id);

	  // update after remove
	  updateAll()
	  console.dir(dt.files);


	});
  }


  // Approximate to the closest prefixed unit

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
  const outputSize =
	exponent === 0
	  ? `${numberOfBytes} bytes`
	  : `${approx.toFixed(3)} ${units[exponent]
	  } (${numberOfBytes} bytes)`;

  //  document.getElementById("fileNum").textContent =   uploadInput.files.length;
  document.getElementById("fileSize").textContent = outputSize;

  // size

  console.log("K**3 ", K ** 3)
  var sizeok = document.getElementById("sizeok")
  var submit = document.getElementById("submit")

  // check
  if (numberOfBytes <= K ** 3) {
	sizeok.textContent = "ok";
	submit.disabled = false;
  } else {
	submit.disabled = true;
	sizeok.textContent = "bad";
  }
}

function setName(name){
  var K = 20;
  if (name.length > K){
	name = name.substring(1, K);
	name = name + "...";
  }
  return name;
}


function validateMyForm() {
  const K = 1024
  let numberOfBytes = 0;
  for (const file of dt.files) {
	numberOfBytes += file.size;
  }

  if (numberOfBytes <= K ** 3) {
	// alert("validations passed");
	return true;


  } else {
	// alert("validation failed false");
	return false;
  }
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