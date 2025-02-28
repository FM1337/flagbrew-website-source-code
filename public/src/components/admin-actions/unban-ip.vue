<template>
  <v-dialog width="unset" v-model="show" persistent>
    <v-card>
      <v-card-title>Unbanning {{ passedIP }} </v-card-title>
      <v-card-subtitle>Are you sure you want to unban this IP?</v-card-subtitle>
      <v-card-text class="text--primary">
        <p>This IP was originally banned for the following reason:</p>
        <p>{{ banReason }}</p>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="blue darken-1" text @click="$emit('close')">
          Cancel
        </v-btn>
        <v-btn color="blue darken-1" text @click="unbanIP">
          Unban
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
export default {
  name: "admin-unban-ip",
  data: () => ({
    show: false,
    banReason: "",
  }),
  props: {
    passedIP: {
      type: String,
      required: true,
    },
  },
  created() {
    const vm = this;
    vm.$api
      .get(`/moderation/ban/${vm.passedIP}`)
      .then((resp) => {
        vm.banReason = resp.data.ban.ban_reason;
        vm.show = true;
      })
      .catch((err) => {
        vm.$store.dispatch("showSnackbar", { text: err.response.data.error, color: "red", timeout: 5000 });
      });
  },
  methods: {
    unbanIP() {
      const vm = this;
      vm.$api
        .delete(`/moderation/ban/${vm.passedIP}`)
        .then((resp) => {
          vm.$store.dispatch("showSnackbar", { text: resp.data.message, color: "green", timeout: 5000 });
          vm.$emit("unbanned");
        })
        .catch((error) => {
          vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
        });
    },
  },
};
</script>

<style></style>
