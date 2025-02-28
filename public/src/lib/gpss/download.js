import JSZip from "jszip";

export function downloadAsZip(pokemons, downloadCode) {
  var zip = new JSZip();

  for (var x = 0; x < pokemons.length; x++) {
    const byteStr = atob(pokemons[x].pokemon);
    const ab = new ArrayBuffer(byteStr.length);

    let ia = new Uint8Array(ab);
    for (let i = 0; i < byteStr.length; i++) {
      ia[i] = byteStr.charCodeAt(i);
    }

    const blob = new Blob([ia]);

    zip.file(pokemons[x].filename, blob);
  }

  zip.generateAsync({ type: "blob" }).then(function(content) {
    const url = window.URL.createObjectURL(content);
    const link = document.createElement("a");
    link.href = url;
    link.setAttribute("download", `${downloadCode}.zip`);

    document.body.appendChild(link);

    link.click();
    link.remove();
  });
}

export function downloadFromBase64(data, filename) {
  const byteStr = atob(data);
  const ab = new ArrayBuffer(byteStr.length);

  let ia = new Uint8Array(ab);
  for (let i = 0; i < byteStr.length; i++) {
    ia[i] = byteStr.charCodeAt(i);
  }

  const blob = new Blob([ia]);

  const url = window.URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.setAttribute("download", filename);

  document.body.appendChild(link);

  link.click();
  link.remove();
}
