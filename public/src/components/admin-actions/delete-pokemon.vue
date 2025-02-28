<template>
  <v-dialog width="unset" v-model="show" persistent>
    <v-card>
      <v-card-title>Deleting {{ deletionData.download_code }} </v-card-title>
      <v-card-subtitle>Are you sure you want to delete this Pokemon?</v-card-subtitle>
      <v-card-text>
        <v-container>
          <v-row>
            <v-col cols="12">
              <v-textarea label="Reason for deletion*" required v-model="deletionReason"></v-textarea>
            </v-col>
          </v-row>
        </v-container>
        <small>*indicates required field</small>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="blue darken-1" text @click="$emit('close')">
          Cancel
        </v-btn>
        <v-btn color="blue darken-1" text @click="deletePkmn" :disabled="deletionReason == ''">
          Delete
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
export default {
  name: "admin-gpss-delete",
  data: () => ({
    show: true,
    deletionReason: "",
  }),
  props: {
    deletionData: {
      type: Object,
      required: true,
    },
  },
  methods: {
    deletePkmn() {
      const vm = this;
      const formData = new FormData();

      formData.append("reason", vm.deletionReason);
      vm.$api
        .delete(`/moderation/delete/gpss/Pokemon/${vm.deletionData.download_code}`, {
          data: formData,
          headers: { "Content-Type": "application/x-www-form-urlencoded" },
        })
        .then((resp) => {
          vm.$store.dispatch("showSnackbar", { text: resp.data.message, color: "green", timeout: 5000 });
          vm.$emit("deleted");
        })
        .catch((error) => {
          vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
        });
    },
  },
};
</script>

<style></style>
