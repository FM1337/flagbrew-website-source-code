<template>
  <v-dialog width="unset" v-model="show" persistent>
    <v-card>
      <v-card-title>Unrestricting {{ passedIP }} </v-card-title>
      <v-card-subtitle>Are you sure you want to unrestrict this IP?</v-card-subtitle>
      <v-card-text class="text--primary">
        <p>This IP was originally banned for the following reason:</p>
        <p>{{ restrictReason }}</p>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="blue darken-1" text @click="$emit('close')">
          Cancel
        </v-btn>
        <v-btn color="blue darken-1" text @click="unrestrictIP">
          Unrestrict
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
export default {
  name: "admin-unrestrict-ip",
  data: () => ({
    show: false,
    restrictReason: "",
  }),
  props: {
    passedIP: {
      type: String | null,
      required: true,
    },
  },
  created() {
    const vm = this;
    vm.$api
      .get(`/moderation/restrict/${vm.passedIP}`)
      .then((resp) => {
        vm.restrictReason = resp.data.reason;
        vm.show = true;
      })
      .catch((err) => {
        vm.$store.dispatch("showSnackbar", { text: err.response.data.error, color: "red", timeout: 5000 });
      });
  },
  methods: {
    unrestrictIP() {
      const vm = this;
      vm.$api
        .delete(`/moderation/restrict/${vm.passedIP}`)
        .then((resp) => {
          vm.$store.dispatch("showSnackbar", { text: resp.data.message, color: "green", timeout: 5000 });
          vm.$emit("unrestricted");
        })
        .catch((error) => {
          vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
        });
    },
  },
};
</script>

<style></style>
