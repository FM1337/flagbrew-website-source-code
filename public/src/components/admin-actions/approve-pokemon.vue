<template>
  <v-dialog width="unset" v-model="show" persistent>
    <v-card>
      <v-card-title>Approving {{ pokemonCode }} </v-card-title>
      <v-card-subtitle>Are you sure you want to approve this Pokemon?</v-card-subtitle>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="blue darken-1" text @click="$emit('close')">
          Cancel
        </v-btn>
        <v-btn color="blue darken-1" text @click="approve">
          Approve
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
export default {
  name: "admin-gpss-approve",
  data: () => ({
    show: true,
  }),
  props: {
    pokemonCode: {
      type: String,
      required: true,
    },
  },
  methods: {
    approve() {
      const vm = this;
      const formData = new FormData();
      formData.append("code", vm.pokemonCode);
      vm.$api
        .post(`/moderation/approve`, formData)
        .then((resp) => {
          vm.$store.dispatch("showSnackbar", { text: resp.data.message, color: "green", timeout: 5000 });
          vm.$emit("approved");
        })
        .catch((error) => {
          vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
        });
    },
  },
};
</script>

<style></style>
