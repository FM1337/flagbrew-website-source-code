<template>
  <v-dialog width="unset" v-model="show" persistent>
    <v-card>
      <v-card-title>Rejecting {{ pokemonCode }} </v-card-title>
      <v-card-subtitle>Are you sure you want to reject this Pokemon?</v-card-subtitle>
      <v-card-text>
        <v-container>
          <v-row>
            <v-col cols="12">
              <v-textarea label="Reason for rejection*" required v-model="rejectionReason"></v-textarea>
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
        <v-btn color="blue darken-1" text @click="reject" :disabled="rejectionReason == ''">
          Reject
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
export default {
  name: "admin-gpss-reject",
  data: () => ({
    show: true,
    rejectionReason: "",
  }),
  props: {
    pokemonCode: {
      type: String,
      required: true,
    },
  },
  methods: {
    reject() {
      const vm = this;
      const formData = new FormData();

      formData.append("reason", vm.rejectionReason);
      formData.append("code", vm.pokemonCode);
      vm.$api
        .post("/moderation/reject", formData)
        .then((resp) => {
          vm.$store.dispatch("showSnackbar", { text: resp.data.message, color: "green", timeout: 5000 });
          vm.$emit("rejected");
        })
        .catch((error) => {
          vm.$store.dispatch("showSnackbar", {
            text: error.response.data.errors
              ? error.response.data.error + ": " + error.response.data.errors.join(" ")
              : error.response.data.error,
            color: "red",
            timeout: 5000,
          });
        });
    },
  },
};
</script>

<style></style>
