<template>
  <v-dialog width="unset" v-model="show" persistent>
    <v-card>
      <v-card-title>{{ passedIP !== "" ? `Restricting ${passedIP}` : "Restricting an IP" }} </v-card-title>
      <v-card-subtitle>
        Are you sure you want to restrict this IP? (This means they will require approvals for their GPSS uploads before they will show up
        in the public list)
      </v-card-subtitle>
      <v-card-text>
        <v-container>
          <v-row>
            <v-col cols="12" v-if="!passedIP">
              <v-text-field label="IP Address*" required v-model="ip"></v-text-field>
            </v-col>
            <v-col cols="12">
              <v-textarea label="Reason for restriction*" required v-model="restrictReason"></v-textarea>
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
        <v-btn
          color="blue darken-1"
          text
          @click="restrictIP"
          :disabled="restrictReason === '' || !(passedIP !== '' ? passedIP !== '' : ip !== '')"
        >
          Restrict
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
export default {
  name: "admin-restrict-ip",
  data: () => ({
    show: true,
    restrictReason: "",
    ip: "",
  }),
  props: {
    passedIP: {
      type: String,
    },
  },
  methods: {
    restrictIP() {
      const vm = this;
      const formData = new FormData();

      formData.append("ip", vm.passedIP !== "" ? vm.passedIP : vm.ip);
      formData.append("reason", vm.restrictReason);
      vm.$api
        .post(`/moderation/restrict`, formData)
        .then((resp) => {
          vm.$store.dispatch("showSnackbar", { text: resp.data.message, color: "green", timeout: 5000 });
          vm.$emit("restricted");
        })
        .catch((error) => {
          vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
        });
    },
  },
};
</script>

<style></style>
