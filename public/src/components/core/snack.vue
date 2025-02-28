<template>
  <v-snackbar v-model="show" :timeout="timeout">
    {{ text }}
    <template v-slot:action="{ attrs }">
      <v-btn :color="color" text v-bind="attrs" @click="show = false">
        Close
      </v-btn>
    </template>
  </v-snackbar>
</template>

<script>
export default {
  name: "core-snackbar",

  data() {
    return {
      text: "",
      color: "",
      timeout: -1,
      show: false,
    };
  },

  created() {
    const vm = this;
    vm.$store.subscribe((mutation, state) => {
      if (mutation.type === "set_snackbar") {
        vm.text = state.snackbar.text;
        vm.color = state.snackbar.color;
        vm.timeout = state.snackbar.timeout;
        vm.show = true;
      }
    });
  },
};
</script>

<style></style>
