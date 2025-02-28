<template>
  <div class="pb-3">
    <v-subheader>Mystery Gift</v-subheader>
    <div class="elevation-1" id="admin-repository">
      <v-toolbar flat dense>
        <v-toolbar-title class="grey--text">Update Mystery Gift Database</v-toolbar-title>
      </v-toolbar>
      <v-row>
        <v-col cols="12">
          <v-file-input accept=".zip" truncate-length="15" v-model="file" label="Mystery Gift Database ZIP File" />
        </v-col>
        <v-col cols="12">
          <v-btn color="success" :disabled="!files || updating" @click="updateMysteryGiftDB">Update Database</v-btn>
        </v-col>
      </v-row>
    </div>
  </div>
</template>
<script>
import { unzip } from "../../lib/core/utils";
export default {
  name: "admin-mystery-gift",
  data: () => ({
    file: null,
    files: null,
    updating: false,
  }),

  watch: {
    file: async function(file) {
      const vm = this;
      if (file) {
        const files = await unzip(file);
        vm.files = files;
      }
    },
  },
  methods: {
    updateMysteryGiftDB: async function() {
      const vm = this;
      let successful = true;
      vm.updating = true;
      for (var i = 0; i < vm.files.length; i++) {
        const formData = new FormData();

        formData.append("file", vm.files[i].data, vm.files[i].filename);
        const success = await vm.$api
          .post("/files/upload/mystery-gift", formData, {
            headers: {
              "Content-Type": "multipart/form-data",
            },
          })
          .then((data) => {
            return true;
          })
          .catch((err) => {
            vm.$store.dispatch("showSnackbar", { text: err.response.data.error, color: "red", timeout: 5000 });
            return false;
          });
        if (!success) {
          successful = false;
          break;
        }
      }

      if (successful) {
        vm.$store.dispatch("showSnackbar", { text: "Mystery Gift DB updated sucessfully", color: "green", timeout: 5000 });
      }
      vm.updating = false;
    },
  },
};
</script>

<style></style>
