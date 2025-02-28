<template>
  <v-dialog width="unset" v-model="show" persistent>
    <v-card>
      <v-card-title>Editing {{ setting.name }} </v-card-title>
      <v-card-subtitle>{{ setting.description }}</v-card-subtitle>
      <v-card-text class="text--primary">
        <v-switch v-model="value" :label="`Value: ${value.toString()}`"></v-switch>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="blue darken-1" text @click="$emit('close')">
          Cancel
        </v-btn>
        <v-btn color="blue darken-1" text @click="saveSetting">
          Save
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
export default {
  name: "admin-edit-setting",
  data: () => ({
    show: true,
    value: null,
  }),
  props: {
    setting: {
      type: Object,
      required: true,
    },
  },
  created() {
    this.value = this.setting.value;
  },
  methods: {
    saveSetting() {
      const vm = this;
      const formData = new FormData();
      formData.set("value", vm.value);

      vm.$api
        .put(`/moderation/settings/${vm.setting.map_key}`, formData)
        .then((resp) => {
          vm.$store.dispatch("showSnackbar", { text: resp.data.message, color: "green", timeout: 5000 });
          vm.$emit("updatedSetting");
        })
        .catch((error) => {
          vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
        });
    },
  },
};
</script>

<style></style>
