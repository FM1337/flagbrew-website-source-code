import client from "./http";
import JSZip from "jszip";

export function checkIPBan(ip) {
  return client
    .get(`/moderation/ban/${ip}`)
    .then((resp) => {
      if (resp.data.ban) {
        return true;
      }
      return false;
    })
    .catch((err) => {
      console.log(err);
      return false;
    });
}

export function checkIPRestrict(ip) {
  return client
    .get(`/moderation/restrict/${ip}`)
    .then((resp) => {
      if (resp.data.restricted) {
        return true;
      }
      return false;
    })
    .catch((err) => {
      console.log(err);
      return false;
    });
}

export function unzip(data) {
  return new Promise((resolve, reject) => {
    const files = [];
    let zip_file = new JSZip();

    zip_file
      .loadAsync(data)
      .then(function(zip) {
        zip.forEach((_, file) => {
          file.async("blob").then((blob) => {
            files.push({
              filename: file.name,
              data: blob,
            });
          });
        });
      })
      .then(() => {
        resolve(files);
      });
  });
}
