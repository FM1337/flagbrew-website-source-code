<template>
  <v-dialog width="unset" v-model="show" persistent>
    <v-card>
      <v-card-title>{{ passedIP !== "" ? `Banning ${passedIP}` : "Banning an IP" }} </v-card-title>
      <v-card-subtitle>Are you sure you want to ban this IP?</v-card-subtitle>
      <v-card-text>
        <v-container>
          <v-row>
            <v-col cols="12" v-if="!passedIP">
              <v-text-field label="IP Address*" required v-model="ip"></v-text-field>
            </v-col>
            <v-col cols="12">
              <v-textarea label="Reason for ban*" required v-model="banReason"></v-textarea>
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
        <v-btn color="blue darken-1" text @click="banIP" :disabled="banReason === '' || !(passedIP !== '' ? passedIP !== '' : ip !== '')">
          Ban
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
export default {
  name: "admin-ban-ip",
  data: () => ({
    show: true,
    banReason: "",
    ip: "",
  }),
  props: {
    passedIP: {
      type: String,
    },
  },
  methods: {
    banIP() {
      const vm = this;
      const formData = new FormData();

      formData.append("ip", vm.passedIP !== "" ? vm.passedIP : vm.ip);
      formData.append("reason", vm.banReason);
      vm.$api
        .post(`/moderation/ban`, formData)
        .then((resp) => {
          vm.$store.dispatch("showSnackbar", { text: resp.data.message, color: "green", timeout: 5000 });
          vm.$emit("banned");
        })
        .catch((error) => {
          vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
        });
    },
  },
};
</script>

<style></style>
